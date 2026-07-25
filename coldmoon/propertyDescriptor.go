package coldmoon

type PropertyDescriptor struct {
	Value           Value
	Writable        bool
	WritableSet     bool
	Get             ObjectType
	GetSet          bool
	Set             ObjectType
	SetSet          bool
	Enumerable      bool
	EnumerableSet   bool
	Configurable    bool
	ConfigurableSet bool
}

func NewFrozenPropertyDescriptor(value Value) *PropertyDescriptor {
	return &PropertyDescriptor{
		Value:           value,
		Writable:        false,
		WritableSet:     true,
		Enumerable:      false,
		EnumerableSet:   true,
		Configurable:    false,
		ConfigurableSet: true,
	}
}

type PropertyDescriptorAttributes struct {
	Writable     bool
	Enumerable   bool
	Configurable bool
}

// 6.2.6.1
func (p *PropertyDescriptor) IsAccessorDescriptor() bool {
	return p.GetSet || p.Get != nil || p.SetSet || p.Set != nil
}

// 6.2.6.2
func (p *PropertyDescriptor) IsDataDescriptor() bool {
	return p.Value != nil || p.WritableSet || p.Writable
}

// 6.2.6.3
func (p *PropertyDescriptor) IsGenericDescriptor() bool {
	return !p.IsAccessorDescriptor() && !p.IsDataDescriptor()
}

// 6.2.6.4
func (p *PropertyDescriptor) FromPropertyDescriptor(agent *Agent, desc *PropertyDescriptor) ObjectType {
	realm := agent.CurrentRealm()

	if desc == nil {
		return nil
	}

	obj := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, []string{})
	Assert(obj.IsExtensible())

	if desc.IsDataDescriptor() {
		value := desc.Value
		if value == nil {
			value = UndefinedValue
		}
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("value"), value)
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("writable"), NewBooleanValue(desc.Writable))
	} else if desc.IsAccessorDescriptor() {
		getter := Value(UndefinedValue)
		if desc.Get != nil {
			getter = desc.Get.ToValue()
		}
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("get"), getter)

		setter := Value(UndefinedValue)
		if desc.Set != nil {
			setter = desc.Set.ToValue()
		}
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("set"), setter)
	}

	obj.CreateDataPropertyOrThrow(NewStringPropertyKey("enumerable"), NewBooleanValue(p.Enumerable))

	obj.CreateDataPropertyOrThrow(NewStringPropertyKey("configurable"), NewBooleanValue(p.Configurable))

	return obj
}

// 6.2.6.6
func (p *PropertyDescriptor) CompletePropertyDescriptor() {
	like := &PropertyDescriptor{
		Value: UndefinedValue,
	}
	if p.IsGenericDescriptor() || p.IsDataDescriptor() {
		if p.Value == nil {
			p.Value = like.Value
		}
		if !p.WritableSet && !p.Writable {
			p.Writable = like.Writable
		}
		p.WritableSet = true
	} else {
		p.GetSet = true
		p.SetSet = true
	}

	if p.Enumerable == false {
		p.Enumerable = like.Enumerable
	}
	p.EnumerableSet = true
	if p.Configurable == false {
		p.Configurable = like.Configurable
	}
	p.ConfigurableSet = true
}

func (p *PropertyDescriptor) IsFullyPopulated() bool {
	if p.IsAccessorDescriptor() {
		return p.GetSet && p.SetSet && p.EnumerableSet && p.ConfigurableSet
	}
	return p.Value != nil && p.WritableSet && p.EnumerableSet && p.ConfigurableSet
}

func (p *PropertyDescriptor) HasFields() bool {
	return p.Value != nil || p.WritableSet || p.Writable || p.GetSet || p.Get != nil || p.SetSet || p.Set != nil ||
		p.EnumerableSet || p.Enumerable || p.ConfigurableSet || p.Configurable
}
