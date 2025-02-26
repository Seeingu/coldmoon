package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

// MARK: - BooleanValue

// SlotBooleanData [[BooleanData]]
const SlotBooleanData = "BooleanData"

type BooleanValue struct {
	Value
	Data bool
}

var (
	FalseValue = &BooleanValue{
		Data: false,
	}
	TrueValue = &BooleanValue{Data: true}
)

func NewBooleanValue(data bool) Value {
	if data {
		return TrueValue
	} else {
		return FalseValue
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

// MARK: - BooleanObject

type BooleanObject struct {
	*Object
	Data bool
}

func (b *BooleanObject) getData() bool {
	return b.Data
}

func NewBooleanObject(agent *Agent, b bool, prototype ObjectType) *BooleanObject {
	o := &BooleanObject{
		Object: NewObject(agent, prototype, "Boolean"),
		Data:   b,
	}
	o.ref = o
	return o
}

func NewBooleanConstructor(realm *Realm) ObjectType {
	// 20.3.1.1
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		value := pkg.SliceSafeGet(argumentsList, 0)
		var b bool
		if value == nil {
			b = false
		} else {
			b = value.ToBoolean()
		}
		if newTarget == nil {
			return NewBooleanValue(b)
		}

		o := OrdinaryCreateFromConstructor(
			realm.Agent,
			newTarget,
			"%Boolean.prototype%",
			[]string{SlotBooleanData})
		booleanObject := &BooleanObject{
			Object: o,
			Data:   b,
		}
		booleanObject.SetSlot(SlotBooleanData, b)
		return booleanObject.ToValue()
	}
	object := CreateBuiltinFunction(
		realm.Agent,
		behavior,
		1, CMString("Boolean"),
		builtinFunctionArgs{
			realm:         realm,
			prototype:     realm.Intrinsics.FunctionPrototype,
			isConstructor: true,
		})

	BindPrototypeAndConstructor(realm.Intrinsics.BooleanPrototype, object)

	return object
}

// MARK: - BooleanPrototype

// 20.3.3
func NewBooleanPrototype(realm *Realm) *BooleanObject {
	object := &BooleanObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "BooleanPrototype"),
		Data:   false,
	}
	object.ref = object
	agent := realm.Agent
	var toString BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		b := ThisBooleanValue(agent, thisArgument)
		if b {
			return NewStringValue("true")
		} else {
			return NewStringValue("false")
		}
	}
	var valueOf BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return NewBooleanValue(ThisBooleanValue(realm.Agent, thisArgument))
	}

	object.defineBuiltinFunction(realm, CMString("toString"), toString, 0)
	object.defineBuiltinFunction(realm, CMString("valueOf"), valueOf, 0)

	return object
}

// ThisBooleanValue
// spec: 20.3.3.3.1
func ThisBooleanValue(agent *Agent, value Value) bool {
	switch o := value.(type) {
	case *BooleanValue:
		return value.ToBoolean()
	case *ObjectValue:
		if b, ok := o.Object.GetSlot(SlotBooleanData); ok {
			return b.(bool)
		}
		if o, ok := o.Object.(*BooleanObject); ok {
			b := o.getData()
			return b
		}
	}
	agent.ThrowException(TypeError, "Not a boolean")
	panic("")
}
