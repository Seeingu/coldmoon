package runtime

import "github.com/Seeingu/coldmoon/coldmoon"

func jsPrint(this coldmoon.Value, arguments []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.Value {
	for _, arg := range arguments {
		println(arg.String())
	}
	return coldmoon.UndefinedValue
}

// https://console.spec.whatwg.org/#console-namespace
// TODO(C): fill up
func CreateConsole(realm *coldmoon.Realm) coldmoon.ObjectType {
	agent := realm.Agent
	proto := coldmoon.OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
	object := coldmoon.NewObject(agent, proto, "console")

	var consoleLog coldmoon.BehaviorFn = func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.Value {
		jsPrint(thisArgument, argumentsList, newTarget)
		return coldmoon.UndefinedValue
	}
	coldmoon.DefineBuiltinFunctionV2(realm, coldmoon.PString("log"), object, consoleLog, 1)

	return object
}

func definePrint(realm *coldmoon.Realm) {
	coldmoon.DefineBuiltinFunctionV2(realm, coldmoon.PString("print"), realm.GlobalObject, jsPrint, 1)
}
