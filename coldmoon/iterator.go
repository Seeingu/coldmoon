package coldmoon

type IteratorRecord struct {
	Iterator   ObjectType
	NextMethod Value
	Done       bool
}

type IteratorKind int

const (
	IteratorKindSync IteratorKind = iota
	IteratorKindAsync
)

func NewIteratorPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "IteratorPrototype")
	var iterator BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) CompletionConvertable[Value] {
		return thisValue
	}

	DefineBuiltinFunction(realm, WellKnownSymbolsIterator, object, iterator, 0)
	return object
}

// GetIteratorFromMethodCompletion creates an iterator record without losing
// exceptions from the iterator method or the next-property lookup.
// spec: 7.4.2
func GetIteratorFromMethodCompletion(agent *Agent, object Value, method ObjectType) (co Completion[*IteratorRecord]) {
	iterator, isAbrupt, rt := ReturnIfAbrupt(method.Call(object, nil), co)
	if isAbrupt {
		return rt
	}
	if iterator == nil || !iterator.IsObject() {
		return co.ThrowTypeError(agent, "iterator method must return an object")
	}
	nextMethod, isAbrupt, rt := ReturnIfAbrupt(GetV(agent, iterator, NewStringPropertyKey("next")), co)
	if isAbrupt {
		return rt
	}
	co.value = &IteratorRecord{
		Iterator:   MustGetObject(iterator),
		NextMethod: nextMethod,
	}
	return
}

// GetIteratorFromMethod is the panic-style compatibility wrapper for callers
// that have not yet migrated to completion-aware iterator acquisition.
func GetIteratorFromMethod(agent *Agent, object Value, method ObjectType) *IteratorRecord {
	result := GetIteratorFromMethodCompletion(agent, object, method)
	if result.IsAbrupt() {
		if result.Error() != nil {
			panic(result.Error())
		}
		panic("GetIteratorFromMethod completed abruptly without an error value")
	}
	return result.Data()
}

// GetIterator
// spec: 7.4.3
func GetIterator(agent *Agent, obj Value, kind IteratorKind) (co Completion[*IteratorRecord]) {
	if IsUndefinedOrNull(obj) {
		return co.ThrowTypeError(agent, "value is not iterable")
	}
	var method ObjectType
	switch kind {
	case IteratorKindSync:
		var isAbrupt bool
		var rt Completion[*IteratorRecord]
		method, isAbrupt, rt = ReturnIfAbrupt(
			GetMethodCompletion(agent, obj, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator])),
			co,
		)
		if isAbrupt {
			return rt
		}
	case IteratorKindAsync:
		var isAbrupt bool
		var rt Completion[*IteratorRecord]
		method, isAbrupt, rt = ReturnIfAbrupt(
			GetMethodCompletion(agent, obj, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsAsyncIterator])),
			co,
		)
		if isAbrupt {
			return rt
		}
		if method == nil {
			syncMethod, isAbrupt, rt := ReturnIfAbrupt(
				GetMethodCompletion(agent, obj, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator])),
				co,
			)
			if isAbrupt {
				return rt
			}
			if syncMethod == nil {
				return co.ThrowTypeError(agent, "GetIterator: async no iterator method")
			}
			syncIteratorRecord, isAbrupt, rt := ReturnIfAbrupt(
				GetIteratorFromMethodCompletion(agent, obj, syncMethod),
				co,
			)
			if isAbrupt {
				return rt
			}
			co.value = CreateAsyncFromSyncIterator(syncIteratorRecord)
			return
		}
	}
	if method == nil {
		return co.ThrowTypeError(agent, "No iterator method")
	}
	return GetIteratorFromMethodCompletion(agent, obj, method)
}

// IteratorNext
// spec: 7.4.4
func (i *IteratorRecord) IteratorNext(value Value) (co Completion[ObjectType]) {
	var result CompletionValue
	if value == nil {
		result = i.NextMethod.Call(i.Iterator.Agent(), i.Iterator.ToValue(), nil)
	} else {
		result = i.NextMethod.Call(i.Iterator.Agent(), i.Iterator.ToValue(), []Value{value})
	}

	resultValue, isAbrupt, rt := ReturnIfAbrupt(result, co)
	if isAbrupt {
		i.Done = true
		return rt
	}
	resultObject, ok := resultValue.(*ObjectValue)
	if !ok {
		i.Done = true
		return co.ThrowTypeError(i.Iterator.Agent(), "IteratorNext method must return an object")
	}
	co.value = resultObject.Object
	return
}

// IteratorComplete
// spec: 7.4.5
func IteratorComplete(iterResult ObjectType) (co Completion[bool]) {
	done, isAbrupt, rt := ReturnIfAbrupt(
		iterResult.internalMethods().Get(
			iterResult,
			NewStringPropertyKey("done"),
			iterResult.ToValue(),
		),
		co,
	)
	if isAbrupt {
		return rt
	}
	co.value = done.ToBoolean()
	return
}

// IteratorValue returns the iterator result's value while preserving an
// abrupt completion produced by an accessor.
// spec: 7.4.6
func IteratorValue(iterResult ObjectType) CompletionValue {
	return iterResult.internalMethods().Get(
		iterResult,
		NewStringPropertyKey("value"),
		iterResult.ToValue(),
	)
}

// IteratorStep
// spec: 7.4.7
// returns Object, false or throw
func (i *IteratorRecord) IteratorStep() (co Completion[ObjectType], isFalse bool) {
	result, isAbrupt, rt := ReturnIfAbrupt(i.IteratorNext(nil), co)
	if isAbrupt {
		co = rt
		return
	}
	done, isAbrupt, rt := ReturnIfAbrupt(IteratorComplete(result), co)
	if isAbrupt {
		i.Done = true
		co = rt
		return
	}
	if done {
		i.Done = true
		isFalse = true
		return
	}
	co.value = result
	return
}

// IteratorStepValue
// spec: 7.4.8
// returns Value or DONE, abrupt
func (i *IteratorRecord) IteratorStepValue() (co CompletionValue, isDone bool) {
	resultCompletion, isDone := i.IteratorStep()
	result, isAbrupt, rt := ReturnIfAbrupt(resultCompletion, co)
	if isAbrupt {
		return rt, false
	}
	if isDone {
		return co, true
	}
	value, isAbrupt, rt := ReturnIfAbrupt(IteratorValue(result), co)
	if isAbrupt {
		i.Done = true
		return rt, false
	}
	co.value = value
	return
}

// IteratorClose
// spec: 7.4.9
func (i *IteratorRecord) IteratorClose(completion CompletionValue) (co CompletionValue) {
	iterator := i.Iterator
	agent := iterator.Agent()
	iteratorValue := iterator.ToValue()
	innerResult := GetV(agent, iteratorValue, NewStringPropertyKey("return"))
	if !innerResult.IsAbrupt() {
		returnMethod := innerResult.Data()
		if IsUndefinedOrNull(returnMethod) {
			return completion
		}
		if !IsCallable(returnMethod) {
			innerResult = co.ThrowTypeError(agent, "iterator return method is not callable")
		} else {
			// Closing is observable even when the original completion is a throw.
			innerResult = returnMethod.Call(agent, iteratorValue, nil)
		}
	}

	// The original throw wins over failures produced while closing, but only
	// after the return getter and method have had their observable effects.
	if completion.IsError() || completion.t == CompletionTypeThrow {
		return completion
	}
	if innerResult.IsAbrupt() {
		return innerResult
	}
	if innerResult.Data() == nil || !innerResult.Data().IsObject() {
		return co.ThrowTypeError(agent, "iterator return method must return an object")
	}
	return completion
}

// AsyncIteratorClose closes an async iterator and awaits the return result
// while preserving an incoming throw completion.
// spec: 7.4.11
func (i *IteratorRecord) AsyncIteratorClose(agent *Agent, completion CompletionValue) (co CompletionValue) {
	iteratorValue := i.Iterator.ToValue()
	returnMethod, isAbrupt, innerResult := ReturnIfAbrupt(
		GetMethodCompletion(agent, iteratorValue, NewStringPropertyKey("return")),
		co,
	)
	if !isAbrupt && returnMethod == nil {
		return completion
	}
	if !isAbrupt {
		innerValue, callAbrupt, callResult := ReturnIfAbrupt(
			returnMethod.Call(iteratorValue, nil),
			co,
		)
		if callAbrupt {
			innerResult = callResult
		} else {
			awaitedValue, awaitAbrupt, awaitResult := ReturnIfAbrupt(Await(agent, innerValue), co)
			if awaitAbrupt {
				innerResult = awaitResult
			} else {
				innerResult.value = awaitedValue
			}
		}
	}

	if completion.IsError() || completion.t == CompletionTypeThrow {
		return completion
	}
	if innerResult.IsAbrupt() {
		return innerResult
	}
	if innerResult.value == nil || !innerResult.value.IsObject() {
		return co.ThrowTypeError(agent, "async iterator return method must return an object")
	}
	return completion
}

// CreateIterResultObject
// spec: 7.4.12
// returns Object
func CreateIterResultObject(agent *Agent, value Value, done bool) ObjectType {
	realm := agent.CurrentRealm()
	obj := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)

	obj.CreateDataPropertyOrThrow(NewStringPropertyKey("value"), value)
	obj.CreateDataPropertyOrThrow(NewStringPropertyKey("done"), NewBooleanValue(done))
	return obj
}

// IteratorToList
// spec: 7.4.14
func (i *IteratorRecord) IteratorToList() (co Completion[[]Value]) {
	var values []Value
	for {
		next, isDone := i.IteratorStepValue()
		if isDone {
			break
		}
		nextValue, isAbrupt, rt := ReturnIfAbrupt(next, co)
		if isAbrupt {
			return rt
		}
		values = append(values, nextValue)
	}
	co.value = values
	return
}

func CreateAsyncFromSyncIterator(iteratorRecord *IteratorRecord) *IteratorRecord {
	agent := iteratorRecord.Iterator.Agent()
	realm := agent.CurrentRealm()
	asyncIterator := NewObject(agent, realm.Intrinsics.AsyncIteratorPrototype, "AsyncFromSyncIterator")

	var next BehaviorFn = func(_ Value, argumentsList []Value, _ ObjectType) CompletionConvertable[Value] {
		promiseCapability := NewPromiseCapability(agent, realm.Intrinsics.Promise.ToValue())
		var result CompletionValue
		var nextValue Value
		if len(argumentsList) != 0 {
			nextValue = argumentAt(argumentsList, 0)
		}
		nextResult := iteratorRecord.IteratorNext(nextValue)
		if nextResult.IsAbrupt() {
			result = CompletionFrom(result, nextResult)
		} else {
			result.value = nextResult.Data().ToValue()
		}
		return AsyncFromSyncIteratorContinuation(agent, result, promiseCapability, iteratorRecord)
	}
	asyncIterator.defineBuiltinFunction(realm, CMString("next"), next, 1)

	var iteratorReturn BehaviorFn = func(_ Value, argumentsList []Value, _ ObjectType) CompletionConvertable[Value] {
		promiseCapability := NewPromiseCapability(agent, realm.Intrinsics.Promise.ToValue())
		value := argumentAt(argumentsList, 0)
		returnMethod := GetMethodCompletion(agent, iteratorRecord.Iterator.ToValue(), NewStringPropertyKey("return"))
		if returnMethod.IsAbrupt() {
			return rejectAsyncFromSyncIteratorPromise(promiseCapability, returnMethod.Error())
		}
		if returnMethod.Data() == nil {
			iteratorRecord.Done = true
			iterResult := CreateIterResultObject(agent, value, true)
			ReturnAssertNormal(promiseCapability.Resolve.Call(UndefinedValue, []Value{iterResult.ToValue()}))
			return promiseCapability.Promise.ToValue()
		}
		result := returnMethod.Data().Call(iteratorRecord.Iterator.ToValue(), []Value{value})
		return AsyncFromSyncIteratorContinuation(agent, result, promiseCapability, iteratorRecord)
	}
	asyncIterator.defineBuiltinFunction(realm, CMString("return"), iteratorReturn, 1)

	var iteratorThrow BehaviorFn = func(_ Value, argumentsList []Value, _ ObjectType) CompletionConvertable[Value] {
		promiseCapability := NewPromiseCapability(agent, realm.Intrinsics.Promise.ToValue())
		value := argumentAt(argumentsList, 0)
		throwMethod := GetMethodCompletion(agent, iteratorRecord.Iterator.ToValue(), NewStringPropertyKey("throw"))
		if throwMethod.IsAbrupt() {
			return rejectAsyncFromSyncIteratorPromise(promiseCapability, throwMethod.Error())
		}
		if throwMethod.Data() == nil {
			iteratorRecord.Done = true
			returnMethod := GetMethodCompletion(agent, iteratorRecord.Iterator.ToValue(), NewStringPropertyKey("return"))
			if returnMethod.IsAbrupt() {
				return rejectAsyncFromSyncIteratorPromise(promiseCapability, returnMethod.Error())
			}
			if returnMethod.Data() != nil {
				returnResult := returnMethod.Data().Call(iteratorRecord.Iterator.ToValue(), nil)
				if returnResult.IsAbrupt() {
					return rejectAsyncFromSyncIteratorPromise(promiseCapability, returnResult.Error())
				}
				if returnResult.Data() == nil || !returnResult.Data().IsObject() {
					reason := agent.ThrowTypeError("sync iterator return method must return an object")
					return rejectAsyncFromSyncIteratorPromise(promiseCapability, reason)
				}
			}
			reason := agent.ThrowTypeError("sync iterator does not provide a throw method")
			return rejectAsyncFromSyncIteratorPromise(promiseCapability, reason)
		}
		result := throwMethod.Data().Call(iteratorRecord.Iterator.ToValue(), []Value{value})
		return AsyncFromSyncIteratorContinuation(agent, result, promiseCapability, iteratorRecord)
	}
	asyncIterator.defineBuiltinFunction(realm, CMString("throw"), iteratorThrow, 1)

	nextMethod := ReturnAssertNormal(GetV(agent, asyncIterator.ToValue(), NewStringPropertyKey("next")))
	return &IteratorRecord{
		Iterator:   asyncIterator,
		NextMethod: nextMethod,
	}
}

// AsyncFromSyncIteratorContinuation converts a synchronous iterator result to
// a promise for an async iterator result, assimilating a promise-valued value.
// spec: 27.1.5.2.1
func AsyncFromSyncIteratorContinuation(
	agent *Agent,
	result CompletionValue,
	promiseCapability *PromiseCapability,
	syncIteratorRecord *IteratorRecord,
) Value {
	if result.IsAbrupt() {
		return rejectAsyncFromSyncIteratorPromise(promiseCapability, result.Error())
	}
	if result.value == nil || !result.value.IsObject() {
		reason := agent.ThrowTypeError("sync iterator method must return an object")
		return rejectAsyncFromSyncIteratorPromise(promiseCapability, reason)
	}

	iterResult := MustGetObject(result.value)
	doneCompletion := IteratorComplete(iterResult)
	if doneCompletion.IsAbrupt() {
		return rejectAsyncFromSyncIteratorPromise(promiseCapability, doneCompletion.Error())
	}
	done := doneCompletion.Data()
	syncIteratorRecord.Done = done
	valueCompletion := IteratorValue(iterResult)
	if valueCompletion.IsAbrupt() {
		return rejectAsyncFromSyncIteratorPromise(promiseCapability, valueCompletion.Error())
	}

	realm := agent.CurrentRealm()
	valueWrapper := PromiseResolve(agent, realm.Intrinsics.Promise, valueCompletion.Data())
	var unwrap BehaviorFn = func(_ Value, argumentsList []Value, _ ObjectType) CompletionConvertable[Value] {
		value := argumentAt(argumentsList, 0)
		return CreateIterResultObject(agent, value, done).ToValue()
	}
	onFulfilled := CreateBuiltinFunction(agent, unwrap, 1, CMString(""), builtinFunctionArgs{realm: realm})
	PerformPromiseThen(
		agent,
		valueWrapper,
		onFulfilled.ToValue(),
		UndefinedValue,
		promiseCapability,
	)
	return promiseCapability.Promise.ToValue()
}

func rejectAsyncFromSyncIteratorPromise(promiseCapability *PromiseCapability, reason Value) Value {
	ReturnAssertNormal(promiseCapability.Reject.Call(UndefinedValue, []Value{reason}))
	return promiseCapability.Promise.ToValue()
}
