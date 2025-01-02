package coldmoon

type BaseValue struct {
	Value
	ref Value
}

var _ Value = (*BaseValue)(nil)

func (b *BaseValue) ToCompletion() CompletionValue {
	return NewCompletionValue(b.ref)
}

func (b *BaseValue) ToPropertyDescriptor() *PropertyDescriptor {
	return &PropertyDescriptor{
		Value: b.ref,
		// TODO(C): check writable
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	}
}

func NewValue(v Value) Value {
	b := &BaseValue{
		ref: v,
	}
	return b
}
