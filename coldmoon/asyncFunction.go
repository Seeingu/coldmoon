package coldmoon

import "fmt"

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
	go AsyncBlockStart(agent, promiseCapability, asyncFunction, asyncContext)
	asyncContext.Suspend()
}

// AsyncBlockStart
// spec: 27.7.5.2
func AsyncBlockStart(agent *Agent, promiseCapability *PromiseCapability, asyncFunction *ECMAScriptFunction, asyncContext *ExecutionContext) {
	asyncContext.awaitCh = make(chan struct{})
	runningContext := agent.RunningExecutionContext()

	closure := func() {
		agent.WG.Add(1)
		defer func() {
			agent.WG.Done()
		}()
		result := asyncFunction.EvaluateBody()
		agent.ExecutionContextStack.Pop()

		if !result.IsAbrupt() && result.t == CompletionTypeNormal {
			ReturnAssertNormal(
				promiseCapability.Resolve.Call(
					UndefinedValue, []Value{UndefinedValue}),
			)
		} else if result.t == CompletionTypeReturn {
			ReturnAssertNormal(
				promiseCapability.Resolve.Call(
					UndefinedValue, []Value{result.value}),
			)
		} else {
			fmt.Println("async block throw", result.Error().String())
			ReturnAssertNormal(
				promiseCapability.Reject.Call(
					UndefinedValue, []Value{result.Error()}),
			)
		}
	}

	agent.ExecutionContextStack.Push(asyncContext)
	closure()
	asyncContext.Resume()
	Assert(runningContext == agent.RunningExecutionContext())
}
