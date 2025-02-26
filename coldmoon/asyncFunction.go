package coldmoon

func NewAsyncFunctionConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		parameterArgs := argumentsList[0 : len(argumentsList)-1]
		bodyArg := argumentsList[len(argumentsList)-1]
		C := agent.ActiveFunctionObject()
		if bodyArg == nil {
			bodyArg = NewStringValue("")
		}
		return CreateDynamicFunction(
			agent,
			C,
			newTarget,
			dynamicFunctionKindAsync,
			parameterArgs,
			bodyArg,
		).ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, CMString("AsyncFunction"), builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionConstructor,
	})

	BindPrototypeAndConstructor(realm.Intrinsics.AsyncFunctionPrototype, object)

	return object
}

func NewAsyncFunctionPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.FunctionPrototype, "AsyncFunctionPrototype")

	object.defineToStringTag("AsyncFunction")
	return object
}

func AsyncFunctionStart(agent *Agent, promiseCapability *PromiseCapability, asyncFunction *ECMAScriptFunction) {
	runningContext := agent.RunningExecutionContext()
	asyncContext := runningContext
	AsyncBlockStart(agent, promiseCapability, asyncFunction, asyncContext)
}

func AsyncBlockStart(agent *Agent, promiseCapability *PromiseCapability, asyncFunction *ECMAScriptFunction, asyncContext *ExecutionContext) {
	runningContext := agent.RunningExecutionContext()

	closure := func() {
		result := asyncFunction.EvaluateBody()
		agent.ExecutionContextStack.Pop()

		if result.t == CompletionTypeNormal {
			promiseCapability.Resolve.Call(
				UndefinedValue, []Value{result.Data()})
		} else {
			panic("AsyncBlockStart: completion type not normal")
		}
	}

	agent.ExecutionContextStack.Push(asyncContext)
	closure()
	Assert(runningContext == agent.RunningExecutionContext())
}
