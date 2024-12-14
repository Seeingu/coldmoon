package coldmoon

func NewAsyncFunctionConstructor(realm *Realm) ObjectType {
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
			dynamicFunctionKindAsync,
			parameterArgs,
			bodyArg,
		)).ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "AsyncFunction", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionConstructor,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        (realm.Intrinsics.AsyncFunctionPrototype).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(realm.Intrinsics.AsyncFunctionPrototype, "constructor", &PropertyDescriptor{
		Value:        (object).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}

func NewAsyncFunctionPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.FunctionPrototype, "AsyncFunctionPrototype")

	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("AsyncFunction"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}

func AsyncFunctionStart(agent *Agent, promiseCapability *PromiseCapability, asyncFunction *ECMAScriptFunction) {
	runningContext := agent.runningExecutionContext()
	asyncContext := runningContext
	AsyncBlockStart(agent, promiseCapability, asyncFunction, asyncContext)
}

func AsyncBlockStart(agent *Agent, promiseCapability *PromiseCapability, asyncFunction *ECMAScriptFunction, asyncContext *ExecutionContext) {
	runningContext := agent.runningExecutionContext()

	closure := func() {
		result := asyncFunction.EvaluateBody()
		agent.ExecutionContextStack.Pop()

		if result.Type == CompletionTypeNormal {
			(promiseCapability.Resolve).ToValue().CallAssumeCallable(
				UndefinedValue, []Value{result.Data()})
		} else {
			panic("AsyncBlockStart: completion type not normal")
		}
	}

	agent.ExecutionContextStack.Push(asyncContext)
	closure()
	Assert(runningContext == agent.runningExecutionContext())
}
