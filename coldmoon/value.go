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

type ArgumentsList []Value

type Value interface {
	String() string
	ToBoolean() bool
	Call(this Value, argumentsList ArgumentsList) Value
	CallNoArgs(this Value) Value
	// --- internal methods ---

	ToCompletion() CompletionValue
	ToPropertyDescriptor() *PropertyDescriptor
}

// TODO(C): we should use convertable
func NewPropertyDescriptorFromValue(value Value) *PropertyDescriptor {
	return &PropertyDescriptor{
		Value:        value,
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	}
}

type StringValue struct {
	Value
	Data string
}

var _ Value = (*StringValue)(nil)

func (s *StringValue) ToCompletion() CompletionValue {
	return NewCompletionValue(s)
}

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

func NewValueFromObject(object ObjectType) Value {
	return &ObjectValue{Object: object}
}

func ToPropertyDescriptor(agent *Agent, value Value) *PropertyDescriptor {
	if value == UndefinedValue {
		return nil
	}
	objectValue, ok := value.(*ObjectValue)
	if !ok {
		panic("TypeError")
	}
	object := objectValue.Object

	desc := &PropertyDescriptor{}

	hasEnumerable := object.HasProperty(NewStringPropertyKey("enumerable"))

	if hasEnumerable {
		enumerable := object.Get(NewStringPropertyKey("enumerable")).ToBoolean()
		desc.Enumerable = enumerable
	}

	hasConfigurable := object.HasProperty(NewStringPropertyKey("configurable"))
	if hasConfigurable {
		configurable := object.Get(NewStringPropertyKey("configurable")).ToBoolean()
		desc.Configurable = configurable
	}

	hasValue := object.HasProperty(NewStringPropertyKey("value"))
	if hasValue {
		desc.Value = object.Get(NewStringPropertyKey("value"))
	}

	hasWritable := object.HasProperty(NewStringPropertyKey("writable"))
	if hasWritable {
		writable := object.Get(NewStringPropertyKey("writable")).ToBoolean()
		desc.Writable = writable
	}

	hasGet := object.HasProperty(NewStringPropertyKey("get"))
	if hasGet {
		get := object.Get(NewStringPropertyKey("get"))
		if !IsCallable(get) && get != UndefinedValue {
			panic("TypeError")
		}
		desc.Get = MustGetObject(get)
	}

	hasSet := object.HasProperty(NewStringPropertyKey("set"))
	if hasSet {
		set := object.Get(NewStringPropertyKey("set"))
		if !IsCallable(set) && set != UndefinedValue {
			panic("TypeError")
		}
		desc.Set = MustGetObject(set)
	}

	if hasGet || hasSet {
		if hasValue || hasWritable {
			panic("TypeError")
		}
	}

	return desc
}

// 7.1.1
func ToPrimitive(agent *Agent, value Value, hint PreferredType) Value {
	if objectValue, isObject := value.(*ObjectValue); isObject {
		symbol := WellKnownSymbols[WellKnownSymbolsToPrimitive]
		exoticToPrim := GetMethod(agent, value, NewSymbolPropertyKey(symbol))
		if exoticToPrim != nil {
			hintString := hint.String()

			result := exoticToPrim.ToValue().Call(value, []Value{
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

// TODO(C): replace with value.ToNumber
func ToNumber(agent *Agent, value Value) *NumberValue {
	switch value := value.(type) {
	case *NumberValue:
		return value
	case *undefinedValue:
		return InfinityValue
	case *nullValue:
		return NewNumberValue(0)
	case *BooleanValue:
		if value.Data {
			return NewNumberValue(1)
		}
		return NewNumberValue(0)
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

func ToIntegerOrInfinity(agent *Agent, value Value) JSInt {
	number := ToNumber(agent, value)
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
	number := ToNumber(agent, value)
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
	number := ToNumber(agent, value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}

	intData := number.Truncate()

	int32bit := math.Mod(float64(intData), POW_2_32)
	return JSInt(uint32(int32bit))
}

func ToInt16(value Value, agent *Agent) int16 {
	number := ToNumber(agent, value)
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
	number := ToNumber(agent, value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int16bit := math.Mod(float64(intData), POW_2_16)
	return uint16(int16bit)
}

func ToInt8(value Value, agent *Agent) int8 {
	number := ToNumber(agent, value)
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
	number := ToNumber(agent, value)
	if !number.IsFinite() || number.Data == 0 {
		return 0
	}
	intData := number.Truncate()
	int8bit := math.Mod(float64(intData), POW_2_8)
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
	case *SymbolValue:
		return NewSymbolObject(agent, v, realm.Intrinsics.SymbolPrototype)
	case *BigIntValue:
		return NewBigIntObject(agent, v, realm.Intrinsics.BigIntPrototype)
	default:
		panic("unimplemented")
	}
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
	if _, ok := bigInt.SetString(value.Data, 10); !ok {
		return nil, false
	}

	return &BigIntValue{
		Data: bigInt,
	}, true
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
	key := ToPrimitive(agent, value, PreferredTypeString)
	if symbolKey, ok := key.(*SymbolValue); ok {
		return NewSymbolPropertyKey(symbolKey)
	}

	keyString := key.String()
	return NewStringPropertyKey(keyString)
}

// 7.1.20
func ToLength(agent *Agent, value Value) JSInt {
	length := ToIntegerOrInfinity(agent, value)

	if length <= 0 {
		return 0
	}

	return JSInt(math.Min(float64(length), POW_2_53-1))
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

// 7.2.11
func SameValueZero(x Value, y Value) bool {
	if ValueType(x) != ValueType(y) {
		return false
	}
	if number, ok := x.(*NumberValue); ok {
		return number.SameValueZero(y.(*NumberValue))
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
func GetV(agent *Agent, value Value, key PropertyKey) Value {
	object := ValueToObject(agent, value)
	return object.InternalMethods().Get(object, key, value)
}

// 7.3.11
// TODO: Should Return UndefinedValue
func GetMethod(agent *Agent, value Value, key PropertyKey) ObjectType {
	fun := GetV(agent, value, key)
	if fun == UndefinedValue || fun == NullValue {
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

// 7.3.20
func CreateListFromArrayLike(agent *Agent, self Value) []Value {
	// TODO: element types
	objectValue, ok := self.(*ObjectValue)
	if !ok {
		panic("TypeError")
	}

	length := objectValue.Object.LengthOfArrayLike()

	var list []Value
	for i := JSInt(0); i < length; i++ {
		index := NewIntegerIndexPropertyKey(i)
		next := GetV(agent, self, index)
		list = append(list, next)
	}

	return list
}

// 7.3.21
func ValueInvoke(agent *Agent, self Value, propertyKey PropertyKey, argumentsList []Value) Value {
	fun := GetV(agent, self, propertyKey)
	return fun.Call(self, argumentsList)
}

// 7.3.21
func OrdinaryHasInstance(agent *Agent, c Value, value Value) Completion[bool] {
	if IsCallable(c) {
		return NewNormalCompletion(false)
	}
	o := MustGetObject(c)
	if b, ok := o.(*BoundFunctionObject); ok {
		bc := b.BoundTargetFunction
		return NewNormalCompletion(InstanceOfOperator(agent, value, bc.ToValue()))
	}

	objectValue, ok := value.(*ObjectValue)
	if !ok {
		return NewNormalCompletion(false)
	}

	proto := o.Get(NewStringPropertyKey("prototype"))
	protoObject, ok := proto.(*ObjectValue)
	if !ok {
		return NewThrowCompletion[bool](agent.ThrowException(TypeError, "prototype is not an object"))
	}

	object := objectValue.Object
	for {
		object = object.InternalMethods().GetPrototypeOf(object)
		if object == nil {
			return NewNormalCompletion(false)
		}
		if protoObject.Object == object {
			return NewNormalCompletion(true)
		}
	}
}

// 7.2.8
func IsRegExp(value Value) bool {
	object, ok := ValueGetObject(value)
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
	if !ValueIsObject(value) {
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
	regexp.DefinePropertyOrThrow(NewStringPropertyKey("lastIndex"), &PropertyDescriptor{
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	return regexp
}

// 22.2.3.3
func RegExpInitialize(agent *Agent, obj ObjectType, pattern Value, flags Value) CompletionObject {
	var p Value
	if pattern == UndefinedValue {
		p = NewStringValue("")
	} else {
		p = ToString(agent, pattern)
	}
	var f Value
	if flags == UndefinedValue {
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
			return NewCompletionObjectError(agent.ThrowException(SyntaxError, "Invalid flags"))
		}
		if flagsBitSet.Test(flagsBit) {
			return NewCompletionObjectError(agent.ThrowException(SyntaxError, "Duplicate flags"))
		}
		flagsBitSet.Set(flagsBit)
	}

	patternText := p.String()
	parseResult, err := ParsePattern(patternText, flagsBitSet.Test(flagsU), flagsBitSet.Test(flagsV))
	if err != nil {
		return NewCompletionObjectError(agent.ThrowException(SyntaxError, "Invalid pattern"))
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

	return NewCompletionObject(obj)
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
	return
}

// Deprecated: use baseValue.TypeString
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

func MustGetObject(value Value) ObjectType {
	return value.(*ObjectValue).Object
}

func ValueIsObject(value Value) bool {
	_, ok := value.(*ObjectValue)
	return ok
}

func ValueIsPromise(value Value) bool {
	objectValue, ok := value.(*ObjectValue)
	if !ok {
		return false
	}
	_, ok = objectValue.Object.(*PromiseObject)
	return ok
}

func ValueGetObject(value Value) (object ObjectType, ok bool) {
	v, ok := ValueGet[*ObjectValue](value)
	if ok {
		return v.Object, true
	}
	return nil, false
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
	if !ValueIsObject(v) {
		return
	}
	o := MustGetObject(v)
	length := o.PropertyStorage().Get(NewStringPropertyKey("length"))
	l = JSInt(MustGetObject(length.Value).(*NumberObject).Data)
	return l, true
}
