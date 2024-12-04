package coldmoon

type SetIteratorObject struct {
	*Object
	SetObject *SetObject
	Kind      objectOwnPropertiesKind
	Index     uint64
}

func CreateSetIterator(agent *Agent, value Value, kind objectOwnPropertiesKind) *SetIteratorObject {
	return &SetIteratorObject{
		Object:    NewObject(agent, agent.CurrentRealm().Intrinsics.SetIteratorPrototype),
		SetObject: RequireInternalSlot[*SetObject](value),
		Kind:      kind,
		Index:     0,
	}
}

func NewSetIteratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.IteratorPrototype)
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		setIterator := MustGetObject(thisValue).(*SetIteratorObject)
		s := setIterator.SetObject
		index := setIterator.Index
		kind := setIterator.Kind

		entries := s.SetValue.Data
		numEntries := uint64(len(entries))
		if index >= numEntries {
			return NewValueFromObject(CreateIterResultObject(realm.Agent, UndefinedValue, true))
		}
		var value Value
		for index < numEntries {
			if _, ok := entries[NewNumberValue(JSNumber(index))]; ok {
				break
			}
			index++
		}
		if index >= numEntries {
			return NewValueFromObject(CreateIterResultObject(realm.Agent, UndefinedValue, true))
		}
		setIterator.Index = index
		key := NewNumberValue(JSNumber(index))
		value = entries[key]
		var result Value
		switch kind {
		case objectOwnPropertiesKindValue:
			result = value
		case objectOwnPropertiesKindKeyAndValue:
			result = NewValueFromObject(CreateArrayFromList(realm.Agent, []Value{value, value}))
		case objectOwnPropertiesKindKey:
			panic("unreachable")
		}
		return NewValueFromObject(CreateIterResultObject(realm.Agent, result, false))
	}
	DefineBuiltinFunction(object, "next", next, 0, realm)
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("SetObject Iterator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
