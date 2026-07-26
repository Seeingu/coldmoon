package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type ReflectObject struct {
	*Object
}

func NewReflectObject(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "Reflect")
	agent := realm.Agent

	var apply BehaviorFn = func(_ Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		target := argumentAt(arguments, 0)
		thisArgument := argumentAt(arguments, 1)
		argumentsList := argumentAt(arguments, 2)

		var co CompletionValue
		if !IsCallable(target) {
			return co.ThrowTypeError(agent, "Reflect.apply called on non-callable")
		}

		args, isAbrupt, rt := ReturnIfAbrupt(CreateListFromArrayLike(agent, argumentsList), co)
		if isAbrupt {
			return rt
		}

		return target.Call(agent, thisArgument, args)
	}
	var construct BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		target := argumentAt(arguments, 0)
		argumentsList := argumentAt(arguments, 1)
		newTarget := argumentAt(arguments, 2)

		var co CompletionValue
		if !IsConstructor(target) {
			return co.ThrowTypeError(agent, "Reflect.construct called on non-constructor")
		}

		if len(arguments) <= 2 {
			newTarget = target
		} else if !IsConstructor(newTarget) {
			return co.ThrowTypeError(agent, "Reflect.construct second argument is not a constructor")
		}

		args, isAbrupt, rt := ReturnIfAbrupt(CreateListFromArrayLike(agent, argumentsList), co)
		if isAbrupt {
			return rt
		}
		return ObjectConstruct(target.(*ObjectValue).Object, args, newTarget.(*ObjectValue).Object).value.ToValue()
	}
	var defineProperty BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		target := argumentAt(arguments, 0)
		propertyKey := argumentAt(arguments, 1)
		attributes := argumentAt(arguments, 2)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.defineProperty called on non-object")
		}

		key, isAbrupt, rt := ReturnIfAbrupt(ToPropertyKey(agent, propertyKey), co)
		if isAbrupt {
			return rt
		}

		desc := attributes.ToPropertyDescriptor(agent)

		targetObject := MustGetObject(target)

		ret, isAbrupt, rt := ReturnIfAbrupt(
			targetObject.internalMethods().DefineOwnProperty(targetObject, key, desc),
			co,
		)
		if isAbrupt {
			return rt
		}
		return NewBooleanValue(ret)
	}
	var deleteProperty BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		target := argumentAt(arguments, 0)
		propertyKey := argumentAt(arguments, 1)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.deleteProperty called on non-object")
		}

		key, isAbrupt, rt := ReturnIfAbrupt(ToPropertyKey(agent, propertyKey), co)
		if isAbrupt {
			return rt
		}

		targetObject := MustGetObject(target)

		ret, isAbrupt, rt := ReturnIfAbrupt(
			targetObject.internalMethods().Delete(targetObject, key),
			co,
		)
		if isAbrupt {
			return rt
		}
		return NewBooleanValue(ret)
	}
	var get BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		target := argumentAt(arguments, 0)
		propertyKey := argumentAt(arguments, 1)
		receiver := pkg.SliceSafeGet(arguments, 2)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.get called on non-object")
		}

		key, isAbrupt, rt := ReturnIfAbrupt(ToPropertyKey(agent, propertyKey), co)
		if isAbrupt {
			return rt
		}

		targetObject := MustGetObject(target)

		return ReturnAssertNormal(
			targetObject.internalMethods().Get(targetObject, key, receiver),
		)
	}
	var getOwnPropertyDescriptor BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		target := argumentAt(arguments, 0)
		propertyKey := argumentAt(arguments, 1)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.getOwnPropertyDescriptor called on non-object")
		}

		key, isAbrupt, rt := ReturnIfAbrupt(ToPropertyKey(agent, propertyKey), co)
		if isAbrupt {
			return rt
		}

		targetObject := MustGetObject(target)
		desc := targetObject.internalMethods().GetOwnProperty(targetObject, key)

		if desc != nil {
			return (desc.FromPropertyDescriptor(agent, desc)).ToValue()
		}

		return UndefinedValue
	}
	var getPrototypeOf BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		target := argumentAt(arguments, 0)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.getPrototypeOf called on non-object")
		}

		targetObject := MustGetObject(target)

		return targetObject.internalMethods().GetPrototypeOf(targetObject).ToValue()
	}
	var has BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		target := argumentAt(arguments, 0)
		propertyKey := argumentAt(arguments, 1)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.has called on non-object")
		}

		key, isAbrupt, rt := ReturnIfAbrupt(ToPropertyKey(agent, propertyKey), co)
		if isAbrupt {
			return rt
		}

		targetObject := MustGetObject(target)

		hasProperty, isAbrupt, rt := ReturnIfAbrupt(
			targetObject.internalMethods().HasProperty(targetObject, key),
			co,
		)
		if isAbrupt {
			return rt
		}
		return NewBooleanValue(hasProperty)
	}
	var isExtensible BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		target := argumentAt(arguments, 0)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.isExtensible called on non-object")
		}

		targetObject := MustGetObject(target)

		return NewBooleanValue(targetObject.internalMethods().IsExtensible(targetObject))
	}
	var ownKeys BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		target := argumentAt(arguments, 0)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.ownKeys called on non-object")
		}

		targetObject := MustGetObject(target)
		propKeys := targetObject.internalMethods().OwnPropertyKeys(targetObject)
		var keys []Value
		for _, key := range propKeys {
			keys = append(keys, key.ToValue())
		}

		return (CreateArrayFromList(agent, keys)).ToValue()
	}
	var preventExtensions BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		target := argumentAt(arguments, 0)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.preventExtensions called on non-object")
		}

		targetObject := MustGetObject(target)

		ret := targetObject.internalMethods().PreventExtensions(targetObject)
		return NewBooleanValue(ret)
	}
	// 28.1.12
	var set BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		target := argumentAt(arguments, 0)
		propertyKey := argumentAt(arguments, 1)
		value := argumentAt(arguments, 2)
		receiver := pkg.SliceSafeGet(arguments, 3)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.set called on non-object")
		}

		key, isAbrupt, rt := ReturnIfAbrupt(ToPropertyKey(agent, propertyKey), co)
		if isAbrupt {
			return rt
		}

		targetObject := MustGetObject(target)

		if receiver == nil {
			receiver = target
		}

		ret, isAbrupt, rt := ReturnIfAbrupt(
			targetObject.internalMethods().Set(targetObject, key, value, receiver),
			co,
		)
		if isAbrupt {
			return rt
		}
		return NewBooleanValue(ret)
	}
	var setPrototypeOf BehaviorFn = func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		target := argumentAt(arguments, 0)
		proto := argumentAt(arguments, 1)

		if !target.IsObject() {
			return agent.ThrowTypeError("Reflect.setPrototypeOf called on non-object")
		}

		targetObject := MustGetObject(target)

		if proto != nil && !proto.IsObject() {
			return agent.ThrowTypeError("Reflect.setPrototypeOf called with non-object prototype")
		}

		ret := targetObject.internalMethods().SetPrototypeOf(targetObject, MustGetObject(proto))
		return NewBooleanValue(ret)
	}

	object.defineBuiltinFunction(realm, CMString("apply"), apply, 3)
	object.defineBuiltinFunction(realm, CMString("construct"), construct, 3)
	object.defineBuiltinFunction(realm, CMString("defineProperty"), defineProperty, 3)
	object.defineBuiltinFunction(realm, CMString("deleteProperty"), deleteProperty, 2)
	object.defineBuiltinFunction(realm, CMString("get"), get, 1)
	object.defineBuiltinFunction(realm, CMString("getOwnPropertyDescriptor"), getOwnPropertyDescriptor, 3)
	object.defineBuiltinFunction(realm, CMString("getPrototypeOf"), getPrototypeOf, 1)
	object.defineBuiltinFunction(realm, CMString("has"), has, 2)
	object.defineBuiltinFunction(realm, CMString("isExtensible"), isExtensible, 1)
	object.defineBuiltinFunction(realm, CMString("ownKeys"), ownKeys, 1)
	object.defineBuiltinFunction(realm, CMString("preventExtensions"), preventExtensions, 1)
	object.defineBuiltinFunction(realm, CMString("set"), set, 3)
	object.defineBuiltinFunction(realm, CMString("setPrototypeOf"), setPrototypeOf, 2)

	object.defineToStringTag("Reflect")

	return object
}
