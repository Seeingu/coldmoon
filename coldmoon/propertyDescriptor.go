package coldmoon

type PropertyDescriptor struct {
	Value      Value
	Writable   bool
	Get        ObjectType
	Set        ObjectType
	Enumerable bool
	// TODO: use nullable
	Configurable bool
}

// 6.2.6.1
func (p *PropertyDescriptor) IsAccessorDescriptor() bool {
	return p.Get != nil || p.Set != nil
}

func (p *PropertyDescriptor) IsDataDescriptor() bool {
	return p.Value != nil || p.Writable
}

func (p *PropertyDescriptor) IsGenericDescriptor() bool {
	return !p.IsAccessorDescriptor() && !p.IsDataDescriptor()
}

func (p *PropertyDescriptor) IsFullyPopulated() bool {
	return p.Value != nil && p.Get != EmptyObject && p.Set != EmptyObject
}

func (p *PropertyDescriptor) HasFields() bool {
	return p.Value != nil || p.Writable || p.Get != EmptyObject || p.Set != EmptyObject || p.Enumerable || p.Configurable
}
