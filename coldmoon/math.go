package coldmoon

import (
	"math"
	"math/bits"
)

type MathObject struct {
	*Object
}

func NewMathObject(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype)

	DefineBuiltinProperty(object, "E", &PropertyDescriptor{
		Value:        NewNumberValue(math.E),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "LN10", &PropertyDescriptor{
		Value:        NewNumberValue(math.Ln10),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "LN2", &PropertyDescriptor{
		Value:        NewNumberValue(math.Ln2),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "LOG2E", &PropertyDescriptor{
		Value:        NewNumberValue(math.Log2E),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "LOG10E", &PropertyDescriptor{
		Value:        NewNumberValue(math.Log10E),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "PI", &PropertyDescriptor{
		Value:        NewNumberValue(math.Pi),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "SQRT1_2", &PropertyDescriptor{
		Value:        NewNumberValue(math.Sqrt(1 / 2)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "SQRT2", &PropertyDescriptor{
		Value:        NewNumberValue(math.Sqrt2),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("Math"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	agent := realm.Agent
	var random BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return NewNumberValue(realm.Rng.Float64())
	}
	var abs BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Abs(n.Data))
	}
	var ceil BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Ceil(n.Data))
	}
	var floor BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Floor(n.Data))
	}
	var pow BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		y := argumentsList[1]
		base := ToNumber(agent, x)
		exponent := ToNumber(agent, y)
		return NewNumberValue(math.Pow(base.Data, exponent.Data))
	}
	var round BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Round(n.Data))
	}
	var trunc BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Trunc(n.Data))
	}
	var clz32 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToUint32(agent, x)
		return NewNumberValue(float64(bits.LeadingZeros32(n)))
	}
	var sign BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		if math.IsNaN(n.Data) {
			return NewNumberValue(n.Data)
		}
		if n.Data == 0 {
			return NewNumberValue(n.Data)
		}
		if n.Data < 0 {
			return NewNumberValue(-1)
		}
		return NewNumberValue(1)
	}
	DefineBuiltinFunction(object, "random", random, 0, realm)
	DefineBuiltinFunction(object, "abs", abs, 1, realm)
	DefineBuiltinFunction(object, "ceil", ceil, 1, realm)
	DefineBuiltinFunction(object, "floor", floor, 1, realm)
	DefineBuiltinFunction(object, "pow", pow, 2, realm)
	DefineBuiltinFunction(object, "round", round, 1, realm)
	DefineBuiltinFunction(object, "trunc", trunc, 1, realm)
	DefineBuiltinFunction(object, "clz32", clz32, 1, realm)
	DefineBuiltinFunction(object, "sign", sign, 1, realm)

	return object
}
