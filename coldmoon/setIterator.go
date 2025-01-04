package coldmoon

type SetIteratorObject struct {
	*Object
	SetObject *SetObject
	Kind      objectOwnPropertiesKind
	Index     JSInt
}

func CreateSetIterator(agent *Agent, value Value, kind objectOwnPropertiesKind) *SetIteratorObject {
	realm := agent.CurrentRealm()
	s := &SetIteratorObject{
		Object:    NewObject(agent, realm.Intrinsics.SetIteratorPrototype, "SetIterator"),
		SetObject: RequireInternalSlot[*SetObject](value),
		Kind:      kind,
		Index:     0,
	}
	s.ref = s
	return s
}

func NewSetIteratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.IteratorPrototype, "SetIteratorPrototype")
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		setIterator := MustGetObject(thisValue).(*SetIteratorObject)
		s := setIterator.SetObject
		index := setIterator.Index
		kind := setIterator.Kind
		entries := s.items()
		numEntries := JSInt(s.size())
		if index >= numEntries {
			return (CreateIterResultObject(realm.Agent, UndefinedValue, true)).ToValue()
		}
		var value Value
		value = entries[index]
		setIterator.Index = index + 1
		var result Value
		switch kind {
		case objectOwnPropertiesKindValue:
			result = value
		case objectOwnPropertiesKindKeyAndValue:
			result = (CreateArrayFromList(realm.Agent, []Value{value, value})).ToValue()
		case objectOwnPropertiesKindKey:
			panic("unreachable")
		}
		return (CreateIterResultObject(realm.Agent, result, false)).ToValue()
	}
	object.defineBuiltinFunction(realm, CMString("next"), next, 0)
	object.defineToStringTag("Set Iterator")
	return object
}
