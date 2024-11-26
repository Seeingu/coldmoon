package coldmoon

// 9.1.1.2
type ObjectEnvironment struct {
	EnvironmentRecord
	BindingObject     ObjectType
	IsWithEnvironment bool
	outerEnv          EnvironmentRecord
}

var _ EnvironmentRecord = (*ObjectEnvironment)(nil)

func (o *ObjectEnvironment) OuterEnv() EnvironmentRecord {
	return o.outerEnv
}

// 9.1.1.2.1
func (o *ObjectEnvironment) HasBinding(name string) bool {
	bindingObject := o.BindingObject
	foundBinding := bindingObject.HasProperty(NewStringPropertyKey(name))
	if !foundBinding {
		return false
	}
	if !o.IsWithEnvironment {
		return true
	}

	symbol := WellKnownSymbols[WellKnownSymbolsUnscopables]
	unscopables := bindingObject.Get(
		NewSymbolPropertyKey(symbol),
	)

	if obj, ok := unscopables.(*ObjectValue); ok {
		blocked := obj.Object.Get(NewStringPropertyKey(name)).ToBoolean()
		if blocked {
			return false
		}
	}
	return true
}

// 9.1.1.2.2
func (o *ObjectEnvironment) CreateMutableBinding(name string, deletable bool) {
	o.BindingObject.DefinePropertyOrThrow(NewStringPropertyKey(name), &PropertyDescriptor{
		Value:        UndefinedValue,
		Writable:     true,
		Enumerable:   true,
		Configurable: deletable,
	})

}

// 9.1.1.2.4
func (o *ObjectEnvironment) InitializeBinding(name string, value Value) {
	o.SetMutableBinding(name, value, false)
}

// 9.1.1.2.5
func (o *ObjectEnvironment) SetMutableBinding(name string, value Value, strict bool) {
	bindingObject := o.BindingObject
	stillExists := bindingObject.HasProperty(NewStringPropertyKey(name))
	if !stillExists && strict {
		panic("ReferenceError")
	}

	var throw = setThrowTypeIgnore
	if strict {
		throw = setThrowTypeThrow
	}
	bindingObject.Set(NewStringPropertyKey(name), value, throw)

}

// 9.1.1.2.6
func (o *ObjectEnvironment) GetBindingValue(name string, strict bool) Value {
	bindingObject := o.BindingObject
	value := bindingObject.HasProperty(NewStringPropertyKey(name))
	if !value {
		if !strict {
			return UndefinedValue
		}
		panic("ReferenceError")
	}
	return bindingObject.Get(NewStringPropertyKey(name))
}

func (o *ObjectEnvironment) DeleteBinding(name string) bool {
	bindingObject := o.BindingObject
	return bindingObject.InternalMethods().Delete(bindingObject, NewStringPropertyKey(name))
}

// 9.1.1.2.8
func (o *ObjectEnvironment) HasThisBinding() bool {
	return false
}

func (o *ObjectEnvironment) HasSuperBinding() bool {
	return false
}

// 9.1.1.2.10
func (o *ObjectEnvironment) WithBaseObject() ObjectType {
	if o.IsWithEnvironment {
		return o.BindingObject
	}
	return nil
}

// 9.1.2.3
func NewObjectEnvironment(obj ObjectType, isWithEnvironment bool, outerEnv EnvironmentRecord) *ObjectEnvironment {
	return &ObjectEnvironment{
		BindingObject:     obj,
		IsWithEnvironment: isWithEnvironment,
		outerEnv:          outerEnv,
	}
}
