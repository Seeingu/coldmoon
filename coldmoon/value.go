package coldmoon

import (
	"errors"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/bits-and-blooms/bitset"
	"github.com/dlclark/regexp2"
	"lukechampine.com/uint128"
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
	prim := ReturnAssertNormal(value.ToPrimitive(agent, PreferredTypeNumber))
	switch p := prim.(type) {
	case *undefinedValue, *nullValue, *NumberValue, *SymbolValue:
		panic("TypeError")
	case *BooleanValue:
		co.value = NewBigIntFromBoolean(p.Data)
		return
	case *BigIntValue:
		co.value = p
		return
	case *StringValue:
		n, ok := StringToBigInt(p)
		Assert(ok)
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

	twoPow64 := uint128.New(0, 1)
	twoPow63 := uint128.New(1<<63, 0)

	int64bit := uint128.FromBig(n.Data).Mod(twoPow64)
	if int64bit.Cmp(twoPow63) >= 0 {
		co.value = int64(int64bit.Sub(twoPow64).Lo)
		return
	} else {
		co.value = int64(int64bit.Lo)
		return
	}
}

// ToBigUint64
// spec: 7.1.16
func ToBigUint64(agent *Agent, value Value) (co Completion[uint64]) {
	n, isAbrupt, rt := ReturnIfAbrupt(ToBigInt(agent, value), co)
	if isAbrupt {
		return rt
	}

	twoPow64 := uint128.New(0, 1)
	int64bit := uint128.FromBig(n.Data).Mod(twoPow64)
	co.value = int64bit.Lo
	return
}

// 7.1.4.1.1
func StringToNumber(value *StringValue) *NumberValue {
	if value.Data == "" {
		return NewNumberValue(0)
	}

	n, err := strconv.ParseFloat(strings.Trim(value.Data, " "), 64)
	if err != nil {
		return NaNValue
	}

	return NewNumberValue(JSNumber(n))
}

// 7.1.14
func StringToBigInt(value *StringValue) (*BigIntValue, bool) {
	bigInt := new(big.Int)
	base := 10
	rawString := value.Data
	if strings.HasPrefix(value.Data, "0x") || strings.HasPrefix(value.Data, "0X") {
		rawString = value.Data[2:]
		base = 16
	} else if strings.HasPrefix(value.Data, "0b") || strings.HasPrefix(value.Data, "0B") {
		rawString = value.Data[2:]
		base = 2
	} else if strings.HasPrefix(value.Data, "0o") || strings.HasSuffix(value.Data, "0O") {
		rawString = value.Data[2:]
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

// 7.1.22
func ToIndex(agent *Agent, value Value) JSInt {
	if IsUndefinedOrNil(value) {
		return 0
	}

	integer := ReturnAssertNormal(ToIntegerOrInfinity(agent, value))
	if integer < 0 || float64(integer) >= POW_2_53 {
		panic("RangeError")
	}
	return integer
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
	o, ok := value.(*ObjectValue)
	if !ok {
		return false
	}
	if _, ok = o.Object.(*ArrayObject); ok {
		return true
	}
	if proxy, ok := o.Object.(*ProxyObject); ok {
		proxy.validateNonRevokedProxy()
		proxyTarget := proxy.Target
		return IsArray(proxyTarget.ToValue())
	}
	return false
}

// 7.2.3
func IsCallable(value Value) bool {
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
func IsConstructor(value Value) bool {
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
		_px, isAbrupt, rt := ReturnIfAbrupt(y.ToPrimitive(agent, PreferredTypeNumber), co)
		if isAbrupt {
			return rt
		}
		px = _px
		_py, isAbrupt, rt := ReturnIfAbrupt(x.ToPrimitive(agent, PreferredTypeNumber), co)
		if isAbrupt {
			return rt
		}
		py = _py
	}
	pxString, isPxString := px.(*StringValue)
	pyString, isPyString := px.(*StringValue)
	if isPxString && isPyString {
		return NewBooleanValue(pxString.Data < pyString.Data).ToCompletion()
	} else {
		nx, isAbrupt, rt := ReturnIfAbrupt(px.ToNumber(agent), co)
		if isAbrupt {
			return rt
		}
		ny, isAbrupt, rt := ReturnIfAbrupt(py.ToNumber(agent), co)
		if isAbrupt {
			return rt
		}
		if nx.IsNaN() || ny.IsNaN() {
			return FalseValue.ToCompletion()
		}
		return NewBooleanValue(nx.Data < ny.Data).ToCompletion()
	}
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
		return IsLooselyEqual(agent, x, ReturnAssertNormal(y.ToPrimitive(agent, PreferredTypeDefault)))
	}

	if xIsObject && (yIsString || yIsNumber || yIsBigInt || yIsSymbol) {
		return IsLooselyEqual(agent, ReturnAssertNormal(x.ToPrimitive(agent, PreferredTypeDefault)), y)
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
	object := value.ToObject(agent).value
	return object.InternalMethods().Get(object, key, value)
}

// 7.3.11
// TODO: Should Return UndefinedValue
func GetMethod(agent *Agent, value Value, key PropertyKey) ObjectType {
	fun := ReturnAssertNormal(GetV(agent, value, key))
	if IsUndefinedOrNull(fun) {
		return nil
	}

	if !IsCallable(fun) {
		panic("TypeError")
	}

	return fun.(*ObjectValue).Object
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

// CreateListFromArrayLike
// spec: 7.3.20
func CreateListFromArrayLike(agent *Agent, self Value) (co Completion[[]Value]) {
	// TODO: element types
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
		IgnoreCase:           flagsBitSet.Test(flagsI),
		Multiline:            flagsBitSet.Test(flagsM),
		Unicode:              flagsBitSet.Test(flagsU),
		DotAll:               flagsBitSet.Test(flagsS),
		UnicodeSets:          flagsBitSet.Test(flagsV),
		CapturingGroupsCount: capturingGroupsCount,
	}
	regexpObject.RegExpRecord = rer
	regexpObject.RegExpMatcher = parseResult

	obj.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(0), setThrowTypeThrow)

	co.value = obj
	return
}

func CountLeftCapturingParensWithin(r *regexp2.Regexp) int {
	// TODO
	return 0
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
	length := o.PropertyStorage().Get(NewStringPropertyKey("length"))
	l = JSInt(MustGetObject(length.Value).(*NumberObject).Data)
	return l, true
}
