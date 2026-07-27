package coldmoon

import (
	"reflect"
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
