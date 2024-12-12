package coldmoon

import (
	"fmt"
	"math"
	"strconv"
)

// MARK: - JSNumber
type JSNumber float64

func (n JSNumber) ToInt() JSInt {
	return JSInt(n)
}

func (n JSNumber) IsNegInf() bool {
	return n == JSNumber(math.Inf(-1))
}

func (n JSNumber) IsPositiveInf() bool {
	return n == JSNumber(math.Inf(1))
}

func (n JSNumber) IsInf() bool {
	return n == JSNumber(math.Inf(0))
}

func (n JSNumber) IsNaN() bool {
	return math.IsNaN(float64(n))
}

func (n JSNumber) Max(b JSNumber) JSNumber {
	return JSNumber(math.Max(float64(n), float64(b)))
}

func (n JSNumber) Min(b JSNumber) JSNumber {
	return JSNumber(math.Min(float64(n), float64(b)))
}

func (n JSNumber) ToFloat() float64 {
	return float64(n)
}

func (n JSNumber) Mod(b JSNumber) JSNumber {
	return JSNumber(math.Mod(float64(n), float64(b)))
}

func (n JSNumber) Floor() JSNumber {
	return JSNumber(math.Floor(float64(n)))
}

var (
	JSNumberInf    = JSNumber(math.Inf(1))
	JSNumberNegInf = JSNumber(math.Inf(-1))
	JSNumberNaN    = JSNumber(math.NaN())
)

// MARK: - NumberValue

type NumberValue struct {
	Value
	Data JSNumber
}

var _ Value = (*NumberValue)(nil)

func NewNumberValue(v JSNumber) *NumberValue {
	return &NumberValue{
		Data: v,
	}
}

func (n *NumberValue) ToCompletion() CompletionValue {
	return NewCompletionValue(n)
}

func (n *NumberValue) String() string {
	return fmt.Sprintf("%f", n.Data)
}

func (n *NumberValue) ToString(radix JSInt) string {
	if math.IsNaN(n.Data.ToFloat()) {
		return "NaN"
	}
	if n.IsPositiveInf() {
		return "Infinity"
	}
	if n.IsNegativeInf() {
		return "-Infinity"
	}
	if n.IsZero() {
		if math.Signbit(n.Data.ToFloat()) {
			return "-0"
		}
		return "0"
	}
	return strconv.FormatFloat(n.Data.ToFloat(), 'f', -1, 64)
}

func (n *NumberValue) ToBoolean() bool {
	if n.Data == 0 || math.IsNaN(n.Data.ToFloat()) {
		return false
	}
	return true
}

func (n *NumberValue) IsNaN() bool {
	return math.IsNaN(n.Data.ToFloat())
}

func (n *NumberValue) IsPositiveInf() bool {
	return math.IsInf(n.Data.ToFloat(), 1)
}

func (n *NumberValue) IsNegativeInf() bool {
	return math.IsInf(float64(n.Data), -1)
}

func (n *NumberValue) IsPositiveZero() bool {
	return n.Data == 0 && math.Signbit(n.Data.ToFloat())
}

func (n *NumberValue) IsNegativeZero() bool {
	return n.Data == 0 && !math.Signbit(n.Data.ToFloat())
}

func (n *NumberValue) IsFinite() bool {
	return !math.IsInf(n.Data.ToFloat(), 0)
}

func (n *NumberValue) Truncate() JSNumber {
	return n.Data
}

func (n *NumberValue) Round() JSNumber {
	return JSNumber(math.Round(n.Data.ToFloat()))
}

func (n *NumberValue) Ceil() JSNumber {
	return JSNumber(math.Ceil(n.Data.ToFloat()))
}

func (n *NumberValue) Floor() JSNumber {
	return JSNumber(math.Floor(n.Data.ToFloat()))
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
		Data: JSNumber(^int64(n.Data)),
	}
}

// 6.1.6.1.3
func (n *NumberValue) Exponentiate(exponent *NumberValue) *NumberValue {
	if exponent.IsNaN() {
		return NewNumberValue(JSNumber(math.NaN()))
	}
	if exponent.IsZero() {
		return NewNumberValue(1)
	}
	if n.IsPositiveInf() {
		if exponent.Data > 0 {
			return NewNumberValue(JSNumber(math.Inf(1)))
		} else {
			return NewNumberValue(0)
		}
	}

	if n.IsNegativeInf() {
		if exponent.Data > 0 {
			if exponent.Data > 0 && math.Mod(exponent.Data.ToFloat(), 2) == 0 {
				return InfinityValue
			} else {
				return NegativeInfinityValue
			}
		} else {
			if math.Mod(exponent.Data.ToFloat(), 2) == 0 {
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
			return NewNumberValue(JSNumberInf)
		}
	}

	if n.IsNegativeZero() {
		if exponent.Data > 0 {
			if math.Mod(exponent.Data.ToFloat(), 2) == 0 {
				return NewNumberValue(0)
			} else {
				return NewNumberValue(-0)
			}
		} else {
			return NewNumberValue(JSNumberInf)
		}
	}

	Assert(n.IsFinite() && !n.IsZero())
	if exponent.IsPositiveInf() {
		if math.Abs(n.Data.ToFloat()) == 1 {
			return NewNumberValue(JSNumberNaN)
		}
		if math.Abs(n.Data.ToFloat()) > 1 {
			return NewNumberValue(JSNumberInf)
		}
		return NewNumberValue(0)
	}
	if exponent.IsNegativeInf() {
		if math.Abs(n.Data.ToFloat()) == 1 {
			return NewNumberValue(JSNumberNaN)
		}
		if math.Abs(n.Data.ToFloat()) > 1 {
			return NewNumberValue(0)
		}
		return NewNumberValue(JSNumberInf)
	}

	Assert(exponent.IsFinite() && !exponent.IsZero())

	return NewNumberValue(JSNumber(math.Pow(n.Data.ToFloat(), exponent.Data.ToFloat())))
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
	return NewNumberValue(JSNumber(math.Mod(n.Data.ToFloat(), other.Data.ToFloat())))
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
	return NewNumberValue(JSNumber(int64(n.Data) << uint64(int64(other.Data))))
}

// 6.1.6.1.10
func (n *NumberValue) SignedRightShift(other *NumberValue) *NumberValue {
	return NewNumberValue(JSNumber(int64(n.Data) >> uint64(int64(other.Data))))
}

// 6.1.6.1.11
func (n *NumberValue) UnsignedRightShift(other *NumberValue) *NumberValue {
	return NewNumberValue(JSNumber(uint64(n.Data) >> uint64(int64(other.Data))))
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
		return NewNumberValue(JSNumber(int64(n.Data) & int64(other.Data)))
	case numberBitwiseOr:
		return NewNumberValue(JSNumber(int64(n.Data) | int64(other.Data)))
	case numberBitwiseXor:
		return NewNumberValue(JSNumber(int64(n.Data) ^ int64(other.Data)))
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
				n.Data = JSNumber(data)
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
			Data:   JSNumber(n.Data),
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
	var isSafeInteger BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		numberValue := arguments[0]

		if IsIntegralNumber(numberValue) {
			data := numberValue.(*NumberValue).Data
			if float64(data) <= POW_2_53-1 {
				return NewBooleanValue(true)
			}
		}
		return NewBooleanValue(false)
	}

	DefineBuiltinFunction(object, "isFinite", isFinite, 1, realm)
	DefineBuiltinFunction(object, "isInteger", isInteger, 1, realm)
	DefineBuiltinFunction(object, "isNaN", isNaN, 1, realm)
	DefineBuiltinFunction(object, "isSafeInteger", isSafeInteger, 1, realm)
	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.NumberPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "EPSILON", &PropertyDescriptor{
		Value:        NewNumberValue(2.220446049250313e-16),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "MAX_SAFE_INTEGER", &PropertyDescriptor{
		Value:        NewNumberValue(9007199254740991),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "MIN_SAFE_INTEGER", &PropertyDescriptor{
		Value:        NewNumberValue(-9007199254740991),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "MAX_VALUE", &PropertyDescriptor{
		Value:        NewNumberValue(1.7976931348623157e+308),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "MIN_VALUE", &PropertyDescriptor{
		Value:        NewNumberValue(5e-324),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "NaN", &PropertyDescriptor{
		Value:        NewNumberValue(JSNumberNaN),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "NEGATIVE_INFINITY", &PropertyDescriptor{
		Value:        NewNumberValue(JSNumberNegInf),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "POSITIVE_INFINITY", &PropertyDescriptor{
		Value:        NewNumberValue(JSNumberInf),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(object, "parseFloat", NewValueFromObject(realm.Intrinsics.ParseFloat))
	DefineBuiltinPropertyV(object, "parseInt", NewValueFromObject(realm.Intrinsics.ParseInt))

	DefineBuiltinPropertyV(realm.Intrinsics.NumberPrototype, "constructor", NewValueFromObject(object))

	return object
}

type NumberObject struct {
	*Object
	Data JSNumber
}

func NewNumberObject(agent *Agent, value JSNumber, prototype ObjectType) *NumberObject {
	object := &NumberObject{
		Object: NewObject(agent, prototype, "Number"),
		Data:   value,
	}
	object.ref = object
	return object
}

func NewNumberPrototype(realm *Realm) *NumberObject {
	object := &NumberObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "NumberPrototype"),
	}
	object.ref = object

	agent := realm.Agent
	var toString BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		radix := arguments[0]

		x := thisNumberValue(agent, this)
		var radixMV JSInt = 10
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
	var toLocaleString BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		return toString(thisNumberValue(agent, this), nil, nil)
	}

	DefineBuiltinFunction(object, "toString", toString, 0, realm)
	DefineBuiltinFunction(object, "valueOf", valueOf, 0, realm)
	DefineBuiltinFunction(object, "toLocaleString", toLocaleString, 0, realm)

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
