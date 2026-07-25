package coldmoon

import (
	"fmt"

	"github.com/Seeingu/coldmoon/pkg"
)

type AsyncGeneratorRequest struct {
	// [[Completion]]
	Completion CompletionValue
	// [[Capability]]
	Capability *PromiseCapability
}

type AsyncGeneratorState int

const (
	AsyncGeneratorStateUndefined AsyncGeneratorState = iota
	AsyncGeneratorStateSuspendedStart
	AsyncGeneratorStateSuspendedYield
	AsyncGeneratorStateExecuting
	AsyncGeneratorStateAwaitingReturn
	AsyncGeneratorStateCompleted
)

type AsyncGeneratorObject struct {
	*Object
	// [[AsyncGeneratorState]]
	AsyncGeneratorState AsyncGeneratorState
	// [[AsyncGeneratorContext]]
	// get from agent

	// [[AsyncGeneratorQueue]]
	AsyncGeneratorQueue pkg.Queue[AsyncGeneratorRequest]
	// [[GeneratorBrand]]
	GeneratorBrand string
	closure        func()
}

// AsyncGeneratorContext is [[AsyncGeneratorContext]]
func (a *AsyncGeneratorObject) AsyncGeneratorContext(agent *Agent) *ExecutionContext {
	return agent.FindExecutionContextById(a.GetId())
}

func NewAsyncGeneratorPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.AsyncIteratorPrototype, "AsyncGeneratorPrototype")
	g := &AsyncGeneratorObject{
		Object: object,
	}
	g.ref = g

	g.defineBuiltinProperty(CMString("constructor"), &PropertyDescriptor{
		Value:        realm.Intrinsics.AsyncGeneratorFunctionPrototype.ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	g.defineToStringTag("AsyncGenerator")

	// spec: 27.6.1.2
	var next BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		value := pkg.SliceSafeGet(argumentsList, 0)
		generatorValue := this
		promiseCapability := NewPromiseCapability(agent, realm.Intrinsics.Promise.ToValue())
		result := AsyncGeneratorValidate(generatorValue, "")
		IfAbruptRejectPromise(agent, result, promiseCapability)
		generator := MustGetObject(generatorValue).(*AsyncGeneratorObject)
		state := generator.AsyncGeneratorState
		if state == AsyncGeneratorStateCompleted {
			iteratorResult := CreateIterResultObject(agent, UndefinedValue, true)
			ReturnAssertNormal(
				promiseCapability.Resolve.Call(UndefinedValue, []Value{iteratorResult.ToValue()}),
			)
			return promiseCapability.Promise.ToValue()
		}
		var completion CompletionValue
		completion.value = value
		AsyncGeneratorEnqueue(agent, generator, completion, promiseCapability)
		if state == AsyncGeneratorStateSuspendedStart || state == AsyncGeneratorStateSuspendedYield {
			AsyncGeneratorResume(agent, generator, completion)
		}
		return promiseCapability.Promise.ToValue()
	}
	g.defineBuiltinFunction(realm, CMString("next"), next, 1)

	var returnFn BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		generatorValue := this
		value := argumentsList[0]
		promiseCapability := NewPromiseCapability(agent, realm.Intrinsics.Promise.ToValue())
		result := AsyncGeneratorValidate(generatorValue, "")
		IfAbruptRejectPromise(agent, result, promiseCapability)
		generator := MustGetObject(generatorValue).(*AsyncGeneratorObject)
		var completion CompletionValue
		completion.t = CompletionTypeReturn
		completion.value = value
		AsyncGeneratorEnqueue(agent, generator, completion, promiseCapability)
		state := generator.AsyncGeneratorState
		if state == AsyncGeneratorStateSuspendedStart || state == AsyncGeneratorStateCompleted {
			generator.AsyncGeneratorState = AsyncGeneratorStateAwaitingReturn
			ReturnAssertNormal(
				AsyncGeneratorAwaitReturn(agent, generator),
			)
		} else if state == AsyncGeneratorStateSuspendedYield {
			AsyncGeneratorResume(agent, generator, completion)
		} else {
			Assert(state == AsyncGeneratorStateExecuting || state == AsyncGeneratorStateAwaitingReturn)
		}
		return promiseCapability.Promise.ToValue()
	}
	g.defineBuiltinFunction(realm, CMString("return"), returnFn, 1)

	return g
}

// AsyncGeneratorValidate
// spec: 27.6.3.3
// returns UNUSED or throw
func AsyncGeneratorValidate(generator Value, generatorBrand string) (co CompletionValue) {
	RequireInternalSlot[*AsyncGeneratorObject](generator)
	return
}

// AsyncGeneratorEnqueue
// spec: 27.6.3.4
// returns UNUSED
func AsyncGeneratorEnqueue(
	agent *Agent,
	generator *AsyncGeneratorObject,
	completion CompletionValue,
	promiseCapability *PromiseCapability,
) {
	request := AsyncGeneratorRequest{
		Completion: completion,
		Capability: promiseCapability,
	}
	generator.AsyncGeneratorQueue.Enqueue(request)
}

// AsyncGeneratorResume
// spec: 27.6.3.6
// returns UNUSED
func AsyncGeneratorResume(agent *Agent, generator *AsyncGeneratorObject, value CompletionValue) {
	Assert(generator.AsyncGeneratorState == AsyncGeneratorStateSuspendedStart || generator.AsyncGeneratorState == AsyncGeneratorStateSuspendedYield)
	genContext := generator.AsyncGeneratorContext(agent)
	callerContext := agent.RunningExecutionContext()
	generator.AsyncGeneratorState = AsyncGeneratorStateExecuting
	fmt.Println("resume")
	agent.ExecutionContextStack.Push(genContext)
	go generator.closure()
	genContext.Suspend()
	fmt.Println("resume after suspended")
	Assert(callerContext == agent.RunningExecutionContext())
}

// AsyncGeneratorAwaitReturn
// spec: 27.6.3.9
// returns UNUSED or throw
func AsyncGeneratorAwaitReturn(agent *Agent, generator *AsyncGeneratorObject) (co CompletionValue) {
	realm := agent.CurrentRealm()
	queue := generator.AsyncGeneratorQueue
	Assert(!queue.IsEmpty())
	next := queue.Data()[0]
	completion := next.Completion
	Assert(completion.t == CompletionTypeReturn)
	promise := PromiseResolve(agent, realm.Intrinsics.Promise, completion.value)
	var fulfilledClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		value := argumentsList[0]
		generator.AsyncGeneratorState = AsyncGeneratorStateCompleted
		var result CompletionValue
		result.value = value
		AsyncGeneratorCompleteStep(agent, generator, result, true)
		AsyncGeneratorDrainQueue(agent, generator)
		return UndefinedValue
	}
	onFulfilled := CreateBuiltinFunction(agent, fulfilledClosure, 1, CMString("onFulfilled"), builtinFunctionArgs{})
	var rejectedClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		reason := argumentsList[0]
		generator.AsyncGeneratorState = AsyncGeneratorStateCompleted
		var result CompletionValue
		result.value = reason
		AsyncGeneratorCompleteStep(agent, generator, result, true)
		AsyncGeneratorDrainQueue(agent, generator)
		return UndefinedValue
	}
	onRejected := CreateBuiltinFunction(agent, rejectedClosure, 1, CMString("onRejected"), builtinFunctionArgs{})
	PerformPromiseThen(agent, promise, onFulfilled.ToValue(), onRejected.ToValue(), nil)
	return
}

// AsyncGeneratorCompleteStep
// spec: 27.6.3.5
// returns UNUSED
func AsyncGeneratorCompleteStep(
	agent *Agent,
	generator *AsyncGeneratorObject,
	completion CompletionValue,
	done bool,
) {
	Assert(!generator.AsyncGeneratorQueue.IsEmpty())
	next := generator.AsyncGeneratorQueue.Dequeue()
	promiseCapability := next.Capability
	value := completion.value
	if completion.t == CompletionTypeThrow {
		ReturnAssertNormal(
			promiseCapability.Reject.Call(UndefinedValue, []Value{value}),
		)
	} else {
		// TODO: check realm is provided from parameters?
		iteratorResult := CreateIterResultObject(agent, value, done)
		ReturnAssertNormal(
			promiseCapability.Resolve.Call(UndefinedValue, []Value{iteratorResult.ToValue()}),
		)
	}
}

// AsyncGeneratorDrainQueue
// spec: 27.6.3.10
func AsyncGeneratorDrainQueue(
	agent *Agent,
	generator *AsyncGeneratorObject,
) {
	Assert(generator.AsyncGeneratorState == AsyncGeneratorStateCompleted)
	queue := generator.AsyncGeneratorQueue
	if queue.IsEmpty() {
		return
	}
	done := false
	for !done {
		next := queue.Dequeue()
		completion := next.Completion
		if completion.t == CompletionTypeReturn {
			generator.AsyncGeneratorState = AsyncGeneratorStateAwaitingReturn
			ReturnAssertNormal(
				AsyncGeneratorAwaitReturn(agent, generator),
			)
			done = true
		} else {
			if completion.t == CompletionTypeNormal {
				var completion CompletionValue
				completion.value = UndefinedValue
				AsyncGeneratorCompleteStep(agent, generator, completion, true)
				if queue.IsEmpty() {
					done = true
				}
			}
		}
	}
}

// AsyncGeneratorStart
// spec: 27.6.3.2
// returns UNUSED
func AsyncGeneratorStart(
	agent *Agent,
	generator *AsyncGeneratorObject,
	generatorBody *FunctionBody,
) {
	Assert(generator.AsyncGeneratorState == AsyncGeneratorStateUndefined)
	genContext := agent.RunningExecutionContext()
	agent.ExecutionContextMap[generator.GetId()] = genContext
	genVM := genContext.VM
	genContext.AsyncGenerator = generator
	genContext.yieldCh = make(chan struct{})

	closure := func() {
		if genContext.isSuspended {
			fmt.Println("closure suspended")
			genContext.yieldCh <- struct{}{}
			return
		}
		fmt.Println("closure first")
		result := generatorBody.Evaluation(genVM)
		fmt.Println("completed")
		// TODO: Assert generator status
		agent.ExecutionContextStack.Pop()
		generator.AsyncGeneratorState = AsyncGeneratorStateCompleted

		if result.t == CompletionTypeNormal {
			result.value = UndefinedValue
		} else if result.t == CompletionTypeReturn {
			result.t = CompletionTypeNormal
		}
		AsyncGeneratorCompleteStep(agent, generator, result, true)
		AsyncGeneratorDrainQueue(agent, generator)
		go genContext.Resume()
	}

	generator.closure = closure
	generator.AsyncGeneratorState = AsyncGeneratorStateSuspendedStart
}

// AsyncGeneratorYield
// spec: 27.6.3.8
func AsyncGeneratorYield(agent *Agent, value Value) (co CompletionValue) {
	fmt.Println("yield")
	genContext := agent.RunningExecutionContext()
	Assert(genContext.AsyncGenerator != nil)
	generator := genContext.AsyncGenerator
	Assert(GetGeneratorKind(agent) == GeneratorKindAsync)
	var completion CompletionValue
	completion.value = value

	// TODO: use previousRealm
	AsyncGeneratorCompleteStep(agent, generator, completion, false)
	queue := generator.AsyncGeneratorQueue
	if !queue.IsEmpty() {
		toYield := queue.Data()[0]
		resumptionValue := toYield.Completion
		return AsyncGeneratorUnwrapYieldResumption(agent, resumptionValue)
	} else {
		generator.AsyncGeneratorState = AsyncGeneratorStateSuspendedYield
		fmt.Println("yield queue empty")
		agent.ExecutionContextStack.Pop()
		genContext.isSuspended = true
		go genContext.Resume()
		<-genContext.yieldCh
		return
	}
}

// AsyncGeneratorUnwrapYieldResumption
// spec: 27.6.3.7
func AsyncGeneratorUnwrapYieldResumption(agent *Agent, resumptionValue CompletionValue) (co CompletionValue) {
	if resumptionValue.t != CompletionTypeReturn {
		return resumptionValue
	}
	awaited := Await(agent, resumptionValue.value)
	if awaited.t == CompletionTypeThrow {
		return awaited
	}
	awaited.t = CompletionTypeReturn
	return awaited
}
