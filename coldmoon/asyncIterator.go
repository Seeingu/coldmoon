package coldmoon

type AsyncIteratorPrototypeObject struct {
	*Object
}

func NewAsyncIteratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype)

	asyncIterator := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		return this
	}
	DefineBuiltinFunction(object, "@@asyncIterator", asyncIterator, 0, realm)

	return object
}
