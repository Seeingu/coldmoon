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
	var iterator BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		return thisValue
	}

	DefineBuiltinFunction(realm, WellKnownSymbolsIterator, object, iterator, 0)
	return object
}

// 7.4.2
func GetIteratorFromMethod(agent *Agent, object Value, method ObjectType) *IteratorRecord {
	iterator := method.Call(object, nil)
	if !iterator.IsObject() {
		panic("TypeError")
	}
	nextMethod := GetV(agent, iterator, NewStringPropertyKey("next"))
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
				panic("TypeError")
			}
			syncIteratorRecord := GetIteratorFromMethod(agent, obj, syncMethod)
			co.value = CreateAsyncFromSyncIterator(syncIteratorRecord)
			return
		}
	}
	if method == nil {
		// TODO: check completion type
		co.err = agent.ThrowException(TypeError, "No iterator method")
		return
	}
	co.value = GetIteratorFromMethod(agent, obj, method)
	return
}

// 7.4.4
func (i *IteratorRecord) IteratorNext(value Value) ObjectType {
	var result Value
	if value == nil {
		result = i.NextMethod.CallNoArgs(i.Iterator.ToValue())
	} else {
		result = i.NextMethod.Call((i.Iterator).ToValue(), []Value{value})
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
	innerResult := GetMethod(agent, iterator.ToValue(), NewStringPropertyKey("return"))

	if innerResult != nil {
		innerResult.ToValue().CallNoArgs(iterator.ToValue())
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
