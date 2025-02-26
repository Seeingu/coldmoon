package coldmoon

import (
	"math"
	"math/bits"
)

type MathObject struct {
	*Object
}

func NewMathObject(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "Math")

	object.defineBuiltinProperty(CMString("E"), NewFrozenPropertyDescriptor(NewNumberValue(math.E)))
	object.defineBuiltinProperty(CMString("LN10"), NewFrozenPropertyDescriptor(NewNumberValue(math.Ln10)))
	object.defineBuiltinProperty(CMString("LN2"), NewFrozenPropertyDescriptor(NewNumberValue(math.Ln2)))
	object.defineBuiltinProperty(CMString("LOG2E"), NewFrozenPropertyDescriptor(NewNumberValue(math.Log2E)))
	object.defineBuiltinProperty(CMString("LOG10E"), NewFrozenPropertyDescriptor(NewNumberValue(math.Log10E)))
	object.defineBuiltinProperty(CMString("PI"), NewFrozenPropertyDescriptor(NewNumberValue(math.Pi)))
	object.defineBuiltinProperty(CMString("SQRT1_2"), NewFrozenPropertyDescriptor(NewNumberValue(JSNumber(math.Sqrt(1/2)))))
	object.defineBuiltinProperty(CMString("SQRT2"), NewFrozenPropertyDescriptor(NewNumberValue(JSNumber(math.Sqrt2))))
	object.defineToStringTag("Math")

	agent := realm.Agent
	var random BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return NewNumberValue(JSNumber(realm.Rng.Float64()))
	}
	var abs BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Abs(n.Data.ToFloat())))
	}
	var ceil BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Ceil(n.Data.ToFloat())))
	}
	var floor BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Floor(float64(n.Data))))
	}
	var pow BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		y := argumentsList[1]
		base := x.ToNumber(agent)
		exponent := y.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Pow(float64(base.Data), float64(exponent.Data))))
	}
	var round BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Round(float64(n.Data))))
	}
	var trunc BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Trunc(float64(n.Data))))
	}
	var clz32 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := ToUint32(agent, x)
		return NewNumberValue(JSNumber(bits.LeadingZeros32(uint32(n))))
	}
	var sign BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		if n.Data.IsNaN() {
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
	var acos BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Acos(float64(n.Data))))
	}
	var acosh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Acosh(float64(n.Data))))
	}
	var asin BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Asin(float64(n.Data))))
	}
	var asinh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Asinh(float64(n.Data))))
	}
	var atan BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Atan(float64(n.Data))))
	}
	var atanh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Atanh(float64(n.Data))))
	}
	var cos BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Cos(float64(n.Data))))
	}
	var cosh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Cosh(float64(n.Data))))
	}
	var sin BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Sin(float64(n.Data))))
	}
	var sinh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Sinh(float64(n.Data))))
	}
	var tan BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Tan(float64(n.Data))))
	}
	var tanh BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Tanh(float64(n.Data))))
	}
	var sqrt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		if n.Data < 0 {
			return NaNValue
		}
		return NewNumberValue(JSNumber(math.Sqrt(float64(n.Data))))
	}
	var cbrt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		if n.Data < 0 {
			return NewNumberValue(JSNumber(-math.Cbrt(float64(-n.Data))))
		}
		return NewNumberValue(JSNumber(math.Cbrt(float64(n.Data))))
	}
	var exp BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Exp(float64(n.Data))))
	}
	var expm1 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		return NewNumberValue(JSNumber(math.Expm1(float64(n.Data))))
	}
	var log BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		if n.Data < 0 {
			return NaNValue
		}
		return NewNumberValue(JSNumber(math.Log(float64(n.Data))))
	}
	var log1p BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		if n.Data < -1 {
			return NaNValue
		}
		return NewNumberValue(JSNumber(math.Log1p(float64(n.Data))))
	}
	var log10 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		if n.Data < 0 {
			return NaNValue
		}
		return NewNumberValue(JSNumber(math.Log10(float64(n.Data))))
	}
	var log2 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0]
		n := x.ToNumber(agent)
		if n.Data < 0 {
			return NaNValue
		}
		return NewNumberValue(JSNumber(math.Log2(float64(n.Data))))
	}
	var mathMax BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		coerced := make([]*NumberValue, len(argumentsList))
		for i, arg := range argumentsList {
			coerced[i] = arg.ToNumber(agent)
		}

		highest := NewNumberValue(JSNumber(math.Inf(-1)))
		for _, number := range coerced {
			if number.IsNaN() {
				return NaNValue
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
	var mathMin BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		coerced := make([]*NumberValue, len(argumentsList))
		for i, arg := range argumentsList {
			coerced[i] = arg.ToNumber(agent)
		}

		lowest := NewNumberValue(JSNumber(math.Inf(1)))
		for _, number := range coerced {
			if number.IsNaN() {
				return NaNValue
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
	var atan2 BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		y := argumentsList[0].ToNumber(agent)
		x := argumentsList[1].ToNumber(agent)
		return NewNumberValue(JSNumber(math.Atan2(float64(y.Data), float64(x.Data))))
	}
	var fround BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		x := argumentsList[0].ToNumber(agent)
		return NewNumberValue(JSNumber(float32(x.Data)))
	}
	var imul BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		a := ToUint32(agent, argumentsList[0])
		b := ToUint32(agent, argumentsList[1])
		product := a * b
		return NewNumberValue(JSNumber(int32(product)))
	}

	object.defineBuiltinFunction(realm, CMString("random"), random, 0)
	object.defineBuiltinFunction(realm, CMString("abs"), abs, 1)
	object.defineBuiltinFunction(realm, CMString("ceil"), ceil, 1)
	object.defineBuiltinFunction(realm, CMString("floor"), floor, 1)
	object.defineBuiltinFunction(realm, CMString("pow"), pow, 2)
	object.defineBuiltinFunction(realm, CMString("round"), round, 1)
	object.defineBuiltinFunction(realm, CMString("trunc"), trunc, 1)
	object.defineBuiltinFunction(realm, CMString("clz32"), clz32, 1)
	object.defineBuiltinFunction(realm, CMString("sign"), sign, 1)
	object.defineBuiltinFunction(realm, CMString("acos"), acos, 1)
	object.defineBuiltinFunction(realm, CMString("acosh"), acosh, 1)
	object.defineBuiltinFunction(realm, CMString("asin"), asin, 1)
	object.defineBuiltinFunction(realm, CMString("asinh"), asinh, 1)
	object.defineBuiltinFunction(realm, CMString("atan"), atan, 1)
	object.defineBuiltinFunction(realm, CMString("atanh"), atanh, 1)
	object.defineBuiltinFunction(realm, CMString("cos"), cos, 1)
	object.defineBuiltinFunction(realm, CMString("cosh"), cosh, 1)
	object.defineBuiltinFunction(realm, CMString("sin"), sin, 1)
	object.defineBuiltinFunction(realm, CMString("sinh"), sinh, 1)
	object.defineBuiltinFunction(realm, CMString("tan"), tan, 1)
	object.defineBuiltinFunction(realm, CMString("tanh"), tanh, 1)
	object.defineBuiltinFunction(realm, CMString("sqrt"), sqrt, 1)
	object.defineBuiltinFunction(realm, CMString("cbrt"), cbrt, 1)
	object.defineBuiltinFunction(realm, CMString("exp"), exp, 1)
	object.defineBuiltinFunction(realm, CMString("expm1"), expm1, 1)
	object.defineBuiltinFunction(realm, CMString("log"), log, 1)
	object.defineBuiltinFunction(realm, CMString("log1p"), log1p, 1)
	object.defineBuiltinFunction(realm, CMString("log10"), log10, 1)
	object.defineBuiltinFunction(realm, CMString("log2"), log2, 1)
	object.defineBuiltinFunction(realm, CMString("atan2"), atan2, 2)
	object.defineBuiltinFunction(realm, CMString("fround"), fround, 1)
	object.defineBuiltinFunction(realm, CMString("imul"), imul, 2)
	object.defineBuiltinFunction(realm, CMString("max"), mathMax, 2)
	object.defineBuiltinFunction(realm, CMString("min"), mathMin, 2)

	return object
}
