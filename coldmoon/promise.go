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

// MARK: - PromiseObject

type PromiseObject struct {
	*Object
	PromiseState     PromiseState
	PromiseResult    Value
	PromiseIsHandled bool
}

func NewPromisePrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)

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
			RejectPromise(_promise, agent.exception)
			return UndefinedValue
		}
		if !ValueIsObject(resolution) {
			FulfillPromise(_promise, resolution)
			return UndefinedValue
		}
		then := MustGetObject(resolution).Get(NewStringPropertyKey("then"))
		if !IsCallable(then) {
			FulfillPromise(_promise, resolution)
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
func FulfillPromise(promise *PromiseObject, value Value) {
	Assert(promise.PromiseState == PromiseStatePending)
	promise.PromiseState = PromiseStateFulfilled
	promise.PromiseResult = value
}

// 27.2.1.7
func RejectPromise(promise *PromiseObject, reason Value) {
	Assert(promise.PromiseState == PromiseStatePending)
	promise.PromiseState = PromiseStateRejected
	promise.PromiseResult = reason
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
