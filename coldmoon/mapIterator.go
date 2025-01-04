package coldmoon

type MapIteratorObject struct {
	*Object
	Map   *MapObject
	Kind  objectOwnPropertiesKind
	Index JSInt
}

// 24.1.5.1
func CreateMapIterator(agent *Agent, value Value, kind objectOwnPropertiesKind) *MapIteratorObject {
	realm := agent.CurrentRealm()
	m := &MapIteratorObject{
		Object: NewObject(agent, realm.Intrinsics.MapIteratorPrototype, "MapIterator"),
		Map:    RequireInternalSlot[*MapObject](value),
		Kind:   kind,
		Index:  0,
	}
	m.ref = m
	return m
}

func NewMapIteratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.IteratorPrototype, "MapIteratorPrototype")
	agent := realm.Agent
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		mapIterator := MustGetObject(thisValue).(*MapIteratorObject)
		m := mapIterator.Map
		index := mapIterator.Index
		kind := mapIterator.Kind

		entries := m.MapValue.Data
		numEntries := JSInt(len(entries))
		if index >= numEntries {
			return CreateIterResultObject(agent, UndefinedValue, true).ToValue()
		}

		for index < numEntries {
			if _, ok := entries[NewNumberValue(index.ToNumber()).Hash()]; ok {
				break
			}
			index++
		}
		if index >= numEntries {
			return CreateIterResultObject(agent, UndefinedValue, true).ToValue()
		}
		mapIterator.Index = index
		key := NewNumberValue(index.ToNumber())
		value := entries[key.Hash()]

		var result Value
		switch kind {
		case objectOwnPropertiesKindKey:
			result = key
		case objectOwnPropertiesKindValue:
			result = value
		case objectOwnPropertiesKindKeyAndValue:
			result = CreateArrayFromList(agent, []Value{key, value}).ToValue()

		}
		return (CreateIterResultObject(agent, result, false)).ToValue()
	}
	object.defineBuiltinFunction(realm, CMString("next"), next, 0)
	object.defineToStringTag("Map Iterator")

	return object
}
