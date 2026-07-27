package coldmoon

type IntlObject struct {
	*Object
}

func NewIntlObject(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "Intl")
	object.defineToStringTag("Intl")
	return object
}
