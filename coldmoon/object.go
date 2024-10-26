package coldmoon

type Data struct {
	prototype       *Object
	extensible      bool
	agent           *Agent
	internalMethods InternalMethods
	propertyStorage PropertyStorage
}

type Object struct {
	data *Data
}

func NewObject(prototype *Object) *Object {
	o := &Object{
		data: &Data{
			prototype:       prototype,
			internalMethods: NewInternalMethods(),
			propertyStorage: NewPropertyStorage(),
		},
	}
	return o
}

var EmptyObject = &Object{}

func (o *Object) Prototype() *Object {
	return o.data.prototype
}
func (o *Object) SetPrototype(p *Object) {
	o.data.prototype = p
}

func (o *Object) Extensible() bool {
	return o.data.extensible
}

func (o *Object) Agent() *Agent {
	return o.data.agent
}

func (o *Object) InternalMethods() *InternalMethods {
	return &o.data.internalMethods
}

func (o *Object) PropertyStorage() *PropertyStorage {
	return &o.data.propertyStorage
}

func IsExtensible(o *Object) bool {
	return o.InternalMethods().IsExtensible(o)
}
