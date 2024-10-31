package coldmoon

// MARK: - BooleanValue
type BooleanValue struct {
	Value
	Data bool
}

func NewBooleanValue(data bool) *BooleanValue {
	return &BooleanValue{
		Data: data,
	}
}

var _ Value = (*BooleanValue)(nil)

func (b *BooleanValue) String() string {
	if b.Data {
		return "true"
	} else {
		return "false"
	}
}

func (b *BooleanValue) ToBoolean() bool {
	return b.Data
}

type booleanType interface {
	getData() bool
}

// MARK: - BooleanObject
type BooleanObject struct {
	booleanType
	*Object
	Data bool
}

func (b *BooleanObject) getData() bool {
	return b.Data
}

func NewBooleanObject(agent *Agent, b bool) *BooleanObject {
	return &BooleanObject{
		Object: &Object{
			data: &Data{
				agent:     agent,
				prototype: agent.CurrentRealm().Intrinsics.BooleanPrototype,
			},
		},
		Data: b,
	}
}

type BooleanConstructor struct {
	booleanType
	*Object
	// [[BooleanData]]
	Data bool
}

func (b *BooleanConstructor) ToObject() *Object {
	return b.Object
}

func (b *BooleanConstructor) getData() bool {
	return b.Data
}

func NewBooleanConstructor(realm *Realm) ObjectType {
	// 20.3.1.1
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		value := argumentsList[0]
		b := value.ToBoolean()
		if newTarget == nil {
			return NewBooleanValue(b)
		}

		o := OrdinaryCreateFromConstructor(
			realm.Agent,
			newTarget,
			"%Boolean.prototype%",
			[]string{})
		booleanObject := &BooleanConstructor{
			Object: o,
			Data:   b,
		}
		return NewValueFromObject(booleanObject)
	}
	object := CreateBuiltinFunction(
		realm.Agent,
		behavior,
		1, "Boolean",
		builtinFunctionArgs{
			realm:         realm,
			prototype:     realm.Intrinsics.FunctionPrototype,
			isConstructor: true,
		})

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.BooleanPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinProperty(realm.Intrinsics.BooleanPrototype, "constructor", NewValueFromObject(object))

	return object
}

// MARK: - BooleanPrototype
type BooleanPrototype struct {
	booleanType
	*Object
	Data bool
}

func (b *BooleanPrototype) ToObject() *Object {
	return b.Object
}

func (b *BooleanPrototype) getData() bool {
	return b.Data
}

// 20.3.3
func NewBooleanPrototype(realm *Realm) *BooleanPrototype {
	object := &BooleanPrototype{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype),
		Data:   false,
	}
	var toString BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		b := thisBooleanValue(realm.Agent, thisArgument)
		if b {
			return NewStringValue("true")
		} else {
			return NewStringValue("false")
		}
	}
	var valueOf BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return NewBooleanValue(thisBooleanValue(realm.Agent, thisArgument))
	}

	DefineBuiltinProperty(object, "constructor",
		NewValueFromObject(realm.Intrinsics.BooleanConstructor))
	DefineBuiltinFunction(object, "toString", toString, 0, realm)
	DefineBuiltinFunction(object, "valueOf", valueOf, 0, realm)

	return object
}

func thisBooleanValue(agent *Agent, value Value) bool {
	switch o := value.(type) {
	case *BooleanValue:
		return value.ToBoolean()
	case *ObjectValue:
		if o, ok := o.Object.(booleanType); ok {
			b := o.getData()
			return b
		}
	}
	agent.ThrowException(TypeError, "Not a boolean")
	panic("")
}
