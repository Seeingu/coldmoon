package coldmoon

type ObjectPrototype struct {
	*Object
}

func NewObjectPrototype(agent *Agent) *ObjectPrototype {
	o := NewObject(agent, nil)
	o.InternalMethods().SetPrototypeOf = ImmutableSetPrototypeOf
	return &ObjectPrototype{o}
}

func (o *ObjectPrototype) ToObject() *Object {
	return o.Object
}
