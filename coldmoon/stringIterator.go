package coldmoon

type StringIteratorObject struct {
	*Object
	Data  string
	Index uint64
}

func NewStringIteratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.IteratorPrototype, "StringIteratorPrototype")
	agent := realm.Agent
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) CompletionConvertable[Value] {
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
	object.defineBuiltinFunction(realm, CMString("next"), next, 0)
	object.defineToStringTag("String Iterator")
	return object
}
