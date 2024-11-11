package coldmoon

import (
	"math"
	"math/big"
	"reflect"

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

type ArgumentsList []Value
type Value interface {
	String() string
	ToBoolean() bool
	CallAssumeCallable(value Value, argumentsList ArgumentsList) Value
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

var NullValue = &nullValue{}

var NaNValue = &NumberValue{Data: math.NaN()}

var InfinityValue = &NumberValue{Data: math.Inf(1)}
var NegativeInfinityValue = NumberValue{Data: math.Inf(-1)}

type ObjectValue struct {
	Value
	Object ObjectType
}

func (o *ObjectValue) CallAssumeCallable(value Value, argumentsList ArgumentsList) Value {
	object := o.Object
	return object.InternalMethods().Call(object, value, argumentsList)
}

func (o *ObjectValue) String() string {
	primValue := ToPrimitive(o.Object.Agent(), o, PreferredTypeString)
	if _, isObject := primValue.(*ObjectValue); isObject {
		panic("")
	}
	return primValue.String()
}

func NewValueFromObject(object ObjectType) Value {
	return &ObjectValue{Object: object}
}

// 7.1.1
func ToPrimitive(agent *Agent, value Value, hint PreferredType) Value {
	if objectValue, isObject := value.(*ObjectValue); isObject {
		symbol := WellKnownSymbols[WellKnownSymbolsToPrimitive]
		exoticToPrim := GetMethod(agent, value, NewSymbolPropertyKey(&symbol))
		if exoticToPrim != nil {
			hintString := hint.String()

			result := NewValueFromObject(exoticToPrim).CallAssumeCallable(value, []Value{
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

func ToNumber(agent *Agent, value Value) *NumberValue {
	switch value := value.(type) {
	case *NumberValue:
		return value
	case *undefinedValue:
		return InfinityValue
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
		primValue := ToPrimitive(agent, value, PreferredTypeNumber)

		if _, ok := primValue.(*ObjectValue); !ok {
			Assert(false)
		}

		return ToNumber(agent, primValue)
	}
	panic("TypeError")

}

// 7.1.3
func ToNumeric(agent *Agent, value Value) Value {
	primValue := ToPrimitive(agent, value, PreferredTypeNumber)
	if bigInt, ok := primValue.(*BigIntValue); ok {
		return bigInt
	}
	return ToNumber(agent, primValue)
}
func ToIntegerOrInfinity(value Value, agent *Agent) float64 {
	number := ToNumber(agent, value)
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

func ToInt32(value Value, agent *Agent) int32 {
	number := ToNumber(agent, value)
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
func ToUint32(agent *Agent, value Value) uint32 {
	number := ToNumber(agent, value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}

	intData := number.Truncate()

	int32bit := math.Mod(intData, POW_2_32)
	return uint32(int32bit)
}
func ToInt16(value Value, agent *Agent) int16 {
	number := ToNumber(agent, value)
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
func ToUint16(value Value, agent *Agent) uint16 {
	number := ToNumber(agent, value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int16bit := math.Mod(intData, POW_2_16)
	return uint16(int16bit)
}
func ToInt8(value Value, agent *Agent) int8 {
	number := ToNumber(agent, value)
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
func ToUint8(value Value, agent *Agent) uint8 {
	number := ToNumber(agent, value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int8bit := math.Mod(intData, POW_2_8)
	return uint8(int8bit)
}
func ToUint8Clamp(value Value, agent *Agent) uint8 {
	number := ToNumber(agent, value)
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
func ToBigInt(agent *Agent, value Value) *BigIntValue {
	prim := ToPrimitive(agent, value, PreferredTypeNumber)
	switch p := prim.(type) {
	case *undefinedValue, *nullValue, *NumberValue, *SymbolValue:
		panic("TypeError")
	case *BooleanValue:
		return NewBigIntFromBoolean(p.Data)
	case *BigIntValue:
		return p
	case *StringValue:
		n, _ := StringToBigInt(p)
		return n
	default:
		panic("unreachable")
	}
}

// 7.1.15
func ToBigInt64(value Value, agent *Agent) int64 {
	n := ToBigInt(agent, value)

	twoPow64 := uint128.New(0, 1)
	twoPow63 := uint128.New(1<<63, 0)

	int64bit := uint128.FromBig(n.Data).Mod(twoPow64)
	if int64bit.Cmp(twoPow63) >= 0 {
		return int64(int64bit.Sub(twoPow64).Lo)
	} else {
		return int64(int64bit.Lo)
	}
}

// 7.1.16
func ToBigUint64(agent *Agent, value Value) uint64 {
	n := ToBigInt(agent, value)

	twoPow64 := uint128.New(0, 1)
	int64bit := uint128.FromBig(n.Data).Mod(twoPow64)
	return int64bit.Lo
}

// 7.1.18
func ValueToObject(agent *Agent, value Value) ObjectType {
	realm := agent.CurrentRealm()
	switch v := value.(type) {
	case *undefinedValue, *nullValue:
		panic("TypeError")
	case *BooleanValue:
		return NewBooleanObject(agent, v.Data, realm.Intrinsics.BooleanPrototype)
	case *ObjectValue:
		return v.Object
	case *StringValue:
		return NewStringObject(agent, v.Data, realm.Intrinsics.StringPrototype)
	case *NumberValue:
		return NewNumberObject(agent, v.Data, realm.Intrinsics.NumberPrototype)
	default:
		panic("unimplemented")
	}
}

// 7.1.4.1.1
func StringToNumber(value *StringValue) *NumberValue {
	return &NumberValue{
		Data: 0,
	}
}

// 7.1.14
func StringToBigInt(value *StringValue) (*BigIntValue, bool) {
	bigInt := new(big.Int)
	if _, ok := bigInt.SetString(value.Data, 10); !ok {
		return nil, false
	}

	return &BigIntValue{
		Data: bigInt,
	}, true
}

// 7.1.19
func ToPropertyKey(agent *Agent, value Value) PropertyKey {
	key := ToPrimitive(agent, value, PreferredTypeString)
	if symbolKey, ok := key.(*SymbolValue); ok {
		return NewSymbolPropertyKey(symbolKey)
	}

	keyString := key.String()
	return NewStringPropertyKey(keyString)
}

// 7.1.20
func ToLength(agent *Agent, value Value) uint64 {
	length := ToIntegerOrInfinity(value, agent)

	if length <= 0 {
		return 0
	}

	return uint64(math.Min(length, POW_2_53-1))
}

// 7.1.22
func ToIndex(value Value, agent *Agent) uint64 {
	if value == UndefinedValue {
		return 0
	}

	integer := ToIntegerOrInfinity(value, agent)
	if integer < 0 || integer >= POW_2_53 {
		panic("RangeError")
	}
	return uint64(integer)
}

// 7.2.2
func isArray(value Value) bool {
	o, ok := value.(*ObjectValue)
	if !ok {
		return false
	}
	if _, ok = o.Object.(*ArrayObject); ok {
		return true
	}
	return false
}

// 7.2.3
func IsCallable(value Value) bool {
	objectValue, isObject := value.(*ObjectValue)

	if !isObject {
		return false
	}
	if objectValue.Object.ToObject().InternalMethods().Call != nil {
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
	if objectValue.Object.ToObject().InternalMethods().Construct != nil {
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

// 7.2.10
func SameValue(x Value, y Value) bool {
	if ValueType(x) != ValueType(y) {
		return false
	}
	if number, ok := x.(*NumberValue); ok {
		return number.SameValue(y.(*NumberValue))
	}

	return SameValueNonNumber(x, y)
}

// 7.2.12
func SameValueNonNumber(x Value, y Value) bool {
	Assert(ValueType(x) == ValueType(y))
	switch x.(type) {
	case *undefinedValue, *nullValue:
		return true
	case *BigIntValue:
		return x.(*BigIntValue).Equal(*y.(*BigIntValue))
	case *StringValue:
		return x.(*StringValue).Data == y.(*StringValue).Data
	case *BooleanValue:
		return x.(*BooleanValue).Data == y.(*BooleanValue).Data
	case *SymbolValue:
		return x.(*SymbolValue).Id == y.(*SymbolValue).Id
	case *ObjectValue:
		return x.(*ObjectValue).Object == y.(*ObjectValue).Object
	default:
		panic("unreachable")
	}
}

type isLessThanOrder int

const (
	IsLessThanOrderLeftFirst isLessThanOrder = iota
	IsLessThanOrderRightFirst
)

// 7.2.13
func IsLessThan(agent *Agent, x, y Value, order isLessThanOrder) bool {
	var px, py Value
	if order == IsLessThanOrderLeftFirst {
		px = ToPrimitive(agent, x, PreferredTypeNumber)
		py = ToPrimitive(agent, y, PreferredTypeNumber)
	} else {
		px = ToPrimitive(agent, y, PreferredTypeNumber)
		py = ToPrimitive(agent, x, PreferredTypeNumber)
	}
	pxString, isPxString := px.(*StringValue)
	pyString, isPyString := px.(*StringValue)
	if isPxString && isPyString {
		return pxString.Data < pyString.Data
	} else {
		nx := ToNumber(agent, px)
		ny := ToNumber(agent, py)
		if nx.IsNaN() || ny.IsNaN() {
			return false
		}
		return nx.Data < ny.Data
	}

}

// 7.2.14
func IsLooselyEqual(agent *Agent, x Value, y Value) bool {
	if reflect.TypeOf(x) == reflect.TypeOf(y) {
		return IsStrictlyEqual(x, y)
	}
	if x == NullValue && y == UndefinedValue {
		return true
	}
	if x == UndefinedValue && y == NullValue {
		return true
	}
	if x == NullValue || y == UndefinedValue {
		return false
	}
	if x == UndefinedValue || y == NullValue {
		return false
	}
	if x == NaNValue && y == NaNValue {
		return true
	}

	_, xIsString := x.(*StringValue)
	_, xIsNumber := x.(*NumberValue)
	_, xIsBigInt := x.(*BigIntValue)
	_, xIsBoolean := x.(*BooleanValue)
	_, xIsObject := x.(*ObjectValue)
	_, xIsSymbol := x.(*SymbolValue)

	yString, yIsString := y.(*StringValue)
	_, yIsNumber := y.(*NumberValue)
	_, yIsBigInt := y.(*BigIntValue)
	_, yIsBoolean := y.(*BooleanValue)
	_, yIsObject := y.(*ObjectValue)
	_, yIsSymbol := y.(*SymbolValue)

	if xIsNumber && yIsString {
		return IsLooselyEqual(agent, x, ToNumber(agent, y))
	}
	if xIsString {
		if yIsNumber {
			return IsLooselyEqual(agent, ToNumber(agent, x), y)
		}
		if yIsBigInt {
			return IsLooselyEqual(agent, y, x)
		}
	}
	if xIsBigInt && yIsString {
		n, ok := StringToBigInt(yString)
		if !ok {
			return false
		}
		return IsLooselyEqual(agent, x, n)
	}
	if xIsBoolean {
		return IsLooselyEqual(agent, ToNumber(agent, x), y)
	}

	if yIsBoolean {
		return IsLooselyEqual(agent, x, ToNumber(agent, y))
	}

	if (xIsString || xIsNumber || xIsBigInt || xIsSymbol) && yIsObject {
		return IsLooselyEqual(agent, x, ToPrimitive(agent, y, PreferredTypeDefault))
	}

	if xIsObject && (yIsString || yIsNumber || yIsBigInt || yIsSymbol) {
		return IsLooselyEqual(agent, ToPrimitive(agent, x, PreferredTypeDefault), y)
	}

	if (xIsBigInt && yIsNumber) || (xIsNumber && yIsBigInt) {
	}

	return false

}

// 7.2.15
func IsStrictlyEqual(x Value, y Value) bool {
	if reflect.TypeOf(x) != reflect.TypeOf(y) {
		return false
	}

	_, xIsNumber := x.(*NumberValue)
	if xIsNumber {
		return x.(*NumberValue).SameValue(y.(*NumberValue))
	}

	return SameValueNonNumber(x, y)
}

// 7.3.3
func GetV(value Value, agent *Agent, key PropertyKey) Value {
	object := ValueToObject(agent, value)
	return object.InternalMethods().Get(object, key, value)
}

// 7.3.11
func GetMethod(agent *Agent, value Value, key PropertyKey) ObjectType {
	fun := GetV(value, agent, key)
	if fun == UndefinedValue || fun == NullValue {
		return nil
	}

	if !IsCallable(fun) {
		panic("TypeError")
	}

	return fun.(*ObjectValue).Object
}

// 7.3.14
func ValueCall(self Value, value Value, argumentsList []Value) Value {
	if !IsCallable(value) {
		panic("TypeError")
	}

	return value.(*ObjectValue).Object.ToObject().InternalMethods().Call(value.(*ObjectValue).Object, self, argumentsList)
}

// 7.3.20
func CreateListFromArrayLike(agent *Agent, self Value) []Value {
	// TODO: element types
	objectValue, ok := self.(*ObjectValue)
	if !ok {
		panic("TypeError")
	}

	length := objectValue.Object.LengthOfArrayLike()

	var list []Value
	for i := uint64(0); i < length; i++ {
		index := NewIntegerIndexPropertyKey(int(i))
		next := GetV(self, agent, index)
		list = append(list, next)
	}

	return list
}

// 7.3.21
func OrdinaryHasInstance(self Value, value Value) bool {
	if IsCallable(self) {
		return false
	}
	selfObject := self.(*ObjectValue).Object

	objectValue, ok := value.(*ObjectValue)
	if !ok {
		return false
	}

	proto := selfObject.Get(NewStringPropertyKey("prototype"))
	protoObject, ok := proto.(*ObjectValue)
	if !ok {
		panic("TypeError")
	}

	object := objectValue.Object
	for {
		object = object.InternalMethods().GetPrototypeOf(object)
		if object == nil {
			return false
		}
		if protoObject.Object == object {
			return true
		}
	}

}

func CallNoArgs(self Value, value Value) Value {
	return ValueCall(self, value, nil)
}

func CallAssumeCallableNoArgs(self, value Value) Value {
	return self.CallAssumeCallable(value, nil)
}

func ValueType(value Value) string {
	switch value.(type) {
	case *undefinedValue:
		return "undefined"
	case *nullValue:
		return "object"
	case *BooleanValue:
		return "boolean"
	case *StringValue:
		return "string"
	case *NumberValue:
		return "number"
	case *BigIntValue:
		return "bigint"
	case *SymbolValue:
		return "symbol"
	case *ObjectValue:
		return "object"
	default:
		panic("unreachable")
	}
}
