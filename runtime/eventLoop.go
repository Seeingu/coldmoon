package runtime

import (
	"time"

	"github.com/Seeingu/coldmoon/coldmoon"
)

type EventLoop interface {
	Poll()
	AddTimer(timer Timer)
}

type CMEventLoop struct {
	timers []Timer
	agent  *coldmoon.Agent
}

var _ EventLoop = (*CMEventLoop)(nil)

func (el *CMEventLoop) Poll() {
	for {
		if len(el.timers) == 0 {
			break
		}
		now := time.Now()
		var nextTimers []Timer
		for _, timer := range el.timers {
			if timer.timeout.After(now) {
				nextTimers = append(nextTimers, timer)
				continue
			}
			if !coldmoon.IsCallable(timer.callback) {
				panic("callback is not callable")
			}
			coldmoon.MustGetObject(timer.callback).Call(nil, nil)
		}
		el.timers = nextTimers

		el.agent.RunJobs()
	}
}

func (el *CMEventLoop) AddTimer(timer Timer) {
	el.timers = append(el.timers, timer)
}

func NewEventLoop(agent *coldmoon.Agent) *CMEventLoop {
	return &CMEventLoop{
		agent: agent,
	}
}
