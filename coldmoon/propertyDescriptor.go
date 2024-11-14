package coldmoon

type PropertyDescriptor struct {
	Value Value
	// TODO: use nullable
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

// 6.2.6.2
func (p *PropertyDescriptor) IsDataDescriptor() bool {
	return p.Value != nil || p.Writable
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

	if p.Value != nil {
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("value"), p.Value)
	}

	// TODO: Writable nil check
	if p.Writable {
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("writable"), NewBooleanValue(p.Writable))
	}

	if p.Get != nil {
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("get"), NewValueFromObject(p.Get))
	}

	if p.Set != nil {
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("set"), NewValueFromObject(p.Set))
	}

	if p.Enumerable {
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("enumerable"), NewBooleanValue(p.Enumerable))
	}

	if p.Configurable {
		obj.CreateDataPropertyOrThrow(NewStringPropertyKey("configurable"), NewBooleanValue(p.Configurable))
	}

	return obj
}

func (p *PropertyDescriptor) IsFullyPopulated() bool {
	return p.Value != nil && p.Get != EmptyObject && p.Set != EmptyObject
}

func (p *PropertyDescriptor) HasFields() bool {
	return p.Value != nil || p.Writable || p.Get != EmptyObject || p.Set != EmptyObject || p.Enumerable || p.Configurable
}
