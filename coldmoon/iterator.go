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

func (i *IteratorRecord) ToCompletion() Completion[*IteratorRecord] {
	return NewNormalCompletion(i)
}

func NewIteratorPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "IteratorPrototype")
	var iterator BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		return thisValue
	}
	DefineBuiltinFunction(object, "@@iterator", iterator, 0, realm)
	return object
}

// 7.4.2
func GetIteratorFromMethod(agent *Agent, object Value, method ObjectType) *IteratorRecord {
	iterator := NewValueFromObject(method).CallAssumeCallable(object, nil)
	if !ValueIsObject(iterator) {
		panic("TypeError")
	}
	nextMethod := GetV(agent, iterator, NewStringPropertyKey("next"))
	iteratorRecord := &IteratorRecord{
		Iterator:   MustGetObject(iterator),
		NextMethod: nextMethod,
	}
	return iteratorRecord
}

// 7.4.3

func GetIterator(agent *Agent, obj Value, kind IteratorKind) Completion[*IteratorRecord] {
	var method ObjectType
	switch kind {
	case IteratorKindSync:
		method = GetMethod(agent, obj, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]))
	case IteratorKindAsync:
		method = GetMethod(agent, obj, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsAsyncIterator]))
		if method == nil {
			syncMethod := GetMethod(agent, obj, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]))
			if syncMethod == nil {
				panic("TypeError")
			}
			syncIteratorRecord := GetIteratorFromMethod(agent, obj, syncMethod)
			return CreateAsyncFromSyncIterator(syncIteratorRecord).ToCompletion()
		}
	}
	if method == nil {
		return NewThrowCompletion[*IteratorRecord](agent.ThrowException(TypeError, "No iterator method"))
	}
	return GetIteratorFromMethod(agent, obj, method).ToCompletion()
}

// 7.4.4
func (i *IteratorRecord) IteratorNext(value Value) ObjectType {
	var result Value
	if value == nil {
		result = CallAssumeCallableNoArgs(i.NextMethod, NewValueFromObject(i.Iterator))
	} else {
		result = i.NextMethod.CallAssumeCallable(NewValueFromObject(i.Iterator), []Value{value})
	}

	resultObject, ok := result.(*ObjectValue)
	if !ok {
		panic("TypeError")
	}
	return resultObject.Object
}

// 7.4.5
func IteratorComplete(iterResult ObjectType) bool {
	return iterResult.Get(NewStringPropertyKey("done")).ToBoolean()
}

// 7.4.6
func IteratorValue(iterResult ObjectType) Value {
	return iterResult.Get(NewStringPropertyKey("value"))
}

// 7.4.7
func (i *IteratorRecord) IteratorStep() ObjectType {
	result := i.IteratorNext(nil)
	done := IteratorComplete(result)
	if done {
		return nil
	}
	return result
}

// 7.4.9
func (i *IteratorRecord) IteratorClose() {
	iterator := i.Iterator
	agent := iterator.Agent()
	innerResult := GetMethod(agent, NewValueFromObject(iterator), NewStringPropertyKey("return"))

	if innerResult != nil {
		CallAssumeCallableNoArgs(NewValueFromObject(innerResult), NewValueFromObject(iterator))
	}
}

// 7.4.12
func CreateIterResultObject(agent *Agent, value Value, done bool) ObjectType {
	realm := agent.CurrentRealm()
	obj := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)

	obj.CreateDataPropertyOrThrow(NewStringPropertyKey("value"), value)
	obj.CreateDataPropertyOrThrow(NewStringPropertyKey("done"), NewBooleanValue(done))
	return obj
}

// 7.4.14
func (i *IteratorRecord) IteratorToList() (values []Value) {
	for {
		next := i.IteratorStep()
		if next == nil {
			break
		}
		values = append(values, IteratorValue(next))
	}
	return
}

func CreateAsyncFromSyncIterator(iteratorRecord *IteratorRecord) *IteratorRecord {
	// TODO
	return iteratorRecord
}
