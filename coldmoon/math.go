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
	var acos BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Acos(n.Data))
	}
	var acosh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Acosh(n.Data))
	}
	var asin BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Asin(n.Data))
	}
	var asinh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Asinh(n.Data))
	}
	var atan BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Atan(n.Data))
	}
	var atanh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Atanh(n.Data))
	}
	var cos BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Cos(n.Data))
	}
	var cosh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Cosh(n.Data))
	}
	var sin BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Sin(n.Data))
	}
	var sinh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Sinh(n.Data))
	}
	var tan BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Tan(n.Data))
	}
	var tanh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Tanh(n.Data))
	}
	var sqrt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		if n.Data < 0 {
			return NewNumberValue(math.NaN())
		}
		return NewNumberValue(math.Sqrt(n.Data))
	}
	var cbrt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		if n.Data < 0 {
			return NewNumberValue(-math.Cbrt(-n.Data))
		}
		return NewNumberValue(math.Cbrt(n.Data))
	}
	var exp BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Exp(n.Data))
	}
	var expm1 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		return NewNumberValue(math.Expm1(n.Data))
	}
	var log BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		if n.Data < 0 {
			return NewNumberValue(math.NaN())
		}
		return NewNumberValue(math.Log(n.Data))
	}
	var log1p BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		if n.Data < -1 {
			return NewNumberValue(math.NaN())
		}
		return NewNumberValue(math.Log1p(n.Data))
	}
	var log10 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		if n.Data < 0 {
			return NewNumberValue(math.NaN())
		}
		return NewNumberValue(math.Log10(n.Data))
	}
	var log2 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := argumentsList[0]
		n := ToNumber(agent, x)
		if n.Data < 0 {
			return NewNumberValue(math.NaN())
		}
		return NewNumberValue(math.Log2(n.Data))
	}
	var mathMax BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		coerced := make([]*NumberValue, len(argumentsList))
		for i, arg := range argumentsList {
			coerced[i] = ToNumber(agent, arg)
		}

		highest := NewNumberValue(math.Inf(-1))
		for _, number := range coerced {
			if number.IsNaN() {
				return NewNumberValue(math.NaN())
			}
			if number.IsPositiveZero() && highest.IsNegativeZero() {
				highest = number
				continue
			}
			if number.Data > highest.Data {
				highest = number
			}
		}
		return highest
	}
	var mathMin BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		coerced := make([]*NumberValue, len(argumentsList))
		for i, arg := range argumentsList {
			coerced[i] = ToNumber(agent, arg)
		}

		lowest := NewNumberValue(math.Inf(1))
		for _, number := range coerced {
			if number.IsNaN() {
				return NewNumberValue(math.NaN())
			}
			if number.IsNegativeZero() && lowest.IsPositiveZero() {
				lowest = number
				continue
			}
			if number.Data < lowest.Data {
				lowest = number
			}
		}
		return lowest
	}
	var atan2 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		y := ToNumber(agent, argumentsList[0])
		x := ToNumber(agent, argumentsList[1])
		return NewNumberValue(math.Atan2(y.Data, x.Data))
	}
	var fround BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		x := ToNumber(agent, argumentsList[0])
		return NewNumberValue(float64(float32(x.Data)))
	}
	var imul BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		a := ToUint32(agent, argumentsList[0])
		b := ToUint32(agent, argumentsList[1])
		product := a * b
		return NewNumberValue(float64(int32(product)))
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
	DefineBuiltinFunction(object, "acos", acos, 1, realm)
	DefineBuiltinFunction(object, "acosh", acosh, 1, realm)
	DefineBuiltinFunction(object, "asin", asin, 1, realm)
	DefineBuiltinFunction(object, "asinh", asinh, 1, realm)
	DefineBuiltinFunction(object, "atan", atan, 1, realm)
	DefineBuiltinFunction(object, "atanh", atanh, 1, realm)
	DefineBuiltinFunction(object, "cos", cos, 1, realm)
	DefineBuiltinFunction(object, "cosh", cosh, 1, realm)
	DefineBuiltinFunction(object, "sin", sin, 1, realm)
	DefineBuiltinFunction(object, "sinh", sinh, 1, realm)
	DefineBuiltinFunction(object, "tan", tan, 1, realm)
	DefineBuiltinFunction(object, "tanh", tanh, 1, realm)
	DefineBuiltinFunction(object, "sqrt", sqrt, 1, realm)
	DefineBuiltinFunction(object, "cbrt", cbrt, 1, realm)
	DefineBuiltinFunction(object, "exp", exp, 1, realm)
	DefineBuiltinFunction(object, "expm1", expm1, 1, realm)
	DefineBuiltinFunction(object, "log", log, 1, realm)
	DefineBuiltinFunction(object, "log1p", log1p, 1, realm)
	DefineBuiltinFunction(object, "log10", log10, 1, realm)
	DefineBuiltinFunction(object, "log2", log2, 1, realm)
	DefineBuiltinFunction(object, "atan2", atan2, 2, realm)
	DefineBuiltinFunction(object, "fround", fround, 1, realm)
	DefineBuiltinFunction(object, "imul", imul, 2, realm)
	DefineBuiltinFunction(object, "max", mathMax, 2, realm)
	DefineBuiltinFunction(object, "min", mathMin, 2, realm)

	return object
}
