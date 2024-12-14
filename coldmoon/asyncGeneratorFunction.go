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
		return (CreateDynamicFunction(
			agent,
			C,
			newTarget,
			dynamicFunctionKindAsyncGenerator,
			parameterArgs,
			bodyArg,
		)).ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "AsyncGeneratorFunction", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionConstructor,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.AsyncGeneratorFunctionPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return object
}

// 27.4.3
func NewAsyncGeneratorFunctionPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.FunctionPrototype, "AsyncGeneratorFunctionPrototype")

	DefineBuiltinPropertyP(object, "constructor", &PropertyDescriptor{
		Value:        (realm.Intrinsics.AsyncGeneratorFunction).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.AsyncGeneratorFunctionPrototypePrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("AsyncGeneratorFunction"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}
