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
		{"globalThis", r.GlobalEnv.GlobalThisValue.ToValue()},
		{"Infinity", InfinityValue},
		{"NaN", NaNValue},
		{"undefined", UndefinedValue},
		{"Boolean", r.Intrinsics.BooleanConstructor.ToValue()},
		{"isFinite", r.Intrinsics.IsFinite.ToValue()},
		{"isNaN", r.Intrinsics.IsNaN.ToValue()},
		{"eval", r.Intrinsics.Eval.ToValue()},
		{"Object", r.Intrinsics.ObjectConstructor.ToValue()},
		{"Function", r.Intrinsics.FunctionConstructor.ToValue()},
		{"Array", r.Intrinsics.ArrayConstructor.ToValue()},
		{"String", r.Intrinsics.StringConstructor.ToValue()},
		{"Number", r.Intrinsics.NumberConstructor.ToValue()},
		{"Symbol", r.Intrinsics.SymbolConstructor.ToValue()},
		{"BigInt", r.Intrinsics.BigIntConstructor.ToValue()},
		{"Math", r.Intrinsics.Math.ToValue()},
		{"Error", r.Intrinsics.ErrorConstructor.ToValue()},
		{"EvalError", r.Intrinsics.EvalErrorConstructor.ToValue()},
		{"RangeError", r.Intrinsics.RangeErrorConstructor.ToValue()},
		{"ReferenceError", r.Intrinsics.ReferenceErrorConstructor.ToValue()},
		{"SyntaxError", r.Intrinsics.SyntaxErrorConstructor.ToValue()},
		{"TypeError", r.Intrinsics.TypeErrorConstructor.ToValue()},
		{"URIError", r.Intrinsics.URIErrorConstructor.ToValue()},
		{"Reflect", r.Intrinsics.Reflect.ToValue()},
		{"Proxy", r.Intrinsics.Proxy.ToValue()},
		{"AggregateError", r.Intrinsics.AggregateErrorConstructor.ToValue()},
		{"Date", r.Intrinsics.DateConstructor.ToValue()},
		{"Map", r.Intrinsics.Map.ToValue()},
		{"Set", r.Intrinsics.Set.ToValue()},
		{"Promise", r.Intrinsics.Promise.ToValue()},
		{"ArrayBuffer", r.Intrinsics.ArrayBufferConstructor.ToValue()},
		{"SharedArrayBuffer", r.Intrinsics.SharedArrayBufferConstructor.ToValue()},
		{"parseInt", r.Intrinsics.ParseInt.ToValue()},
		{"parseFloat", r.Intrinsics.ParseFloat.ToValue()},
		{"RegExp", r.Intrinsics.RegExpConstructor.ToValue()},
		{"DataView", r.Intrinsics.DataViewConstructor.ToValue()},
		{"decodeURI", r.Intrinsics.DecodeURI.ToValue()},
		{"decodeURIComponent", r.Intrinsics.DecodeURIComponent.ToValue()},
		{"encodeURI", r.Intrinsics.EncodeURI.ToValue()},
		{"encodeURIComponent", r.Intrinsics.EncodeURIComponent.ToValue()},
		{"Intl", r.Intrinsics.Intl.ToValue()},
		{"JSON", r.Intrinsics.JSON.ToValue()},
		{"BigInt64Array", r.Intrinsics.BigInt64ArrayConstructor.ToValue()},
		{"BigUint64Array", r.Intrinsics.BigUint64ArrayConstructor.ToValue()},
		{"Int8Array", r.Intrinsics.Int8ArrayConstructor.ToValue()},
		{"Uint8Array", r.Intrinsics.Uint8ArrayConstructor.ToValue()},
		{"Uint8ClampedArray", r.Intrinsics.Uint8ClampedArrayConstructor.ToValue()},
		{"Int16Array", r.Intrinsics.Int16ArrayConstructor.ToValue()},
		{"Uint16Array", r.Intrinsics.Uint16ArrayConstructor.ToValue()},
		{"Int32Array", r.Intrinsics.Int32ArrayConstructor.ToValue()},
		{"Uint32Array", r.Intrinsics.Uint32ArrayConstructor.ToValue()},
		{"Float32Array", r.Intrinsics.Float32ArrayConstructor.ToValue()},
		{"Float64Array", r.Intrinsics.Float64ArrayConstructor.ToValue()},
		{"Atomics", r.Intrinsics.Atomics.ToValue()},
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
// TODO(C): should use ToString
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
		return NewNumberValue(JSNumber(sign) * JSNumber(mathInt))
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
		return NewNumberValue(JSNumber(f))
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

func NewEncodeURI(realm *Realm) ObjectType {
	agent := realm.Agent
	var encodeURI BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		uriString := ToString(agent, arguments[0]).String()
		extraUnescaped := ";/?:@&=+$,#"
		return NewStringValue(encode(agent, uriString, extraUnescaped))
	}
	return CreateBuiltinFunction(agent, encodeURI, 1, "encodeURI", builtinFunctionArgs{
		realm: realm,
	})
}

func NewEncodeURIComponent(realm *Realm) ObjectType {
	agent := realm.Agent
	var encodeURIComponent BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		uriString := ToString(agent, arguments[0]).String()
		extraUnescaped := ""
		return NewStringValue(encode(agent, uriString, extraUnescaped))
	}
	return CreateBuiltinFunction(agent, encodeURIComponent, 1, "encodeURIComponent", builtinFunctionArgs{
		realm: realm,
	})
}

func decode(agent *Agent, uriString string, reservedSet string) string {
	// TODO:
	return url.QueryEscape(uriString)
}

func encode(agent *Agent, s string, extraUnescaped string) string {
	// TODO:
	r, err := url.QueryUnescape(s)
	if err != nil {
		panic("url.QueryUnescape error")
	}
	return r
}
