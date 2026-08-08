package runtime

import (
	"math"
	"time"

	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/pkg"
)

// maxTimerDelayMilliseconds is the largest delay the host timer accepts
// (2^31-1 ms, matching the HTML timer contract).
const maxTimerDelayMilliseconds = 1<<31 - 1

func setTimeout(agent *coldmoon.Agent, this coldmoon.Value, arguments []coldmoon.Value, newTarget coldmoon.ObjectType) (co coldmoon.CompletionValue) {
	callback := pkg.SliceSafeGet(arguments, 0)
	if callback == nil {
		callback = coldmoon.UndefinedValue
	}
	if !coldmoon.IsCallable(callback) {
		return co.ThrowTypeError(agent, "setTimeout callback is not callable")
	}
	delay := pkg.SliceSafeGet(arguments, 1)
	if delay == nil {
		delay = coldmoon.NewNumberValue(0)
	}
	// The delay follows ToIntegerOrInfinity (NaN→0, ±Inf preserved) and is then
	// clamped to [0, 2^31-1] ms per the HTML timer contract. The conversion to
	// JSInt must happen after the Infinity clamp: converting +Inf to an integer
	// yields a huge negative value on Go, which used to overflow time.Duration
	// and fire the timer immediately.
	delayNumber, isAbrupt, rt := coldmoon.ReturnIfAbrupt(delay.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	delayMilliseconds := delayNumber.Data.ToFloat()
	if math.IsNaN(delayMilliseconds) || math.IsInf(delayMilliseconds, -1) {
		delayMilliseconds = 0
	} else if math.IsInf(delayMilliseconds, 1) {
		delayMilliseconds = maxTimerDelayMilliseconds
	} else {
		delayMilliseconds = math.Trunc(delayMilliseconds)
		if delayMilliseconds < 0 {
			delayMilliseconds = 0
		} else if delayMilliseconds > maxTimerDelayMilliseconds {
			delayMilliseconds = maxTimerDelayMilliseconds
		}
	}
	id := agent.Scheduler.ScheduleTimer(time.Duration(delayMilliseconds)*time.Millisecond, func() coldmoon.CompletionValue {
		return coldmoon.MustGetObject(callback).Call(coldmoon.UndefinedValue, nil)
	})
	return id.ToValue().ToCompletion()
}

func clearTimeout(agent *coldmoon.Agent, this coldmoon.Value, arguments []coldmoon.Value, newTarget coldmoon.ObjectType) (co coldmoon.CompletionValue) {
	idValue := pkg.SliceSafeGet(arguments, 0)
	if idValue == nil {
		return coldmoon.UndefinedValue.ToCompletion()
	}
	idNumber, isAbrupt, rt := coldmoon.ReturnIfAbrupt(idValue.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	agent.Scheduler.CancelTimer(idNumber.Data.ToInt())
	return coldmoon.UndefinedValue.ToCompletion()
}

func CreateSetTimeout(realm *coldmoon.Realm) {
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("setTimeout"), realm.GlobalObject, func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		return setTimeout(realm.Agent, thisArgument, argumentsList, newTarget)
	}, 2)
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("clearTimeout"), realm.GlobalObject, func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		return clearTimeout(realm.Agent, thisArgument, argumentsList, newTarget)
	}, 1)
}
