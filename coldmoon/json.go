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

	DefineToStringTagBuiltinProperty(object, "JSON")

	var parse BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
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
	stringify := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
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
		if obj, ok := ValueGetObject(replacer); ok {
			if IsCallable(replacer) {
				replacerFunction = obj
			} else {
				isArray := IsArray(replacer)
				if isArray {
					length := obj.LengthOfArrayLike()
					for k := range length {
						prop := NewIntegerIndexPropertyKey(k)
						v := obj.Get(prop)
						var item string
						switch vv := v.(type) {
						case *StringValue:
							item = vv.Data
						case *NumberValue:
							item = vv.ToString(10)
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

		if obj, ok := ValueGetObject(space); ok {
			if ObjectIs[*NumberObject](obj) {
				space = ToNumber(agent, space)
			} else if ObjectIs[*StringObject](obj) {
				space = ToString(agent, space)
			}
		}

		var gap string
		if ValueIs[*NumberValue](space) {
			spaceMV := ToIntegerOrInfinity(agent, space)
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
		return NewStringValue(SerializeJSONProperty(agent, state, NewStringPropertyKey(""), wrapper))
	}

	DefineBuiltinFunction(object, "parse", parse, 2, realm)
	DefineBuiltinFunction(object, "stringify", stringify, 3, realm)

	j := &JSON{
		Object: object,
	}
	j.ref = j
	return j
}

// InternalizeJSONProperty
func InternalizeJSONProperty(agent *Agent, holder ObjectType, name PropertyKey, reviver Value) Value {
	value := holder.Get(name)
	if ValueIsObject(value) {
		obj := MustGetObject(value)
		isArray := IsArray(value)
		if isArray {
			length := obj.LengthOfArrayLike()
			for i := JSInt(0); i < length; i++ {
				prop := NewIntegerIndexPropertyKey(i)
				newElement := InternalizeJSONProperty(agent, obj, prop, reviver)
				if newElement == UndefinedValue {
					obj.InternalMethods().Delete(obj, prop)
				} else {
					obj.CreateDataProperty(prop, newElement)
				}
			}
		} else {
			keys := obj.EnumerableOwnProperties(objectOwnPropertiesKindKey)
			for _, key := range keys {
				prop := ToPropertyKey(agent, key)
				newElement := InternalizeJSONProperty(agent, obj, prop, reviver)
				if newElement == UndefinedValue {
					obj.InternalMethods().Delete(obj, prop)
				} else {
					obj.CreateDataProperty(prop, newElement)
				}
			}
		}
	}
	return reviver.Call(holder.ToValue(), []Value{name.ToValue(), value})
}

func SerializeJSONProperty(agent *Agent, state *JSONSerializationRecord, key PropertyKey, holder ObjectType) string {
	value := holder.Get(key)
	if ValueIsObject(value) || ValueIs[*BigIntValue](value) {
		toJSON := GetV(agent, value, NewStringPropertyKey("toJSON"))
		if IsCallable(toJSON) {
			value = toJSON.Call(value, []Value{key.ToValue()})
		}
	}

	if state.ReplacerFunction != nil {
		value = state.ReplacerFunction.ToValue().Call(holder.ToValue(), []Value{key.ToValue(), value})
	}

	if obj, ok := ValueGetObject(value); ok {
		if ObjectIs[*NumberObject](obj) {
			value = ToNumber(agent, value)
		} else if ObjectIs[*StringObject](obj) {
			value = ToString(agent, value)
		} else if ObjectIs[*BooleanObject](obj) {
			value = NewBooleanValue(value.ToBoolean())
		} else if ObjectIs[*BigIntObject](obj) {
			value = ToBigInt(agent, value)
		}
	}

	switch v := value.(type) {
	case *nullValue:
		return "null"
	case *BooleanValue:
		return value.String()
	case *NumberValue:
		return value.String()
	case *StringValue:
		return value.String()
	case *BigIntValue:
		panic("TypeError")
	case *ObjectValue:
		if !IsCallable(value) {
			isArray := IsArray(value)
			if isArray {
				return SerializeJSONArray(agent, state, v.Object)
			}
			return SerializeJSONObject(agent, state, v.Object)
		}
	}
	return ""
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

func SerializeJSONObject(agent *Agent, state *JSONSerializationRecord, value ObjectType) string {
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
			converted[i] = ToPropertyKey(agent, key)
		}
		K = converted
	} else {
		K = state.PropertyList
	}

	var partial []string
	for _, P := range K {
		strP := SerializeJSONProperty(agent, state, P, value)
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

	return final
}

func SerializeJSONArray(agent *Agent, state *JSONSerializationRecord, value ObjectType) string {
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
	length := value.LengthOfArrayLike()
	for i := JSInt(0); i < length; i++ {
		strI := SerializeJSONProperty(agent, state, NewIntegerIndexPropertyKey(i), value)
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

	return final
}
