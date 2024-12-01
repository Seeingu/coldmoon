package coldmoon

import (
	"net/url"
	"strconv"
	"strings"
)

type constructorProperties struct {
	Name               string
	PropertyDescriptor *PropertyDescriptor
}

// 19.1
func GlobalObjectProperties(r *Realm) []constructorProperties {
	propNameValues := []struct {
		name  string
		value Value
	}{
		{"globalThis", NewValueFromObject(r.GlobalEnv.GlobalThisValue)},
		{"Infinity", InfinityValue},
		{"NaN", NaNValue},
		{"undefined", UndefinedValue},
		{"Boolean", NewValueFromObject(r.Intrinsics.BooleanConstructor)},
		{"isFinite", NewValueFromObject(r.Intrinsics.IsFinite)},
		{"isNaN", NewValueFromObject(r.Intrinsics.IsNaN)},
		{"eval", NewValueFromObject(r.Intrinsics.Eval)},
		{"Object", NewValueFromObject(r.Intrinsics.ObjectConstructor)},
		{"Function", NewValueFromObject(r.Intrinsics.FunctionConstructor)},
		{"Array", NewValueFromObject(r.Intrinsics.ArrayConstructor)},
		{"String", NewValueFromObject(r.Intrinsics.StringConstructor)},
		{"Number", NewValueFromObject(r.Intrinsics.NumberConstructor)},
		{"Symbol", NewValueFromObject(r.Intrinsics.SymbolConstructor)},
		{"BigInt", NewValueFromObject(r.Intrinsics.BigIntConstructor)},
		{"Math", NewValueFromObject(r.Intrinsics.Math)},
		{"Error", NewValueFromObject(r.Intrinsics.ErrorConstructor)},
		{"EvalError", NewValueFromObject(r.Intrinsics.EvalErrorConstructor)},
		{"RangeError", NewValueFromObject(r.Intrinsics.RangeErrorConstructor)},
		{"ReferenceError", NewValueFromObject(r.Intrinsics.ReferenceErrorConstructor)},
		{"SyntaxError", NewValueFromObject(r.Intrinsics.SyntaxErrorConstructor)},
		{"TypeError", NewValueFromObject(r.Intrinsics.TypeErrorConstructor)},
		{"URIError", NewValueFromObject(r.Intrinsics.URIErrorConstructor)},
		{"Reflect", NewValueFromObject(r.Intrinsics.Reflect)},
		{"Proxy", NewValueFromObject(r.Intrinsics.Proxy)},
		{"AggregateError", NewValueFromObject(r.Intrinsics.AggregateErrorConstructor)},
		{"Date", NewValueFromObject(r.Intrinsics.DateConstructor)},
		{"Map", NewValueFromObject(r.Intrinsics.Map)},
		{"Set", NewValueFromObject(r.Intrinsics.Set)},
		{"Promise", NewValueFromObject(r.Intrinsics.Promise)},
		{"ArrayBuffer", r.Intrinsics.ArrayBufferConstructor.ToValue()},
		{"parseInt", NewValueFromObject(r.Intrinsics.ParseInt)},
		{"parseFloat", NewValueFromObject(r.Intrinsics.ParseFloat)},
		{"RegExp", NewValueFromObject(r.Intrinsics.RegExpConstructor)},
		{"DataView", NewValueFromObject(r.Intrinsics.DataViewConstructor)},
		{"decodeURI", NewValueFromObject(r.Intrinsics.DecodeURI)},
	}

	var properties []constructorProperties
	for _, prop := range propNameValues {
		properties = append(properties, constructorProperties{
			Name: prop.name,
			PropertyDescriptor: &PropertyDescriptor{
				Value:        prop.value,
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		})
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

func NewDecodeURI(realm *Realm) ObjectType {
	agent := realm.Agent
	var decodeURI BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		uriString := ToString(agent, arguments[0]).String()
		preserveEscapeSet := ";/?:@&=+$,#"
		return NewStringValue(decode(agent, uriString, preserveEscapeSet))
	}
	return CreateBuiltinFunction(agent, decodeURI, 1, "decodeURI", builtinFunctionArgs{
		realm: realm,
	})
}

func NewDecodeURIComponent(realm *Realm) ObjectType {
	agent := realm.Agent
	var decodeURIComponent BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		uriString := ToString(agent, arguments[0]).String()
		preserveEscapeSet := ""
		return NewStringValue(decode(agent, uriString, preserveEscapeSet))
	}
	return CreateBuiltinFunction(agent, decodeURIComponent, 1, "decodeURIComponent", builtinFunctionArgs{
		realm: realm,
	})
}

func decode(agent *Agent, uriString string, reservedSet string) string {
	// TODO:
	return url.QueryEscape(uriString)
}
