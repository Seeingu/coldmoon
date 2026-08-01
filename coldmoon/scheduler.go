package coldmoon

import (
	"sync"
	"time"
)

// Clock is the Scheduler's only time seam. Production uses RealClock while
// tests can supply a deterministic clock without sleeping.
type Clock interface {
	Now() time.Time
	After(time.Duration) <-chan time.Time
}

// RealClock implements Clock with the process wall clock.
type RealClock struct{}

// Now returns the current wall-clock time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// After waits for the requested duration.
func (RealClock) After(delay time.Duration) <-chan time.Time {
	return time.After(delay)
}

type scheduledTask struct {
	id       JSInt
	deadline time.Time
	order    uint64
	run      func() CompletionValue
}

// Scheduler owns an Agent's promise jobs, timers, and finite asynchronous
// tasks. Keeping these queues together makes "run until idle" a single runtime
// invariant instead of a protocol shared by callers.
type Scheduler struct {
	agent *Agent
	clock Clock

	mu          sync.Mutex
	jobs        []*QueuedPromiseJob
	timers      []scheduledTask
	activeTasks int
	nextTimerID JSInt
	nextOrder   uint64
	wake        chan struct{}
}

// NewScheduler creates a scheduler for agent. A nil clock selects RealClock.
func NewScheduler(agent *Agent, clock Clock) *Scheduler {
	if clock == nil {
		clock = RealClock{}
	}
	return &Scheduler{
		agent: agent,
		clock: clock,
		wake:  make(chan struct{}, 1),
	}
}

// EnqueuePromiseJob appends a promise job to the FIFO microtask queue.
func (s *Scheduler) EnqueuePromiseJob(job *Job, realm *Realm) {
	s.mu.Lock()
	s.jobs = append(s.jobs, &QueuedPromiseJob{job: job, realm: realm})
	s.mu.Unlock()
	s.signal()
}

// ScheduleTimer schedules run after delay and returns an Agent-local timer id.
// Negative delays are clamped to zero, matching the host timer contract.
func (s *Scheduler) ScheduleTimer(delay time.Duration, run func() CompletionValue) JSInt {
	if delay < 0 {
		delay = 0
	}

	s.mu.Lock()
	s.nextTimerID++
	s.nextOrder++
	id := s.nextTimerID
	s.timers = append(s.timers, scheduledTask{
		id:       id,
		deadline: s.clock.Now().Add(delay),
		order:    s.nextOrder,
		run:      run,
	})
	s.mu.Unlock()
	s.signal()
	return id
}

// StartTask starts a finite asynchronous task and records it before launching
// the goroutine. Recording first avoids the Add/Wait race of sync.WaitGroup.
func (s *Scheduler) StartTask(run func()) {
	s.mu.Lock()
	s.activeTasks++
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			s.activeTasks--
			s.mu.Unlock()
			s.signal()
		}()
		run()
	}()
}

// RunJobs drains promise jobs in FIFO order. Each job is removed before it is
// run, so repeated drains never execute a completed job again.
func (s *Scheduler) RunJobs() {
	for {
		job := s.takeJob()
		if job == nil {
			return
		}
		s.runJob(job)
	}
}

// RunUntilIdle drains promise jobs, timers, and finite asynchronous tasks until
// no work remains. Jobs queued by other jobs are handled in the same turn.
func (s *Scheduler) RunUntilIdle() {
	for {
		s.RunJobs()
		if s.runDueTimers() {
			continue
		}

		hasJobs, activeTasks, nextDeadline, hasTimer := s.snapshot()
		if hasJobs {
			continue
		}
		if activeTasks == 0 && !hasTimer {
			return
		}

		if hasTimer {
			delay := nextDeadline.Sub(s.clock.Now())
			if delay <= 0 {
				continue
			}
			if activeTasks == 0 {
				<-s.clock.After(delay)
				continue
			}
			select {
			case <-s.wake:
			case <-s.clock.After(delay):
			}
			continue
		}

		// With no timer, only a tracked asynchronous task can produce more work.
		<-s.wake
	}
}

func (s *Scheduler) signal() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Scheduler) takeJob() *QueuedPromiseJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.jobs) == 0 {
		return nil
	}
	job := s.jobs[0]
	s.jobs = s.jobs[1:]
	return job
}

func (s *Scheduler) runJob(job *QueuedPromiseJob) {
	func() {
		var hostScope *ExecutionContextScope
		if !s.agent.hasExecutionContext() {
			Assert(job.realm != nil)
			hostScope = s.agent.enterExecutionContext(&ExecutionContext{
				Realm: job.realm,
				ch:    make(chan struct{}),
			})
			defer hostScope.Leave()
		}

		context := s.agent.RunningExecutionContext()
		previousRealm := context.Realm
		if job.realm != nil {
			context.Realm = job.realm
		}
		defer func() {
			context.Realm = previousRealm
		}()
		job.job.Fun(job.job.Captures)
	}()
	s.agent.waitForAsyncContinuations()
}

func (s *Scheduler) runDueTimers() bool {
	ran := false
	for {
		timer, ok := s.takeDueTimer()
		if !ok {
			return ran
		}
		ran = true
		result := timer.run()
		if result.IsAbrupt() {
			panic(result.Error())
		}
		s.RunJobs()
	}
}

func (s *Scheduler) takeDueTimer() (scheduledTask, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock.Now()
	index := -1
	for i := range s.timers {
		timer := s.timers[i]
		if timer.deadline.After(now) {
			continue
		}
		if index == -1 ||
			timer.deadline.Before(s.timers[index].deadline) ||
			(timer.deadline.Equal(s.timers[index].deadline) && timer.order < s.timers[index].order) {
			index = i
		}
	}
	if index == -1 {
		return scheduledTask{}, false
	}

	timer := s.timers[index]
	s.timers = append(s.timers[:index], s.timers[index+1:]...)
	return timer, true
}

func (s *Scheduler) snapshot() (hasJobs bool, activeTasks int, nextDeadline time.Time, hasTimer bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hasJobs = len(s.jobs) > 0
	activeTasks = s.activeTasks
	for _, timer := range s.timers {
		if !hasTimer || timer.deadline.Before(nextDeadline) {
			nextDeadline = timer.deadline
			hasTimer = true
		}
	}
	return
}
