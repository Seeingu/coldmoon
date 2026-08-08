package runtime

import (
	"math"
	"testing"
	"time"

	"github.com/Seeingu/coldmoon/coldmoon"
)

// p1Clock advances immediately on After so scheduler drains never block.
type p1Clock struct {
	now time.Time
}

func (c *p1Clock) Now() time.Time {
	return c.now
}

func (c *p1Clock) After(delay time.Duration) <-chan time.Time {
	c.now = c.now.Add(delay)
	ready := make(chan time.Time, 1)
	ready <- c.now
	return ready
}

func newTimerTestAgent() (*coldmoon.Agent, *coldmoon.Realm, *p1Clock) {
	coldmoon.InitializeConstants()
	clock := &p1Clock{now: time.Unix(0, 0)}
	agent := coldmoon.NewAgentWithClock(clock)
	coldmoon.InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	CreateSetTimeout(realm)
	return agent, realm, clock
}

func noopCallback(realm *coldmoon.Realm) coldmoon.Value {
	// %parseInt% is a callable intrinsic; its behavior is irrelevant here.
	return realm.Intrinsics.ParseInt.ToValue()
}

// TestSetTimeoutClampsDelay verifies the host timer contract: NaN maps to 0,
// negative values clamp to 0, and Infinity or huge finite delays clamp to
// 2^31-1 ms instead of overflowing time.Duration and firing immediately.
// The advancing clock records how far the drain advanced to reach the timer.
func TestSetTimeoutClampsDelay(t *testing.T) {
	maxDelay := time.Duration(1<<31-1) * time.Millisecond

	cases := []struct {
		name        string
		delay       coldmoon.Value
		wantAdvance time.Duration
	}{
		{"zero", coldmoon.NewNumberValue(0), 0},
		{"negative", coldmoon.NewNumberValue(-5), 0},
		{"nan", coldmoon.NewNumberValue(coldmoon.JSNumber(math.NaN())), 0},
		{"infinity", coldmoon.NewNumberValue(coldmoon.JSNumber(math.Inf(1))), maxDelay},
		{"huge finite", coldmoon.NewNumberValue(1e15), maxDelay},
	}

	for _, tc := range cases {
		agent, realm, clock := newTimerTestAgent()
		callback := noopCallback(realm)

		completion := setTimeout(agent, coldmoon.UndefinedValue, []coldmoon.Value{callback, tc.delay}, nil)
		if completion.IsAbrupt() {
			t.Fatalf("%s: setTimeout failed: %v", tc.name, completion.Error())
		}
		agent.Scheduler.RunUntilIdle()

		advanced := clock.now.Sub(time.Unix(0, 0))
		if advanced != tc.wantAdvance {
			t.Fatalf("%s: drain advanced %v, want %v", tc.name, advanced, tc.wantAdvance)
		}
	}
}

// TestClearTimeoutRemovesPendingTimer verifies the clearTimeout host binding
// removes the matching timer: the drain then has nothing to wait for.
func TestClearTimeoutRemovesPendingTimer(t *testing.T) {
	agent, realm, clock := newTimerTestAgent()
	callback := noopCallback(realm)

	idCompletion := setTimeout(agent, coldmoon.UndefinedValue, []coldmoon.Value{callback, coldmoon.NewNumberValue(100)}, nil)
	if idCompletion.IsAbrupt() {
		t.Fatal("setTimeout failed")
	}
	id := idCompletion.Data()

	clearCompletion := clearTimeout(agent, coldmoon.UndefinedValue, []coldmoon.Value{id}, nil)
	if clearCompletion.IsAbrupt() {
		t.Fatal("clearTimeout failed")
	}
	agent.Scheduler.RunUntilIdle()

	if advanced := clock.now.Sub(time.Unix(0, 0)); advanced != 0 {
		t.Fatalf("drain advanced %v after clearTimeout, want 0", advanced)
	}
}
