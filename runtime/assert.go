package runtime

import (
	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/pkg"
)

func defineAssert(realm *coldmoon.Realm) {
	object := realm.GlobalObject

	var assert coldmoon.BehaviorFn = func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.Value {
		equality := pkg.SliceSafeGet(argumentsList, 0)
		msg := pkg.SliceSafeGet(argumentsList, 1)
		if equality == nil {
			panic("assert: no arguments")
		}
		if equality.ToBoolean() {
			return coldmoon.UndefinedValue
		}
		txt := "assertion failed"
		if msg == nil {
			panic(txt)
		}
		panic(txt + ": " + msg.String())
	}

	var assertEqual coldmoon.BehaviorFn = func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.Value {
		a := pkg.SliceSafeGet(argumentsList, 0)
		b := pkg.SliceSafeGet(argumentsList, 1)
		equality := coldmoon.IsStrictlyEqual(a, b)
		msg := pkg.SliceSafeGet(argumentsList, 2)
		if equality {
			return nil
		}
		txt := "assertion failed: a is " + a.String() + ", b is " + b.String()
		if msg == nil {
			panic(txt)
		}
		panic(txt + "\nmsg: " + msg.String())
	}
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("assert"), object, assert, 1)
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("assertEqual"), object, assertEqual, 2)
}
