package coldmoon

func NewAsyncGeneratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.AsyncIteratorPrototype, "AsyncGeneratorPrototype")

	DefineBuiltinPropertyP(object, "constructor", &PropertyDescriptor{
		Value:        (realm.Intrinsics.AsyncGeneratorFunctionPrototype).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("AsyncGenerator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}
