package coldmoon

type ObjectPrototype struct {
	*Object
}

func NewObjectPrototype(realm *Realm) *ObjectPrototype {
	o := NewObject(realm.Agent, nil)
	o.InternalMethods().SetPrototypeOf = ImmutableSetPrototypeOf
	return &ObjectPrototype{o}
}

func (o *ObjectPrototype) ToObject() *Object {
	return o.Object
}
