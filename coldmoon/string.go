package coldmoon

import (
	"sort"
	"strings"

	"github.com/samber/lo"
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
	var defineOwnProperty DefineOwnPropertyFn = func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) bool {
		s := o.(*StringObject)
		stringDesc := StringGetOwnProperty(s, p)
		if stringDesc != nil {
			extensible := s.Extensible()
			return IsCompatiblePropertyDescriptor(extensible, desc, stringDesc)
		}
		return OrdinaryDefineOwnProperty(o, p, desc)
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
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		value := argumentsList[0]
		var s string
		if len(argumentsList) > 0 {
			symbolValue, isSymbol := value.(*SymbolValue)
			if newTarget == nil && isSymbol {
				return NewStringValue(symbolValue.SymbolDescriptiveString())
			}
			s = value.String()
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

	var fromCharCode BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		var s string
		for _, arg := range argumentsList {
			s += string(rune(int(ToIntegerOrInfinity(realm.Agent, arg))))
		}
		return NewStringValue(s)
	}
	var fromCodePoint BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		codePoints := make([]rune, len(argumentsList))
		for i, arg := range argumentsList {
			codePoints[i] = rune(ToIntegerOrInfinity(realm.Agent, arg))
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
	var toString BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		s := thisArgument.ThisStringValue()
		return NewStringValue(s)
	}
	var valueOf BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return NewStringValue(thisArgument.ThisStringValue())
	}
	toLowerCase := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		s := this.ThisStringValue()
		return NewStringValue(strings.ToLower(s))
	}
	toUpperCase := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		s := this.ThisStringValue()
		return NewStringValue(strings.ToUpper(s))
	}
	trim := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		s := this.ThisStringValue()
		return NewStringValue(strings.TrimSpace(s))
	}
	trimEnd := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		s := this.ThisStringValue()
		return NewStringValue(strings.TrimRightFunc(s, func(r rune) bool {
			return strings.ContainsRune(" \t\n\v\f\r", r)
		}))
	}
	trimStart := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		s := this.ThisStringValue()
		return NewStringValue(strings.TrimLeftFunc(s, func(r rune) bool {
			return strings.ContainsRune(" \t\n\v\f\r", r)
		}))
	}
	var charAt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		position := int(ToIntegerOrInfinity(agent, argumentsList[0]))
		size := len(s)
		if position < 0 || position >= size {
			return NewStringValue("")
		}
		return NewStringValue(string(s[position]))
	}
	var charCodeAt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		position := int(ToIntegerOrInfinity(agent, argumentsList[0]))
		size := len(s)
		if position < 0 || position >= size {
			return NaNValue
		}
		return NewNumberValue(JSNumber(s[position]))
	}
	var iterator BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		stringIteratorObject := &StringIteratorObject{
			Object: NewObject(agent, realm.Intrinsics.StringIteratorPrototype, "String Iterator"),
			Data:   s,
		}
		stringIteratorObject.ref = stringIteratorObject
		return (stringIteratorObject).ToValue()
	}
	var at BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		index := argumentsList[0]
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		length := len(s)
		relativeIndex := int(ToIntegerOrInfinity(agent, index))
		if relativeIndex < 0 || relativeIndex >= length {
			return UndefinedValue
		}
		k := relativeIndex
		return NewStringValue(string(s[k]))
	}
	var slice BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		intStart := ToIntegerOrInfinity(agent, argumentsList[0])
		intEnd := ToIntegerOrInfinity(agent, argumentsList[1])

		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		length := JSInt(len(s))

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
	var repeat BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		n := ToIntegerOrInfinity(agent, argumentsList[0])
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		if n < 0 || n.IsInf() {
			panic("RangeError")
		}
		if n == 0 {
			return NewStringValue("")
		}
		return NewStringValue(strings.Repeat(s, int(n)))
	}
	var concat BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		for _, arg := range argumentsList {
			s += arg.String()
		}
		return NewStringValue(s)
	}
	var search BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		regexp := argumentsList[0]
		o := RequireObjectCoercible(agent, thisArgument)
		if regexp != UndefinedValue && regexp != NullValue {
			searcher := GetMethod(agent, regexp, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsSearch]))
			if searcher != nil {
				return searcher.Call(regexp, []Value{o})
			}
		}

		s := ToString(agent, o)
		rx := RegExpCreate(agent, regexp, UndefinedValue)
		return ValueInvoke(agent, rx.Data().ToValue(), NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsSearch]), []Value{s})
	}
	var matchAll BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
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
				return matcher.Call(regexp, []Value{o})
			}
		}
		s := ToString(agent, o)
		rx := RegExpCreate(agent, regexp, UndefinedValue)
		return ValueInvoke(agent, rx.Data().ToValue(), NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsMatchAll]), []Value{s})
	}
	indexOf := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		searchString := argumentsList[0]
		position := ToIntegerOrInfinity(agent, argumentsList[1])
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		searchStr := ToString(agent, searchString)
		pos := int(position)
		length := len(s)
		start := lo.Clamp(pos, 0, length)
		return NewNumberValue(JSNumber(StringIndexOf(s, searchStr.Data, start)))
	}
	lastIndexOf := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		searchString := argumentsList[0]
		position := ToIntegerOrInfinity(agent, argumentsList[1])
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		searchStr := ToString(agent, searchString).Data
		pos := int(position)
		searchLen := len(s)
		start := lo.Clamp(pos, 0, searchLen)
		if len(searchStr) == 0 {
			return NewNumberValue(JSNumber(start))
		}
		return NewNumberValue(JSNumber(strings.LastIndex(s[start:], searchStr)))
	}
	startsWith := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		searchString := argumentsList[0]
		var position Value
		if len(argumentsList) > 1 {
			position = argumentsList[1]
		}
		o := RequireObjectCoercible(agent, this)
		s := ToString(agent, o)
		isRegExp := IsRegExp(searchString)
		if isRegExp {
			panic("TypeError")
		}
		searchStr := ToString(agent, searchString)
		length := len(s.Data)
		var pos int
		if position == nil {
			pos = 0
		} else {
			pos = int(ToIntegerOrInfinity(agent, position))
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
	endsWith := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		searchString := argumentsList[0]
		var position Value
		if len(argumentsList) > 1 {
			position = argumentsList[1]
		}
		o := RequireObjectCoercible(agent, thisArgument)
		s := ToString(agent, o)
		isRegExp := IsRegExp(searchString)
		if isRegExp {
			panic("TypeError")
		}
		searchStr := ToString(agent, searchString)
		length := len(s.Data)
		var pos int
		if position == nil {
			pos = length
		} else {
			pos = int(ToIntegerOrInfinity(agent, position))
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
	var includes BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		searchString := argumentsList[0]
		var position Value
		if len(argumentsList) > 1 {
			position = argumentsList[1]
		}
		o := RequireObjectCoercible(agent, this)
		s := ToString(agent, o)
		isRegExp := IsRegExp(searchString)
		if isRegExp {
			panic("TypeError")
		}
		searchStr := ToString(agent, searchString)
		length := len(s.Data)
		var pos int
		if position == nil {
			pos = 0
		} else {
			pos = int(ToIntegerOrInfinity(agent, position))
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
	var codePointAt BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		pos := argumentsList[0]
		o := RequireObjectCoercible(agent, this)
		s := ToString(agent, o)
		position := ToIntegerOrInfinity(agent, pos)
		size := len(s.Data)
		if position < 0 || int(position) >= size {
			return UndefinedValue
		}
		return NewNumberValue(JSNumber(s.Data[int(position)]))
	}
	substring := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		start := ToIntegerOrInfinity(agent, argumentsList[0])
		end := ToIntegerOrInfinity(agent, argumentsList[1])
		o := RequireObjectCoercible(agent, this)
		s := ToString(agent, o)
		length := JSInt(len(s.Data))
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
	stringPrototype.defineBuiltinFunction(realm, CMString("search"), search, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("matchAll"), matchAll, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("indexOf"), indexOf, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("lastIndexOf"), lastIndexOf, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("startsWith"), startsWith, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("endsWith"), endsWith, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("includes"), includes, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("codePointAt"), codePointAt, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("toLowerCase"), toLowerCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("toUpperCase"), toUpperCase, 0)
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
	stringPrototype.defineBuiltinFunction(realm, CMString("search"), search, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("matchAll"), matchAll, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("indexOf"), indexOf, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("lastIndexOf"), lastIndexOf, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("startsWith"), startsWith, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("endsWith"), endsWith, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("includes"), includes, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("codePointAt"), codePointAt, 1)
	stringPrototype.defineBuiltinFunction(realm, CMString("toLowerCase"), toLowerCase, 0)
	stringPrototype.defineBuiltinFunction(realm, CMString("toUpperCase"), toUpperCase, 0)
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
	return strings.Index(s[position:], searchString)
}
