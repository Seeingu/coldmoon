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

// GetIteratorFromMethod
// spec: 7.4.2
func GetIteratorFromMethod(agent *Agent, object Value, method ObjectType) *IteratorRecord {
	iterator := method.Call(object, nil).value
	if !iterator.IsObject() {
		panic("TypeError")
	}
	nextMethod := ReturnAssertNormal(GetV(agent, iterator, NewStringPropertyKey("next")))
	// TODO: not standard, for debug
	Assert(nextMethod != UndefinedValue)
	iteratorRecord := &IteratorRecord{
		Iterator:   MustGetObject(iterator),
		NextMethod: nextMethod,
	}
	return iteratorRecord
}

// GetIterator
// spec: 7.4.3
func GetIterator(agent *Agent, obj Value, kind IteratorKind) (co Completion[*IteratorRecord]) {
	var method ObjectType
	switch kind {
	case IteratorKindSync:
		method = GetMethod(agent, obj, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]))
	case IteratorKindAsync:
		method = GetMethod(agent, obj, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsAsyncIterator]))
		if method == nil {
			syncMethod := GetMethod(agent, obj, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]))
			if syncMethod == nil {
				return co.ThrowTypeError(agent, "GetIterator: async no iterator method")
			}
			syncIteratorRecord := GetIteratorFromMethod(agent, obj, syncMethod)
			co.value = CreateAsyncFromSyncIterator(syncIteratorRecord)
			return
		}
	}
	if method == nil {
		return co.ThrowTypeError(agent, "No iterator method")
	}
	co.value = GetIteratorFromMethod(agent, obj, method)
	return
}

// IteratorNext
// spec: 7.4.4
func (i *IteratorRecord) IteratorNext(value Value) (co Completion[ObjectType]) {
	var result CompletionValue
	if value == nil {
		result = i.NextMethod.CallNoArgs(i.Iterator.ToValue())
	} else {
		result = i.NextMethod.Call(i.Iterator.Agent(), i.Iterator.ToValue(), []Value{value})
	}

	resultObject, ok := result.value.(*ObjectValue)
	if !ok {
		return co.ThrowTypeError(i.Iterator.Agent(), "IteratorNext method must return an object")
	}
	co.value = resultObject.Object
	return
}

// IteratorComplete
// spec: 7.4.5
func IteratorComplete(iterResult ObjectType) (co Completion[bool]) {
	co.value = iterResult.Get(NewStringPropertyKey("done")).ToBoolean()
	return
}

// 7.4.6
func IteratorValue(iterResult ObjectType) Value {
	return iterResult.Get(NewStringPropertyKey("value"))
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
		co = rt
		return
	}
	if done {
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
	result := i.IteratorNext(nil)
	if result.t == CompletionTypeThrow {
		i.Done = true
		return CompletionFrom(co, result), false
	}
	done := IteratorComplete(result.value)
	if done.t == CompletionTypeThrow {
		i.Done = true
		return CompletionFrom(co, done), false
	}
	if done.value {
		i.Done = true
		return co, true
	}
	value := result.value.Get(CMString("value").ToPropertyKey())
	// TODO(BM): value is throw
	co.value = value
	return
}

// TODO(BM): handle completion
// IteratorClose
// spec: 7.4.9
func (i *IteratorRecord) IteratorClose(completion CompletionValue) (co CompletionValue) {
	iterator := i.Iterator
	agent := iterator.Agent()
	innerResult := GetMethod(agent, iterator.ToValue(), NewStringPropertyKey("return"))

	if innerResult != nil {
		innerResult.ToValue().CallNoArgs(iterator.ToValue())
	}
	return
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
	// TODO
	return iteratorRecord
}
