package coldmoon

import (
	"errors"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/bits-and-blooms/bitset"
	"github.com/dlclark/regexp2"
)

// MARK: - PreferredType

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

type ArgumentsList = []Value

func NewValueFromObject(object ObjectType) Value {
	ov := &ObjectValue{Object: object}
	ov.Value = NewBaseValue(ov)
	return ov
}

// ToNumeric
// spec: 7.1.3
func ToNumeric(agent *Agent, value Value) (co CompletionValue) {
	primValue, isAbrupt, rt := ReturnIfAbrupt(value.ToPrimitive(agent, PreferredTypeNumber), co)
	if isAbrupt {
		return rt
	}
	if bigInt, ok := primValue.(*BigIntValue); ok {
		return bigInt.ToCompletion()
	}
	if n, isAbrupt, rt := ReturnIfAbrupt(primValue.ToNumber(agent), co); isAbrupt {
		return rt
	} else {
		co.value = n
	}
	return
}

// ToIntegerOrInfinity
// spec: 7.1.5
func ToIntegerOrInfinity(agent *Agent, value Value) (co Completion[JSInt]) {
	number, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	if number.IsNaN() {
		co.value = 0
		return
	}
	if number.IsPositiveInf() {
		co.value = JSInt(math.Inf(1))
		return
	}
	if number.IsNegativeInf() {
		co.value = JSInt(math.Inf(-1))
		return
	}
	co.value = JSInt(number.Truncate())
	return
}

var (
	POW_2_53 = math.Pow(2, 53)
	POW_2_32 = math.Pow(2, 32)
	POW_2_31 = math.Pow(2, 31)
	POW_2_16 = math.Pow(2, 16)
	POW_2_15 = math.Pow(2, 15)
	POW_2_8  = math.Pow(2, 8)
	POW_2_7  = math.Pow(2, 7)
)

func ToInt32(agent *Agent, value Value) (co Completion[JSInt]) {
	number, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	if !number.IsFinite() || number.Data == 0 {
		co.value = 0
		return
	}

	intData := number.Truncate()

	int32bit := math.Mod(float64(intData), POW_2_32)
	if int32bit >= POW_2_31 {
		co.value = JSInt(int32(int32bit - POW_2_32))
		return
	} else {
		co.value = JSInt(int32(int32bit))
		return
	}
}

func ToUint32(agent *Agent, value Value) (co Completion[JSInt]) {
	number, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	if !number.IsFinite() || number.Data == 0 {
		co.value = 0
		return
	}

	intData := number.Truncate()

	int32bit := math.Mod(float64(intData), POW_2_32)
	co.value = JSInt(uint32(int32bit))
	return
}

func ToInt16(value Value, agent *Agent) (co Completion[int16]) {
	number, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	if !number.IsFinite() || number.Data == 0 {
		co.value = 0
		return
	}

	intData := number.Truncate()

	int16bit := math.Mod(float64(intData), POW_2_16)

	if int16bit >= POW_2_15 {
		co.value = int16(int16bit - POW_2_16)
		return
	} else {
		co.value = int16(int16bit)
		return
	}
}

func ToUint16(value Value, agent *Agent) (co Completion[uint16]) {
	number, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	if !number.IsFinite() || number.Data == 0 {
		co.value = 0
		return
	}
	intData := number.Truncate()
	int16bit := math.Mod(float64(intData), POW_2_16)
	co.value = uint16(int16bit)
	return
}

func ToInt8(value Value, agent *Agent) (co Completion[int8]) {
	number, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	if !number.IsFinite() || number.Data == 0 {
		co.value = 0
		return
	}
	intData := number.Truncate()
	int8bit := math.Mod(float64(intData), POW_2_8)
	if int8bit >= POW_2_7 {
		co.value = int8(int8bit - POW_2_8)
		return
	} else {
		co.value = int8(int8bit)
		return
	}
}

func ToUint8(value Value, agent *Agent) (co Completion[uint8]) {
	number, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	if !number.IsFinite() || number.Data == 0 {
		co.value = 0
		return
	}
	intData := number.Truncate()
	int8bit := math.Mod(float64(intData), POW_2_8)
	co.value = uint8(int8bit)
	return
}

func ToUint8Clamp(value Value, agent *Agent) (co Completion[uint8]) {
	number, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	if number.IsNaN() {
		co.value = 0
		return
	}
	if number.Data <= 0 {
		co.value = 0
		return
	}
	if number.Data >= 255 {
		co.value = 255
		return
	}

	f := math.Floor(number.Data.ToFloat())
	fInt := uint8(f)

	if f+0.5 < number.Data.ToFloat() {
		co.value = fInt + 1
		return
	}
	if number.Data.ToFloat() < f+0.5 {
		co.value = fInt
		return
	}

	if fInt%2 != 0 {
		co.value = fInt + 1
		return
	}

	co.value = fInt
	return
}

// ToBigInt
// spec: 7.1.13
func ToBigInt(agent *Agent, value Value) (co Completion[*BigIntValue]) {
	if value == nil {
		value = UndefinedValue
	}
	prim, isAbrupt, rt := ReturnIfAbrupt(value.ToPrimitive(agent, PreferredTypeNumber), co)
	if isAbrupt {
		return rt
	}
	switch p := prim.(type) {
	case *undefinedValue, *nullValue, *NumberValue, *SymbolValue:
		return co.ThrowTypeError(agent, "Cannot convert value to BigInt")
	case *BooleanValue:
		co.value = NewBigIntFromBoolean(p.Data)
		return
	case *BigIntValue:
		co.value = p
		return
	case *StringValue:
		n, ok := StringToBigInt(p)
		if !ok {
			return co.ThrowError(agent, SyntaxError, "Cannot convert string to BigInt")
		}
		co.value = n
		return
	default:
		panic("unreachable")
	}
}

// ToBigInt64
// spec: 7.1.15
func ToBigInt64(value Value, agent *Agent) (co Completion[int64]) {
	n, isAbrupt, rt := ReturnIfAbrupt(ToBigInt(agent, value), co)
	if isAbrupt {
		return rt
	}

	twoPow64 := new(big.Int).Lsh(big.NewInt(1), 64)
	twoPow63 := new(big.Int).Lsh(big.NewInt(1), 63)
	int64bit := new(big.Int).Mod(new(big.Int).Set(n.Data), twoPow64)
	if int64bit.Cmp(twoPow63) >= 0 {
		int64bit.Sub(int64bit, twoPow64)
	}
	co.value = int64bit.Int64()
	return
}

// ToBigUint64
// spec: 7.1.16
func ToBigUint64(agent *Agent, value Value) (co Completion[uint64]) {
	n, isAbrupt, rt := ReturnIfAbrupt(ToBigInt(agent, value), co)
	if isAbrupt {
		return rt
	}

	twoPow64 := new(big.Int).Lsh(big.NewInt(1), 64)
	uint64bit := new(big.Int).Mod(new(big.Int).Set(n.Data), twoPow64)
	co.value = uint64bit.Uint64()
	return
}

// 7.1.4.1.1
func StringToNumber(value *StringValue) *NumberValue {
	if value.Data == "" {
		return NewNumberValue(0)
	}

	n, err := strconv.ParseFloat(strings.TrimFunc(value.Data, isECMAScriptWhitespace), 64)
	if err != nil && !math.IsInf(n, 0) {
		return NaNValue
	}

	return NewNumberValue(JSNumber(n))
}

// 7.1.14
func StringToBigInt(value *StringValue) (*BigIntValue, bool) {
	bigInt := new(big.Int)
	base := 10
	rawString := strings.TrimSpace(value.Data)
	if rawString == "" {
		return NewBigIntValue(bigInt), true
	}
	if strings.HasPrefix(rawString, "0x") || strings.HasPrefix(rawString, "0X") {
		rawString = rawString[2:]
		base = 16
	} else if strings.HasPrefix(rawString, "0b") || strings.HasPrefix(rawString, "0B") {
		rawString = rawString[2:]
		base = 2
	} else if strings.HasPrefix(rawString, "0o") || strings.HasPrefix(rawString, "0O") {
		rawString = rawString[2:]
		base = 8
	}
	if _, ok := bigInt.SetString(rawString, base); !ok {
		return nil, false
	}

	return NewBigIntValue(bigInt), true
}

func GetPrivateName(agent *Agent, value Value) (*PrivateName, bool) {
	if symbol, ok := value.(*SymbolValue); ok && symbol.IsPrivate {
		return &PrivateName{
			Symbol: symbol,
		}, true
	}
	return nil, false
}

// ToPropertyKey
// spec: 7.1.19
func ToPropertyKey(agent *Agent, value Value) (co Completion[PropertyKey]) {
	key, isAbrupt, rt := ReturnIfAbrupt(value.ToPrimitive(agent, PreferredTypeString), co)
	if isAbrupt {
		return rt
	}
	Assert(key != nil)
	if symbolKey, ok := key.(*SymbolValue); ok {
		co.value = NewSymbolPropertyKey(symbolKey)
		return
	}

	keyString := key.String()
	co.value = NewStringPropertyKey(keyString)
	return
}

// ToLength
// spec: 7.1.20
func ToLength(agent *Agent, value Value) (co Completion[JSInt]) {
	length, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, value), co)
	if isAbrupt {
		return rt
	}

	if length <= 0 {
		co.value = 0
		return
	}
	co.value = JSInt(math.Min(float64(length), POW_2_53-1))
	return
}

// ToIndex
// spec: 7.1.22
func ToIndex(agent *Agent, value Value) (co Completion[JSInt]) {
	if IsUndefinedOrNil(value) {
		co.value = 0
		return
	}

	integer, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, value), co)
	if isAbrupt {
		return rt
	}
	if integer < 0 || float64(integer) >= POW_2_53 {
		return co.ThrowError(agent, RangeError, "ToIndex: value is out of range")
	}
	co.value = integer
	return
}

// 7.2.1
func RequireObjectCoercible(agent *Agent, value Value) Value {
	if value == UndefinedValue || value == NullValue {
		panic("TypeError")
	}
	return value
}

// 7.2.2
func IsArray(value Value) bool {
	return ReturnAssertNormal(IsArrayCompletion(value))
}

func IsArrayCompletion(value Value) (co Completion[bool]) {
	o, ok := value.(*ObjectValue)
	if !ok {
		return
	}
	if _, ok = o.Object.(*ArrayObject); ok {
		co.value = true
		return
	}
	if proxy, ok := o.Object.(*ProxyObject); ok {
		if proxy.Target == nil {
			return co.ThrowTypeError(proxy.Agent(), "IsArray called on a revoked Proxy")
		}
		Assert(proxy.Handler != nil)
		proxyTarget := proxy.Target
		return IsArrayCompletion(proxyTarget.ToValue())
	}
	return
}

// 7.2.3
func IsCallable(value Value) bool {
	objectValue, isObject := value.(*ObjectValue)

	if !isObject {
		return false
	}
	if objectValue.Object.internalMethods().Call != nil {
		return true
	}

	return false
}

// 7.2.4
func IsConstructor(value Value) bool {
	objectValue, isObject := value.(*ObjectValue)

	if !isObject {
		return false
	}
	if objectValue.Object.internalMethods().Construct != nil {
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
	if x.TypeString() != y.TypeString() {
		return false
	}
	if number, _, ok := x.NumberOrBigInt(); ok {
		return number.SameValue(y.(*NumberValue))
	}

	return SameValueNonNumber(x, y)
}

// 7.2.11
func SameValueZero(x Value, y Value) bool {
	if x.TypeString() != y.TypeString() {
		return false
	}
	if number, _, ok := x.NumberOrBigInt(); ok {
		return number.SameValueZero(y.(*NumberValue))
	}

	return SameValueNonNumber(x, y)
}

// 7.2.12
func SameValueNonNumber(x Value, y Value) bool {
	Assert(x.TypeString() == y.TypeString())
	switch x.(type) {
	case *undefinedValue, *nullValue:
		return true
	case *BigIntValue:
		return x.(*BigIntValue).Equal(*y.(*BigIntValue))
	case *StringValue:
		return x.(*StringValue).Data == y.(*StringValue).Data
	case *BooleanValue:
		return x == y
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

// IsLessThan
// spec: 7.2.13
// returns either a Boolean or undefined, or throw
func IsLessThan(agent *Agent, x, y Value, order isLessThanOrder) (co CompletionValue) {
	var px, py Value
	if order == IsLessThanOrderLeftFirst {
		_px, isAbrupt, rt := ReturnIfAbrupt(x.ToPrimitive(agent, PreferredTypeNumber), co)
		if isAbrupt {
			return rt
		}
		px = _px
		_py, isAbrupt, rt := ReturnIfAbrupt(y.ToPrimitive(agent, PreferredTypeNumber), co)
		if isAbrupt {
			return rt
		}
		py = _py
	} else {
		_py, isAbrupt, rt := ReturnIfAbrupt(y.ToPrimitive(agent, PreferredTypeNumber), co)
		if isAbrupt {
			return rt
		}
		py = _py
		_px, isAbrupt, rt := ReturnIfAbrupt(x.ToPrimitive(agent, PreferredTypeNumber), co)
		if isAbrupt {
			return rt
		}
		px = _px
	}
	pxString, isPxString := px.(*StringValue)
	pyString, isPyString := py.(*StringValue)
	if isPxString && isPyString {
		return NewBooleanValue(utf16LessThan(pxString.Data, pyString.Data)).ToCompletion()
	}
	pxBigInt, isPxBigInt := px.(*BigIntValue)
	pyBigInt, isPyBigInt := py.(*BigIntValue)
	if isPxBigInt && isPyString {
		ny, ok := StringToBigInt(pyString)
		if !ok {
			return UndefinedValue.ToCompletion()
		}
		return NewBooleanValue(pxBigInt.LessThan(ny)).ToCompletion()
	}
	if isPxString && isPyBigInt {
		nx, ok := StringToBigInt(pxString)
		if !ok {
			return UndefinedValue.ToCompletion()
		}
		return NewBooleanValue(nx.LessThan(pyBigInt)).ToCompletion()
	}

	nx, isAbrupt, rt := ReturnIfAbrupt(ToNumeric(agent, px), co)
	if isAbrupt {
		return rt
	}
	ny, isAbrupt, rt := ReturnIfAbrupt(ToNumeric(agent, py), co)
	if isAbrupt {
		return rt
	}
	nxNumber, nxIsNumber := nx.(*NumberValue)
	nyNumber, nyIsNumber := ny.(*NumberValue)
	if (nxIsNumber && nxNumber.IsNaN()) || (nyIsNumber && nyNumber.IsNaN()) {
		return UndefinedValue.ToCompletion()
	}
	if nxIsNumber && nyIsNumber {
		return NewBooleanValue(nxNumber.Data < nyNumber.Data).ToCompletion()
	}
	nxBigInt, nxIsBigInt := nx.(*BigIntValue)
	nyBigInt, nyIsBigInt := ny.(*BigIntValue)
	if nxIsBigInt && nyIsBigInt {
		return NewBooleanValue(nxBigInt.LessThan(nyBigInt)).ToCompletion()
	}
	Assert((nxIsNumber && nyIsBigInt) || (nxIsBigInt && nyIsNumber))
	if (nxIsNumber && nxNumber.IsNegativeInf()) || (nyIsNumber && nyNumber.IsPositiveInf()) {
		return TrueValue.ToCompletion()
	}
	if (nxIsNumber && nxNumber.IsPositiveInf()) || (nyIsNumber && nyNumber.IsNegativeInf()) {
		return FalseValue.ToCompletion()
	}
	return NewBooleanValue(numericRat(nx).Cmp(numericRat(ny)) < 0).ToCompletion()
}

// utf16LessThan compares ECMAScript String code units rather than Go's UTF-8
// bytes or Unicode scalar values.
func utf16LessThan(x, y string) bool {
	xUnits := utf16.Encode([]rune(x))
	yUnits := utf16.Encode([]rune(y))
	limit := min(len(xUnits), len(yUnits))
	for i := range limit {
		if xUnits[i] != yUnits[i] {
			return xUnits[i] < yUnits[i]
		}
	}
	return len(xUnits) < len(yUnits)
}

// numericRat gives finite Number and BigInt values one exact mathematical
// representation for the mixed-type comparison required by IsLessThan.
func numericRat(value Value) *big.Rat {
	switch value := value.(type) {
	case *NumberValue:
		if result := new(big.Rat).SetFloat64(float64(value.Data)); result != nil {
			return result
		}
	case *BigIntValue:
		return new(big.Rat).SetInt(value.Data)
	}
	panic("IsLessThan received a non-finite or non-numeric value")
}

// 7.2.14
func IsLooselyEqual(agent *Agent, x Value, y Value) (co Completion[bool]) {
	if reflect.TypeOf(x) == reflect.TypeOf(y) {
		co.value = IsStrictlyEqual(x, y)
		return
	}
	if x == NullValue && y == UndefinedValue {
		co.value = true
		return
	}
	if x == UndefinedValue && y == NullValue {
		co.value = true
		return
	}
	if x == NullValue || y == UndefinedValue {
		co.value = false
		return
	}
	if x == UndefinedValue || y == NullValue {
		co.value = false
		return
	}
	if x == NaNValue && y == NaNValue {
		co.value = true
		return
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
		yn, isAbrupt, rt := ReturnIfAbrupt(y.ToNumber(agent), co)
		if isAbrupt {
			return rt
		}
		return IsLooselyEqual(agent, x, yn)
	}
	if xIsString {
		if yIsNumber {
			xn, isAbrupt, rt := ReturnIfAbrupt(x.ToNumber(agent), co)
			if isAbrupt {
				return rt
			}
			return IsLooselyEqual(agent, xn, y)
		}
		if yIsBigInt {
			return IsLooselyEqual(agent, y, x)
		}
	}
	if xIsBigInt && yIsString {
		n, ok := StringToBigInt(yString)
		if !ok {
			co.value = false
			return
		}
		return IsLooselyEqual(agent, x, n)
	}
	if xIsBoolean {
		xn, isAbrupt, rt := ReturnIfAbrupt(x.ToNumber(agent), co)
		if isAbrupt {
			return rt
		}
		return IsLooselyEqual(agent, xn, y)
	}

	if yIsBoolean {
		yn, isAbrupt, rt := ReturnIfAbrupt(y.ToNumber(agent), co)
		if isAbrupt {
			return rt
		}
		return IsLooselyEqual(agent, x, yn)
	}

	if (xIsString || xIsNumber || xIsBigInt || xIsSymbol) && yIsObject {
		yPrimitive, isAbrupt, rt := ReturnIfAbrupt(y.ToPrimitive(agent, PreferredTypeDefault), co)
		if isAbrupt {
			return rt
		}
		return IsLooselyEqual(agent, x, yPrimitive)
	}

	if xIsObject && (yIsString || yIsNumber || yIsBigInt || yIsSymbol) {
		xPrimitive, isAbrupt, rt := ReturnIfAbrupt(x.ToPrimitive(agent, PreferredTypeDefault), co)
		if isAbrupt {
			return rt
		}
		return IsLooselyEqual(agent, xPrimitive, y)
	}

	if (xIsBigInt && yIsNumber) || (xIsNumber && yIsBigInt) {
	}

	co.value = false
	return
}

// 7.2.15
func IsStrictlyEqual(x Value, y Value) bool {
	if reflect.TypeOf(x) != reflect.TypeOf(y) {
		return false
	}

	if xNumber, xIsNumber := x.(*NumberValue); xIsNumber {
		return xNumber.SameValue(y.(*NumberValue))
	}
	return SameValueNonNumber(x, y)
}

// 7.3.3
func GetV(agent *Agent, value Value, key PropertyKey) CompletionValue {
	var co CompletionValue
	object, isAbrupt, rt := ReturnIfAbrupt(value.ToObject(agent), co)
	if isAbrupt {
		return rt
	}
	return object.internalMethods().Get(object, key, value)
}

// GetMethodCompletion returns a callable property while preserving failures
// from accessors and the TypeError required for non-callable properties.
// spec: 7.3.11
func GetMethodCompletion(agent *Agent, value Value, key PropertyKey) (co Completion[ObjectType]) {
	fun, isAbrupt, rt := ReturnIfAbrupt(GetV(agent, value, key), co)
	if isAbrupt {
		return rt
	}
	if IsUndefinedOrNull(fun) {
		return
	}

	if !IsCallable(fun) {
		return co.ThrowTypeError(agent, "method is not callable")
	}

	co.value = MustGetObject(fun)
	return
}

// GetMethod is the panic-style compatibility wrapper used by algorithms that
// have not yet migrated to completion-aware method lookup. Language errors are
// rethrown as Values so builtin invocation can preserve them as completions.
func GetMethod(agent *Agent, value Value, key PropertyKey) ObjectType {
	result := GetMethodCompletion(agent, value, key)
	if result.IsAbrupt() {
		if result.Error() != nil {
			panic(result.Error())
		}
		panic("GetMethod completed abruptly without an error value")
	}
	return result.Data()
}

// 7.3.18
func CreateArrayFromList(agent *Agent, elements []Value) ObjectType {
	array := ArrayCreate(agent, JSInt(len(elements)), nil)

	for i, element := range elements {
		var propKey PropertyKey
		if float64(i) <= POW_2_53 {
			propKey = NewIntegerIndexPropertyKey(JSInt(i))
		} else {
			propKey = NewStringPropertyKey(string(rune(i)))
		}

		array.CreateDataPropertyOrThrow(propKey, element)
	}

	return array
}

type ArrayLikeElementTypes int

const (
	ArrayLikeElementTypesAll ArrayLikeElementTypes = iota
	ArrayLikeElementTypesPropertyKey
)

// CreateListFromArrayLike
// spec: 7.3.19
func CreateListFromArrayLike(agent *Agent, self Value, elementTypes ...ArrayLikeElementTypes) (co Completion[[]Value]) {
	Assert(len(elementTypes) <= 1)
	validElementTypes := ArrayLikeElementTypesAll
	if len(elementTypes) == 1 {
		validElementTypes = elementTypes[0]
	}
	Assert(validElementTypes == ArrayLikeElementTypesAll || validElementTypes == ArrayLikeElementTypesPropertyKey)

	objectValue, ok := self.(*ObjectValue)
	if !ok {
		return co.ThrowTypeError(agent, "TypeError")
	}

	length, isAbrupt, rt := ReturnIfAbrupt(objectValue.Object.LengthOfArrayLike(), co)
	if isAbrupt {
		return rt
	}

	var list []Value
	for i := JSInt(0); i < length; i++ {
		index := NewIntegerIndexPropertyKey(i)
		next, isAbrupt, rt := ReturnIfAbrupt(GetV(agent, self, index), co)
		if isAbrupt {
			return rt
		}
		if validElementTypes == ArrayLikeElementTypesPropertyKey {
			switch next.(type) {
			case *StringValue, *SymbolValue:
			default:
				return co.ThrowTypeError(agent, "array-like element is not a property key")
			}
		}
		list = append(list, next)
	}

	co.value = list
	return
}

// ValueInvoke Invoke
// spec: 7.3.20
func ValueInvoke(agent *Agent, self Value, propertyKey PropertyKey, argumentsList []Value) (co CompletionValue) {
	fun, isAbrupt, rt := ReturnIfAbrupt(GetV(agent, self, propertyKey), co)
	if isAbrupt {
		return rt
	}
	return fun.Call(agent, self, argumentsList)
}

// 7.2.8
func IsRegExp(value Value) bool {
	object, ok := value.GetObject()
	if !ok {
		return false
	}
	matcher := object.Get(NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsMatch]))
	if matcher != UndefinedValue {
		return matcher.ToBoolean()
	}
	if _, ok := object.(*RegExpObject); ok {
		return true
	}
	return false
}

// 10.1.15
func RequireInternalSlot[O ObjectType](value Value) O {
	if !value.IsObject() {
		panic("TypeError")
	}
	o := MustGetObject(value)
	if !ObjectIs[O](o) {
		panic("TypeError")
	}
	return o.(O)
}

// 22.2.3.2
func RegExpAlloc(agent *Agent, newTarget ObjectType) ObjectType {
	obj := OrdinaryCreateFromConstructor(agent, newTarget, "%RegExp.prototype%", nil)
	regexp := &RegExpObject{
		Object: obj,
	}
	regexp.ref = regexp
	regexp.DefinePropertyOrThrow(NewStringPropertyKey("lastIndex"), &PropertyDescriptor{
		Value:        NewNumberValue(0),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	return regexp
}

// 22.2.3.3
func RegExpInitialize(agent *Agent, obj ObjectType, pattern Value, flags Value) (co Completion[ObjectType]) {
	var p Value
	if pattern == UndefinedValue {
		p = NewStringValue("")
	} else {
		p = ToString(agent, pattern)
	}
	var f Value
	if IsUndefinedOrNil(flags) {
		f = NewStringValue("")
	} else {
		f = ToString(agent, flags)
	}

	var flagsBitSet bitset.BitSet
	flagsD := uint(1)
	flagsG := uint(1 << 1)
	flagsI := uint(1 << 2)
	flagsM := uint(1 << 3)
	flagsS := uint(1 << 4)
	flagsU := uint(1 << 5)
	flagsY := uint(1 << 6)
	flagsV := uint(1 << 7)
	for _, c := range f.String() {
		var flagsBit uint
		switch c {
		case 'd':
			flagsBit = flagsD
		case 'g':
			flagsBit = flagsG
		case 'i':
			flagsBit = flagsI
		case 'm':
			flagsBit = flagsM
		case 's':
			flagsBit = flagsS
		case 'u':
			flagsBit = flagsU
		case 'y':
			flagsBit = flagsY
		case 'v':
			flagsBit = flagsV
		default:
			co.err = agent.ThrowException(SyntaxError, "Invalid flags")
			return
		}
		if flagsBitSet.Test(flagsBit) {
			co.err = agent.ThrowException(SyntaxError, "Duplicate flags")
			return
		}
		flagsBitSet.Set(flagsBit)
	}

	patternText := p.String()
	parseResult, err := ParsePattern(patternText, flagsBitSet.Test(flagsU), flagsBitSet.Test(flagsV))
	if err != nil {
		co.err = agent.ThrowException(SyntaxError, "Invalid pattern")
		return
	}

	regexpObject := obj.(*RegExpObject)
	regexpObject.OriginalSource = p.String()
	regexpObject.OriginalFlags = f.String()

	capturingGroupsCount := CountLeftCapturingParensWithin(parseResult)
	rer := &RegExpRecord{
		HasIndices:           flagsBitSet.Test(flagsD),
		Global:               flagsBitSet.Test(flagsG),
		IgnoreCase:           flagsBitSet.Test(flagsI),
		Multiline:            flagsBitSet.Test(flagsM),
		Unicode:              flagsBitSet.Test(flagsU),
		DotAll:               flagsBitSet.Test(flagsS),
		UnicodeSets:          flagsBitSet.Test(flagsV),
		Sticky:               flagsBitSet.Test(flagsY),
		CapturingGroupsCount: capturingGroupsCount,
	}
	regexpObject.RegExpRecord = rer
	regexpObject.RegExpMatcher = parseResult

	obj.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(0), setThrowTypeThrow)

	co.value = obj
	return
}

func CountLeftCapturingParensWithin(r *regexp2.Regexp) int {
	// regexp2 includes group zero (the entire match) in its group-number
	// metadata. Every remaining number represents one capturing parenthesis,
	// including named groups while excluding non-capturing assertions/groups.
	count := len(r.GetGroupNumbers()) - 1
	if count < 0 {
		return 0
	}
	return count
}

// 22.2.3.4
func ParsePattern(pattern string, unicode bool, unicodeSets bool) (r *regexp2.Regexp, err error) {
	if unicode && unicodeSets {
		err = errors.New("SyntaxError")
		return
	}
	if unicodeSets || unicode {
		return regexp2.Compile(pattern, regexp2.Unicode)
	}
	return regexp2.Compile(pattern, regexp2.None)
}

func MustGetObject(value Value) ObjectType {
	return value.(*ObjectValue).Object
}

func ValueIs[Type Value](value Value) bool {
	_, ok := value.(Type)
	return ok
}

func ValueGet[Type Value](value Value) (Type, bool) {
	v, ok := value.(Type)
	return v, ok
}

func ValueGetLength(v Value) (l JSInt, ok bool) {
	if !v.IsObject() {
		return
	}
	o := MustGetObject(v)
	length := o.propertyStorage().Get(NewStringPropertyKey("length"))
	l = JSInt(MustGetObject(length.Value).(*NumberObject).Data)
	return l, true
}
