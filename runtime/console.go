package runtime

import "github.com/Seeingu/coldmoon/coldmoon"

func DefineConsole(realm *coldmoon.Realm) {
	jsPrint := func(this coldmoon.Value, arguments []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.Value {
		for _, arg := range arguments {
			println(arg.String())
		}
		return coldmoon.UndefinedValue
	}
	coldmoon.DefineBuiltinFunctionV2(realm, coldmoon.PString("print"), realm.GlobalObject, jsPrint, 1)
}
