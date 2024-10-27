package coldmoon

import (
	"math"
	"math/big"

	"lukechampine.com/uint128"
)

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

func ToNumber(value Value) *NumberValue {
	switch value := value.(type) {
	case *NumberValue:
		return value
	case *undefinedValue:
		return &InfinityValue
	case *nullValue:
		return &NumberValue{Data: 0}
	case *BooleanValue:
		if value.Data {
			return &NumberValue{Data: 1}
		}
		return &NumberValue{Data: 0}
	case *StringValue:
		return StringToNumber(value)
	case *ObjectValue:
		primValue := ToPrimitive(value, PreferredTypeNumber)

		if _, ok := primValue.(*ObjectValue); !ok {
			Assert(false)
		}

		return ToNumber(primValue)
	}
	panic("TypeError")

}

// 7.1.3
func ToNumeric(value Value) Value {
	primValue := ToPrimitive(value, PreferredTypeNumber)
	if bigInt, ok := primValue.(*BigInt); ok {
		return bigInt
	}
	return ToNumber(primValue)
}
func ToIntegerOrInfinity(value Value) float64 {
	number := ToNumber(value)
	if number.IsNaN() {
		return 0
	}
	if number.IsPositiveInf() {
		return math.Inf(1)
	}
	if number.IsNegativeInf() {
		return math.Inf(-1)
	}
	return number.Truncate()
}

var POW_2_53 = math.Pow(2, 53)
var POW_2_32 = math.Pow(2, 32)
var POW_2_31 = math.Pow(2, 31)
var POW_2_16 = math.Pow(2, 16)
var POW_2_15 = math.Pow(2, 15)
var POW_2_8 = math.Pow(2, 8)
var POW_2_7 = math.Pow(2, 7)

func ToInt32(value Value) int32 {
	number := ToNumber(value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}

	intData := number.Truncate()

	int32bit := math.Mod(intData, POW_2_32)
	if int32bit >= POW_2_31 {
		return int32(int32bit - POW_2_32)
	} else {
		return int32(int32bit)
	}

}
func ToUint32(value Value) uint32 {
	number := ToNumber(value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}

	intData := number.Truncate()

	int32bit := math.Mod(intData, POW_2_32)
	return uint32(int32bit)
}
func ToInt16(value Value) int16 {
	number := ToNumber(value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}

	intData := number.Truncate()

	int16bit := math.Mod(intData, POW_2_16)

	if int16bit >= POW_2_15 {
		return int16(int16bit - POW_2_16)
	} else {
		return int16(int16bit)
	}
}
func ToUint16(value Value) uint16 {
	number := ToNumber(value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int16bit := math.Mod(intData, POW_2_16)
	return uint16(int16bit)
}
func ToInt8(value Value) int8 {
	number := ToNumber(value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int8bit := math.Mod(intData, POW_2_8)
	if int8bit >= POW_2_7 {
		return int8(int8bit - POW_2_8)
	} else {
		return int8(int8bit)
	}
}
func ToUint8(value Value) uint8 {
	number := ToNumber(value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int8bit := math.Mod(intData, POW_2_8)
	return uint8(int8bit)
}
func ToUint8Clamp(value Value) uint8 {
	number := ToNumber(value)
	if number.IsNaN() {
		return 0
	}
	if number.Data <= 0 {
		return 0
	}
	if number.Data >= 255 {
		return 255
	}

	f := math.Floor(number.Data)
	fInt := uint8(f)

	if f+0.5 < number.Data {
		return fInt + 1
	}
	if number.Data < f+0.5 {
		return fInt
	}

	if fInt%2 != 0 {
		return fInt + 1
	}

	return fInt
}
func ToBigInt(value Value) BigInt {
	prim := ToPrimitive(value, PreferredTypeNumber)
	switch p := prim.(type) {
	case *undefinedValue, *nullValue, *NumberValue, *Symbol:
		panic("TypeError")
	case *BooleanValue:
		return NewBigIntFromBoolean(p.Data)
	case *BigInt:
		return *p
	case *StringValue:
		n := StringToBigInt(p)
		return n
	default:
		panic("unreachable")
	}
}

// 7.1.15
func ToBigInt64(value Value) int64 {
	n := ToBigInt(value)

	twoPow64 := uint128.New(0, 1)
	twoPow63 := uint128.New(1<<63, 0)

	int64bit := uint128.FromBig(&n.Data).Mod(twoPow64)
	if int64bit.Cmp(twoPow63) >= 0 {
		return int64(int64bit.Sub(twoPow64).Lo)
	} else {
		return int64(int64bit.Lo)
	}
}

// 7.1.16
func ToBigUint64(value Value) uint64 {
	n := ToBigInt(value)

	twoPow64 := uint128.New(0, 1)
	int64bit := uint128.FromBig(&n.Data).Mod(twoPow64)
	return int64bit.Lo
}

// 7.1.4.1.1
func StringToNumber(value *StringValue) *NumberValue {
	return &NumberValue{
		Data: 0,
	}
}

// 7.1.14
func StringToBigInt(value *StringValue) BigInt {
	return BigInt{
		Data: *big.NewInt(0),
	}
}

// 7.1.19
func ToPropertyKey(value Value) PropertyKey {
	key := ToPrimitive(value, PreferredTypeString)
	if symbolKey, ok := key.(*Symbol); ok {
		return NewSymbolPropertyKey(symbolKey)
	}

	keyString := key.String()
	return NewStringPropertyKey(keyString)
}

// 7.1.20
func ToLength(value Value) uint64 {
	length := ToIntegerOrInfinity(value)

	if length <= 0 {
		return 0
	}

	return uint64(math.Min(length, math.Pow(2, 53)-1))
}

// 7.1.22
func ToIndex(value Value) uint64 {
	if value == UndefinedValue {
		return 0
	}

	integer := ToIntegerOrInfinity(value)
	if integer < 0 || integer >= math.Pow(2, 53) {
		panic("RangeError")
	}
	return uint64(integer)
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

// 7.2.6
func IsIntegralNumber(value Value) bool {
	n, ok := value.(*NumberValue)
	if !ok {
		return false
	}

	if !n.IsFinite() {
		return false
	}
	if n.Truncate() != n.Data {
		return false
	}
	return true
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
