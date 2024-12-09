package coldmoon

type GeneratorObject struct {
	*Object
}

func NewGeneratorPrototype(realm *Realm) *GeneratorObject {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.IteratorPrototype, "GeneratorPrototype")
	g := &GeneratorObject{
		Object: object,
	}

	DefineBuiltinPropertyP(object, "constructor", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.GeneratorFunctionPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("Generator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return g
}
