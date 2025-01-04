package coldmoon

type BaseValue struct {
	Value
}

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

func (b *BaseValue) ToNumber(agent *Agent) *NumberValue {
	switch value := b.Value.(type) {
	case *NumberValue:
		return value
	case *undefinedValue:
		return InfinityValue
	case *nullValue:
		return NewNumberValue(0)
	case *BooleanValue:
		if value.Data {
			return NewNumberValue(1)
		}
		return NewNumberValue(0)
	case *StringValue:
		return StringToNumber(value)
	case *ObjectValue:
		primValue := ToPrimitive(agent, value, PreferredTypeNumber)

		if _, ok := primValue.(*ObjectValue); !ok {
			Assert(false)
		}

		return ToNumber(agent, primValue)
	}
	agent.ThrowTypeError("TypeError")
	return nil
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

func (b *BaseValue) ToBoolean() bool {
	switch v := b.Value.(type) {
	case *BooleanValue:
		return v.Data
	case *NumberValue:
		return v.Data != 0 && !v.Data.IsNaN()
	case *StringValue:
		return v.Data != ""
	case *SymbolValue:
		return true
	case *BigIntValue:
		return v.Data.Cmp(bigZero) != 0
	case *ObjectValue:
		return true
	case *nullValue:
		return false
	case *undefinedValue:
		return false
	default:
		panic("unreachable")
	}
}

func (b *BaseValue) String() string {
	return b.Value.String()
}

func NewBaseValue(v Value) Value {
	b := &BaseValue{
		Value: v,
	}
	return b
}
