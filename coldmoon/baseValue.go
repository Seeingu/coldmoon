package coldmoon

type BaseValue struct {
	Value
}

func (b *BaseValue) ToCompletion() CompletionValue {
	return NewCompletionValue(b.Value)
}

func (b *BaseValue) Hash() string {
	switch v := b.Value.(type) {
	case *undefinedValue, *nullValue, *BooleanValue, *StringValue, *SymbolValue, *BigIntValue:
		return v.String()
	case *NumberValue:
		return v.String()
	default:
		panic("unimplemented")
	}
}

func (b *BaseValue) GetValue(agent *Agent) Value {
	if r, ok := b.Value.(*ReferenceRecordValue); ok {
		return r.ReferenceRecord.GetValue(agent)
	}
	return b.Value
}

// TODO(BM): return a string completion or abrupt completion
func (b *BaseValue) ThisStringValue() string {
	switch v := b.Value.(type) {
	case *StringValue:
		return v.Data
	case *ObjectValue:
		s, ok := v.Object.(*StringObject)
		if ok {
			return s.Data
		}
	}
	panic("TypeError")
}

// TODO(I): this is a backdoor method, every value should implement this method
func (b *BaseValue) ToString() CMString {
	return CMString(b.String())
}

// TODO: type error handling
func (b *BaseValue) ToObject(agent *Agent) ObjectType {
	realm := agent.CurrentRealm()
	switch v := b.Value.(type) {
	case *undefinedValue, *nullValue:
		agent.ThrowTypeError("TypeError")
	case *BooleanValue:
		return NewBooleanObject(agent, v.Data, realm.Intrinsics.BooleanPrototype)
	case *ObjectValue:
		return v.Object
	case *StringValue:
		return NewStringObject(agent, v.Data, realm.Intrinsics.StringPrototype)
	case *NumberValue:
		return NewNumberObject(agent, v.Data, realm.Intrinsics.NumberPrototype)
	case *SymbolValue:
		return NewSymbolObject(agent, v, realm.Intrinsics.SymbolPrototype)
	case *BigIntValue:
		return NewBigIntObject(agent, v, realm.Intrinsics.BigIntPrototype)
	default:
		panic("unimplemented")
	}
	panic("unreachable")
}

func (b *BaseValue) ToPropertyDescriptor() *PropertyDescriptor {
	return &PropertyDescriptor{
		Value: b.Value,
		// TODO(C): check writable
		Writable:     true,
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

func (b *BaseValue) TypeString() string {
	switch v := b.Value.(type) {
	case *undefinedValue:
		return "undefined"
	case *nullValue:
		return "object"
	case *BooleanValue:
		return "boolean"
	case *NumberValue:
		return "number"
	case *StringValue:
		return "string"
	case *SymbolValue:
		return "symbol"
	case *BigIntValue:
		return "bigint"
	case *ObjectValue:
		if v.Object.InternalMethods().Call != nil {
			return "function"
		} else {
			return "object"
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
