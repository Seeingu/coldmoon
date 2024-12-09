package coldmoon

type IntlObject struct {
	*Object
}

func NewIntlObject(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "Intl")
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("Intl"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
