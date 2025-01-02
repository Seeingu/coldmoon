package coldmoon

type BaseValue struct {
	Value
}

var _ Value = (*BaseValue)(nil)

func (b *BaseValue) ToCompletion() CompletionValue {
	return NewCompletionValue(b.Value)
}

func (b *BaseValue) ToPropertyDescriptor() *PropertyDescriptor {
	return &PropertyDescriptor{
		Value: b.Value,
		// TODO(C): check writable
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	}
}

func (b *BaseValue) TypeString() *StringValue {
	switch v := b.Value.(type) {
	case *undefinedValue:
		return NewStringValue("undefined")
	case *nullValue:
		return NewStringValue("object")
	case *BooleanValue:
		return NewStringValue("boolean")
	case *NumberValue:
		return NewStringValue("number")
	case *StringValue:
		return NewStringValue("string")
	case *SymbolValue:
		return NewStringValue("symbol")
	case *BigIntValue:
		return NewStringValue("bigint")
	case *ObjectValue:
		if v.Object.InternalMethods().Call != nil {
			return NewStringValue("function")
		} else {
			return NewStringValue("object")
		}
	default:
		panic("unreachable")
	}
}

func NewValue(v Value) Value {
	b := &BaseValue{
		Value: v,
	}
	return b
}
