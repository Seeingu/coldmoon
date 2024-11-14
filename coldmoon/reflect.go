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

	DefineBuiltinFunction(object, "apply", apply, 3, realm)
	DefineBuiltinFunction(object, "construct", construct, 3, realm)

	return object
}
