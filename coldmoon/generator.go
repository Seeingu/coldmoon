package coldmoon

func NewGeneratorPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.IteratorPrototype)

	DefineBuiltinProperty(object, "constructor", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.GeneratorFunctionPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	DefineBuiltinProperty(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("Generator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}
