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
	g.ref = g

	DefineBuiltinPropertyP(object, "constructor", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.GeneratorFunctionPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	DefineToStringTagBuiltinProperty(object, "Generator")

	return g
}
