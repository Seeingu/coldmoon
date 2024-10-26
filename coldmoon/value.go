package coldmoon

import "math"

type PreferredType int

const (
	PreferredTypeString PreferredType = iota
	PreferredTypeNumber
	PreferredTypeDefault
)

func (hint PreferredType) String() string {
	switch hint {
	case PreferredTypeString:
		return "string"
	case PreferredTypeNumber:
		return "number"
	default:
		return "default"
	}
}

type Value interface {
	String() string
	ToBoolean() bool
}

type undefinedValue struct {
	Value
}

var _ Value = (*undefinedValue)(nil)

func (u *undefinedValue) String() string {
	return "undefined"
}
func (u *undefinedValue) ToBoolean() bool {
	return false
}

var UndefinedValue = &undefinedValue{}

type nullValue struct {
	Value
}

var _ Value = (*nullValue)(nil)

func (n *nullValue) String() string {
	return "null"
}

func (n *nullValue) ToBoolean() bool {
	return false
}

type StringValue struct {
	Value
	Data string
}

var _ Value = (*StringValue)(nil)

func (s *StringValue) String() string {
	return s.Data
}

func (s *StringValue) ToBoolean() bool {
	if len(s.Data) == 0 {
		return false
	}
	return true
}

func NewStringValue(value string) *StringValue {
	return &StringValue{Data: value}
}

var NullValue = nullValue{}

var NaNValue = NumberValue{Data: math.NaN()}

var InfinityValue = NumberValue{Data: math.Inf(1)}
var NegativeInfinityValue = NumberValue{Data: math.Inf(-1)}

type ObjectValue struct {
	Value
	Object *Object
}

func (o ObjectValue) String() string {
	primValue := ToPrimitive(o, PreferredTypeString)
	if _, isObject := primValue.(*ObjectValue); isObject {
		panic("")
	}
	return primValue.String()
}

func NewValueFromObject(object *Object) Value {
	return ObjectValue{Object: object}
}

// 7.1.1
func ToPrimitive(value Value, hint PreferredType) Value {
	if objectValue, isObject := value.(*ObjectValue); isObject {
		// TODO:
		exoticToPrim := UndefinedValue
		if exoticToPrim != UndefinedValue {
			hintString := hint.String()

			result := Call(exoticToPrim, value, []Value{
				NewStringValue(hintString),
			})
			if _, isObject = result.(*ObjectValue); !isObject {
				return result
			}
			panic("TypeError")
		}
		preferredType := hint
		if preferredType == PreferredTypeDefault {
			preferredType = PreferredTypeNumber
		}
		return objectValue.Object.OrdinaryToPrimitive(preferredType)
	}

	return value
}

// 7.2.2
func isArray(value Value) bool {
	return false
}

// 7.2.3
func isCallable(value Value) bool {
	objectValue, isObject := value.(*ObjectValue)

	if !isObject {
		return false
	}
	if objectValue.Object.InternalMethods().Call != nil {
		return true
	}

	return false
}

// 7.2.4
func isConstructor(value Value) bool {
	objectValue, isObject := value.(*ObjectValue)

	if !isObject {
		return false
	}
	if objectValue.Object.InternalMethods().Construct != nil {
		return true
	}

	return false
}

// 7.3.14
func Call(self Value, value Value, argumentsList []Value) Value {
	if !isCallable(value) {
		panic("TypeError")
	}

	return value.(*ObjectValue).Object.InternalMethods().Call(value.(*ObjectValue).Object, self, argumentsList)
}

func CallNoArgs(self Value, value Value) Value {
	return Call(self, value, nil)
}

func CallAssumeCallable(self Value, value Value, argumentsList []Value) Value {
	return self.(*ObjectValue).Object.InternalMethods().Call(self.(*ObjectValue).Object, self, argumentsList)
}

func CallAssumeCallableNoArgs(self, value Value) Value {
	return CallAssumeCallable(self, self, nil)
}
