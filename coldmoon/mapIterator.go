package coldmoon

type MapIteratorObject struct {
	*Object
	Map   *MapObject
	Kind  objectOwnPropertiesKind
	Index JSInt
}

// 24.1.5.1
func CreateMapIterator(agent *Agent, value Value, kind objectOwnPropertiesKind) *MapIteratorObject {
	return &MapIteratorObject{
		Object: NewObject(agent, agent.CurrentRealm().Intrinsics.MapIteratorPrototype, "MapIterator"),
		Map:    RequireInternalSlot[*MapObject](value),
		Kind:   kind,
		Index:  0,
	}
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
			return NewValueFromObject(CreateIterResultObject(agent, UndefinedValue, true))
		}

		for index < numEntries {
			if _, ok := entries[NewNumberValue(index.ToNumber())]; ok {
				break
			}
			index++
		}
		if index >= numEntries {
			return NewValueFromObject(CreateIterResultObject(agent, UndefinedValue, true))
		}
		mapIterator.Index = index
		key := NewNumberValue(index.ToNumber())
		value := entries[key]

		var result Value
		switch kind {
		case objectOwnPropertiesKindKey:
			result = key
		case objectOwnPropertiesKindValue:
			result = value
		case objectOwnPropertiesKindKeyAndValue:
			result = NewValueFromObject(CreateArrayFromList(agent, []Value{key, value}))

		}
		return NewValueFromObject(CreateIterResultObject(agent, result, false))
	}
	DefineBuiltinFunction(object, "next", next, 0, realm)
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("Map Iterator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}
