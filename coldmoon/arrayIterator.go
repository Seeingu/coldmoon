package coldmoon

type ArrayIteratorObject struct {
	*Object
	Array *ArrayObject
	Kind  objectOwnPropertiesKind
	Index uint64
}

func NewArrayIteratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.IteratorPrototype)
	agent := realm.Agent
	// 23.1.5.2.1
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		arrayIterator := MustGetObject(thisValue).(*ArrayIteratorObject)
		array := arrayIterator.Array
		index := arrayIterator.Index
		kind := arrayIterator.Kind

		length := array.LengthOfArrayLike()
		if index >= length {
			return NewValueFromObject(CreateIterResultObject(agent, UndefinedValue, true))
		}
		indexNumber := NewNumberValue(float64(index))
		var result Value
		if kind == objectOwnPropertiesKindKey {
			result = indexNumber
		} else {
			elementKey := NewIntegerIndexPropertyKey(int(index))
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
	DefineBuiltinProperty(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("Array Iterator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}

// 23.1.5.1
func CreateArrayIterator(agent *Agent, array *ArrayObject, kind objectOwnPropertiesKind) *ArrayIteratorObject {
	return &ArrayIteratorObject{
		Object: NewObject(agent, agent.CurrentRealm().Intrinsics.ArrayIteratorPrototype),
		Array:  array,
		Kind:   kind,
		Index:  0,
	}
}
