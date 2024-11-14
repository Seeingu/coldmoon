package coldmoon

import "fmt"

type ReflectObject struct {
	*Object
}

func NewReflectObject(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype)
	agent := realm.Agent

	var apply BehaviorFn = func(_ Value, arguments []Value, newTarget ObjectType) Value {
		target := arguments[0]
		thisArgument := arguments[1]
		argumentsList := arguments[2]

		if !IsCallable(target) {
			panic("TypeError")
		}

		args := CreateListFromArrayLike(agent, argumentsList)

		return ValueCall(target, thisArgument, args)
	}
	var construct BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) Value {
		target := arguments[0]
		argumentsList := arguments[1]
		newTarget := arguments[2]

		if !IsConstructor(target) {
			panic("TypeError")
		}

		if len(arguments) <= 2 {
			newTarget = target
		} else if !IsConstructor(newTarget) {
			panic("TypeError")
		}

		args := CreateListFromArrayLike(agent, argumentsList)

		return NewValueFromObject(ObjectConstruct(target.(*ObjectValue).Object, args, newTarget.(*ObjectValue).Object))
	}
	var defineProperty BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) Value {
		target := arguments[0]
		propertyKey := arguments[1]
		attributes := arguments[2]

		if !ValueIsObject(target) {
			panic("TypeError")
		}

		key := ToPropertyKey(agent, propertyKey)

		desc := ToPropertyDescriptor(agent, attributes)

		targetObject := MustGetObject(target)

		ret := targetObject.InternalMethods().DefineOwnProperty(
			targetObject, key, desc,
		)
		return NewBooleanValue(ret)
	}
	var deleteProperty BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) Value {
		target := arguments[0]
		propertyKey := arguments[1]

		if !ValueIsObject(target) {
			panic("TypeError")
		}

		key := ToPropertyKey(agent, propertyKey)

		targetObject := MustGetObject(target)

		ret := targetObject.InternalMethods().Delete(targetObject, key)
		return NewBooleanValue(ret)
	}
	var get BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) Value {
		target := arguments[0]
		propertyKey := arguments[1]
		receiver := arguments[2]

		if !ValueIsObject(target) {
			panic("TypeError")
		}

		key := ToPropertyKey(agent, propertyKey)

		targetObject := MustGetObject(target)

		return targetObject.InternalMethods().Get(targetObject, key, receiver)
	}
	var getOwnPropertyDescriptor BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) Value {
		target := arguments[0]
		propertyKey := arguments[1]

		if !ValueIsObject(target) {
			panic("TypeError")
		}

		key := ToPropertyKey(agent, propertyKey)

		targetObject := MustGetObject(target)
		desc := targetObject.InternalMethods().GetOwnProperty(targetObject, key)

		if desc != nil {
			return NewValueFromObject(desc.FromPropertyDescriptor(agent, desc))
		}

		panic("return undefined")
	}
	var getPrototypeOf BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) Value {
		target := arguments[0]

		if !ValueIsObject(target) {
			panic("TypeError")
		}

		targetObject := MustGetObject(target)

		return NewValueFromObject(targetObject.InternalMethods().GetPrototypeOf(targetObject))
	}
	var has BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) Value {
		target := arguments[0]
		propertyKey := arguments[1]

		if !ValueIsObject(target) {
			panic("TypeError")
		}

		key := ToPropertyKey(agent, propertyKey)

		targetObject := MustGetObject(target)

		return NewBooleanValue(targetObject.InternalMethods().HasProperty(targetObject, key))
	}
	var isExtensible BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) Value {
		target := arguments[0]

		if !ValueIsObject(target) {
			panic("TypeError")
		}

		targetObject := MustGetObject(target)

		return NewBooleanValue(targetObject.InternalMethods().IsExtensible(targetObject))
	}
	var ownKeys BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) Value {
		target := arguments[0]

		if !ValueIsObject(target) {
			panic("TypeError")
		}

		targetObject := MustGetObject(target)
		propKeys := targetObject.InternalMethods().OwnPropertyKeys(targetObject)
		var keys []Value
		for _, key := range propKeys {
			switch k := key.(type) {
			case StringPropertyKey:
				keys = append(keys, NewStringValue(k.Value))
			case SymbolPropertyKey:
				keys = append(keys, k.Value)
			case IntegerIndexPropertyKey:
				keys = append(keys, NewStringValue(fmt.Sprintf("%d", k.Value)))
			}
		}

		return NewValueFromObject(CreateArrayFromList(agent, keys))
	}

	DefineBuiltinFunction(object, "apply", apply, 3, realm)
	DefineBuiltinFunction(object, "construct", construct, 3, realm)
	DefineBuiltinFunction(object, "defineProperty", defineProperty, 3, realm)
	DefineBuiltinFunction(object, "deleteProperty", deleteProperty, 2, realm)
	DefineBuiltinFunction(object, "get", get, 1, realm)
	DefineBuiltinFunction(object, "getOwnPropertyDescriptor", getOwnPropertyDescriptor, 3, realm)
	DefineBuiltinFunction(object, "getPrototypeOf", getPrototypeOf, 1, realm)
	DefineBuiltinFunction(object, "has", has, 2, realm)
	DefineBuiltinFunction(object, "isExtensible", isExtensible, 1, realm)
	DefineBuiltinFunction(object, "ownKeys", ownKeys, 1, realm)

	return object
}
