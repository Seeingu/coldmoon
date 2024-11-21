package coldmoon

type PromiseState int

const (
	PromiseStatePending PromiseState = iota
	PromiseStateFulfilled
	PromiseStateRejected
)

type PromiseObject struct {
	*Object
	PromiseState     PromiseState
	PromiseResult    Value
	PromiseIsHandled bool
}

func NewPromisePrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)
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
		executor.CallAssumeCallable(UndefinedValue, []Value{NewValueFromObject(resolvingFunctions.Resolve)})
		return NewValueFromObject(promise)
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "Promise", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.PromisePrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(realm.Intrinsics.PromisePrototype, "constructor", NewValueFromObject(object))

	return object
}

type ResolvingFunctions struct {
	Resolve ObjectType
	Reject  ObjectType
}

type AlreadyResolved struct {
	Value bool
}
type AdditionalFields struct {
	Promise         *PromiseObject
	AlreadyResolved *AlreadyResolved
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
		Resolve: resolve,
		Reject:  reject,
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
