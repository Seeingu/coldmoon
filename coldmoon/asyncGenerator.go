package coldmoon

func NewAsyncGeneratorPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.AsyncIteratorPrototype)

	DefineBuiltinProperty(object, "constructor", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.AsyncGeneratorFunctionPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	DefineBuiltinProperty(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("AsyncGenerator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}
