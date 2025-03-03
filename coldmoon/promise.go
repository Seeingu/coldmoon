package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

// MARK: - PromiseState

type PromiseState int

const (
	PromiseStatePending PromiseState = iota
	PromiseStateFulfilled
	PromiseStateRejected
)

// 27.2.1.1
type PromiseCapability struct {
	Promise *PromiseObject
	Resolve ObjectType
	Reject  ObjectType
}

func (p *PromiseCapability) ToImportedModulePayload() ImportedModulePayload {
	return ImportedModulePayload{
		PromiseCapability: p,
	}
}

// 27.2.1.5
func NewPromiseCapability(agent *Agent, constructor Value) *PromiseCapability {
	if !IsConstructor(constructor) {
		agent.ThrowTypeError("is not a constructor")
		return nil
	}
	var executorClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return executorClosure(agent, thisArgument, argumentsList, newTarget)
	}
	resolvingFunctions := &ResolvingFunctions{
		Resolve: UndefinedValue,
		Reject:  UndefinedValue,
	}
	executor := CreateBuiltinFunction(agent, executorClosure, 2, CMString(""), builtinFunctionArgs{
		additionalFieldsV2: resolvingFunctions,
	})
	promise := MustGetObject(constructor).Construct([]Value{executor.ToValue()}, nil)
	if !IsCallable(resolvingFunctions.Resolve) {
		agent.ThrowTypeError("TypeError")
		return nil
	}
	if !IsCallable(resolvingFunctions.Reject) {
		agent.ThrowTypeError("TypeError")
		return nil
	}
	return &PromiseCapability{
		Promise: promise.value.(*PromiseObject),
		Resolve: MustGetObject(resolvingFunctions.Resolve),
		Reject:  MustGetObject(resolvingFunctions.Reject),
	}
}

func executorClosure(agent *Agent, thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
	resolve := argumentsList[0]
	reject := argumentsList[1]
	function := agent.ActiveFunctionObject()
	resolvingFunctions := function.(*BuiltinFunction).AdditionalFieldsV2.(*ResolvingFunctions)
	if !IsUndefinedOrNil(resolvingFunctions.Resolve) {
		return agent.ThrowTypeError("already called")
	}
	if !IsUndefinedOrNil(resolvingFunctions.Reject) {
		return agent.ThrowTypeError("already called")
	}
	resolvingFunctions.Resolve = resolve
	resolvingFunctions.Reject = reject
	return UndefinedValue
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
	HostDefined *HostDefined
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
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "PromisePrototype")

	// 27.2.5.4
	var then BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return promiseThen(agent, this, arguments, newTarget)
	}
	var catch BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		onRejected := arguments[0]
		promise := this
		return ValueInvoke(agent, promise, NewStringPropertyKey("then"), []Value{UndefinedValue, onRejected})
	}
	var finally BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return promiseFinally(agent, this, arguments, newTarget)
	}
	object.defineBuiltinFunction(realm, CMString("then"), then, 2)
	object.defineBuiltinFunction(realm, CMString("catch"), catch, 1)
	object.defineBuiltinFunction(realm, CMString("finally"), finally, 1)

	object.defineToStringTag("Promise")
	return object
}

// 27.2.5.4
func promiseThen(agent *Agent, this Value, arguments []Value, newTarget ObjectType) Value {
	realm := agent.CurrentRealm()
	onFulfilled := arguments[0]
	onRejected := pkg.SliceSafeGet(arguments, 1)
	if onRejected == nil {
		onRejected = UndefinedValue
	}
	promise := this
	if !promise.IsPromise() {
		return agent.ThrowTypeError("Promise.prototype.then called on incompatible receiver")
	}
	promiseObject := MustGetObject(promise)
	C := promiseObject.SpeciesConstructor(realm.Intrinsics.Promise)
	resultCapability := NewPromiseCapability(agent, C.Data().ToValue())
	return PerformPromiseThen(agent, promiseObject, onFulfilled, onRejected, resultCapability)
}

// spec: 27.2.5.3
func promiseFinally(agent *Agent, this Value, arguments []Value, newTarget ObjectType) (co CompletionValue) {
	promise := this
	realm := agent.CurrentRealm()
	onFinally := arguments[0]
	if !promise.IsObject() {
		return co.ThrowTypeError(agent, "Promise.prototype.finally called on incompatible receiver")
	}
	C := MustGetObject(promise).SpeciesConstructor(realm.Intrinsics.Promise)
	Assert(IsConstructor(C.Data().ToValue()))
	var thenFinally Value
	var catchFinally Value
	if !IsCallable(onFinally) {
		thenFinally = onFinally
		catchFinally = onFinally
	} else {
		thenFinally = (realm.Intrinsics.FunctionPrototype).ToValue()
		catchFinally = (realm.Intrinsics.FunctionPrototype).ToValue()
		captures := &PromiseThenFinallyCaptures{
			OnFinally:   onFinally,
			Constructor: C.Data(),
		}
		thenFinallyClosure := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
			return thenFinallyClosure(agent, this, arguments, newTarget)
		}

		thenFinally = CreateBuiltinFunction(agent, thenFinallyClosure, 1, CMString(""), builtinFunctionArgs{
			additionalFieldsV2: captures,
		}).ToValue()

		catchFinallyClosure := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
			return catchFinallyClosure(agent, this, arguments, newTarget)
		}
		catchFinally = CreateBuiltinFunction(agent, catchFinallyClosure, 1, CMString(""), builtinFunctionArgs{
			additionalFieldsV2: captures,
		}).ToValue()
	}
	return ValueInvoke(agent, promise, NewStringPropertyKey("then"), []Value{thenFinally, catchFinally}).ToCompletion()
}

func thenFinallyClosure(agent *Agent, this Value, arguments []Value, newTarget ObjectType) CompletionValue {
	function := agent.ActiveFunctionObject()
	captures := function.(*BuiltinFunction).AdditionalFieldsV2.(*PromiseThenFinallyCaptures)
	onFinally := captures.OnFinally
	result := ReturnAssertNormal(onFinally.CallNoArgs(UndefinedValue))
	p := PromiseResolve(agent, captures.Constructor, result)

	var returnValue BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		f := agent.ActiveFunctionObject()
		return f.(*BuiltinFunction).AdditionalFieldsV2.(Value)
	}

	valueThunk := CreateBuiltinFunction(agent, returnValue, 0, CMString(""), builtinFunctionArgs{
		additionalFieldsV2: arguments[0],
	})
	return ValueInvoke(agent, p.ToValue(), NewStringPropertyKey("then"), []Value{(valueThunk).ToValue()})
}

func catchFinallyClosure(agent *Agent, this Value, arguments []Value, newTarget ObjectType) CompletionValue {
	function := agent.ActiveFunctionObject()
	captures := function.(*BuiltinFunction).AdditionalFieldsV2.(*PromiseThenFinallyCaptures)
	onFinally := captures.OnFinally
	result := ReturnAssertNormal(onFinally.CallNoArgs(UndefinedValue))
	p := PromiseResolve(agent, captures.Constructor, result)
	reason := arguments[0]

	var throwReason BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		f := agent.ActiveFunctionObject()
		_reason := f.(*BuiltinFunction).AdditionalFieldsV2.(Value)
		agent.exception = _reason
		return _reason
	}

	thrower := CreateBuiltinFunction(agent, throwReason, 0, CMString(""), builtinFunctionArgs{
		additionalFieldsV2: reason,
	})
	return ValueInvoke(agent, p.ToValue(), NewStringPropertyKey("then"), []Value{(thrower).ToValue()})
}

func NewPromiseConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		executor := arguments[0]
		if newTarget == nil {
			return agent.ThrowTypeError("TypeError")
		}
		if !IsCallable(executor) {
			return agent.ThrowTypeError("TypeError")
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%Promise.prototype", nil)
		promise := &PromiseObject{
			Object:           o,
			PromiseState:     PromiseStatePending,
			PromiseResult:    UndefinedValue,
			PromiseIsHandled: false,
		}
		promise.ref = promise

		resolvingFunctions := CreateResolvingFunctions(agent, promise)
		executor.Call(agent, UndefinedValue, []Value{resolvingFunctions.Resolve, resolvingFunctions.Reject})
		return promise.ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, CMString("Promise"), builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	var reject BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		reason := arguments[0]
		C := this
		capability := NewPromiseCapability(agent, C)
		capability.Reject.Call(UndefinedValue, []Value{reason})
		return capability.Promise.ToValue()
	}
	var resolve BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		resolution := arguments[0]
		C := this
		if !C.IsObject() {
			return agent.ThrowTypeError("TypeError")
		}
		return PromiseResolve(agent, MustGetObject(C), resolution).ToValue()
	}
	race := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		iterable := arguments[0]
		C := this
		promiseCapability := NewPromiseCapability(agent, C)
		promiseResolveCompletion := GetPromiseResolve(agent, MustGetObject(C))
		if !IfAbruptRejectPromise(agent, promiseResolveCompletion, promiseCapability) {
			return UndefinedValue
		}
		promiseResolve := promiseResolveCompletion.Data()

		iteratorRecordCompletion := GetIterator(agent, iterable, IteratorKindSync)
		if !IfAbruptRejectPromise(agent, iteratorRecordCompletion, promiseCapability) {
			return UndefinedValue
		}
		iteratorRecord := iteratorRecordCompletion.Data()

		result := PerformPromiseRace(agent, iteratorRecord, MustGetObject(C), promiseCapability, promiseResolve)
		if result.IsAbrupt() {
			if !iteratorRecord.Done {
				iteratorRecord.IteratorClose()
			}
			if !IfAbruptRejectPromise(agent, result, promiseCapability) {
				return UndefinedValue
			}
		}
		return result.Data()
	}
	all := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		iterable := arguments[0]
		C := this
		promiseCapability := NewPromiseCapability(agent, C)
		promiseResolveCompletion := GetPromiseResolve(agent, MustGetObject(C))
		if !IfAbruptRejectPromise(agent, promiseResolveCompletion, promiseCapability) {
			return UndefinedValue
		}
		promiseResolve := promiseResolveCompletion.Data()

		iteratorRecordCompletion := GetIterator(agent, iterable, IteratorKindSync)
		if !IfAbruptRejectPromise(agent, iteratorRecordCompletion, promiseCapability) {
			return UndefinedValue
		}
		iteratorRecord := iteratorRecordCompletion.Data()

		result := PerformPromiseAll(agent, iteratorRecord, MustGetObject(C), promiseCapability, promiseResolve)
		if result.IsAbrupt() {
			if !iteratorRecord.Done {
				iteratorRecord.IteratorClose()
			}
			if !IfAbruptRejectPromise(agent, result, promiseCapability) {
				return UndefinedValue
			}
		}
		return result.Data()
	}
	allSettled := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		iterable := arguments[0]
		C := this
		promiseCapability := NewPromiseCapability(agent, C)
		promiseResolveCompletion := GetPromiseResolve(agent, MustGetObject(C))
		if !IfAbruptRejectPromise(agent, promiseResolveCompletion, promiseCapability) {
			return UndefinedValue
		}
		promiseResolve := promiseResolveCompletion.Data()

		iteratorRecordCompletion := GetIterator(agent, iterable, IteratorKindSync)
		if !IfAbruptRejectPromise(agent, iteratorRecordCompletion, promiseCapability) {
			return UndefinedValue
		}
		iteratorRecord := iteratorRecordCompletion.Data()

		result := PerformPromiseAllSettled(agent, iteratorRecord, MustGetObject(C), promiseCapability, promiseResolve)
		if result.IsAbrupt() {
			if !iteratorRecord.Done {
				iteratorRecord.IteratorClose()
			}
			if !IfAbruptRejectPromise(agent, result, promiseCapability) {
				return UndefinedValue
			}
		}
		return result.Data()
	}
	promiseAny := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		iterable := arguments[0]
		C := this
		promiseCapability := NewPromiseCapability(agent, C)
		promiseResolveCompletion := GetPromiseResolve(agent, MustGetObject(C))
		if !IfAbruptRejectPromise(agent, promiseResolveCompletion, promiseCapability) {
			return UndefinedValue
		}
		promiseResolve := promiseResolveCompletion.Data()

		iteratorRecordCompletion := GetIterator(agent, iterable, IteratorKindSync)
		if !IfAbruptRejectPromise(agent, iteratorRecordCompletion, promiseCapability) {
			return UndefinedValue
		}
		iteratorRecord := iteratorRecordCompletion.Data()

		result := PerformPromiseAny(agent, iteratorRecord, MustGetObject(C), promiseCapability, promiseResolve)
		if result.IsAbrupt() {
			if !iteratorRecord.Done {
				iteratorRecord.IteratorClose()
			}
			if !IfAbruptRejectPromise(agent, result, promiseCapability) {
				return UndefinedValue
			}
		}
		return result.Data()
	}
	object.defineBuiltinFunction(realm, CMString("reject"), reject, 1)
	object.defineBuiltinFunction(realm, CMString("resolve"), resolve, 1)
	object.defineBuiltinFunction(realm, CMString("race"), race, 1)
	object.defineBuiltinFunction(realm, CMString("all"), all, 1)
	object.defineBuiltinFunction(realm, CMString("allSettled"), allSettled, 1)
	object.defineBuiltinFunction(realm, CMString("any"), promiseAny, 1)

	object.defineBuiltinAccessor(realm, WellKnownSymbolsSpecies, builtinAccessorParams{
		Getter: func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
			return this
		},
	})

	BindPrototypeAndConstructor(realm.Intrinsics.PromisePrototype, object)
	return object
}

type ResolvingFunctions struct {
	Resolve Value
	Reject  Value
}

type AlreadyResolved struct {
	Value bool
}

type ArgGetterSetterCaptures struct {
	Name string
	Env  EnvironmentRecord
}

// TODO: maybe can be removed
type AdditionalFields struct {
	ClassConstructorFields *ClassConstructorFields
}

func IfAbruptRejectPromise[T any](agent *Agent, value Completion[T], capability *PromiseCapability) bool {
	if value.IsAbrupt() {
		capability.Reject.Call(UndefinedValue, []Value{value.Error()})
		return false
	}
	return true
}

// 27.2.1.3.1
func stepsReject(
	agent *Agent, this Value, arguments []Value, newTarget ObjectType,
) Value {
	F := agent.ActiveFunctionObject()
	additionalFields := F.(*BuiltinFunction).AdditionalFieldsV2.(*stepsRejectAdditionalFields)
	alreadyResolved := additionalFields.AlreadyResolved
	promise := additionalFields.Promise
	if alreadyResolved.Value {
		return UndefinedValue
	}
	alreadyResolved.Value = true
	promise.PromiseState = PromiseStateRejected
	promise.PromiseResult = arguments[0]
	return UndefinedValue
}

// 27.2.1.3.2
func stepsResolve(agent *Agent, this Value, arguments []Value, newTarget ObjectType) Value {
	resolution := arguments[0]
	F := agent.ActiveFunctionObject()
	additionalFields := F.(*BuiltinFunction).AdditionalFieldsV2.(*stepsResolveAdditionalFields)
	promise := additionalFields.Promise
	alreadyResolved := additionalFields.AlreadyResolved

	if alreadyResolved.Value {
		return UndefinedValue
	}
	alreadyResolved.Value = true
	if SameValue(resolution, promise.ToValue()) {
		agent.ThrowException(TypeError, "self resolution")
		RejectPromise(agent, promise, agent.exception)
		return UndefinedValue
	}
	if !resolution.IsObject() {
		FulfillPromise(agent, promise, resolution)
		return UndefinedValue
	}
	then := MustGetObject(resolution).Get(NewStringPropertyKey("then"))
	if !IsCallable(then) {
		FulfillPromise(agent, promise, resolution)
		return UndefinedValue
	}

	thenJobCallback := agent.HostHooks.HostMakeJobCallback(MustGetObject(then))
	job := NewPromiseResolveThenableJob(agent, promise, MustGetObject(resolution), thenJobCallback)
	agent.HostHooks.HostEnqueuePromiseJob(agent, job.Job, job.Realm)
	return UndefinedValue
}

type stepsResolveAdditionalFields struct {
	Promise         *PromiseObject
	AlreadyResolved *AlreadyResolved
}
type stepsRejectAdditionalFields struct {
	Promise         *PromiseObject
	AlreadyResolved *AlreadyResolved
}

// 27.2.1.3
func CreateResolvingFunctions(agent *Agent, promise *PromiseObject) *ResolvingFunctions {
	realm := agent.CurrentRealm()
	alreadyResolved := &AlreadyResolved{Value: false}

	lengthResolve := JSInt(1)
	resolveAdditionalFields := &stepsResolveAdditionalFields{
		Promise:         promise,
		AlreadyResolved: alreadyResolved,
	}

	resolve := CreateBuiltinFunction(agent, func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return stepsResolve(agent, thisArgument, argumentsList, newTarget)
	}, lengthResolve, CMString(""), builtinFunctionArgs{
		realm:              realm,
		additionalFieldsV2: resolveAdditionalFields,
	})

	var stepsReject BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return stepsReject(agent, this, arguments, newTarget)
	}
	lengthReject := JSInt(1)
	rejectAdditionalFields := &stepsRejectAdditionalFields{
		Promise:         promise,
		AlreadyResolved: alreadyResolved,
	}
	reject := CreateBuiltinFunction(agent, stepsReject, lengthReject, CMString(""), builtinFunctionArgs{
		realm:              realm,
		additionalFieldsV2: rejectAdditionalFields,
	})
	return &ResolvingFunctions{
		Resolve: resolve.ToValue(),
		Reject:  reject.ToValue(),
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
	if !promise.PromiseIsHandled {
		agent.HostHooks.HostPromiseRejectionTracker(promise, PromiseRejectionTrackerOperationReject)
	}
	TriggerPromiseReactions(agent, reactions, reason)
}

// 27.2.1.8
func TriggerPromiseReactions(agent *Agent, reactions []*PromiseReaction, argument Value) {
	for _, reaction := range reactions {
		job := NewPromiseReactionJob(agent, reaction, argument)
		agent.HostHooks.HostEnqueuePromiseJob(agent, job.Job, job.Realm)
	}
}

type PromiseJobReactionCaptures struct {
	Agent    *Agent
	Reaction *PromiseReaction
	Argument Value
}
type PromiseJobThenableReactionCaptures struct {
	Agent            *Agent
	PromiseToResolve *PromiseObject
	Thenable         ObjectType
	ThenJobCallback  *JobCallback
}
type PromiseThenFinallyCaptures struct {
	OnFinally   Value
	Constructor ObjectType
}

type Job struct {
	Fun      func(captures any) Value
	Captures any
}

type PromiseReactionJob struct {
	Realm *Realm
	Job   *Job
}

// 27.2.2.1
func NewPromiseReactionJob(agent *Agent, reaction *PromiseReaction, argument Value) *PromiseReactionJob {
	captures := &PromiseJobReactionCaptures{
		Agent:    agent,
		Reaction: reaction,
		Argument: argument,
	}

	// TODO(BM): use outer env directly
	fun := func(_captures any) Value {
		captures := _captures.(*PromiseJobReactionCaptures)
		agent := captures.Agent
		reaction := captures.Reaction
		argument := captures.Argument
		promiseCapability := reaction.Capability
		t := reaction.Type
		handler := reaction.Handler
		var handlerResult CompletionValue
		if handler == nil {
			if t == PromiseReactionTypeFulfill {
				handlerResult.value = argument
			} else {
				handlerResult.err = agent.exception
			}
		} else {
			handlerResult.value = agent.HostHooks.HostCallJobCallback(handler, UndefinedValue, []Value{argument}).value
		}
		if promiseCapability == nil {
			return UndefinedValue
		}
		if handlerResult.IsError() {
			reason := handlerResult.Error()
			return promiseCapability.Reject.Call(
				UndefinedValue, []Value{reason},
			).value
		} else {
			value := handlerResult.Data()
			return promiseCapability.Resolve.Call(
				UndefinedValue, []Value{value},
			).value
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
	if x.IsPromise() {
		xConstructor := MustGetObject(x).Get(NewStringPropertyKey("constructor"))
		if SameValue(xConstructor, (constructor).ToValue()) {
			return MustGetObject(x)
		}
	}

	promiseCapability := NewPromiseCapability(agent, constructor.ToValue())
	promiseCapability.Resolve.Call(UndefinedValue, []Value{x})
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
		return resultCapability.Promise.ToValue()
	}
}

type promiseAdditionalFields struct {
	// [[AlreadyCalled]]
	alreadyCalled bool
	index         uint64
	// [[Values]] or [[Errors]] in Promise.any
	values            []Value
	capability        *PromiseCapability
	RemainingElements *RemainingElements
}

func PerformPromiseAll(
	agent *Agent,
	iterator *IteratorRecord,
	constructor ObjectType,
	resultCapability *PromiseCapability,
	promiseResolve ObjectType,
) (co Completion[Value]) {
	values := []Value{}
	remainingElements := &RemainingElements{Value: 1}
	var index int = 0
	for {
		next := iterator.IteratorStep()
		if next == nil {
			iterator.Done = true
			remainingElements.Value--
			if remainingElements.Value == 0 {
				valuesArray := CreateArrayFromList(agent, values)
				resultCapability.Resolve.Call(UndefinedValue, []Value{valuesArray.ToValue()})
			}
			co.value = resultCapability.Promise.ToValue()
			return
		}

		nextValue := IteratorValue(next)
		values = append(values, nextValue)
		nextPromise := promiseResolve.Call(constructor.ToValue(), []Value{nextValue})

		steps := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
			F := agent.ActiveFunctionObject()
			additionalFields := F.(*BuiltinFunction).AdditionalFieldsV2.(*promiseAdditionalFields)
			if additionalFields.alreadyCalled {
				return UndefinedValue
			}
			additionalFields.alreadyCalled = true
			_index := additionalFields.index
			_values := additionalFields.values

			_values[_index] = arguments[0]
			remainingElements.Value--
			if remainingElements.Value == 0 {
				valuesArray := CreateArrayFromList(agent, _values)
				return resultCapability.Resolve.Call(UndefinedValue, []Value{valuesArray.ToValue()}).value
			}
			return UndefinedValue
		}
		length := JSInt(1)
		onFulfilled := CreateBuiltinFunction(agent, steps, length, CMString(""), builtinFunctionArgs{
			additionalFieldsV2: &promiseAdditionalFields{
				alreadyCalled:     false,
				index:             uint64(index),
				values:            values,
				capability:        resultCapability,
				RemainingElements: remainingElements,
			},
		})

		remainingElements.Value++
		ValueInvoke(agent, nextPromise.value, NewStringPropertyKey("then"), []Value{onFulfilled.ToValue(), resultCapability.Reject.ToValue()})
		index++
	}
}

func PerformPromiseAllSettled(
	agent *Agent,
	iterator *IteratorRecord,
	constructor ObjectType,
	resultCapability *PromiseCapability,
	promiseResolve ObjectType,
) (co Completion[Value]) {
	var values []Value
	remainingElements := &RemainingElements{Value: 1}
	var index int = 0
	for {
		next := iterator.IteratorStep()
		if next == nil {
			iterator.Done = true
			remainingElements.Value--
			if remainingElements.Value == 0 {
				valuesArray := CreateArrayFromList(agent, values)
				resultCapability.Resolve.Call(UndefinedValue, []Value{valuesArray.ToValue()})
			}
			co.value = resultCapability.Promise.ToValue()
			return
		}

		nextValue := IteratorValue(next)
		values = append(values, nextValue)
		nextPromise := promiseResolve.Call(constructor.ToValue(), []Value{nextValue})

		stepsFulfilled := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
			F := agent.ActiveFunctionObject()
			additionalFields := F.(*BuiltinFunction).AdditionalFieldsV2.(*promiseAdditionalFields)
			if additionalFields.alreadyCalled {
				return UndefinedValue
			}
			additionalFields.alreadyCalled = true
			_index := additionalFields.index
			_values := additionalFields.values
			obj := OrdinaryObjectCreate(agent, agent.CurrentRealm().Intrinsics.ObjectPrototype, nil)
			obj.CreateDataPropertyOrThrow(NewStringPropertyKey("status"), NewStringValue("fulfilled"))
			obj.CreateDataPropertyOrThrow(NewStringPropertyKey("value"), arguments[0])
			_values[_index] = obj.ToValue()
			remainingElements.Value--
			if remainingElements.Value == 0 {
				valuesArray := CreateArrayFromList(agent, _values)
				return resultCapability.Resolve.Call(UndefinedValue, []Value{valuesArray.ToValue()}).value
			}
			return UndefinedValue
		}
		lengthFulfilled := JSInt(1)
		onFulfilled := CreateBuiltinFunction(agent, stepsFulfilled, lengthFulfilled, CMString(""), builtinFunctionArgs{
			additionalFieldsV2: &promiseAdditionalFields{
				alreadyCalled:     false,
				index:             uint64(index),
				values:            values,
				capability:        resultCapability,
				RemainingElements: remainingElements,
			},
		})

		stepsReject := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
			F := agent.ActiveFunctionObject()
			additionalFields := F.(*BuiltinFunction).AdditionalFieldsV2.(*promiseAdditionalFields)
			if additionalFields.alreadyCalled {
				return UndefinedValue
			}
			additionalFields.alreadyCalled = true
			_index := additionalFields.index
			_values := additionalFields.values
			obj := OrdinaryObjectCreate(agent, agent.CurrentRealm().Intrinsics.ObjectPrototype, nil)
			obj.CreateDataPropertyOrThrow(NewStringPropertyKey("status"), NewStringValue("rejected"))
			obj.CreateDataPropertyOrThrow(NewStringPropertyKey("reason"), arguments[0])
			_values[_index] = obj.ToValue()
			remainingElements.Value--
			if remainingElements.Value == 0 {
				valuesArray := CreateArrayFromList(agent, _values)
				return resultCapability.Resolve.Call(UndefinedValue, []Value{valuesArray.ToValue()}).value
			}
			return UndefinedValue
		}
		lengthRejected := JSInt(1)
		onRejected := CreateBuiltinFunction(agent, stepsReject, lengthRejected, CMString(""), builtinFunctionArgs{
			additionalFieldsV2: &promiseAdditionalFields{
				alreadyCalled:     false,
				index:             uint64(index),
				values:            values,
				capability:        resultCapability,
				RemainingElements: remainingElements,
			},
		})

		remainingElements.Value++
		ValueInvoke(agent, nextPromise.value, NewStringPropertyKey("then"), []Value{onFulfilled.ToValue(), onRejected.ToValue()})
		index++
	}
}

func PerformPromiseAny(
	agent *Agent,
	iterator *IteratorRecord,
	constructor ObjectType,
	resultCapability *PromiseCapability,
	promiseResolve ObjectType,
) (co Completion[Value]) {
	var errors []Value
	remainingElements := &RemainingElements{Value: 1}
	var index int = 0
	for {
		next := iterator.IteratorStep()
		if next == nil {
			iterator.Done = true
			remainingElements.Value--
			if remainingElements.Value == 0 {
				err := MustGetObject(agent.ThrowException(AggregateError, "No promises in Promise.any were resolved"))
				err.DefinePropertyOrThrow(NewStringPropertyKey("errors"), &PropertyDescriptor{
					Value:        CreateArrayFromList(agent, errors).ToValue(),
					Writable:     true,
					Enumerable:   false,
					Configurable: true,
				})
				co.err = err.ToValue()
				return
			}
			co.value = resultCapability.Promise.ToValue()
			return
		}

		nextValue := IteratorValue(next)
		errors = append(errors, UndefinedValue)
		nextPromise := promiseResolve.Call(constructor.ToValue(), []Value{nextValue})

		stepsRejected := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
			F := agent.ActiveFunctionObject()
			additionalFields := F.(*BuiltinFunction).AdditionalFieldsV2.(*promiseAdditionalFields)
			if additionalFields.alreadyCalled {
				return UndefinedValue
			}
			additionalFields.alreadyCalled = true
			_index := additionalFields.index
			_errors := additionalFields.values
			_errors[_index] = arguments[0]
			remainingElements.Value--
			if remainingElements.Value == 0 {
				err := MustGetObject(agent.ThrowException(AggregateError, "All promises in Promise.any were rejected"))
				err.DefinePropertyOrThrow(NewStringPropertyKey("errors"), &PropertyDescriptor{
					Value:        CreateArrayFromList(agent, _errors).ToValue(),
					Writable:     true,
					Enumerable:   false,
					Configurable: true,
				})
				return resultCapability.Reject.Call(UndefinedValue, []Value{err.ToValue()}).value
			}
			return UndefinedValue
		}
		length := JSInt(1)
		onFulfilled := CreateBuiltinFunction(agent, stepsRejected, length, CMString(""), builtinFunctionArgs{
			additionalFieldsV2: &promiseAdditionalFields{
				alreadyCalled:     false,
				index:             uint64(index),
				values:            errors,
				capability:        resultCapability,
				RemainingElements: remainingElements,
			},
		})

		remainingElements.Value++
		ValueInvoke(agent, nextPromise.value, NewStringPropertyKey("then"), []Value{onFulfilled.ToValue(), resultCapability.Reject.ToValue()})
		index++
	}
}

// 27.2.2.2
func NewPromiseResolveThenableJob(agent *Agent, promise *PromiseObject, thenable ObjectType, thenJobCallback *JobCallback) *PromiseReactionJob {
	captures := &PromiseJobThenableReactionCaptures{
		Agent:            agent,
		PromiseToResolve: promise,
		Thenable:         thenable,
		ThenJobCallback:  thenJobCallback,
	}
	fun := func(_captures any) Value {
		captures := _captures.(*PromiseJobThenableReactionCaptures)
		agent := captures.Agent
		promiseToResolve := captures.PromiseToResolve
		thenable := captures.Thenable
		thenJobCallback := captures.ThenJobCallback

		resolvingFunctions := CreateResolvingFunctions(agent, promiseToResolve)
		thenCallResult := agent.HostHooks.HostCallJobCallback(thenJobCallback, thenable.ToValue(), []Value{resolvingFunctions.Resolve, resolvingFunctions.Reject})
		return thenCallResult.value
	}
	job := &Job{
		Fun:      fun,
		Captures: captures,
	}
	getThenRealmResult := thenJobCallback.Callback.GetFunctionRealm()
	var thenRealm *Realm
	if getThenRealmResult != nil {
		thenRealm = getThenRealmResult
	} else {
		thenRealm = agent.CurrentRealm()
	}

	return &PromiseReactionJob{
		Realm: thenRealm,
		Job:   job,
	}
}

func PerformPromiseRace(
	agent *Agent,
	iterator *IteratorRecord,
	constructor ObjectType,
	resultCapability *PromiseCapability,
	promiseResolve ObjectType,
) (co Completion[Value]) {
	for {
		next := iterator.IteratorStep()
		if next == nil {
			iterator.Done = true
			co.value = resultCapability.Promise.ToValue()
			return
		}

		nextValue := IteratorValue(next)
		nextPromise := promiseResolve.Call(constructor.ToValue(), []Value{nextValue})
		ValueInvoke(agent, nextPromise.value, NewStringPropertyKey("then"), []Value{resultCapability.Resolve.ToValue(), resultCapability.Reject.ToValue()})
	}
}

// 27.2.4.1.1
func GetPromiseResolve(agent *Agent, promiseConstructor ObjectType) (co Completion[ObjectType]) {
	promiseResolve := promiseConstructor.Get(NewStringPropertyKey("resolve"))
	if !IsCallable(promiseResolve) {
		co.err = agent.ThrowException(TypeError, "Promise.resolve is not callable")
		return
	}
	co.value = MustGetObject(promiseResolve)
	return
}
