package coldmoon

func NewGeneratorFunctionConstructor(realm *Realm) ObjectType {
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
				dynamicFunctionKindGenerator,
				parameterArgs,
				bodyArg,
			))

	}
	object := CreateBuiltinFunction(agent, behavior, 1, "GeneratorFunction", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionConstructor,
	})
	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.GeneratorFunctionPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	return object
}

func NewGeneratorFunctionPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.FunctionPrototype)
	DefineBuiltinProperty(object, "constructor", NewValueFromObject(realm.Intrinsics.GeneratorFunctionConstructor))
	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.GeneratorFunctionPrototypePrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	DefineBuiltinProperty(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("GeneratorFunction"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
