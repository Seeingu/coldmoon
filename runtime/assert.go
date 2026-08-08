package runtime

import (
	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/pkg"
)

func defineAssert(realm *coldmoon.Realm) {
	object := realm.GlobalObject

	// failure converts an assertion failure into a Value panic carrying an
	// Error object. The builtin boundary converts it to a normal JS throw, so
	// try/catch can observe it; a bare string panic would crash the host.
	failure := func(txt string) coldmoon.Value {
		errorObject := realm.Intrinsics.ErrorConstructor.Construct(
			[]coldmoon.Value{coldmoon.NewStringValue(txt)},
			nil,
		).Data()
		return errorObject.ToValue()
	}
	var assert coldmoon.BehaviorFn = func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		equality := pkg.SliceSafeGet(argumentsList, 0)
		msg := pkg.SliceSafeGet(argumentsList, 1)
		if equality == nil {
			panic(realm.Agent.ThrowException(coldmoon.TypeError, "assert: no arguments"))
		}
		if equality.ToBoolean() {
			return coldmoon.UndefinedValue
		}
		txt := "assertion failed"
		if msg == nil {
			panic(failure(txt))
		}
		panic(failure(txt + ": " + msg.String()))
	}

	var assertEqual coldmoon.BehaviorFn = func(thisArgument coldmoon.Value, argumentsList []coldmoon.Value, newTarget coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		a := pkg.SliceSafeGet(argumentsList, 0)
		b := pkg.SliceSafeGet(argumentsList, 1)
		equality := coldmoon.IsStrictlyEqual(a, b)
		msg := pkg.SliceSafeGet(argumentsList, 2)
		if equality {
			return coldmoon.UndefinedValue
		}
		txt := "assertion failed: a is " + a.String() + ", b is " + b.String()
		if msg == nil {
			panic(failure(txt))
		}
		panic(failure(txt + "\nmsg: " + msg.String()))
	}
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("assert"), object, assert, 1)
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("assertEqual"), object, assertEqual, 2)
}
