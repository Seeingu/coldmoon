package coldmoon

type StringIteratorObject struct {
	*Object
	Data  string
	Index uint64
}

func NewStringIteratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.IteratorPrototype, "StringIteratorPrototype")
	agent := realm.Agent
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		stringIterator := MustGetObject(thisValue).(*StringIteratorObject)
		data := stringIterator.Data
		length := len(data)
		if stringIterator.Index >= uint64(length) {
			return NewValueFromObject(CreateIterResultObject(agent, UndefinedValue, true))
		}
		result := NewStringValue(string(data[stringIterator.Index]))
		stringIterator.Index += 1
		return NewValueFromObject(CreateIterResultObject(agent, result, false))
	}
	DefineBuiltinFunction(object, "next", next, 0, realm)
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("String Iterator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
