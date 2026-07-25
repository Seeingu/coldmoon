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

func (b *BaseValue) GetValue(agent *Agent) CompletionValue {
	if r, ok := b.Value.ReferenceRecord(); ok {
		return r.GetValue(agent)
	}
	return b.Value.ToCompletion()
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
	return ReturnAssertNormal(ToPropertyDescriptorCompletion(agent, b.Value))
}

func ToPropertyDescriptorCompletion(agent *Agent, value Value) (co Completion[*PropertyDescriptor]) {
	objectValue, ok := value.(*ObjectValue)
	if !ok {
		return co.ThrowTypeError(agent, "Property descriptor is not an object")
	}
	object := objectValue.Object

	desc := &PropertyDescriptor{}

	hasEnumerable, isAbrupt, rt := ReturnIfAbrupt(
		object.InternalMethods().HasProperty(object, NewStringPropertyKey("enumerable")),
		co,
	)
	if isAbrupt {
		return rt
	}
	if hasEnumerable {
		enumerableValue, isAbrupt, rt := ReturnIfAbrupt(
			object.InternalMethods().Get(object, NewStringPropertyKey("enumerable"), value),
			co,
		)
		if isAbrupt {
			return rt
		}
		enumerable := enumerableValue.ToBoolean()
		desc.Enumerable = enumerable
		desc.EnumerableSet = true
	}

	hasConfigurable, isAbrupt, rt := ReturnIfAbrupt(
		object.InternalMethods().HasProperty(object, NewStringPropertyKey("configurable")),
		co,
	)
	if isAbrupt {
		return rt
	}
	if hasConfigurable {
		configurableValue, isAbrupt, rt := ReturnIfAbrupt(
			object.InternalMethods().Get(object, NewStringPropertyKey("configurable"), value),
			co,
		)
		if isAbrupt {
			return rt
		}
		configurable := configurableValue.ToBoolean()
		desc.Configurable = configurable
		desc.ConfigurableSet = true
	}

	hasValue, isAbrupt, rt := ReturnIfAbrupt(
		object.InternalMethods().HasProperty(object, NewStringPropertyKey("value")),
		co,
	)
	if isAbrupt {
		return rt
	}
	if hasValue {
		desc.Value, isAbrupt, rt = ReturnIfAbrupt(
			object.InternalMethods().Get(object, NewStringPropertyKey("value"), value),
			co,
		)
		if isAbrupt {
			return rt
		}
	}

	hasWritable, isAbrupt, rt := ReturnIfAbrupt(
		object.InternalMethods().HasProperty(object, NewStringPropertyKey("writable")),
		co,
	)
	if isAbrupt {
		return rt
	}
	if hasWritable {
		writableValue, isAbrupt, rt := ReturnIfAbrupt(
			object.InternalMethods().Get(object, NewStringPropertyKey("writable"), value),
			co,
		)
		if isAbrupt {
			return rt
		}
		writable := writableValue.ToBoolean()
		desc.Writable = writable
		desc.WritableSet = true
	}

	hasGet, isAbrupt, rt := ReturnIfAbrupt(
		object.InternalMethods().HasProperty(object, NewStringPropertyKey("get")),
		co,
	)
	if isAbrupt {
		return rt
	}
	if hasGet {
		get, isAbrupt, rt := ReturnIfAbrupt(
			object.InternalMethods().Get(object, NewStringPropertyKey("get"), value),
			co,
		)
		if isAbrupt {
			return rt
		}
		if !IsCallable(get) && get != UndefinedValue {
			return co.ThrowTypeError(agent, "Property descriptor getter is not callable")
		}
		if get != UndefinedValue {
			desc.Get = MustGetObject(get)
		}
	}

	hasSet, isAbrupt, rt := ReturnIfAbrupt(
		object.InternalMethods().HasProperty(object, NewStringPropertyKey("set")),
		co,
	)
	if isAbrupt {
		return rt
	}
	if hasSet {
		set, isAbrupt, rt := ReturnIfAbrupt(
			object.InternalMethods().Get(object, NewStringPropertyKey("set"), value),
			co,
		)
		if isAbrupt {
			return rt
		}
		if !IsCallable(set) && set != UndefinedValue {
			return co.ThrowTypeError(agent, "Property descriptor setter is not callable")
		}
		if set != UndefinedValue {
			desc.Set = MustGetObject(set)
		}
	}

	if hasGet || hasSet {
		if hasValue || hasWritable {
			return co.ThrowTypeError(agent, "Invalid mixed data and accessor descriptor")
		}
	}

	co.value = desc
	return
}

// Call
// spec: 7.3.13
func (b *BaseValue) Call(agent *Agent, this Value, argumentsList ArgumentsList) (co CompletionValue) {
	if !IsCallable(b.Value) {
		return co.ThrowTypeError(agent, "value is not callable")
	}
	panic("internal error: call should not be called directly on BaseValue")
}

// ToPrimitive
// spec: 7.1.1
func (b *BaseValue) ToPrimitive(agent *Agent, hint PreferredType) (co CompletionValue) {
	value := b.Value
	if objectValue, isObject := value.(*ObjectValue); isObject {
		symbol := WellKnownSymbols[WellKnownSymbolsToPrimitive]
		exoticToPrim := GetMethod(agent, value, NewSymbolPropertyKey(symbol))
		if exoticToPrim != nil {
			hintString := hint.String()

			r := exoticToPrim.Call(value, []Value{
				NewStringValue(hintString),
			})
			result, isAbrupt, rt := ReturnIfAbrupt(r, co)
			if isAbrupt {
				return rt
			}
			if _, isObject = result.(*ObjectValue); !isObject {
				return result.ToCompletion()
			}

			return co.ThrowTypeError(agent, "could not convert object to primitive")
		}
		preferredType := hint
		if preferredType == PreferredTypeDefault {
			preferredType = PreferredTypeNumber
		}
		return objectValue.Object.OrdinaryToPrimitive(preferredType)
	}

	return value.ToCompletion()
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
		return co.ThrowTypeError(agent, "ToObject")
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

// ToNumber
// spec: 7.1.4
func (b *BaseValue) ToNumber(agent *Agent) (co Completion[*NumberValue]) {
	switch value := b.Value.(type) {
	case *NumberValue:
		co.value = value
		return
	case *undefinedValue:
		co.value = NaNValue
		return
	case *nullValue:
		co.value = NewNumberValue(0)
		return
	case *BooleanValue:
		if value.Data {
			co.value = NewNumberValue(1)
			return
		}
		co.value = NewNumberValue(0)
		return
	case *StringValue:
		co.value = StringToNumber(value)
		return
	case *ObjectValue:
		primValue, isAbrupt, rt := ReturnIfAbrupt(value.ToPrimitive(agent, PreferredTypeNumber), co)
		if isAbrupt {
			return rt
		}
		return primValue.ToNumber(agent)
	}
	return co.ThrowTypeError(agent, "TypeError")
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

// FIXME: will have recursive call
func NewBaseValue(v Value) Value {
	b := &BaseValue{
		Value: v,
	}
	return b
}
