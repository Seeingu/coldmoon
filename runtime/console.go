package runtime

import (
	"strings"

	"github.com/Seeingu/coldmoon/coldmoon"
)

func jsPrint(this coldmoon.Value, arguments []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
	var sb []string
	for _, arg := range arguments {
		sb = append(sb, string(arg.ToString()))
	}
	println(strings.Join(sb, " "))
	return coldmoon.UndefinedValue
}

// https://console.spec.whatwg.org/#console-namespace
// TODO(BM): fill up
func CreateConsole(realm *coldmoon.Realm) coldmoon.ObjectType {
	agent := realm.Agent
	proto := coldmoon.OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
	object := coldmoon.NewObject(agent, proto, "console")

	var consoleLog coldmoon.BehaviorFn = func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		jsPrint(thisArgument, argumentsList, newTarget)
		return coldmoon.UndefinedValue
	}
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("log"), object, consoleLog, 1)

	return object
}

func definePrint(realm *coldmoon.Realm) {
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("print"), realm.GlobalObject, jsPrint, 1)
}
