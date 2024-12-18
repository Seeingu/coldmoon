package coldmoon

func NewAsyncGeneratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.AsyncIteratorPrototype, "AsyncGeneratorPrototype")

	DefineBuiltinPropertyP(object, "constructor", &PropertyDescriptor{
		Value:        (realm.Intrinsics.AsyncGeneratorFunctionPrototype).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	DefineToStringTagBuiltinProperty(object, "AsyncGenerator")

	return object
}
