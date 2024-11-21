package coldmoon

// MARK: - PromiseState

type PromiseState int

const (
	PromiseStatePending PromiseState = iota
	PromiseStateFulfilled
	PromiseStateRejected
)

// 27.2.1.1
type PromiseCapability struct {
	Promise ObjectType
	Resolve ObjectType
	Reject  ObjectType
}

func NewPromiseCapability(agent *Agent, constructor Value) *PromiseCapability {
	if !IsConstructor(constructor) {
		panic("TypeError")
	}
	var executorClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		resolve := argumentsList[0]
		reject := argumentsList[1]
		function := agent.ActiveFunctionObject()
		additionalFields := function.(*BuiltinFunction).AdditionalFields
		resolvingFunctions := additionalFields.ResolvingFunctions
		if resolvingFunctions.Resolve != nil {
			panic("TypeError")
		}
		if resolvingFunctions.Reject != nil {
			panic("TypeError")
		}
		resolvingFunctions.Resolve = resolve
		resolvingFunctions.Reject = reject
		return UndefinedValue
	}
	additionalFields := &AdditionalFields{
		ResolvingFunctions: &ResolvingFunctions{
			Resolve: UndefinedValue,
			Reject:  UndefinedValue,
		},
	}
	executor := CreateBuiltinFunction(agent, executorClosure, 2, "", builtinFunctionArgs{
		additionalFields: additionalFields,
	})
	promise := MustGetObject(constructor).Construct([]Value{NewValueFromObject(executor)}, nil)
	if !IsCallable(additionalFields.ResolvingFunctions.Resolve) {
		panic("TypeError")
	}
	if !IsCallable(additionalFields.ResolvingFunctions.Reject) {
		panic("TypeError")
	}
	return &PromiseCapability{
		Promise: promise,
		Resolve: MustGetObject(additionalFields.ResolvingFunctions.Resolve),
		Reject:  MustGetObject(additionalFields.ResolvingFunctions.Reject),
	}

}

// MARK: - PromiseReaction

type PromiseReactionType int

const (
	PromiseReactionTypeFulfill PromiseReactionType = iota
	PromiseReactionTypeReject
)

type PromiseReaction struct {
	Capability *PromiseCapability
	Type       PromiseReactionType
	Handler    *JobCallback
}

// MARK: - Job

type JobCallback struct {
	Callback    ObjectType
	HostDefined interface{}
}

type RemainingElements struct {
	Value int
}

// MARK: - PromiseObject

type PromiseObject struct {
	*Object
	PromiseState            PromiseState
	PromiseResult           Value
	PromiseFulfillReactions []*PromiseReaction
	PromiseRejectReactions  []*PromiseReaction
	PromiseIsHandled        bool
}

func NewPromisePrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)

	var then BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		onFulfilled := arguments[0]
		onRejected := arguments[1]
		promise := this
		if !ValueIsPromise(promise) {
			panic("TypeError")
		}
		promiseObject := MustGetObject(promise)
		C := promiseObject.SpeciesConstructor(realm.Intrinsics.Promise)
		resultCapability := NewPromiseCapability(agent, C.Value)
		return PerformPromiseThen(agent, promiseObject, onFulfilled, onRejected, resultCapability)
	}
	var catch BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		onRejected := arguments[0]
		promise := this
		return ValueInvoke(agent, promise, NewStringPropertyKey("then"), []Value{UndefinedValue, onRejected})
	}
	DefineBuiltinFunction(object, "then", then, 2, realm)
	DefineBuiltinFunction(object, "catch", catch, 1, realm)

	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("Promise"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}

func NewPromiseConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		executor := arguments[0]
		if newTarget == nil {
			panic("TypeError")
		}
		if !IsCallable(executor) {
			panic("TypeError")
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%Promise.prototype", nil)
		promise := &PromiseObject{
			Object:           o,
			PromiseState:     PromiseStatePending,
			PromiseResult:    UndefinedValue,
			PromiseIsHandled: false,
		}

		resolvingFunctions := CreateResolvingFunctions(agent, promise)
		executor.CallAssumeCallable(UndefinedValue, []Value{resolvingFunctions.Resolve})
		return NewValueFromObject(promise)
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "Promise", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})

	var reject BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		reason := arguments[0]
		C := this
		capability := NewPromiseCapability(agent, C)
		ValueCall(NewValueFromObject(capability.Reject), UndefinedValue, []Value{reason})
		return NewValueFromObject(capability.Promise)
	}
	var resolve BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		resolution := arguments[0]
		C := this
		if !ValueIsObject(C) {
			panic("TypeError")
		}
		return NewValueFromObject(PromiseResolve(agent, MustGetObject(C), resolution))
	}
	DefineBuiltinFunction(object, "reject", reject, 1, realm)
	DefineBuiltinFunction(object, "resolve", resolve, 1, realm)

	var getter BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		return this
	}
	DefineBuiltinAccessor(realm, object, "@@species", getter, nil)

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.PromisePrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.PromisePrototype, "constructor", NewValueFromObject(object))

	return object
}

type ResolvingFunctions struct {
	Resolve Value
	Reject  Value
}

type AlreadyResolved struct {
	Value bool
}
type AdditionalFields struct {
	Promise            *PromiseObject
	AlreadyResolved    *AlreadyResolved
	ResolvingFunctions *ResolvingFunctions
}

// 27.2.1.3
func CreateResolvingFunctions(agent *Agent, promise *PromiseObject) *ResolvingFunctions {
	realm := agent.CurrentRealm()
	alreadyResolved := &AlreadyResolved{Value: false}

	var stepsResolve BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		resolution := arguments[0]
		F := agent.ActiveFunctionObject()
		additionalFields := F.(*BuiltinFunction).AdditionalFields
		_promise := additionalFields.Promise
		_alreadyResolved := additionalFields.AlreadyResolved

		if _alreadyResolved.Value {
			return UndefinedValue
		}
		_alreadyResolved.Value = true
		if SameValue(resolution, NewValueFromObject(_promise)) {
			agent.ThrowException(TypeError, "self resolution")
			RejectPromise(agent, _promise, agent.exception)
			return UndefinedValue
		}
		if !ValueIsObject(resolution) {
			FulfillPromise(agent, _promise, resolution)
			return UndefinedValue
		}
		then := MustGetObject(resolution).Get(NewStringPropertyKey("then"))
		if !IsCallable(then) {
			FulfillPromise(agent, _promise, resolution)
			return UndefinedValue
		}
		return UndefinedValue
	}
	lengthResolve := 1
	resolveAdditionalFields := &AdditionalFields{
		Promise:         promise,
		AlreadyResolved: alreadyResolved,
	}
	resolve := CreateBuiltinFunction(agent, stepsResolve, float64(lengthResolve), "", builtinFunctionArgs{
		realm:            realm,
		additionalFields: resolveAdditionalFields,
	})

	var stepsReject BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		if alreadyResolved.Value {
			return UndefinedValue
		}
		alreadyResolved.Value = true
		promise.PromiseState = PromiseStateRejected
		promise.PromiseResult = arguments[0]
		return UndefinedValue
	}
	lengthReject := 1
	rejectAdditionalFields := &AdditionalFields{
		Promise:         promise,
		AlreadyResolved: alreadyResolved,
	}
	reject := CreateBuiltinFunction(agent, stepsReject, float64(lengthReject), "", builtinFunctionArgs{
		realm:            realm,
		additionalFields: rejectAdditionalFields,
	})
	return &ResolvingFunctions{
		Resolve: NewValueFromObject(resolve),
		Reject:  NewValueFromObject(reject),
	}

}

// 27.2.1.4
func FulfillPromise(agent *Agent, promise *PromiseObject, value Value) {
	Assert(promise.PromiseState == PromiseStatePending)
	reactions := promise.PromiseFulfillReactions
	promise.PromiseResult = value
	promise.PromiseState = PromiseStateFulfilled
	TriggerPromiseReactions(agent, reactions, value)
}

// 27.2.1.7
func RejectPromise(agent *Agent, promise *PromiseObject, reason Value) {
	Assert(promise.PromiseState == PromiseStatePending)
	reactions := promise.PromiseRejectReactions
	promise.PromiseState = PromiseStateRejected
	promise.PromiseResult = reason
	TriggerPromiseReactions(agent, reactions, reason)
}

// 27.2.1.8
func TriggerPromiseReactions(agent *Agent, reactions []*PromiseReaction, argument Value) {
	for _, reaction := range reactions {
		job := NewPromiseReactionJob(agent, reaction, argument)
		agent.HostHooks.HostEnqueuePromiseJob(agent, job.Job, job.Realm)
	}
}

type JobCaptures struct {
	Agent    *Agent
	Reaction *PromiseReaction
	Argument Value
}
type Job struct {
	Fun      func(captures *JobCaptures) Value
	Captures *JobCaptures
}

type PromiseReactionJob struct {
	Realm *Realm
	Job   *Job
}

// 27.2.2.1
func NewPromiseReactionJob(agent *Agent, reaction *PromiseReaction, argument Value) *PromiseReactionJob {
	captures := &JobCaptures{
		Agent:    agent,
		Reaction: reaction,
		Argument: argument,
	}

	var fun = func(captures *JobCaptures) Value {
		agent := captures.Agent
		reaction := captures.Reaction
		argument := captures.Argument
		promiseCapability := reaction.Capability
		t := reaction.Type
		handler := reaction.Handler
		var handlerResult *CompletionRecord
		if handler == nil {
			if t == PromiseReactionTypeFulfill {
				handlerResult = NewNormalCompletion(argument)
			} else {
				handlerResult = NewThrowCompletion(agent.exception)
			}
		} else {
			handlerResult = NewNormalCompletion(agent.HostHooks.HostCallJobCallback(handler, UndefinedValue, []Value{argument}))
		}
		if promiseCapability == nil {
			return UndefinedValue
		}
		if handlerResult.Type != CompletionTypeNormal {
			reason := handlerResult.Value
			return promiseCapability.Reject.ToValue().CallAssumeCallable(
				UndefinedValue, []Value{reason},
			)
		} else {
			value := handlerResult.Value
			return promiseCapability.Resolve.ToValue().CallAssumeCallable(
				UndefinedValue, []Value{value},
			)
		}
	}
	job := &Job{
		Fun:      fun,
		Captures: captures,
	}
	var handlerRealm *Realm
	if reaction.Handler != nil {
		getHandlerRealmResult := reaction.Handler.Callback.GetFunctionRealm()
		if getHandlerRealmResult != nil {
			handlerRealm = getHandlerRealmResult
		} else {
			handlerRealm = agent.CurrentRealm()
		}
	}

	return &PromiseReactionJob{
		Realm: handlerRealm,
		Job:   job,
	}
}

// 27.2.4.7.1
func PromiseResolve(agent *Agent, constructor ObjectType, x Value) ObjectType {
	if ValueIsPromise(x) {
		xConstructor := MustGetObject(x).Get(NewStringPropertyKey("constructor"))
		if SameValue(xConstructor, NewValueFromObject(constructor)) {
			return MustGetObject(x)
		}
	}

	promiseCapability := NewPromiseCapability(agent, NewValueFromObject(constructor))
	NewValueFromObject(promiseCapability.Resolve).CallAssumeCallable(UndefinedValue, []Value{x})
	return promiseCapability.Promise
}

// 27.2.5.4.1
func PerformPromiseThen(agent *Agent, promise ObjectType, onFulfilled Value, onRejected Value, resultCapability *PromiseCapability) Value {
	var onFulfilledJobCallback *JobCallback
	if IsCallable(onFulfilled) {
		onFulfilledJobCallback = agent.HostHooks.HostMakeJobCallback(MustGetObject(onFulfilled))
	}
	var onRejectedJobCallback *JobCallback
	if IsCallable(onRejected) {
		onRejectedJobCallback = agent.HostHooks.HostMakeJobCallback(MustGetObject(onRejected))
	}

	fulfillReaction := &PromiseReaction{
		Capability: resultCapability,
		Type:       PromiseReactionTypeFulfill,
		Handler:    onFulfilledJobCallback,
	}
	rejectReaction := &PromiseReaction{
		Capability: resultCapability,
		Type:       PromiseReactionTypeReject,
		Handler:    onRejectedJobCallback,
	}
	p := promise.(*PromiseObject)
	switch p.PromiseState {
	case PromiseStatePending:
		p.PromiseFulfillReactions = append(p.PromiseFulfillReactions, fulfillReaction)
		p.PromiseRejectReactions = append(p.PromiseRejectReactions, rejectReaction)
	case PromiseStateFulfilled:
		value := p.PromiseResult
		fulfillJob := NewPromiseReactionJob(agent, fulfillReaction, value)
		agent.HostHooks.HostEnqueuePromiseJob(agent, fulfillJob.Job, fulfillJob.Realm)
	case PromiseStateRejected:
		reason := p.PromiseResult
		if !p.PromiseIsHandled {
			agent.HostHooks.HostPromiseRejectionTracker(p, PromiseRejectionTrackerOperationReject)
		}

		rejectJob := NewPromiseReactionJob(agent, rejectReaction, reason)
		agent.HostHooks.HostEnqueuePromiseJob(agent, rejectJob.Job, rejectJob.Realm)
	}

	p.PromiseIsHandled = true
	if resultCapability == nil {
		return UndefinedValue
	} else {
		return NewValueFromObject(resultCapability.Promise)
	}
}
