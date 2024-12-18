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
			return (CreateIterResultObject(agent, UndefinedValue, true)).ToValue()
		}
		result := NewStringValue(string(data[stringIterator.Index]))
		stringIterator.Index += 1
		return (CreateIterResultObject(agent, result, false)).ToValue()
	}
	DefineBuiltinFunction(object, "next", next, 0, realm)
	DefineToStringTagBuiltinProperty(object, "String Iterator")
	return object
}
