package coldmoon

import (
	"encoding/json"
	"strings"
)

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
			var i int
			for d.More() {
				arr.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(i), decodeJSON(agent, d))
				i++
			}
			return arr.ToValue()
		case '{':
			realm := agent.CurrentRealm()
			obj := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectConstructor, nil)
			for d.More() {
				key := decodeJSON(agent, d)
				value := decodeJSON(agent, d)
				obj.CreateDataPropertyOrThrow(NewStringPropertyKey(key.String()), value)
			}
			return obj.ToValue()
		default:
			return UndefinedValue
		}
	case bool:
		return NewBooleanValue(t)
	case float64:
		return NewNumberValue(t)
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
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)

	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("JSON"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

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
	DefineBuiltinFunction(object, "parse", parse, 2, realm)

	return &JSON{
		Object: object,
	}
}

func InternalizeJSONProperty(agent *Agent, holder ObjectType, name PropertyKey, reviver Value) Value {
	value := holder.Get(name)
	if ValueIsObject(value) {
		obj := MustGetObject(value)
		isArray := IsArray(value)
		if isArray {
			length := obj.LengthOfArrayLike()
			for i := 0; i < int(length); i++ {
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
	return reviver.CallAssumeCallable(holder.ToValue(), []Value{name.ToValue(), value})
}
