package coldmoon

type AsyncIteratorPrototypeObject struct {
	*Object
}

func NewAsyncIteratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "AsyncIteratorPrototype")

	asyncIterator := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return this
	}
	object.defineBuiltinFunction(realm, WellKnownSymbolsAsyncIterator, asyncIterator, 0)

	return object
}
