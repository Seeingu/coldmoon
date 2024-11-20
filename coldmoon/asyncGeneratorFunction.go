package coldmoon

// 27.4.2
func NewAsyncGeneratorFunctionConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		parameterArgs := argumentsList[0 : len(argumentsList)-1]
		bodyArg := argumentsList[len(argumentsList)-1]

		C := agent.ActiveFunctionObject()
		if bodyArg == nil {
			bodyArg = NewStringValue("")
		}
		return NewValueFromObject(
			CreateDynamicFunction(
				agent,
				C,
				newTarget,
				dynamicFunctionKindAsyncGenerator,
				parameterArgs,
				bodyArg,
			))
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "AsyncGeneratorFunction", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionConstructor,
	})

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.AsyncGeneratorFunctionPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return object
}

// 27.4.3
func NewAsyncGeneratorFunctionPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.FunctionPrototype)

	DefineBuiltinProperty(object, "constructor", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.AsyncGeneratorFunction),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.AsyncGeneratorFunctionPrototypePrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinProperty(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("AsyncGeneratorFunction"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}
