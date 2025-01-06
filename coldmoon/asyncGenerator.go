package coldmoon

func NewAsyncGeneratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.AsyncIteratorPrototype, "AsyncGeneratorPrototype")

	object.defineBuiltinProperty(CMString("constructor"), &PropertyDescriptor{
		Value:        realm.Intrinsics.AsyncGeneratorFunctionPrototype.ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	object.defineToStringTag("AsyncGenerator")

	return object
}
