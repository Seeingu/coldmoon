package coldmoon

// MARK: - BooleanValue

type BooleanValue struct {
	Value
	Data bool
}

var (
	FalseValue = &BooleanValue{Data: false}
	TrueValue  = &BooleanValue{Data: true}
)

func NewBooleanValue(data bool) *BooleanValue {
	if data {
		return TrueValue
	} else {
		return FalseValue
	}
}

var _ Value = (*BooleanValue)(nil)

func (b *BooleanValue) ToCompletion() CompletionValue {
	return NewCompletionValue(b)
}

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

// MARK: - BooleanObject

type BooleanObject struct {
	*Object
	Data bool
}

func (b *BooleanObject) getData() bool {
	return b.Data
}

func NewBooleanObject(agent *Agent, b bool, prototype ObjectType) *BooleanObject {
	return &BooleanObject{
		Object: NewObject(agent, prototype, "Boolean"),
		Data:   b,
	}
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
		booleanObject := &BooleanObject{
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

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.BooleanPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinPropertyV(realm.Intrinsics.BooleanPrototype, "constructor", NewValueFromObject(object))

	return object
}

// MARK: - BooleanPrototype

// 20.3.3
func NewBooleanPrototype(realm *Realm) *BooleanObject {
	object := &BooleanObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "BooleanPrototype"),
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

	DefineBuiltinPropertyV(object, "constructor",
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
		if o, ok := o.Object.(*BooleanObject); ok {
			b := o.getData()
			return b
		}
	}
	agent.ThrowException(TypeError, "Not a boolean")
	panic("")
}
