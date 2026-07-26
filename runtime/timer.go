package runtime

import (
	"time"

	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/pkg"
)

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
	delayNumber, isAbrupt, rt := coldmoon.ReturnIfAbrupt(delay.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	id := agent.Scheduler.ScheduleTimer(time.Duration(delayNumber.Data.ToInt())*time.Millisecond, func() coldmoon.CompletionValue {
		return coldmoon.MustGetObject(callback).Call(coldmoon.UndefinedValue, nil)
	})
	return id.ToValue().ToCompletion()
}

func CreateSetTimeout(realm *coldmoon.Realm) {
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("setTimeout"), realm.GlobalObject, func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		return setTimeout(realm.Agent, thisArgument, argumentsList, newTarget)
	}, 2)
}
