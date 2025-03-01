package runtime

import (
	"time"

	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/pkg"
)

type Timer struct {
	id       coldmoon.JSInt
	callback coldmoon.Value
	timeout  time.Time
}

var uniqueTimerId int

func getNextTimerId() coldmoon.JSInt {
	uniqueTimerId++
	return coldmoon.JSInt(uniqueTimerId)
}

func setTimeout(e EventLoop, agent *coldmoon.Agent, this coldmoon.Value, arguments []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
	callback := arguments[0]
	delay := pkg.SliceSafeGet(arguments, 1)
	if delay == nil {
		delay = coldmoon.NewNumberValue(0)
	}
	delayNumber := coldmoon.ReturnAssertNormal(delay.ToNumber(agent))
	timeout := time.Now().Add(time.Duration(delayNumber.Data.ToInt()) * time.Millisecond)
	id := getNextTimerId()
	e.AddTimer(Timer{
		id:       id,
		callback: callback,
		timeout:  timeout,
	})
	return id.ToValue()
}

func CreateSetTimeout(eventLoop EventLoop, realm *coldmoon.Realm) {
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("setTimeout"), realm.GlobalObject, func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		return setTimeout(eventLoop, realm.Agent, thisArgument, argumentsList, newTarget)
	}, 2)
}
