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
		return (CreateDynamicFunction(
			agent,
			C,
			newTarget,
			dynamicFunctionKindGenerator,
			parameterArgs,
			bodyArg,
		)).ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "GeneratorFunction", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionConstructor,
	})
	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        (realm.Intrinsics.GeneratorFunctionPrototype).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	return object
}

func NewGeneratorFunctionPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.FunctionPrototype, "GeneratorFunctionPrototype")
	DefineBuiltinPropertyP(object, "constructor", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.GeneratorFunctionConstructor),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.GeneratorFunctionPrototypePrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("GeneratorFunction"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
