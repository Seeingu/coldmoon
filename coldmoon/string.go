package coldmoon

import (
	"sort"
	"strings"

	"github.com/Seeingu/coldmoon/pkg"
	"github.com/samber/lo"
	"golang.org/x/text/unicode/norm"
)

// MARK: - String Object

type StringObject struct {
	*Object
	Data string
}

func StringGetOwnProperty(s *StringObject, p PropertyKey) *PropertyDescriptor {
	index, err := p.GetIndex()
	if err != nil {
		return nil
	}

	str := s.Data
	l := JSInt(len(str))
	if l <= index {
		return nil
	}

	resultStr := str[index : index+1]
	return &PropertyDescriptor{
		Value:        NewStringValue(resultStr),
		Writable:     false,
		Enumerable:   true,
		Configurable: false,
	}
}

func NewStringObject(agent *Agent, s string, prototype ObjectType) *StringObject {
	// 10.4.3.1
	var getOwnProperty GetOwnPropertyFn = func(o ObjectType, p PropertyKey) *PropertyDescriptor {
		desc := OrdinaryGetOwnProperty(o, p)

		if desc != nil {
			return desc
		}
		return StringGetOwnProperty(o.(*StringObject), p)
	}
	var defineOwnProperty DefineOwnPropertyFn = func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) (co Completion[bool]) {
		s := o.(*StringObject)
		stringDesc := StringGetOwnProperty(s, p)
		if stringDesc != nil {
			extensible := s.Extensible()
			co.value = IsCompatiblePropertyDescriptor(extensible, desc, stringDesc)
			return
		}
		co.value = OrdinaryDefineOwnProperty(o, p, desc)
		return
	}
	// 10.4.3.3
	ownPropertyKeys := func(o ObjectType) []PropertyKey {
		propertiesMap := o.PropertyStorage().Properties
		str := o.(*StringObject).Data
		length := JSInt(len(str))
		keys := make([]PropertyKey, length+JSInt(len(propertiesMap)))
		for i := range length {
			keys[i] = NewIntegerIndexPropertyKey(i)
		}

		keysGreaterThanLength := lo.Filter(lo.Keys(propertiesMap), func(pk PropertyKey, index int) bool {
			if index, err := pk.GetIndex(); err == nil {
				return index >= length
			}
			return false
		})
		sort.Slice(keysGreaterThanLength, func(i, j int) bool {
			ii, err := keysGreaterThanLength[i].GetIndex()
			if err != nil {
				panic(err)
			}
			jj, err := keysGreaterThanLength[j].GetIndex()
			if err != nil {
				panic(err)
			}
			return ii < jj
		})
		copy(keys[length:], keysGreaterThanLength)

		for pk := range propertiesMap {
			if _, ok := pk.(StringPropertyKey); ok {
				keys = append(keys, pk)
			}
		}
		for pk := range propertiesMap {
			if _, ok := pk.(SymbolPropertyKey); ok {
				keys = append(keys, pk)
			}
		}
		return keys
	}

	stringObject := &StringObject{
		Object: NewObject(agent, prototype, "String"),
		Data:   s,
	}
	stringObject.ref = stringObject
	stringObject.InternalMethods().GetOwnProperty = getOwnProperty
	stringObject.InternalMethods().DefineOwnProperty = defineOwnProperty
	stringObject.InternalMethods().OwnPropertyKeys = ownPropertyKeys

	length := JSInt(len(s))
	stringObject.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(length.ToNumber()),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return stringObject
}

// 10.4.3.4
var StringCreate = NewStringObject

// MARK: - StringConstructor

func NewStringConstructor(realm *Realm) ObjectType {
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		var s string
		if len(argumentsList) > 0 {
			value := argumentsList[0]
			symbolValue, isSymbol := value.(*SymbolValue)
			if newTarget == nil && isSymbol {
				return NewStringValue(symbolValue.SymbolDescriptiveString())
			}
			stringValue, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(realm.Agent, value), co)
			if isAbrupt {
				return rt
			}
			s = stringValue.Data
		}

		if newTarget == nil {
			return NewStringValue(s)
		}
		return StringCreate(realm.Agent, s, GetPrototypeFromConstructor(newTarget, "%String.prototype%")).ToValue()
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 1, CMString("String"), builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	BindPrototypeAndConstructor(realm.Intrinsics.StringPrototype, object)

	var fromCharCode BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		var s string
		for _, arg := range argumentsList {
			n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(realm.Agent, arg), co)
			if isAbrupt {
				return rt
			}
			s += string(rune(int(n)))
		}
		return NewStringValue(s)
	}
	var fromCodePoint BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		codePoints := make([]rune, len(argumentsList))
		for i, arg := range argumentsList {
			n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(realm.Agent, arg), co)
			if isAbrupt {
				return rt
			}
			codePoints[i] = rune(n)
		}
		return NewStringValue(string(codePoints))
	}
	object.defineBuiltinFunction(realm, CMString("fromCharCode"), fromCharCode, 1)
	object.defineBuiltinFunction(realm, CMString("fromCodePoint"), fromCodePoint, 1)

	return object
}

func NewStringPrototype(realm *Realm) *StringObject {
	stringPrototype := &StringObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "StringPrototype"),
		Data:   "",
	}
	stringPrototype.ref = stringPrototype

	agent := realm.Agent
	var toString BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		switch value := thisArgument.(type) {
		case *StringValue:
			return value
		case *ObjectValue:
			if stringObject, ok := value.Object.(*StringObject); ok {
				return NewStringValue(stringObject.Data)
			}
		}
		return co.ThrowTypeError(agent, "String.prototype.toString receiver is not a String")
	}
	var valueOf BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return toString(thisArgument, argumentsList, newTarget)
	}
	toLowerCase := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String case conversion called on null or undefined")
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		return NewStringValue(strings.ToLower(s.Data))
	}
	toUpperCase := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String case conversion called on null or undefined")
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		return NewStringValue(strings.ToUpper(s.Data))
	}
	isTrimWhiteSpace := func(r rune) bool {
		switch r {
		case '\u0009', '\u000A', '\u000B', '\u000C', '\u000D', '\u0020',
			'\u00A0', '\u1680', '\u2028', '\u2029', '\u202F', '\u205F',
			'\u3000', '\uFEFF':
			return true
		}
		return r >= '\u2000' && r <= '\u200A'
	}
	trimString := func(this Value, where string) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNil(this) || this == NullValue {
			return co.ThrowTypeError(agent, "String trim called on null or undefined")
		}
		stringValue, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		s := stringValue.Data
		switch where {
		case "start":
			s = strings.TrimLeftFunc(s, isTrimWhiteSpace)
		case "end":
			s = strings.TrimRightFunc(s, isTrimWhiteSpace)
		default:
			s = strings.TrimFunc(s, isTrimWhiteSpace)
		}
		return NewStringValue(s)
	}
	trim := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return trimString(this, "both")
	}
	trimEnd := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return trimString(this, "end")
	}
	trimStart := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return trimString(this, "start")
	}
	var charAt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(thisArgument) {
			return co.ThrowTypeError(agent, "String.prototype.charAt called on null or undefined")
		}
		positionValue := pkg.SliceSafeGet(argumentsList, 0)
		if positionValue == nil {
			positionValue = UndefinedValue
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, thisArgument), co)
		if isAbrupt {
			return rt
		}
		n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, positionValue), co)
		if isAbrupt {
			return rt
		}
		position := int(n)
		size := len(s.Data)
		if position < 0 || position >= size {
			return NewStringValue("")
		}
		return NewStringValue(string(s.Data[position]))
	}
	var charCodeAt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(thisArgument) {
			return co.ThrowTypeError(agent, "String.prototype.charCodeAt called on null or undefined")
		}
		positionValue := pkg.SliceSafeGet(argumentsList, 0)
		if positionValue == nil {
			positionValue = UndefinedValue
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, thisArgument), co)
		if isAbrupt {
			return rt
		}
		n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, positionValue), co)
		if isAbrupt {
			return rt
		}
		position := int(n)
		size := len(s.Data)
		if position < 0 || position >= size {
			return NaNValue
		}
		return NewNumberValue(JSNumber(s.Data[position]))
	}
	var iterator BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		stringIteratorObject := &StringIteratorObject{
			Object: NewObject(agent, realm.Intrinsics.StringIteratorPrototype, "String Iterator"),
			Data:   s,
		}
		stringIteratorObject.ref = stringIteratorObject
		return (stringIteratorObject).ToValue()
	}
	var at BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		index := argumentsList[0]
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		length := len(s)
		n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, index), co)
		if isAbrupt {
			return rt
		}
		relativeIndex := int(n)
		if relativeIndex < 0 || relativeIndex >= length {
			return UndefinedValue
		}
		k := relativeIndex
		return NewStringValue(string(s[k]))
	}
	var slice BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(thisArgument) {
			return co.ThrowTypeError(agent, "String.prototype.slice called on null or undefined")
		}
		start := pkg.SliceSafeGet(argumentsList, 0)
		end := pkg.SliceSafeGet(argumentsList, 1)
		if start == nil {
			start = UndefinedValue
		}
		if end == nil {
			end = UndefinedValue
		}
		stringValue, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, thisArgument), co)
		if isAbrupt {
			return rt
		}
		intStart, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, start), co)
		if isAbrupt {
			return rt
		}
		s := stringValue.Data
		length := JSInt(len(s))
		var intEnd JSInt
		if end == UndefinedValue {
			intEnd = length
		} else {
			intEnd, isAbrupt, rt = ReturnIfAbrupt(ToIntegerOrInfinity(agent, end), co)
			if isAbrupt {
				return rt
			}
		}

		var from JSInt
		if intStart.IsNegInf() {
			from = 0
		} else if intStart < 0 {
			from = (length + intStart).Max(0)
		} else {
			from = intStart.Min(length)
		}

		var to JSInt
		if intEnd.IsNegInf() {
			to = 0
		} else if intEnd < 0 {
			to = (length + intEnd).Max(0)
		} else {
			to = intEnd.Min(length)
		}
		if from >= to {
			return NewStringValue("")
		}
		return NewStringValue(s[int(from):int(to)])
	}
	var repeat BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(thisArgument) {
			return co.ThrowTypeError(agent, "String.prototype.repeat called on null or undefined")
		}
		count := pkg.SliceSafeGet(argumentsList, 0)
		if count == nil {
			count = UndefinedValue
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, thisArgument), co)
		if isAbrupt {
			return rt
		}
		n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, count), co)
		if isAbrupt {
			return rt
		}
		if n < 0 || n.IsInf() {
			return co.ThrowRangeError(agent, "repeat count must be finite and non-negative")
		}
		if n == 0 {
			return NewStringValue("")
		}
		return NewStringValue(strings.Repeat(s.Data, int(n)))
	}
	var concat BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		o := RequireObjectCoercible(agent, thisArgument)
		stringValue, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, o), co)
		if isAbrupt {
			return rt
		}
		s := stringValue.Data
		for _, arg := range argumentsList {
			stringValue, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, arg), co)
			if isAbrupt {
				return rt
			}
			s += stringValue.Data
		}
		return NewStringValue(s)
	}
	var split BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		separator := pkg.SliceSafeGet(argumentsList, 0)
		limit := pkg.SliceSafeGet(argumentsList, 1)
		if separator == nil {
			separator = UndefinedValue
		}
		if limit == nil {
			limit = UndefinedValue
		}
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String.prototype.split called on null or undefined")
		}
		if !IsUndefinedOrNull(separator) {
			splitter := GetMethod(agent, separator, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsSplit]))
			if splitter != nil {
				return splitter.Call(separator, []Value{this, limit})
			}
		}
		stringValue, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		var lim JSInt = JSInt(POW_2_32 - 1)
		if limit != UndefinedValue {
			lim, isAbrupt, rt = ReturnIfAbrupt(ToUint32(agent, limit), co)
			if isAbrupt {
				return rt
			}
		}
		result := ArrayCreate(agent, 0, nil)
		if lim == 0 {
			return result.ToValue()
		}
		if separator == UndefinedValue {
			result.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(0), stringValue)
			return result.ToValue()
		}
		separatorValue, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, separator), co)
		if isAbrupt {
			return rt
		}
		parts := strings.Split(stringValue.Data, separatorValue.Data)
		if JSInt(len(parts)) > lim {
			parts = parts[:int(lim)]
		}
		for index, part := range parts {
			result.CreateDataPropertyOrThrow(
				NewIntegerIndexPropertyKey(JSInt(index)),
				NewStringValue(part),
			)
		}
		return result.ToValue()
	}
	getStringSubstitution := func(matched, source string, position int, replacement string) string {
		var result strings.Builder
		for index := 0; index < len(replacement); index++ {
			if replacement[index] != '$' || index+1 >= len(replacement) {
				result.WriteByte(replacement[index])
				continue
			}
			switch replacement[index+1] {
			case '$':
				result.WriteByte('$')
				index++
			case '&':
				result.WriteString(matched)
				index++
			case '`':
				result.WriteString(source[:position])
				index++
			case '\'':
				result.WriteString(source[position+len(matched):])
				index++
			default:
				result.WriteByte('$')
			}
		}
		return result.String()
	}
	replaceString := func(this Value, argumentsList []Value, replaceAll bool) CompletionConvertable[Value] {
		var co CompletionValue
		searchValue := pkg.SliceSafeGet(argumentsList, 0)
		replaceValue := pkg.SliceSafeGet(argumentsList, 1)
		if searchValue == nil {
			searchValue = UndefinedValue
		}
		if replaceValue == nil {
			replaceValue = UndefinedValue
		}
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String replacement called on null or undefined")
		}
		if !IsUndefinedOrNull(searchValue) {
			if replaceAll && IsRegExp(searchValue) {
				flags := MustGetObject(searchValue).Get(NewStringPropertyKey("flags"))
				flagsString, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, flags), co)
				if isAbrupt {
					return rt
				}
				if !strings.Contains(flagsString.Data, "g") {
					return co.ThrowTypeError(agent, "replaceAll requires a global RegExp")
				}
			}
			replacer := GetMethod(agent, searchValue, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsReplace]))
			if replacer != nil {
				return replacer.Call(searchValue, []Value{this, replaceValue})
			}
		}
		sourceValue, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		searchString, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, searchValue), co)
		if isAbrupt {
			return rt
		}
		functionalReplace := IsCallable(replaceValue)
		replacement := ""
		if !functionalReplace {
			replacementValue, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, replaceValue), co)
			if isAbrupt {
				return rt
			}
			replacement = replacementValue.Data
		}
		source := sourceValue.Data
		search := searchString.Data
		positions := []int{}
		if replaceAll {
			if search == "" {
				for position := 0; position <= len(source); position++ {
					positions = append(positions, position)
				}
			} else {
				for start := 0; start <= len(source)-len(search); {
					offset := strings.Index(source[start:], search)
					if offset < 0 {
						break
					}
					position := start + offset
					positions = append(positions, position)
					start = position + len(search)
				}
			}
		} else if position := strings.Index(source, search); position >= 0 {
			positions = append(positions, position)
		}
		if len(positions) == 0 {
			return sourceValue
		}
		var result strings.Builder
		endOfLastMatch := 0
		for _, position := range positions {
			result.WriteString(source[endOfLastMatch:position])
			replacementText := replacement
			if functionalReplace {
				replaced, isAbrupt, rt := ReturnIfAbrupt(
					replaceValue.Call(
						agent,
						UndefinedValue,
						[]Value{searchString, NewNumberValue(JSNumber(position)), sourceValue},
					),
					co,
				)
				if isAbrupt {
					return rt
				}
				replacedString, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, replaced), co)
				if isAbrupt {
					return rt
				}
				replacementText = replacedString.Data
			} else {
				replacementText = getStringSubstitution(search, source, position, replacement)
			}
			result.WriteString(replacementText)
			endOfLastMatch = position + len(search)
		}
		result.WriteString(source[endOfLastMatch:])
		return NewStringValue(result.String())
	}
	var replace BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return replaceString(this, argumentsList, false)
	}
	var replaceAll BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return replaceString(this, argumentsList, true)
	}
	stringPad := func(this Value, argumentsList []Value, atStart bool) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String padding called on null or undefined")
		}
		source, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		targetLengthValue := pkg.SliceSafeGet(argumentsList, 0)
		if targetLengthValue == nil {
			targetLengthValue = UndefinedValue
		}
		targetLength, isAbrupt, rt := ReturnIfAbrupt(ToLength(agent, targetLengthValue), co)
		if isAbrupt {
			return rt
		}
		if targetLength <= JSInt(len(source.Data)) {
			return source
		}
		fillValue := pkg.SliceSafeGet(argumentsList, 1)
		fillString := " "
		if fillValue != nil && fillValue != UndefinedValue {
			fill, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, fillValue), co)
			if isAbrupt {
				return rt
			}
			fillString = fill.Data
		}
		if fillString == "" {
			return source
		}
		fillLength := int(targetLength) - len(source.Data)
		repeated := strings.Repeat(fillString, (fillLength/len(fillString))+1)[:fillLength]
		if atStart {
			return NewStringValue(repeated + source.Data)
		}
		return NewStringValue(source.Data + repeated)
	}
	var padStart BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return stringPad(this, argumentsList, true)
	}
	var padEnd BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return stringPad(this, argumentsList, false)
	}
	var localeCompare BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String.prototype.localeCompare called on null or undefined")
		}
		source, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		that := pkg.SliceSafeGet(argumentsList, 0)
		if that == nil {
			that = UndefinedValue
		}
		other, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, that), co)
		if isAbrupt {
			return rt
		}
		return NewNumberValue(JSNumber(strings.Compare(source.Data, other.Data)))
	}
	var normalizeString BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String.prototype.normalize called on null or undefined")
		}
		source, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		formValue := pkg.SliceSafeGet(argumentsList, 0)
		form := "NFC"
		if formValue != nil && formValue != UndefinedValue {
			converted, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, formValue), co)
			if isAbrupt {
				return rt
			}
			form = converted.Data
		}
		switch form {
		case "NFC":
			return NewStringValue(norm.NFC.String(source.Data))
		case "NFD":
			return NewStringValue(norm.NFD.String(source.Data))
		case "NFKC":
			return NewStringValue(norm.NFKC.String(source.Data))
		case "NFKD":
			return NewStringValue(norm.NFKD.String(source.Data))
		default:
			return co.ThrowRangeError(agent, "invalid normalization form")
		}
	}
	var search BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		regexp := argumentsList[0]
		o := RequireObjectCoercible(agent, thisArgument)
		if regexp != UndefinedValue && regexp != NullValue {
			searcher := GetMethod(agent, regexp, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsSearch]))
			if searcher != nil {
				return searcher.Call(regexp, []Value{o}).value
			}
		}

		s := ToString(agent, o)
		rx := RegExpCreate(agent, regexp, UndefinedValue)
		return ValueInvoke(agent, rx.Data().ToValue(), NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsSearch]), []Value{s})
	}
	var match BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String.prototype.match called on null or undefined")
		}
		regexp := pkg.SliceSafeGet(argumentsList, 0)
		if regexp == nil {
			regexp = UndefinedValue
		}
		if !IsUndefinedOrNull(regexp) {
			matcher := GetMethod(agent, regexp, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsMatch]))
			if matcher != nil {
				return matcher.Call(regexp, []Value{this})
			}
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		rx, isAbrupt, rt := ReturnIfAbrupt(RegExpCreate(agent, regexp, UndefinedValue), co)
		if isAbrupt {
			return rt
		}
		return ValueInvoke(
			agent,
			rx.ToValue(),
			NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsMatch]),
			[]Value{s},
		)
	}
	var matchAll BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		regexp := argumentsList[0]
		o := RequireObjectCoercible(agent, thisArgument)
		if regexp != UndefinedValue && regexp != NullValue {
			isRegExp := IsRegExp(regexp)
			if isRegExp {
				flags := MustGetObject(regexp).Get(NewStringPropertyKey("flags"))
				RequireObjectCoercible(agent, flags)
				if !strings.Contains(ToString(agent, flags).Data, "g") {
					panic("TypeError")
				}
			}
			matcher := GetMethod(agent, regexp, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsMatchAll]))
			if matcher != nil {
				return matcher.Call(regexp, []Value{o}).value
			}
		}
		s := ToString(agent, o)
		rx := RegExpCreate(agent, regexp, UndefinedValue)
		return ValueInvoke(agent, rx.Data().ToValue(), NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsMatchAll]), []Value{s})
	}
	indexOf := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(thisArgument) {
			return co.ThrowTypeError(agent, "String.prototype.indexOf called on null or undefined")
		}
		searchString := pkg.SliceSafeGet(argumentsList, 0)
		positionValue := pkg.SliceSafeGet(argumentsList, 1)
		if searchString == nil {
			searchString = UndefinedValue
		}
		if positionValue == nil {
			positionValue = UndefinedValue
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, thisArgument), co)
		if isAbrupt {
			return rt
		}
		searchStr, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, searchString), co)
		if isAbrupt {
			return rt
		}
		position, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, positionValue), co)
		if isAbrupt {
			return rt
		}
		pos := int(position)
		length := len(s.Data)
		start := lo.Clamp(pos, 0, length)
		return NewNumberValue(JSNumber(StringIndexOf(s.Data, searchStr.Data, start)))
	}
	lastIndexOf := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(thisArgument) {
			return co.ThrowTypeError(agent, "String.prototype.lastIndexOf called on null or undefined")
		}
		searchString := pkg.SliceSafeGet(argumentsList, 0)
		positionValue := pkg.SliceSafeGet(argumentsList, 1)
		if searchString == nil {
			searchString = UndefinedValue
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, thisArgument), co)
		if isAbrupt {
			return rt
		}
		searchStr, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, searchString), co)
		if isAbrupt {
			return rt
		}
		position := JSInt(len(s.Data))
		if positionValue != nil && positionValue != UndefinedValue {
			position, isAbrupt, rt = ReturnIfAbrupt(ToIntegerOrInfinity(agent, positionValue), co)
			if isAbrupt {
				return rt
			}
		}
		start := lo.Clamp(int(position), 0, len(s.Data))
		if len(searchStr.Data) == 0 {
			return NewNumberValue(JSNumber(start))
		}
		end := min(start+len(searchStr.Data), len(s.Data))
		return NewNumberValue(JSNumber(strings.LastIndex(s.Data[:end], searchStr.Data)))
	}
	startsWith := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String.prototype.startsWith called on null or undefined")
		}
		searchString := pkg.SliceSafeGet(argumentsList, 0)
		position := pkg.SliceSafeGet(argumentsList, 1)
		if searchString == nil {
			searchString = UndefinedValue
		}
		isRegExp := IsRegExp(searchString)
		if isRegExp {
			return co.ThrowTypeError(agent, "startsWith search string cannot be a RegExp")
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		searchStr, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, searchString), co)
		if isAbrupt {
			return rt
		}
		length := len(s.Data)
		var pos int
		if position == nil {
			pos = 0
		} else {
			n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, position), co)
			if isAbrupt {
				return rt
			}
			pos = int(n)
		}

		start := lo.Clamp(pos, 0, length)
		searchLength := len(searchStr.Data)

		if searchLength == 0 {
			return TrueValue
		}

		end := start + searchLength
		if end > length {
			return FalseValue
		}

		substring := s.Data[start:end]
		if substring == searchStr.Data {
			return TrueValue
		}

		return FalseValue
	}
	endsWith := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(thisArgument) {
			return co.ThrowTypeError(agent, "String.prototype.endsWith called on null or undefined")
		}
		searchString := pkg.SliceSafeGet(argumentsList, 0)
		position := pkg.SliceSafeGet(argumentsList, 1)
		if searchString == nil {
			searchString = UndefinedValue
		}
		isRegExp := IsRegExp(searchString)
		if isRegExp {
			return co.ThrowTypeError(agent, "endsWith search string cannot be a RegExp")
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, thisArgument), co)
		if isAbrupt {
			return rt
		}
		searchStr, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, searchString), co)
		if isAbrupt {
			return rt
		}
		length := len(s.Data)
		var pos int
		if position == nil {
			pos = length
		} else {
			n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, position), co)
			if isAbrupt {
				return rt
			}
			pos = int(n)
		}

		end := lo.Clamp(pos, 0, length)
		searchLength := len(searchStr.Data)

		if searchLength == 0 {
			return TrueValue
		}

		start := end - searchLength
		if start < 0 {
			return FalseValue
		}

		substring := s.Data[start:end]
		if substring == searchStr.Data {
			return TrueValue
		}

		return FalseValue
	}
	var includes BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String.prototype.includes called on null or undefined")
		}
		searchString := pkg.SliceSafeGet(argumentsList, 0)
		position := pkg.SliceSafeGet(argumentsList, 1)
		if searchString == nil {
			searchString = UndefinedValue
		}
		isRegExp := IsRegExp(searchString)
		if isRegExp {
			return co.ThrowTypeError(agent, "includes search string cannot be a RegExp")
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		searchStr, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, searchString), co)
		if isAbrupt {
			return rt
		}
		length := len(s.Data)
		var pos int
		if position == nil {
			pos = 0
		} else {
			n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, position), co)
			if isAbrupt {
				return rt
			}
			pos = int(n)
		}

		start := lo.Clamp(pos, 0, length)
		searchLength := len(searchStr.Data)

		if searchLength == 0 {
			return TrueValue
		}

		end := length - searchLength
		if start > end {
			return FalseValue
		}

		if strings.Contains(s.Data[start:], searchStr.Data) {
			return TrueValue
		}

		return FalseValue
	}
	var codePointAt BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String.prototype.codePointAt called on null or undefined")
		}
		pos := pkg.SliceSafeGet(argumentsList, 0)
		if pos == nil {
			pos = UndefinedValue
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		position, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, pos), co)
		if isAbrupt {
			return rt
		}
		size := len(s.Data)
		if position < 0 || int(position) >= size {
			return UndefinedValue
		}
		return NewNumberValue(JSNumber(s.Data[int(position)]))
	}
	substring := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if IsUndefinedOrNull(this) {
			return co.ThrowTypeError(agent, "String.prototype.substring called on null or undefined")
		}
		startValue := pkg.SliceSafeGet(argumentsList, 0)
		endValue := pkg.SliceSafeGet(argumentsList, 1)
		if startValue == nil {
			startValue = UndefinedValue
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, this), co)
		if isAbrupt {
			return rt
		}
		start, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, startValue), co)
		if isAbrupt {
			return rt
		}
		length := JSInt(len(s.Data))
		end := length
		if endValue != nil && endValue != UndefinedValue {
			end, isAbrupt, rt = ReturnIfAbrupt(ToIntegerOrInfinity(agent, endValue), co)
			if isAbrupt {
				return rt
			}
		}
		finalStart := lo.Clamp(start, 0, length)
		finalEnd := lo.Clamp(end, 0, length)
		from := finalStart.Min(finalEnd)
		to := finalStart.Max(finalEnd)
		return NewStringValue(s.Data[int(from):int(to)])
	}

	stringPrototype.defineBuiltinFunction(realm, CMString("toString"), toString, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("valueOf"), valueOf, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("charAt"), charAt, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("charCodeAt"), charCodeAt, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("iterator"), iterator, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("at"), at, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("slice"), slice, 2)
	stringPrototype.defineBuiltinFunction(realm, CMString("repeat"), repeat, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("concat"), concat, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("split"), split, 2)
	stringPrototype.defineBuiltinFunction(realm, CMString("replace"), replace, 2)
	stringPrototype.defineBuiltinFunction(realm, CMString("replaceAll"), replaceAll, 2)
	stringPrototype.defineBuiltinFunction(realm, CMString("padStart"), padStart, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("padEnd"), padEnd, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("localeCompare"), localeCompare, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("normalize"), normalizeString, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("search"), search, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("match"), match, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("matchAll"), matchAll, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("indexOf"), indexOf, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("lastIndexOf"), lastIndexOf, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("startsWith"), startsWith, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("endsWith"), endsWith, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("includes"), includes, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("codePointAt"), codePointAt, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("toLowerCase"), toLowerCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("toUpperCase"), toUpperCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("toLocaleLowerCase"), toLowerCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("toLocaleUpperCase"), toUpperCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("trim"), trim, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("trimEnd"), trimEnd, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("trimStart"), trimStart, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("substring"), substring, 2)
	stringPrototype.defineBuiltinFunction(realm, CMString("toString"), toString, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("valueOf"), valueOf, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("charAt"), charAt, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("charCodeAt"), charCodeAt, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("iterator"), iterator, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("at"), at, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("slice"), slice, 2)
	stringPrototype.defineBuiltinFunction(realm, CMString("repeat"), repeat, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("concat"), concat, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("split"), split, 2)
	stringPrototype.defineBuiltinFunction(realm, CMString("replace"), replace, 2)
	stringPrototype.defineBuiltinFunction(realm, CMString("replaceAll"), replaceAll, 2)
	stringPrototype.defineBuiltinFunction(realm, CMString("padStart"), padStart, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("padEnd"), padEnd, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("localeCompare"), localeCompare, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("normalize"), normalizeString, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("search"), search, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("match"), match, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("matchAll"), matchAll, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("indexOf"), indexOf, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("lastIndexOf"), lastIndexOf, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("startsWith"), startsWith, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("endsWith"), endsWith, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("includes"), includes, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("codePointAt"), codePointAt, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("toLowerCase"), toLowerCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("toUpperCase"), toUpperCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("toLocaleLowerCase"), toLowerCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("toLocaleUpperCase"), toUpperCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("trim"), trim, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("trimEnd"), trimEnd, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("trimStart"), trimStart, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("substring"), substring, 2)

	return stringPrototype
}

func StringIndexOf(s string, searchString string, position int) int {
	if position < 0 {
		position = 0
	}
	if position > len(s) {
		return -1
	}
	index := strings.Index(s[position:], searchString)
	if index < 0 {
		return -1
	}
	return position + index
}
