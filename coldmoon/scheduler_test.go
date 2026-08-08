package coldmoon

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

type advancingClock struct {
	now time.Time
}

func (c *advancingClock) Now() time.Time {
	return c.now
}

func (c *advancingClock) After(delay time.Duration) <-chan time.Time {
	c.now = c.now.Add(delay)
	ready := make(chan time.Time, 1)
	ready <- c.now
	return ready
}

// newSchedulerTestAgent creates a fully initialized Agent with deterministic
// time so scheduler tests exercise the same Realm switching as production.
func newSchedulerTestAgent(t *testing.T) (*Agent, *advancingClock) {
	t.Helper()
	InitializeConstants()
	clock := &advancingClock{now: time.Unix(0, 0)}
	agent := NewAgentWithClock(clock)
	InitializeHostDefinedRealm(agent, nil)
	return agent, clock
}

func TestSchedulerDrainsPromiseJobsExactlyOnceInFIFOOrder(t *testing.T) {
	agent, _ := newSchedulerTestAgent(t)
	realm := agent.CurrentRealm()
	var order []int

	nested := &Job{Fun: func(any) Value {
		order = append(order, 3)
		return UndefinedValue
	}}
	first := &Job{Fun: func(any) Value {
		order = append(order, 1)
		agent.Scheduler.EnqueuePromiseJob(nested, realm)
		return UndefinedValue
	}}
	second := &Job{Fun: func(any) Value {
		order = append(order, 2)
		return UndefinedValue
	}}

	agent.Scheduler.EnqueuePromiseJob(first, realm)
	agent.Scheduler.EnqueuePromiseJob(second, realm)
	agent.Scheduler.RunUntilIdle()
	agent.Scheduler.RunUntilIdle()

	if !reflect.DeepEqual(order, []int{1, 2, 3}) {
		t.Fatalf("promise job order = %v, want [1 2 3]", order)
	}
}

func TestSchedulerRunsTimersByDeadlineAndInsertionOrder(t *testing.T) {
	agent, _ := newSchedulerTestAgent(t)
	var order []int
	schedule := func(delay time.Duration, value int) {
		agent.Scheduler.ScheduleTimer(delay, func() CompletionValue {
			order = append(order, value)
			return UndefinedValue.ToCompletion()
		})
	}

	schedule(10*time.Millisecond, 1)
	schedule(5*time.Millisecond, 0)
	schedule(10*time.Millisecond, 2)
	agent.Scheduler.RunUntilIdle()

	if !reflect.DeepEqual(order, []int{0, 1, 2}) {
		t.Fatalf("timer order = %v, want [0 1 2]", order)
	}
}

// TestSchedulerDrainsRemainingTimersBeforeRethrowing verifies that one expected
// JavaScript failure cannot strand later work in the same Agent scheduler.
func TestSchedulerDrainsRemainingTimersBeforeRethrowing(t *testing.T) {
	agent, _ := newSchedulerTestAgent(t)
	var order []int
	agent.Scheduler.ScheduleTimer(0, func() (completion CompletionValue) {
		order = append(order, 1)
		return completion.ThrowTypeError(agent, "first timer failed")
	})
	agent.Scheduler.ScheduleTimer(0, func() CompletionValue {
		order = append(order, 2)
		return UndefinedValue.ToCompletion()
	})

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		agent.Scheduler.RunUntilIdle()
	}()

	if _, ok := recovered.(Value); !ok {
		t.Fatalf("recovered = %T, want JavaScript Value", recovered)
	}
	if !reflect.DeepEqual(order, []int{1, 2}) {
		t.Fatalf("timer order = %v, want [1 2]", order)
	}
	// A second drain proves the first call left no failed or successful timer
	// behind after reporting the recorded language error.
	agent.Scheduler.RunUntilIdle()
}

// TestSchedulerRejectsNonThrowTimerAbruptCompletion keeps internal control-flow
// completions from being mislabeled (or silently discarded) as JavaScript errors.
func TestSchedulerRejectsNonThrowTimerAbruptCompletion(t *testing.T) {
	agent, _ := newSchedulerTestAgent(t)
	agent.Scheduler.ScheduleTimer(0, func() CompletionValue {
		return CompletionValue{t: CompletionTypeBreak}
	})

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		agent.Scheduler.RunUntilIdle()
	}()

	if recovered == nil {
		t.Fatal("RunUntilIdle accepted a non-throw abrupt timer completion")
	}
	if _, ok := recovered.(Value); ok {
		t.Fatalf("recovered = %T, want an invariant panic rather than a JavaScript Value", recovered)
	}
}

func TestSchedulerWaitsForTrackedTasksAndTheirJobs(t *testing.T) {
	agent, _ := newSchedulerTestAgent(t)
	realm := agent.CurrentRealm()
	ran := false

	agent.Scheduler.StartTask(func() {
		agent.Scheduler.EnqueuePromiseJob(&Job{Fun: func(any) Value {
			ran = true
			return UndefinedValue
		}}, realm)
	})
	agent.Scheduler.RunUntilIdle()

	if !ran {
		t.Fatal("tracked task's promise job did not run")
	}
}

// TestConcurrentAgentsShareWellKnownSymbolsWithoutRacing guards the
// initWellKnownSymbols registry: hosts such as the Test262 runner create
// agents in parallel, and the shared map must be populated once under -race.
func TestConcurrentAgentsShareWellKnownSymbolsWithoutRacing(t *testing.T) {
	const agents = 16
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range agents {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			agent := NewAgent()
			if WellKnownSymbols[WellKnownSymbolsIterator] == nil {
				t.Error("well-known symbols were not initialized")
			}
			InitializeHostDefinedRealm(agent, nil)
		}()
	}
	close(start)
	wg.Wait()
}

// TestSchedulerDoesNotWaitForAwaitSuspendedTasks reproduces the 1.2 deadlock:
// a task goroutine suspended on an await that no promise job will ever settle
// must not keep idle detection alive. RunUntilIdle returns, and a later wake
// source can resume the continuation through ResumeAwait.
func TestSchedulerDoesNotWaitForAwaitSuspendedTasks(t *testing.T) {
	agent, _ := newSchedulerTestAgent(t)
	wake := make(chan struct{})
	agent.Scheduler.StartTask(func() {
		agent.Scheduler.SuspendAwait()
		<-wake
		agent.Scheduler.ResumeAwait()
	})

	// RunUntilIdle must return even though the task goroutine is alive but
	// suspended waiting for a future wake source.
	agent.Scheduler.RunUntilIdle()

	close(wake)
	agent.Scheduler.RunUntilIdle()
}

// TestAbortAsyncContinuationsReleasesWaiters verifies that a panicked async
// task closes the completion channels of suspended continuations, so promise
// jobs sequencing through waitForAsyncContinuations cannot hang the main
// thread on top of the crash.
func TestAbortAsyncContinuationsReleasesWaiters(t *testing.T) {
	agent := NewAgent()
	context := &ExecutionContext{
		awaitCh:               make(chan struct{}),
		asyncContinuationDone: make(chan struct{}),
	}
	agent.asyncContinuationContexts[context] = struct{}{}

	agent.abortAsyncContinuations()

	if context.asyncContinuationDone != nil {
		t.Fatal("asyncContinuationDone was not cleared")
	}
	if _, registered := agent.asyncContinuationContexts[context]; !registered {
		t.Fatal("continuation was dropped from the registry")
	}
	if context.Result.t != CompletionTypeThrow {
		t.Fatalf("Result type = %v, want throw", context.Result.t)
	}
}
