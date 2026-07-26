package coldmoon

import (
	"math"
	"net/url"
	"strconv"
	"strings"
)

type constructorProperties struct {
	Name               string
	PropertyDescriptor *PropertyDescriptor
}

var globalIntrinsicBindings = []struct {
	name      string
	intrinsic IntrinsicName
}{
	{"Boolean", IntrinsicNameBoolean},
	{"isFinite", IntrinsicNameIsFinite},
	{"isNaN", IntrinsicNameIsNaN},
	{"eval", IntrinsicNameEval},
	{"Object", IntrinsicNameObject},
	{"Function", IntrinsicNameFunction},
	{"Array", IntrinsicNameArray},
	{"String", IntrinsicNameString},
	{"Number", IntrinsicNameNumber},
	{"Symbol", IntrinsicNameSymbol},
	{"BigInt", IntrinsicNameBigInt},
	{"Math", IntrinsicNameMath},
	{"Error", IntrinsicNameError},
	{"EvalError", IntrinsicNameEvalError},
	{"RangeError", IntrinsicNameRangeError},
	{"ReferenceError", IntrinsicNameReferenceError},
	{"SyntaxError", IntrinsicNameSyntaxError},
	{"TypeError", IntrinsicNameTypeError},
	{"URIError", IntrinsicNameURIError},
	{"Reflect", IntrinsicNameReflect},
	{"Proxy", IntrinsicNameProxy},
	{"AggregateError", IntrinsicNameAggregateError},
	{"Date", IntrinsicNameDate},
	{"Map", IntrinsicNameMap},
	{"Set", IntrinsicNameSet},
	{"Promise", IntrinsicNamePromise},
	{"ArrayBuffer", IntrinsicNameArrayBuffer},
	{"SharedArrayBuffer", IntrinsicNameSharedArrayBuffer},
	{"parseInt", IntrinsicNameParseInt},
	{"parseFloat", IntrinsicNameParseFloat},
	{"RegExp", IntrinsicNameRegExp},
	{"DataView", IntrinsicNameDataView},
	{"decodeURI", IntrinsicNameDecodeURI},
	{"decodeURIComponent", IntrinsicNameDecodeURIComponent},
	{"encodeURI", IntrinsicNameEncodeURI},
	{"encodeURIComponent", IntrinsicNameEncodeURIComponent},
	{"Intl", IntrinsicNameIntl},
	{"JSON", IntrinsicNameJSON},
	{"BigInt64Array", IntrinsicNameBigInt64Array},
	{"BigUint64Array", IntrinsicNameBigUint64Array},
	{"Int8Array", IntrinsicNameInt8Array},
	{"Uint8Array", IntrinsicNameUint8Array},
	{"Uint8ClampedArray", IntrinsicNameUint8ClampedArray},
	{"Int16Array", IntrinsicNameInt16Array},
	{"Uint16Array", IntrinsicNameUint16Array},
	{"Int32Array", IntrinsicNameInt32Array},
	{"Uint32Array", IntrinsicNameUint32Array},
	{"Float32Array", IntrinsicNameFloat32Array},
	{"Float64Array", IntrinsicNameFloat64Array},
	{"Atomics", IntrinsicNameAtomics},
}

// 19.1
func GlobalObjectProperties(r *Realm) []constructorProperties {
	Assert(r.IsReady())
	propNameValues := []struct {
		name  string
		value Value
	}{
		{"globalThis", r.GlobalEnv.GlobalThisValue.ToValue()},
		{"Infinity", InfinityValue},
		{"NaN", NaNValue},
		{"undefined", UndefinedValue},
	}
	for _, binding := range globalIntrinsicBindings {
		propNameValues = append(propNameValues, struct {
			name  string
			value Value
		}{
			name:  binding.name,
			value: r.Intrinsics.Get(binding.intrinsic).ToValue(),
		})
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
	var isFinite BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		number, isAbrupt, rt := ReturnIfAbrupt(argumentAt(args, 0).ToNumber(realm.Agent), CompletionValue{})
		if isAbrupt {
			return rt
		}
		return NewBooleanValue(number.IsFinite())
	}

	return CreateBuiltinFunction(realm.Agent, isFinite, 1, CMString("isFinite"), builtinFunctionArgs{
		realm: realm,
	})
}

func NewIsNaN(realm *Realm) ObjectType {
	var isNaN BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		number, isAbrupt, rt := ReturnIfAbrupt(argumentAt(args, 0).ToNumber(realm.Agent), CompletionValue{})
		if isAbrupt {
			return rt
		}
		return NewBooleanValue(number.IsNaN())
	}

	return CreateBuiltinFunction(realm.Agent, isNaN, 1, CMString("isNaN"), builtinFunctionArgs{
		realm: realm,
	})
}

func NewEval(realm *Realm) ObjectType {
	var eval BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		if len(args) == 0 {
			return UndefinedValue
		}
		return PerformEval(realm.Agent, argumentAt(args, 0), false, false)
	}
	return CreateBuiltinFunction(realm.Agent, eval, 1, CMString("eval"), builtinFunctionArgs{
		realm: realm,
	})
}

// 7.1.17
func ToStringCompletion(agent *Agent, v Value) (co Completion[*StringValue]) {
	switch v.(type) {
	case *StringValue:
		co.value = v.(*StringValue)
	case *NumberValue:
		co.value = NewStringValue(v.String())
	case *BooleanValue:
		co.value = NewStringValue(v.String())
	case *SymbolValue:
		return co.ThrowTypeError(agent, "Cannot convert a Symbol value to a string")
	case *BigIntValue:
		co.value = NewStringValue(v.String())
	case *undefinedValue:
		co.value = NewStringValue("undefined")
	case *nullValue:
		co.value = NewStringValue("null")
	default:
		Assert(v.IsObject())
		primValue, isAbrupt, rt := ReturnIfAbrupt(v.ToPrimitive(agent, PreferredTypeString), co)
		if isAbrupt {
			return rt
		}
		Assert(!primValue.IsObject())
		return ToStringCompletion(agent, primValue)
	}
	return
}

func ToString(agent *Agent, v Value) *StringValue {
	return ReturnAssertNormal(ToStringCompletion(agent, v))
}

// 19.2.5
func NewParseInt(realm *Realm) ObjectType {
	agent := realm.Agent
	var parseInt BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		stringValue := argumentAt(arguments, 0)
		radix := argumentAt(arguments, 1)
		inputString := ToString(agent, stringValue)
		S := inputString.TrimString()
		sign := 1
		if S[0] == '-' {
			sign = -1
		}
		if S[0] == '+' || S[0] == '-' {
			S = S[1:]
		}
		R, isAbrupt, rt := ReturnIfAbrupt(ToInt32(agent, radix), co)
		if isAbrupt {
			return rt
		}
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

	return CreateBuiltinFunction(agent, parseInt, 2, CMString("parseInt"), builtinFunctionArgs{
		realm: realm,
	})
}

func NewParseFloat(realm *Realm) ObjectType {
	agent := realm.Agent
	var parseFloat BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		stringValue := argumentAt(arguments, 0)
		inputString := ToString(agent, stringValue)
		trimmedString := inputString.TrimString()

		f, err := strconv.ParseFloat(trimmedString, 64)
		if err != nil && !math.IsInf(f, 0) {
			return NaNValue
		}
		return NewNumberValue(JSNumber(f))
	}
	return CreateBuiltinFunction(agent, parseFloat, 1, CMString("parseFloat"), builtinFunctionArgs{
		realm: realm,
	})
}

func NewDecodeURI(realm *Realm) ObjectType {
	agent := realm.Agent
	var decodeURI BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		uriString := ToString(agent, argumentAt(arguments, 0)).String()
		preserveEscapeSet := ";/?:@&=+$,#"
		return NewStringValue(decode(agent, uriString, preserveEscapeSet))
	}
	return CreateBuiltinFunction(agent, decodeURI, 1, CMString("decodeURI"), builtinFunctionArgs{
		realm: realm,
	})
}

func NewDecodeURIComponent(realm *Realm) ObjectType {
	agent := realm.Agent
	var decodeURIComponent BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		uriString := ToString(agent, argumentAt(arguments, 0)).String()
		preserveEscapeSet := ""
		return NewStringValue(decode(agent, uriString, preserveEscapeSet))
	}
	return CreateBuiltinFunction(agent, decodeURIComponent, 1, CMString("decodeURIComponent"), builtinFunctionArgs{
		realm: realm,
	})
}

func NewEncodeURI(realm *Realm) ObjectType {
	agent := realm.Agent
	var encodeURI BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		uriString := ToString(agent, argumentAt(arguments, 0)).String()
		extraUnescaped := ";/?:@&=+$,#"
		return NewStringValue(encode(agent, uriString, extraUnescaped))
	}
	return CreateBuiltinFunction(agent, encodeURI, 1, CMString("encodeURI"), builtinFunctionArgs{
		realm: realm,
	})
}

func NewEncodeURIComponent(realm *Realm) ObjectType {
	agent := realm.Agent
	var encodeURIComponent BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		uriString := ToString(agent, argumentAt(arguments, 0)).String()
		extraUnescaped := ""
		return NewStringValue(encode(agent, uriString, extraUnescaped))
	}
	return CreateBuiltinFunction(agent, encodeURIComponent, 1, CMString("encodeURIComponent"), builtinFunctionArgs{
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
