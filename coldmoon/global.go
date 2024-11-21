package coldmoon

import (
	"strconv"
	"strings"
)

type constructorProperties struct {
	Name               string
	PropertyDescriptor *PropertyDescriptor
}

// 19.1
func GlobalObjectProperties(r *Realm) []constructorProperties {
	properties := []constructorProperties{
		{
			"globalThis",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.GlobalEnv.GlobalThisValue),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Infinity",
			&PropertyDescriptor{
				Value:        InfinityValue,
				Writable:     false,
				Enumerable:   false,
				Configurable: false,
			},
		},
		{
			"NaN",
			&PropertyDescriptor{
				Value:        NaNValue,
				Writable:     false,
				Enumerable:   false,
				Configurable: false,
			},
		},
		{
			"undefined",
			&PropertyDescriptor{
				Value:        UndefinedValue,
				Writable:     false,
				Enumerable:   false,
				Configurable: false,
			},
		},
		{
			"Boolean",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.BooleanConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"isFinite",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.IsFinite),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"isNaN",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.IsNaN),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"eval",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Eval),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Object",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ObjectConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Function",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.FunctionConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Array",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ArrayConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"String",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.StringConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Number",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.NumberConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Symbol",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.SymbolConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"BigInt",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.BigIntConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Math",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Math),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Error",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"EvalError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.EvalErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"RangeError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.RangeErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"ReferenceError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ReferenceErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"SyntaxError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.SyntaxErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"TypeError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.TypeErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"URIError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.URIErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Reflect",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Reflect),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Proxy",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Proxy),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"AggregateError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.AggregateErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Date",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.DateConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Map",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Map),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"SetObject",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Set),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Promise",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Promise),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"ArrayBuffer",
			&PropertyDescriptor{
				Value:        r.Intrinsics.ArrayBufferConstructor.ToValue(),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"parseInt",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ParseInt),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"parseFloat",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ParseFloat),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
	}
	return properties
}

func NewIsFinite(realm *Realm) ObjectType {
	var isFinite BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		number := ToNumber(realm.Agent, args[0])
		return NewBooleanValue(number.IsFinite())
	}

	return CreateBuiltinFunction(realm.Agent, isFinite, 1, "isFinite", builtinFunctionArgs{
		realm: realm,
	})
}

func NewIsNaN(realm *Realm) ObjectType {
	var isNaN BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		number := ToNumber(realm.Agent, args[0])
		return NewBooleanValue(number.IsNaN())
	}

	return CreateBuiltinFunction(realm.Agent, isNaN, 1, "isNaN", builtinFunctionArgs{
		realm: realm,
	})
}

func NewEval(realm *Realm) ObjectType {
	var eval BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		if len(args) == 0 {
			return nil
		}
		return PerformEval(realm.Agent, args[0], false, false)
	}
	return CreateBuiltinFunction(realm.Agent, eval, 1, "eval", builtinFunctionArgs{
		realm: realm,
	})
}

// 7.1.17
func ToString(agent *Agent, v Value) *StringValue {
	switch v.(type) {
	case *StringValue:
		return v.(*StringValue)
	case *NumberValue:
		return NewStringValue(v.String())
	case *BooleanValue:
		return NewStringValue(v.String())
	case *SymbolValue:
		return NewStringValue(v.String())
	case *BigIntValue:
		return NewStringValue(v.String())
	case *undefinedValue:
		return NewStringValue("undefined")
	case *nullValue:
		return NewStringValue("null")
	default:
		Assert(ValueIsObject(v))
		primValue := ToPrimitive(agent, v, PreferredTypeString)
		Assert(!ValueIsObject(primValue))
		return ToString(agent, primValue)
	}

}

// 19.2.5
func NewParseInt(realm *Realm) ObjectType {
	agent := realm.Agent
	var parseInt BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		stringValue := arguments[0]
		radix := arguments[1]
		inputString := ToString(agent, stringValue)
		S := inputString.TrimString()
		sign := 1
		if S[0] == '-' {
			sign = -1
		}
		if S[0] == '+' || S[0] == '-' {
			S = S[1:]
		}
		R := ToInt32(agent, radix)
		stripPrefix := true
		if R != 0 {
			if R < 2 || R > 36 {
				return NaNValue
			}
			if R != 16 {
				stripPrefix = false
			}
		} else {
			R = 10
		}
		if stripPrefix {
			if strings.HasPrefix(S, "0x") || strings.HasPrefix(S, "0X") {
				S = S[2:]
				R = 16
			}
		}

		var mathInt int64
		for _, c := range S {
			if c >= '0' && c <= '9' {
				c -= '0'
			} else if c >= 'a' && c <= 'z' {
				c -= 'a' - 10
			} else if c >= 'A' && c <= 'Z' {
				c -= 'A' - 10
			} else {
				break
			}
			if int(c) >= int(R) {
				break
			}
			mathInt = mathInt*int64(R) + int64(c)
		}
		return NewNumberValue(float64(sign) * float64(mathInt))
	}

	return CreateBuiltinFunction(agent, parseInt, 2, "parseInt", builtinFunctionArgs{
		realm: realm,
	})
}

func NewParseFloat(realm *Realm) ObjectType {
	agent := realm.Agent
	var parseFloat BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		stringValue := arguments[0]
		inputString := ToString(agent, stringValue)
		trimmedString := inputString.TrimString()

		f, err := strconv.ParseFloat(trimmedString, 64)
		if err != nil {
			return NaNValue
		}
		return NewNumberValue(f)
	}
	return CreateBuiltinFunction(agent, parseFloat, 1, "parseFloat", builtinFunctionArgs{
		realm: realm,
	})

}
