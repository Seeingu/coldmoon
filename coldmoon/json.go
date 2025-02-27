package coldmoon

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Seeingu/coldmoon/pkg"
	"github.com/samber/lo"
)

type JSONSerializationRecord struct {
	// [[ReplacerFunction]]
	ReplacerFunction ObjectType
	// [[PropertyList]]
	PropertyList []PropertyKey
	// [[Gap]]
	Gap string
	// [[Stack]]
	Stack *pkg.Stack[ObjectType]
	// [[Indent]]
	Indent string
}

type JSON struct {
	*Object
}

func decodeJSON(agent *Agent, d *json.Decoder) Value {
	t, err := d.Token()
	if err != nil {
		return UndefinedValue
	}
	switch t := t.(type) {
	case json.Delim:
		switch t {
		case '[':
			arr := ArrayCreate(agent, 0, nil)
			var i JSInt
			for d.More() {
				arr.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(i), decodeJSON(agent, d))
				i++
			}
			return (arr).ToValue()
		case '{':
			realm := agent.CurrentRealm()
			obj := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectConstructor, nil)
			for d.More() {
				key := decodeJSON(agent, d)
				value := decodeJSON(agent, d)
				obj.CreateDataPropertyOrThrow(NewStringPropertyKey(key.String()), value)
			}
			return (obj).ToValue()
		default:
			return UndefinedValue
		}
	case bool:
		return NewBooleanValue(t)
	case float64:
		return NewNumberValue(JSNumber(t))
	case string:
		return NewStringValue(t)
	case nil:
		return NullValue
	default:
		return UndefinedValue
	}
}

func NewJSON(realm *Realm) *JSON {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "JSON")

	object.defineToStringTag("JSON")

	var parse BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		text := argumentsList[0]
		var reviver Value
		if len(argumentsList) > 1 {
			reviver = argumentsList[1]
		}
		jsonString := ToString(agent, text)

		r := strings.NewReader(jsonString.Data)
		decoder := json.NewDecoder(r)
		var unfiltered Value = decodeJSON(agent, decoder)
		if IsCallable(reviver) {
			root := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
			rootName := NewStringPropertyKey("")
			root.CreateDataPropertyOrThrow(rootName, unfiltered)
			return InternalizeJSONProperty(agent, root, rootName, reviver)
		} else {
			return unfiltered
		}
	}
	stringify := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		value := argumentsList[0]
		var replacer Value = UndefinedValue
		var space Value = UndefinedValue
		if len(argumentsList) > 1 {
			replacer = argumentsList[1]
		}
		if len(argumentsList) > 2 {
			space = argumentsList[2]
		}
		indent := ""
		var propertyList []PropertyKey

		var replacerFunction ObjectType
		if obj, ok := replacer.GetObject(); ok {
			if IsCallable(replacer) {
				replacerFunction = obj
			} else {
				isArray := IsArray(replacer)
				if isArray {
					length := ReturnAssertNormal(obj.LengthOfArrayLike())
					for k := range length {
						prop := NewIntegerIndexPropertyKey(k)
						v := obj.Get(prop)
						var item string
						switch vv := v.(type) {
						case *StringValue:
							item = vv.Data
						case *NumberValue:
							item = string(vv.ToString())
						case *ObjectValue:
							if ObjectIs[*StringObject](vv.Object) || ObjectIs[*NumberObject](vv.Object) {
								item = ToString(agent, v).Data
							}
						}
						if item != "" {
							propertyList = append(propertyList, NewStringPropertyKey(item))
						}
					}
				}
			}
		}

		if obj, ok := space.GetObject(); ok {
			if ObjectIs[*NumberObject](obj) {
				_space, isAbrupt, rt := ReturnIfAbrupt(space.ToNumber(agent), co)
				if isAbrupt {
					return rt
				}
				space = _space
			} else if ObjectIs[*StringObject](obj) {
				space = ToString(agent, space)
			}
		}

		var gap string
		if ValueIs[*NumberValue](space) {
			spaceMV, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, space), co)
			if isAbrupt {
				return rt
			}
			if spaceMV < 1 {
				gap = ""
			} else {
				gap = strconv.Itoa(int(spaceMV))
			}
		} else if s, ok := ValueGet[*StringValue](space); ok {
			if len(s.Data) <= 10 {
				gap = s.Data
			} else {
				gap = s.Data[:10]
			}
		}

		wrapper := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
		wrapper.CreateDataPropertyOrThrow(NewStringPropertyKey(""), value)

		stack := pkg.NewStack[ObjectType]()
		state := &JSONSerializationRecord{
			ReplacerFunction: replacerFunction,
			PropertyList:     propertyList,
			Gap:              gap,
			Stack:            stack,
			Indent:           indent,
		}
		s, isAbrupt, rt := ReturnIfAbrupt(SerializeJSONProperty(agent, state, NewStringPropertyKey(""), wrapper), co)
		if isAbrupt {
			return rt
		}
		return NewStringValue(s)
	}

	object.defineBuiltinFunction(realm, CMString("parse"), parse, 2)
	object.defineBuiltinFunction(realm, CMString("stringify"), stringify, 3)

	j := &JSON{
		Object: object,
	}
	j.ref = j
	return j
}

// InternalizeJSONProperty
// spec: 25.5.1.1
func InternalizeJSONProperty(agent *Agent, holder ObjectType, name PropertyKey, reviver Value) (co CompletionValue) {
	value := holder.Get(name)
	if value.IsObject() {
		obj := MustGetObject(value)
		isArray := IsArray(value)
		if isArray {
			length := ReturnAssertNormal(obj.LengthOfArrayLike())
			for i := JSInt(0); i < length; i++ {
				prop := NewIntegerIndexPropertyKey(i)
				newElement, isAbrupt, rt := ReturnIfAbrupt(InternalizeJSONProperty(agent, obj, prop, reviver), co)
				if isAbrupt {
					return rt
				}
				if newElement == UndefinedValue {
					obj.InternalMethods().Delete(obj, prop)
				} else {
					obj.CreateDataProperty(prop, newElement)
				}
			}
		} else {
			keys := obj.EnumerableOwnProperties(objectOwnPropertiesKindKey)
			for _, key := range keys {
				prop, isAbrupt, rt := ReturnIfAbrupt(ToPropertyKey(agent, key), co)
				if isAbrupt {
					return rt
				}
				newElement, isAbrupt, rt := ReturnIfAbrupt(InternalizeJSONProperty(agent, obj, prop, reviver), co)
				if isAbrupt {
					return rt
				}
				if newElement == UndefinedValue {
					obj.InternalMethods().Delete(obj, prop)
				} else {
					obj.CreateDataProperty(prop, newElement)
				}
			}
		}
	}
	return reviver.Call(agent, holder.ToValue(), []Value{name.ToValue(), value})
}

func SerializeJSONProperty(agent *Agent, state *JSONSerializationRecord, key PropertyKey, holder ObjectType) (co Completion[string]) {
	value := holder.Get(key)
	if value.IsObject() || ValueIs[*BigIntValue](value) {
		toJSON := ReturnAssertNormal(GetV(agent, value, NewStringPropertyKey("toJSON")))
		if IsCallable(toJSON) {
			value = ReturnAssertNormal(
				toJSON.Call(agent, value, []Value{key.ToValue()}),
			)
		}
	}

	if state.ReplacerFunction != nil {
		value = state.ReplacerFunction.Call(holder.ToValue(), []Value{key.ToValue(), value}).value
	}

	if obj, ok := value.GetObject(); ok {
		if ObjectIs[*NumberObject](obj) {
			_value, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
			if isAbrupt {
				return rt
			}
			value = _value
		} else if ObjectIs[*StringObject](obj) {
			value = ToString(agent, value)
		} else if ObjectIs[*BooleanObject](obj) {
			value = NewBooleanValue(value.ToBoolean())
		} else if ObjectIs[*BigIntObject](obj) {
			v, isAbrupt, rt := ReturnIfAbrupt(ToBigInt(agent, value), co)
			if isAbrupt {
				return rt
			}
			value = v
		}
	}

	switch v := value.(type) {
	case *nullValue:
		co.value = "null"
		return
	case *BooleanValue:
		co.value = value.String()
		return
	case *NumberValue:
		co.value = value.String()
		return
	case *StringValue:
		co.value = value.String()
	case *BigIntValue:
		return co.ThrowTypeError(agent, "TypeError")
	case *ObjectValue:
		if !IsCallable(value) {
			isArray := IsArray(value)
			if isArray {
				return SerializeJSONArray(agent, state, v.Object)
			}
			return SerializeJSONObject(agent, state, v.Object)
		}
	}
	co.value = ""
	return
}

func QuoteJSONString(agent *Agent, value string) string {
	product := "\""
	for _, c := range value {
		if c == '"' || c == '\\' {
			product += "\\" + string(c)
		} else if c == '\b' {
			product += "\\b"
		} else if c == '\f' {
			product += "\\f"
		} else if c == '\n' {
			product += "\\n"
		} else if c == '\r' {
			product += "\\r"
		} else if c == '\t' {
			product += "\\t"
		} else if c < 0x20 {
			product += UnicodeEscape(uint16(c))
		} else {
			product += string(c)
		}
	}
	product += "\""
	return product
}

func UnicodeEscape(c uint16) string {
	// TODO
	return ""
}

func SerializeJSONObject(agent *Agent, state *JSONSerializationRecord, value ObjectType) (co Completion[string]) {
	if lo.Contains(state.Stack.Data(), value) {
		panic("TypeError")
	}
	state.Stack.Push(value)
	defer state.Stack.Pop()

	stepBack := state.Indent
	defer func() {
		state.Indent = stepBack
	}()
	state.Indent = state.Indent + state.Gap

	var K []PropertyKey
	if state.PropertyList == nil {
		keys := value.EnumerableOwnProperties(objectOwnPropertiesKindKey)
		converted := make([]PropertyKey, len(keys))
		for i, key := range keys {
			v, isAbrupt, rt := ReturnIfAbrupt(ToPropertyKey(agent, key), co)
			if isAbrupt {
				return rt
			}
			converted[i] = v
		}
		K = converted
	} else {
		K = state.PropertyList
	}

	var partial []string
	for _, P := range K {
		strP, isAbrupt, rt := ReturnIfAbrupt(SerializeJSONProperty(agent, state, P, value), co)
		if isAbrupt {
			return rt
		}
		if strP != "" {
			member := QuoteJSONString(agent, P.ToValue().String()) + ": " + strP
			partial = append(partial, member)
		}
	}

	var final string
	if len(partial) == 0 {
		final = "{}"
	} else {
		sep := ",\n" + state.Indent
		properties := strings.Join(partial, sep)
		final = "{\n" + state.Indent + properties + "\n" + stepBack + "}"
	}

	co.value = final
	return
}

func SerializeJSONArray(agent *Agent, state *JSONSerializationRecord, value ObjectType) (co Completion[string]) {
	if lo.Contains(state.Stack.Data(), value) {
		panic("TypeError")
	}
	state.Stack.Push(value)
	defer state.Stack.Pop()

	stepBack := state.Indent
	defer func() {
		state.Indent = stepBack
	}()
	state.Indent = state.Indent + state.Gap

	var partial []string
	length := ReturnAssertNormal(value.LengthOfArrayLike())
	for i := JSInt(0); i < length; i++ {
		strI, isAbrupt, rt := ReturnIfAbrupt(SerializeJSONProperty(agent, state, NewIntegerIndexPropertyKey(i), value), co)
		if isAbrupt {
			return rt
		}
		if strI == "" {
			strI = "null"
		}
		partial = append(partial, strI)
	}

	var final string
	if len(partial) == 0 {
		final = "[]"
	} else {
		sep := ",\n" + state.Indent
		elements := strings.Join(partial, sep)
		final = "[\n" + state.Indent + elements + "\n" + stepBack + "]"
	}

	co.value = final
	return
}
