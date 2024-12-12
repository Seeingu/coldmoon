package coldmoon

type ArrayIteratorObject struct {
	*Object
	Array ObjectType
	Kind  objectOwnPropertiesKind
	Index JSInt
}

func NewArrayIteratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.IteratorPrototype, "ArrayIteratorPrototype")
	agent := realm.Agent
	// 23.1.5.2.1
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		arrayIterator := MustGetObject(thisValue).(*ArrayIteratorObject)
		array := arrayIterator.Array
		index := arrayIterator.Index
		kind := arrayIterator.Kind

		var length JSInt
		if typedArray, ok := array.(*TypedArrayObject); ok {
			taRecord := MakeTypedArrayWithBufferWitnessRecord(typedArray, SeqCst)
			if IsTypedArrayOutOfBounds(taRecord) {
				panic("TypeError")
			}
			length = TypedArrayLength(taRecord)
		} else {
			length = array.LengthOfArrayLike()
		}
		if index >= length {
			return NewValueFromObject(CreateIterResultObject(agent, UndefinedValue, true))
		}
		indexNumber := NewNumberValue(index.ToNumber())
		var result Value
		if kind == objectOwnPropertiesKindKey {
			result = indexNumber
		} else {
			elementKey := NewIntegerIndexPropertyKey(index)
			elementValue := array.Get(elementKey)
			if kind == objectOwnPropertiesKindValue {
				result = elementValue
			} else {
				result = NewValueFromObject(CreateArrayFromList(agent, []Value{indexNumber, elementValue}))
			}
		}
		arrayIterator.Index += 1
		return NewValueFromObject(CreateIterResultObject(agent, result, false))
	}
	DefineBuiltinFunction(object, "next", next, 0, realm)
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("Array Iterator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}

// 23.1.5.1
func CreateArrayIterator(agent *Agent, array ObjectType, kind objectOwnPropertiesKind) *ArrayIteratorObject {
	a := &ArrayIteratorObject{
		Object: NewObject(agent, agent.CurrentRealm().Intrinsics.ArrayIteratorPrototype, "ArrayIterator"),
		Array:  array,
		Kind:   kind,
		Index:  0,
	}
	a.ref = a
	return a
}
