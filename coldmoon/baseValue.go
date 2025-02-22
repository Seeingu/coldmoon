package coldmoon

type BaseValue struct {
	Value
}

func (b *BaseValue) ToCompletion() (co CompletionValue) {
	co.value = b.Value
	return
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
	if r, ok := b.Value.ReferenceRecord(); ok {
		return r.GetValue(agent)
	}
	return b.Value
}

func (b *BaseValue) ToPropertyKey() PropertyKey {
	switch v := b.Value.(type) {
	case *StringValue:
		return NewStringPropertyKey(v.Data)
	case *SymbolValue:
		return NewSymbolPropertyKey(v)
	default:
		panic("unimplemented")
	}
}

// ToPropertyDescriptor
// spec: 6.2.6.5
func (b *BaseValue) ToPropertyDescriptor(agent *Agent) *PropertyDescriptor {
	value := b.Value
	if value == UndefinedValue {
		return nil
	}
	objectValue, ok := value.(*ObjectValue)
	if !ok {
		agent.ThrowTypeError("Value is not an object")
		return nil
	}
	object := objectValue.Object

	desc := &PropertyDescriptor{}

	hasEnumerable := object.HasProperty(NewStringPropertyKey("enumerable"))

	if hasEnumerable {
		enumerable := object.Get(NewStringPropertyKey("enumerable")).ToBoolean()
		desc.Enumerable = enumerable
	}

	hasConfigurable := object.HasProperty(NewStringPropertyKey("configurable"))
	if hasConfigurable {
		configurable := object.Get(NewStringPropertyKey("configurable")).ToBoolean()
		desc.Configurable = configurable
	}

	hasValue := object.HasProperty(NewStringPropertyKey("value"))
	if hasValue {
		desc.Value = object.Get(NewStringPropertyKey("value"))
	}

	hasWritable := object.HasProperty(NewStringPropertyKey("writable"))
	if hasWritable {
		writable := object.Get(NewStringPropertyKey("writable")).ToBoolean()
		desc.Writable = writable
	}

	hasGet := object.HasProperty(NewStringPropertyKey("get"))
	if hasGet {
		get := object.Get(NewStringPropertyKey("get"))
		if !IsCallable(get) && get != UndefinedValue {
			panic("TypeError")
		}
		desc.Get = MustGetObject(get)
	}

	hasSet := object.HasProperty(NewStringPropertyKey("set"))
	if hasSet {
		set := object.Get(NewStringPropertyKey("set"))
		if !IsCallable(set) && set != UndefinedValue {
			panic("TypeError")
		}
		desc.Set = MustGetObject(set)
	}

	if hasGet || hasSet {
		if hasValue || hasWritable {
			panic("TypeError")
		}
	}

	return desc
}

// ToPrimitive
// spec: 7.1.1
func (b *BaseValue) ToPrimitive(agent *Agent, hint PreferredType) Value {
	value := b.Value
	if objectValue, isObject := value.(*ObjectValue); isObject {
		symbol := WellKnownSymbols[WellKnownSymbolsToPrimitive]
		exoticToPrim := GetMethod(agent, value, NewSymbolPropertyKey(symbol))
		if exoticToPrim != nil {
			hintString := hint.String()

			result := exoticToPrim.Call(value, []Value{
				NewStringValue(hintString),
			}).value
			if _, isObject = result.(*ObjectValue); !isObject {
				return result
			}

			return agent.ThrowTypeError("could not convert object to primitive")
		}
		preferredType := hint
		if preferredType == PreferredTypeDefault {
			preferredType = PreferredTypeNumber
		}
		return objectValue.Object.OrdinaryToPrimitive(preferredType)
	}

	return value
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

// spec: 7.1.18
// TODO: type error handling
func (b *BaseValue) ToObject(agent *Agent) (co Completion[ObjectType]) {
	realm := agent.CurrentRealm()
	switch v := b.Value.(type) {
	case *undefinedValue, *nullValue:
		co.ThrowTypeError(agent, "ToObject")
	case *BooleanValue:
		co.value = NewBooleanObject(agent, v.Data, realm.Intrinsics.BooleanPrototype)
	case *ObjectValue:
		co.value = v.Object
	case *StringValue:
		co.value = NewStringObject(agent, v.Data, realm.Intrinsics.StringPrototype)
	case *NumberValue:
		co.value = NewNumberObject(agent, v.Data, realm.Intrinsics.NumberPrototype)
	case *SymbolValue:
		co.value = NewSymbolObject(agent, v, realm.Intrinsics.SymbolPrototype)
	case *BigIntValue:
		co.value = NewBigIntObject(agent, v, realm.Intrinsics.BigIntPrototype)
	default:
		panic("unimplemented")
	}
	return
}

func (b *BaseValue) ToBuiltinPropertyDescriptor() *PropertyDescriptor {
	return &PropertyDescriptor{
		Value: b.Value,
		// TODO: check writable
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	}
}

func (b *BaseValue) IsObject() bool {
	_, ok := b.GetObject()
	return ok
}

func (b *BaseValue) IsPromise() bool {
	objectValue, ok := b.GetObject()
	if !ok {
		return false
	}
	_, ok = objectValue.(*PromiseObject)
	return ok
}

func (b *BaseValue) GetObject() (object ObjectType, ok bool) {
	v, ok := ValueGet[*ObjectValue](b.Value)
	if ok {
		return v.Object, true
	}
	return nil, false
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
		primValue := value.ToPrimitive(agent, PreferredTypeNumber)
		return primValue.ToNumber(agent)
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

func (b *BaseValue) NumberOrBigInt() (num *NumberValue, bigInt *BigIntValue, ok bool) {
	switch v := b.Value.(type) {
	case *NumberValue:
		num = v
		ok = true
		return
	case *BigIntValue:
		bigInt = v
		ok = true
		return
	default:
		return
	}
}

func (b *BaseValue) ReferenceRecord() (*ReferenceRecord, bool) {
	if r, ok := b.Value.(*ReferenceRecordValue); ok {
		return r.record, true
	}
	return nil, false
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
