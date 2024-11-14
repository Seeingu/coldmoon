package coldmoon

import (
	"fmt"
	"math"
	"strconv"
)

type NumberValue struct {
	Value
	Data float64
}

var _ Value = (*NumberValue)(nil)

func NewNumberValue(v float64) *NumberValue {
	return &NumberValue{
		Data: v,
	}
}

var _ Value = (*NumberValue)(nil)

func (n *NumberValue) String() string {
	return fmt.Sprintf("%f", n.Data)
}

func (n *NumberValue) ToString(radix float64) string {
	if math.IsNaN(n.Data) {
		return "NaN"
	}
	if n.IsPositiveInf() {
		return "Infinity"
	}
	if n.IsNegativeInf() {
		return "-Infinity"
	}
	if n.IsZero() {
		if math.Signbit(n.Data) {
			return "-0"
		}
		return "0"
	}
	return strconv.FormatFloat(n.Data, 'f', -1, 64)

}

func (n *NumberValue) ToBoolean() bool {
	if n.Data == 0 || math.IsNaN(n.Data) {
		return false
	}
	return true
}

func (n *NumberValue) IsNaN() bool {
	return math.IsNaN(n.Data)
}

func (n *NumberValue) IsPositiveInf() bool {
	return math.IsInf(n.Data, 1)
}

func (n *NumberValue) IsNegativeInf() bool {
	return math.IsInf(n.Data, -1)
}

func (n *NumberValue) IsPositiveZero() bool {
	return n.Data == 0 && math.Signbit(n.Data)
}

func (n *NumberValue) IsNegativeZero() bool {
	return n.Data == 0 && !math.Signbit(n.Data)
}

func (n *NumberValue) IsFinite() bool {
	return !math.IsInf(n.Data, 0)
}

func (n *NumberValue) Truncate() float64 {
	return n.Data
}

func (n *NumberValue) Round() float64 {
	return math.Round(n.Data)
}

func (n *NumberValue) Ceil() float64 {
	return math.Ceil(n.Data)
}

func (n *NumberValue) Floor() float64 {
	return math.Floor(n.Data)
}

// 6.1.6.1.1
func (n *NumberValue) UnaryMinus() *NumberValue {
	return &NumberValue{
		Data: -n.Data,
	}
}

// 6.1.6.1.2
func (n *NumberValue) BitwiseNot() *NumberValue {
	return &NumberValue{
		Data: float64(^int64(n.Data)),
	}
}

// 6.1.6.1.3
func (n *NumberValue) Exponentiate(exponent *NumberValue) *NumberValue {
	if exponent.IsNaN() {
		return NewNumberValue(math.NaN())
	}
	if exponent.IsZero() {
		return NewNumberValue(1)
	}
	if n.IsPositiveInf() {
		if exponent.Data > 0 {
			return NewNumberValue(math.Inf(1))
		} else {
			return NewNumberValue(0)
		}
	}

	if n.IsNegativeInf() {
		if exponent.Data > 0 {
			if exponent.Data > 0 && math.Mod(exponent.Data, 2) == 0 {
				return NewNumberValue(math.Inf(1))
			} else {
				return NewNumberValue(math.Inf(-1))
			}
		} else {
			if math.Mod(exponent.Data, 2) == 0 {
				return NewNumberValue(0)
			} else {
				return NewNumberValue(-0)
			}
		}
	}

	if n.IsPositiveZero() {
		if exponent.Data > 0 {
			return NewNumberValue(0)
		} else {
			return NewNumberValue(math.Inf(1))
		}
	}

	if n.IsNegativeZero() {
		if exponent.Data > 0 {
			if math.Mod(exponent.Data, 2) == 0 {
				return NewNumberValue(0)
			} else {
				return NewNumberValue(-0)
			}
		} else {
			return NewNumberValue(math.Inf(1))
		}
	}

	Assert(n.IsFinite() && !n.IsZero())
	if exponent.IsPositiveInf() {
		if math.Abs(n.Data) == 1 {
			return NewNumberValue(math.NaN())
		}
		if math.Abs(n.Data) > 1 {
			return NewNumberValue(math.Inf(1))
		}
		return NewNumberValue(0)
	}
	if exponent.IsNegativeInf() {
		if math.Abs(n.Data) == 1 {
			return NewNumberValue(math.NaN())
		}
		if math.Abs(n.Data) > 1 {
			return NewNumberValue(0)
		}
		return NewNumberValue(math.Inf(1))
	}

	Assert(exponent.IsFinite() && !exponent.IsZero())

	return NewNumberValue(math.Pow(n.Data, exponent.Data))
}

// 6.1.6.1.4
func (n *NumberValue) Multiply(other *NumberValue) *NumberValue {
	return NewNumberValue(n.Data * other.Data)
}

// 6.1.6.1.5
func (n *NumberValue) Divide(other *NumberValue) *NumberValue {
	return NewNumberValue(n.Data / other.Data)
}

// 6.1.6.1.6
func (n *NumberValue) Remainder(other *NumberValue) *NumberValue {
	return NewNumberValue(math.Mod(n.Data, other.Data))
}

// 6.1.6.1.7
func (n *NumberValue) Add(other *NumberValue) *NumberValue {
	return NewNumberValue(n.Data + other.Data)
}

// 6.1.6.1.8
func (n *NumberValue) Subtract(other *NumberValue) *NumberValue {
	return NewNumberValue(n.Data - other.Data)
}

// 6.1.6.1.9
func (n *NumberValue) LeftShift(other *NumberValue) *NumberValue {
	return NewNumberValue(float64(int64(n.Data) << uint64(int64(other.Data))))
}

// 6.1.6.1.10
func (n *NumberValue) SignedRightShift(other *NumberValue) *NumberValue {
	return NewNumberValue(float64(int64(n.Data) >> uint64(int64(other.Data))))
}

// 6.1.6.1.11
func (n *NumberValue) UnsignedRightShift(other *NumberValue) *NumberValue {
	return NewNumberValue(float64(uint64(n.Data) >> uint64(int64(other.Data))))
}

// 6.1.6.1.12
func (n *NumberValue) LessThan(other NumberValue) bool {
	if n.IsNaN() || other.IsNaN() {
		return false
	}
	return n.Data < other.Data
}

// 6.1.6.1.14
func (n *NumberValue) SameValue(other *NumberValue) bool {
	if n.IsNaN() && other.IsNaN() {
		return true
	}
	if n.IsPositiveZero() && other.IsNegativeZero() {
		return false
	}
	if n.IsNegativeZero() && other.IsPositiveZero() {
		return false
	}
	return n.Data == other.Data
}

// 6.1.6.1.15
func (n *NumberValue) SameValueZero(other *NumberValue) bool {
	if n.IsNaN() && other.IsNaN() {
		return true
	}
	return n.Data == other.Data
}

// 6.1.6.1.16
type numberBitwiseOp int

const (
	numberBitwiseAnd numberBitwiseOp = iota
	numberBitwiseXor
	numberBitwiseOr
)

func (n *NumberValue) NumberBitwiseOp(op numberBitwiseOp, other *NumberValue) *NumberValue {
	switch op {
	case numberBitwiseAnd:
		return NewNumberValue(float64(int64(n.Data) & int64(other.Data)))
	case numberBitwiseOr:
		return NewNumberValue(float64(int64(n.Data) | int64(other.Data)))
	case numberBitwiseXor:
		return NewNumberValue(float64(int64(n.Data) ^ int64(other.Data)))
	}
	panic("unreachable")
}

// 6.1.6.1.17
func (n *NumberValue) BitwiseAnd(other *NumberValue) *NumberValue {
	return n.NumberBitwiseOp(numberBitwiseAnd, other)
}

// 6.1.6.1.18
func (n *NumberValue) BitwiseXor(other *NumberValue) *NumberValue {
	return n.NumberBitwiseOp(numberBitwiseXor, other)
}

// 6.1.6.1.19
func (n *NumberValue) BitwiseOr(other *NumberValue) *NumberValue {
	return n.NumberBitwiseOp(numberBitwiseOr, other)
}

func (n *NumberValue) IsZero() bool {
	return n.Data == 0
}

// MARK: - Number Object

func NewNumberConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		value := argumentsList[0]
		n := NewNumberValue(0)
		if len(argumentsList) > 0 {
			prim := ToNumeric(agent, value)

			bigintPrim, isBigInt := prim.(*BigIntValue)
			if isBigInt {
				data, _ := strconv.ParseFloat(bigintPrim.String(), 64)
				n.Data = data
			} else {
				n = ToNumber(agent, prim)
			}
		}

		if newTarget == nil {
			return n
		}

		object := OrdinaryCreateFromConstructor(
			agent,
			newTarget,
			"%Number.prototype", nil)
		numberObject := &NumberObject{
			Object: object,
			Data:   n.Data,
		}
		return NewValueFromObject(numberObject)

	}
	object := CreateBuiltinFunction(realm.Agent, behavior, 1, "Number", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})

	var isFinite BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		value := arguments[0]
		numberValue, ok := value.(*NumberValue)
		if !ok {
			return NewBooleanValue(false)
		}
		if !numberValue.IsFinite() {
			return NewBooleanValue(false)
		}
		return NewBooleanValue(true)
	}
	var isInteger BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		value := arguments[0]
		return NewBooleanValue(IsIntegralNumber(value))
	}
	var isNaN BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		value := arguments[0]
		numberValue, ok := value.(*NumberValue)
		if !ok {
			return NewBooleanValue(false)
		}
		return NewBooleanValue(numberValue.IsNaN())
	}

	DefineBuiltinFunction(object, "isFinite", isFinite, 1, realm)
	DefineBuiltinFunction(object, "isInteger", isInteger, 1, realm)
	DefineBuiltinFunction(object, "isNaN", isNaN, 1, realm)
	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.NumberPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(realm.Intrinsics.NumberPrototype, "constructor", NewValueFromObject(object))

	return object
}

type NumberObject struct {
	*Object
	Data float64
}

func NewNumberObject(agent *Agent, value float64, prototype ObjectType) *NumberObject {
	object := &NumberObject{
		Object: NewObject(agent, prototype),
		Data:   value,
	}
	return object
}

func NewNumberPrototype(realm *Realm) *NumberObject {
	object := &NumberObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype),
	}

	agent := realm.Agent
	var toString BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		radix := arguments[0]

		x := thisNumberValue(agent, this)
		var radixMV float64 = 10
		if radix != nil {
			radixMV = ToIntegerOrInfinity(agent, radix)
		}

		if radixMV < 2 || radixMV > 36 {
			panic("RangeError")
		}
		return NewStringValue(x.ToString(radixMV))
	}

	var valueOf BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		x := thisNumberValue(agent, this)
		return x
	}

	DefineBuiltinFunction(object, "toString", toString, 0, realm)
	DefineBuiltinFunction(object, "valueOf", valueOf, 0, realm)

	return object
}

func thisNumberValue(agent *Agent, this Value) *NumberValue {
	switch v := this.(type) {
	case *NumberValue:
		return v
	case *ObjectValue:
		numberObject, ok := v.Object.(*NumberObject)
		if ok {
			return NewNumberValue(numberObject.Data)
		}
	}
	panic("TypeError")
}
