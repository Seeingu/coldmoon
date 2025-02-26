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

// 7.1.3
func ToNumeric(agent *Agent, value Value) Value {
	primValue := value.ToPrimitive(agent, PreferredTypeNumber)
	if bigInt, ok := primValue.(*BigIntValue); ok {
		return bigInt
	}
	return primValue.ToNumber(agent)
}

func ToIntegerOrInfinity(agent *Agent, value Value) JSInt {
	number := value.ToNumber(agent)
	if number.IsNaN() {
		return 0
	}
	if number.IsPositiveInf() {
		return JSInt(math.Inf(1))
	}
	if number.IsNegativeInf() {
		return JSInt(math.Inf(-1))
	}
	return JSInt(number.Truncate())
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

func ToInt32(agent *Agent, value Value) JSInt {
	number := value.ToNumber(agent)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}

	intData := number.Truncate()

	int32bit := math.Mod(float64(intData), POW_2_32)
	if int32bit >= POW_2_31 {
		return JSInt(int32(int32bit - POW_2_32))
	} else {
		return JSInt(int32(int32bit))
	}
}

func ToUint32(agent *Agent, value Value) JSInt {
	number := value.ToNumber(agent)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}

	intData := number.Truncate()

	int32bit := math.Mod(float64(intData), POW_2_32)
	return JSInt(uint32(int32bit))
}

func ToInt16(value Value, agent *Agent) int16 {
	number := value.ToNumber(agent)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}

	intData := number.Truncate()

	int16bit := math.Mod(float64(intData), POW_2_16)

	if int16bit >= POW_2_15 {
		return int16(int16bit - POW_2_16)
	} else {
		return int16(int16bit)
	}
}

func ToUint16(value Value, agent *Agent) uint16 {
	number := value.ToNumber(agent)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int16bit := math.Mod(float64(intData), POW_2_16)
	return uint16(int16bit)
}

func ToInt8(value Value, agent *Agent) int8 {
	number := value.ToNumber(agent)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int8bit := math.Mod(float64(intData), POW_2_8)
	if int8bit >= POW_2_7 {
		return int8(int8bit - POW_2_8)
	} else {
		return int8(int8bit)
	}
}

func ToUint8(value Value, agent *Agent) uint8 {
	number := value.ToNumber(agent)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int8bit := math.Mod(float64(intData), POW_2_8)
	return uint8(int8bit)
}

func ToUint8Clamp(value Value, agent *Agent) uint8 {
	number := value.ToNumber(agent)
	if number.IsNaN() {
		return 0
	}
	if number.Data <= 0 {
		return 0
	}
	if number.Data >= 255 {
		return 255
	}

	f := math.Floor(number.Data.ToFloat())
	fInt := uint8(f)

	if f+0.5 < number.Data.ToFloat() {
		return fInt + 1
	}
	if number.Data.ToFloat() < f+0.5 {
		return fInt
	}

	if fInt%2 != 0 {
		return fInt + 1
	}

	return fInt
}

func ToBigInt(agent *Agent, value Value) *BigIntValue {
	prim := value.ToPrimitive(agent, PreferredTypeNumber)
	switch p := prim.(type) {
	case *undefinedValue, *nullValue, *NumberValue, *SymbolValue:
		panic("TypeError")
	case *BooleanValue:
		return NewBigIntFromBoolean(p.Data)
	case *BigIntValue:
		return p
	case *StringValue:
		n, ok := StringToBigInt(p)
		Assert(ok)
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

// 7.1.19
func ToPropertyKey(agent *Agent, value Value) PropertyKey {
	key := value.ToPrimitive(agent, PreferredTypeString)
	if symbolKey, ok := key.(*SymbolValue); ok {
		return NewSymbolPropertyKey(symbolKey)
	}

	keyString := key.String()
	return NewStringPropertyKey(keyString)
}

// ToLength
// spec: 7.1.20
func ToLength(agent *Agent, value Value) (co Completion[JSInt]) {
	length := ToIntegerOrInfinity(agent, value)

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

	integer := ToIntegerOrInfinity(agent, value)
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

// TODO: standardalize
// 7.2.13
func IsLessThanV2(agent *Agent, x, y Value, order isLessThanOrder) Value {
	var px, py Value
	if order == IsLessThanOrderLeftFirst {
		px = x.ToPrimitive(agent, PreferredTypeNumber)
		py = y.ToPrimitive(agent, PreferredTypeNumber)
	} else {
		px = y.ToPrimitive(agent, PreferredTypeNumber)
		py = x.ToPrimitive(agent, PreferredTypeNumber)
	}
	pxString, isPxString := px.(*StringValue)
	pyString, isPyString := px.(*StringValue)
	if isPxString && isPyString {
		return NewBooleanValue(pxString.Data < pyString.Data)
	} else {
		nx := px.ToNumber(agent)
		ny := py.ToNumber(agent)
		if nx.IsNaN() || ny.IsNaN() {
			return FalseValue
		}
		return NewBooleanValue(nx.Data < ny.Data)
	}
}

// 7.2.13
func IsLessThan(agent *Agent, x, y Value, order isLessThanOrder) bool {
	return IsLessThanV2(agent, x, y, order).ToBoolean()
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
		return IsLooselyEqual(agent, x, y.ToNumber(agent))
	}
	if xIsString {
		if yIsNumber {
			return IsLooselyEqual(agent, x.ToNumber(agent), y)
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
		return IsLooselyEqual(agent, x.ToNumber(agent), y)
	}

	if yIsBoolean {
		return IsLooselyEqual(agent, x, y.ToNumber(agent))
	}

	if (xIsString || xIsNumber || xIsBigInt || xIsSymbol) && yIsObject {
		return IsLooselyEqual(agent, x, y.ToPrimitive(agent, PreferredTypeDefault))
	}

	if xIsObject && (yIsString || yIsNumber || yIsBigInt || yIsSymbol) {
		return IsLooselyEqual(agent, x.ToPrimitive(agent, PreferredTypeDefault), y)
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

// TODO(BM): returns completion
// CreateListFromArrayLike
// spec: 7.3.20
func CreateListFromArrayLike(agent *Agent, self Value) []Value {
	var co Completion[[]Value]
	// TODO: element types
	objectValue, ok := self.(*ObjectValue)
	if !ok {
		panic(co.ThrowTypeError(agent, "TypeError"))
	}

	length, isAbrupt, rt := ReturnIfAbrupt(objectValue.Object.LengthOfArrayLike(), co)
	if isAbrupt {
		panic(rt)
	}

	var list []Value
	for i := JSInt(0); i < length; i++ {
		index := NewIntegerIndexPropertyKey(i)
		next := ReturnAssertNormal(GetV(agent, self, index))
		list = append(list, next)
	}

	return list
}

// ValueInvoke Invoke
// 7.3.20
func ValueInvoke(agent *Agent, self Value, propertyKey PropertyKey, argumentsList []Value) Value {
	fun := ReturnAssertNormal(GetV(agent, self, propertyKey))
	return ReturnAssertNormal(fun.Call(self, argumentsList))
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
