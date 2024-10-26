package coldmoon

type PropertyDescriptor struct {
	Value        Value
	Writable     bool
	Get          *Object
	Set          *Object
	Enumerable   bool
	Configurable bool
}

func (p *PropertyDescriptor) IsAccessorDescriptor() bool {
	return p.Get != EmptyObject || p.Set != EmptyObject
}

func (p *PropertyDescriptor) IsDataDescriptor() bool {
	return p.Value != nil || p.Writable
}

func (p *PropertyDescriptor) IsGenericDescriptor() bool {
	return !p.IsAccessorDescriptor() && !p.IsDataDescriptor()
}
