package coldmoon

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

	DefineBuiltinFunction(object, "apply", apply, 3, realm)
	DefineBuiltinFunction(object, "construct", construct, 3, realm)
	DefineBuiltinFunction(object, "defineProperty", defineProperty, 3, realm)

	return object
}
