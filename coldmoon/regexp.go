package coldmoon

import (
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/Seeingu/coldmoon/pkg"

	"github.com/dlclark/regexp2"
)

type RegExpObject struct {
	*Object
	OriginalSource string
	OriginalFlags  string
	RegExpRecord   *RegExpRecord
	RegExpMatcher  *regexp2.Regexp
}

type RegExpRecord struct {
	HasIndices           bool
	Global               bool
	IgnoreCase           bool
	Multiline            bool
	Unicode              bool
	DotAll               bool
	UnicodeSets          bool
	Sticky               bool
	CapturingGroupsCount int
}

func NewRegExpPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "RegExp")

	dotAll := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		flag := RegExpHasFlag(agent, this, "s")
		return flag.Data()
	}
	global := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		flag := RegExpHasFlag(agent, this, "g")
		return flag.Data()
	}
	hasIndices := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		flag := RegExpHasFlag(agent, this, "d")
		return flag.Data()
	}
	ignoreCase := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		flag := RegExpHasFlag(agent, this, "i")
		return flag.Data()
	}
	multiline := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		flag := RegExpHasFlag(agent, this, "m")
		return flag.Data()
	}
	sticky := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		flag := RegExpHasFlag(agent, this, "y")
		return flag.Data()
	}
	unicode := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		flag := RegExpHasFlag(agent, this, "u")
		return flag.Data()
	}
	unicodeSets := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		flag := RegExpHasFlag(agent, this, "v")
		return flag.Data()
	}
	flags := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		if !this.IsObject() {
			return agent.ThrowException(TypeError, "RegExp.prototype.flags: 'this' is not an object")
		}
		r, ok := MustGetObject(this).(*RegExpObject)
		if !ok {
			return agent.ThrowException(TypeError, "RegExp.prototype.flags: 'this' is not a RegExp object")
		}
		var f string
		if r.Object.Get(NewStringPropertyKey("hasIndices")).ToBoolean() {
			f += "d"
		}
		if r.Object.Get(NewStringPropertyKey("global")).ToBoolean() {
			f += "g"
		}
		if r.Object.Get(NewStringPropertyKey("ignoreCase")).ToBoolean() {
			f += "i"
		}
		if r.Object.Get(NewStringPropertyKey("multiline")).ToBoolean() {
			f += "m"
		}
		if r.Object.Get(NewStringPropertyKey("dotAll")).ToBoolean() {
			f += "s"
		}
		if r.Object.Get(NewStringPropertyKey("unicode")).ToBoolean() {
			f += "u"
		}
		if r.Object.Get(NewStringPropertyKey("unicodeSets")).ToBoolean() {
			f += "v"
		}
		if r.Object.Get(NewStringPropertyKey("sticky")).ToBoolean() {
			f += "y"
		}
		return NewStringValue(f)
	}
	source := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		if !this.IsObject() {
			return agent.ThrowException(TypeError, "RegExp.prototype.source: 'this' is not an object")
		}
		r, ok := MustGetObject(this).(*RegExpObject)
		if !ok {
			return agent.ThrowException(TypeError, "RegExp.prototype.source: 'this' is not a RegExp object")
		}
		src := r.OriginalSource
		_flags := r.OriginalFlags
		return NewStringValue(EscapeRegExpPattern(src, _flags))
	}
	toString := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		if !this.IsObject() {
			return agent.ThrowException(TypeError, "RegExp.prototype.toString: 'this' is not an object")
		}
		r, ok := MustGetObject(this).(*RegExpObject)
		if !ok {
			return agent.ThrowException(TypeError, "RegExp.prototype.toString: 'this' is not a RegExp object")
		}
		src := ToString(agent, r.Get(NewStringPropertyKey("source")))
		_flags := ToString(agent, r.Get(NewStringPropertyKey("flags")))
		return NewStringValue("/" + src.String() + "/" + _flags.String())
	}
	exec := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		if !this.IsObject() {
			return agent.ThrowException(TypeError, "RegExp.prototype.exec: 'this' is not an object")
		}
		r, ok := MustGetObject(this).(*RegExpObject)
		if !ok {
			return agent.ThrowException(TypeError, "RegExp.prototype.exec: 'this' is not a RegExp object")
		}
		s := ToString(agent, argumentAt(arguments, 0))
		result := RegExpBuiltinExec(agent, r, s.Data)
		if result.Data() == NullValue {
			return NullValue
		} else if result.IsError() {
			panic("RegExp.prototype.exec: error")
		} else {
			return result.Data()
		}
	}
	test := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		if !this.IsObject() {
			return agent.ThrowException(TypeError, "RegExp.prototype.test: 'this' is not an object")
		}
		r, ok := MustGetObject(this).(*RegExpObject)
		if !ok {
			return agent.ThrowException(TypeError, "RegExp.prototype.test: 'this' is not a RegExp object")
		}
		s := ToString(agent, argumentAt(arguments, 0))
		m := RegExpExec(agent, r, s.Data)
		if m.Data() == NullValue {
			return FalseValue
		}
		return TrueValue
	}
	search := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		rx := this
		if !rx.IsObject() {
			panic("TypeError")
		}
		rxObject := MustGetObject(rx)
		S := ToString(agent, argumentAt(arguments, 0))
		previousLastIndex := rxObject.Get(NewStringPropertyKey("lastIndex"))
		if !SameValue(previousLastIndex, NewNumberValue(0)) {
			rxObject.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(0), setThrowTypeThrow)
		}
		result := RegExpExec(agent, rxObject.(*RegExpObject), S.Data)
		currentLastIndex := rxObject.Get(NewStringPropertyKey("lastIndex"))
		if !SameValue(currentLastIndex, previousLastIndex) {
			rxObject.Set(NewStringPropertyKey("lastIndex"), previousLastIndex, setThrowTypeThrow)
		}
		if result.Data() == NullValue {
			return NewNumberValue(-1)
		}
		return MustGetObject(result.Data()).Get(NewStringPropertyKey("index"))
	}
	match := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		rxValue, ok := this.(*ObjectValue)
		if !ok {
			return co.ThrowTypeError(agent, "RegExp match receiver is not an object")
		}
		rx, ok := rxValue.Object.(*RegExpObject)
		if !ok {
			return co.ThrowTypeError(agent, "RegExp match receiver is not a RegExp")
		}
		input := pkg.SliceSafeGet(arguments, 0)
		if input == nil {
			input = UndefinedValue
		}
		s, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, input), co)
		if isAbrupt {
			return rt
		}
		if !rx.Get(NewStringPropertyKey("global")).ToBoolean() {
			return RegExpExec(agent, rx, s.Data)
		}
		fullUnicode := rx.Get(NewStringPropertyKey("unicode")).ToBoolean() ||
			rx.Get(NewStringPropertyKey("unicodeSets")).ToBoolean()
		_, isAbrupt, rt = ReturnIfAbrupt(
			rx.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(0), setThrowTypeThrow),
			co,
		)
		if isAbrupt {
			return rt
		}
		result := ArrayCreate(agent, 0, nil)
		count := JSInt(0)
		for {
			next, isAbrupt, rt := ReturnIfAbrupt(RegExpExec(agent, rx, s.Data), co)
			if isAbrupt {
				return rt
			}
			if next == NullValue {
				if count == 0 {
					return NullValue
				}
				return result.ToValue()
			}
			nextObject := MustGetObject(next)
			matched, isAbrupt, rt := ReturnIfAbrupt(
				ToStringCompletion(agent, nextObject.Get(NewStringPropertyKey("0"))),
				co,
			)
			if isAbrupt {
				return rt
			}
			result.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(count), matched)
			count++
			if matched.Data == "" {
				lastIndex, isAbrupt, rt := ReturnIfAbrupt(
					ToLength(agent, rx.Get(NewStringPropertyKey("lastIndex"))),
					co,
				)
				if isAbrupt {
					return rt
				}
				nextIndex := AdvanceStringIndex(s.Data, lastIndex, fullUnicode)
				_, isAbrupt, rt = ReturnIfAbrupt(
					rx.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(nextIndex.ToNumber()), setThrowTypeThrow),
					co,
				)
				if isAbrupt {
					return rt
				}
			}
		}
	}
	matchAll := func(this Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		if !this.IsObject() {
			panic("TypeError")
		}
		rx, ok := this.(*ObjectValue)
		if !ok {
			panic("TypeError")
		}
		r := rx.Object
		s := ToString(agent, argumentAt(arguments, 0))
		c := r.SpeciesConstructor(realm.Intrinsics.RegExpConstructor)
		_flags := ToString(agent, r.Get(NewStringPropertyKey("flags")))
		matcher := c.Data().Construct([]Value{this, _flags}, nil).value
		lastIndex := ReturnAssertNormal(ToLength(agent, r.Get(NewStringPropertyKey("lastIndex"))))
		matcher.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(JSNumber(lastIndex)), setThrowTypeThrow)
		_global := strings.Contains(_flags.Data, "g")
		_fullUnicode := strings.Contains(_flags.Data, "u") || strings.Contains(_flags.Data, "v")
		return CreateRegExpStringIterator(agent, matcher.(*RegExpObject), s.Data, _global, _fullUnicode).ToValue()
	}

	object.defineBuiltinFunction(realm, CMString("toString"), toString, 0)
	object.defineBuiltinFunction(realm, CMString("exec"), exec, 1)
	object.defineBuiltinFunction(realm, CMString("test"), test, 1)
	object.defineBuiltinFunction(realm, WellKnownSymbolsSearch, search, 1)
	object.defineBuiltinFunction(realm, WellKnownSymbolsMatch, match, 1)
	object.defineBuiltinFunction(realm, WellKnownSymbolsMatchAll, matchAll, 1)
	object.defineBuiltinAccessor(realm, CMString("dotAll"), builtinAccessorParams{
		Getter: dotAll,
	})
	object.defineBuiltinAccessor(realm, CMString("global"), builtinAccessorParams{
		Getter: global,
	})
	object.defineBuiltinAccessor(realm, CMString("hasIndices"), builtinAccessorParams{
		Getter: hasIndices,
	})
	object.defineBuiltinAccessor(realm, CMString("ignoreCase"), builtinAccessorParams{
		Getter: ignoreCase,
	})
	object.defineBuiltinAccessor(realm, CMString("multiline"), builtinAccessorParams{
		Getter: multiline,
	})
	object.defineBuiltinAccessor(realm, CMString("sticky"), builtinAccessorParams{
		Getter: sticky,
	})
	object.defineBuiltinAccessor(realm, CMString("unicode"), builtinAccessorParams{
		Getter: unicode,
	})
	object.defineBuiltinAccessor(realm, CMString("unicodeSets"), builtinAccessorParams{
		Getter: unicodeSets,
	})
	object.defineBuiltinAccessor(realm, CMString("flags"), builtinAccessorParams{
		Getter: flags,
	})
	object.defineBuiltinAccessor(realm, CMString("source"), builtinAccessorParams{
		Getter: source,
	})

	return object
}

func NewRegExpConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	// 22.2.4.1
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, target ObjectType) CompletionConvertable[Value] {
		pattern := pkg.SliceSafeGet(argumentsList, 0)
		if pattern == nil {
			pattern = UndefinedValue
		}
		flags := pkg.SliceSafeGet(argumentsList, 1)
		patternIsRegexp := IsRegExp(pattern)
		var newTarget ObjectType
		if target == nil {
			newTarget = agent.ActiveFunctionObject()
			if patternIsRegexp && IsUndefinedOrNil(flags) {
				patternConstructor := MustGetObject(pattern).Get(NewStringPropertyKey("constructor"))
				if SameValue((newTarget).ToValue(), patternConstructor) {
					return pattern
				}
			}
		} else {
			newTarget = target
		}
		var p Value
		var f Value
		if pattern.IsObject() {
			if regexp, ok := MustGetObject(pattern).(*RegExpObject); ok {
				p = NewStringValue(regexp.OriginalSource)
				if flags == UndefinedValue {
					f = NewStringValue(regexp.OriginalFlags)
				} else {
					f = flags
				}
			}
		} else if patternIsRegexp {
			p = pattern
			if flags == UndefinedValue {
				f = NewStringValue("")
			} else {
				f = flags
			}
		} else {
			p = pattern
			f = flags
		}
		o := RegExpAlloc(agent, target)
		initialize := RegExpInitialize(agent, o, p, f)
		return initialize.Data().ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 2, CMString("RegExp"), builtinFunctionArgs{
		realm:         realm,
		isConstructor: true,
		prototype:     realm.Intrinsics.FunctionPrototype,
	})

	object.defineBuiltinAccessor(realm, WellKnownSymbolsSpecies, builtinAccessorParams{
		Getter: func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
			return this
		},
	})

	BindPrototypeAndConstructor(realm.Intrinsics.RegExpPrototype, object)
	return object
}

// 22.2.3.1
func RegExpCreate(agent *Agent, pattern Value, flags Value) Completion[ObjectType] {
	realm := agent.CurrentRealm()

	obj := RegExpAlloc(agent, realm.Intrinsics.RegExpConstructor)
	return RegExpInitialize(agent, obj, pattern, flags)
}

// 22.2.6.4.1
func RegExpHasFlag(agent *Agent, R Value, flag string) CompletionValue {
	if !R.IsObject() {
		return agent.ThrowException(TypeError, "RegExpHasFlag: R is not an object").ToCompletion()
	}
	r, ok := MustGetObject(R).(*RegExpObject)
	if !ok {
		return agent.ThrowException(TypeError, "RegExpHasFlag: R is not a RegExp object").ToCompletion()
	}
	var f bool
	switch flag {
	case "d":
		f = r.RegExpRecord.HasIndices
	case "g":
		f = r.RegExpRecord.Global
	case "s":
		f = r.RegExpRecord.DotAll
	case "m":
		f = r.RegExpRecord.Multiline
	case "i":
		f = r.RegExpRecord.IgnoreCase
	case "u":
		f = r.RegExpRecord.Unicode
	case "v":
		f = r.RegExpRecord.UnicodeSets
	case "y":
		f = r.RegExpRecord.Sticky
	default:
		f = false
	}
	return NewBooleanValue(f).ToCompletion()
}

func EscapeRegExpPattern(P string, F string) string {
	// The flags participate in the specification's parse-text validation, but
	// the concrete source escapes below are identical for every flag set.
	_ = F
	if P == "" {
		return "(?:)"
	}
	replacer := strings.NewReplacer(
		"/", `\/`,
		"\n", `\n`,
		"\r", `\r`,
		"\u2028", `\u2028`,
		"\u2029", `\u2029`,
	)
	return replacer.Replace(P)
}

type MatchRecord struct {
	StartIndex JSInt
	EndIndex   JSInt
}

// 22.2.7.1
func RegExpExec(agent *Agent, regExp *RegExpObject, s string) (co Completion[Value]) {
	exec := regExp.Get(NewStringPropertyKey("exec"))
	if IsCallable(exec) {
		result := ReturnAssertNormal(
			exec.Call(agent, regExp.ToValue(), []Value{NewStringValue(s)}),
		)
		if !result.IsObject() && result != NullValue {
			co.err = agent.ThrowException(TypeError, "RegExpExec: exec is not an object")
			return
		}
		if result == NullValue {
			co.value = NullValue
			return
		}
		co.value = result
		return
	}
	return RegExpBuiltinExec(agent, regExp, s)
}

// 22.2.7.2
// return null, object, or throw
func RegExpBuiltinExec(agent *Agent, regExp *RegExpObject, s string) (co Completion[Value]) {
	codeUnits := utf16.Encode([]rune(s))
	length := JSInt(len(codeUnits))
	lastIndex := ReturnAssertNormal(ToLength(agent, regExp.Get(NewStringPropertyKey("lastIndex"))))

	flags := regExp.OriginalFlags
	global := strings.Contains(flags, "g")
	sticky := strings.Contains(flags, "y")
	hasIndices := strings.Contains(flags, "d")

	if !global && !sticky {
		lastIndex = 0
	}

	matcher := regExp.RegExpMatcher
	fullUnicode := strings.Contains(flags, "u") || strings.Contains(flags, "v")
	input := regexpInputCharacters(codeUnits, fullUnicode)
	var match *regexp2.Match
	for {
		if lastIndex > length {
			if global || sticky {
				_, isAbrupt, rt := ReturnIfAbrupt(
					regExp.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(0), setThrowTypeThrow),
					co,
				)
				if isAbrupt {
					return rt
				}
			}
			co.value = NullValue
			return
		}
		inputIndex := int(lastIndex)
		if fullUnicode {
			inputIndex = codePointIndexForStringIndex(codeUnits, lastIndex)
		}
		r, err := matcher.FindRunesMatchStartingAt(input, inputIndex)
		if err != nil {
			co.err = NewStringValue(err.Error())
			return
		}
		if r == nil || (sticky && r.Index != inputIndex) {
			if global || sticky {
				_, isAbrupt, rt := ReturnIfAbrupt(
					regExp.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(0), setThrowTypeThrow),
					co,
				)
				if isAbrupt {
					return rt
				}
			}
			co.value = NullValue
			return
		}
		match = r
		break
	}

	matchRecord := regexpMatchRecord(s, fullUnicode, match.Index, match.Length)
	lastIndex = matchRecord.StartIndex
	e := matchRecord.EndIndex
	if fullUnicode {
		Assert(e <= length)
	}
	if global || sticky {
		_, isAbrupt, rt := ReturnIfAbrupt(
			regExp.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(JSNumber(e)), setThrowTypeThrow),
			co,
		)
		if isAbrupt {
			return rt
		}
	}
	matchedGroups := match.Groups()
	n := JSInt(len(matchedGroups))
	Assert(int(n)-1 == regExp.RegExpRecord.CapturingGroupsCount)
	Assert(float64(n) < POW_2_32-1)

	A := ArrayCreate(agent, n, nil)
	A.CreateDataPropertyOrThrow(
		NewStringPropertyKey("index"),
		NewNumberValue(matchRecord.StartIndex.ToNumber()))
	A.CreateDataPropertyOrThrow(NewStringPropertyKey("input"), NewStringValue(s))

	indices := make([]*MatchRecord, n)
	indices[0] = matchRecord

	matchedSubstr := NewStringValue(GetMatchString(agent, s, matchRecord))
	A.CreateDataPropertyOrThrow(NewStringPropertyKey("0"), matchedSubstr)

	var groups Value
	var hasGroups bool
	for i := 1; i < len(matchedGroups); i++ {
		if _, err := strconv.Atoi(matchedGroups[i].Name); err != nil {
			hasGroups = true
			break
		}
	}
	if hasGroups {
		groups = (OrdinaryObjectCreate(agent, nil, nil)).ToValue()
	} else {
		groups = UndefinedValue
	}

	A.CreateDataPropertyOrThrow(NewStringPropertyKey("groups"), groups)
	i := JSInt(1)
	groupNames := make([]string, n-1)
	for i < n {
		var captureI *MatchRecord
		var capturedValue Value
		group := matchedGroups[i]
		if len(group.Captures) == 0 {
			capturedValue = UndefinedValue
		} else {
			captureI = regexpMatchRecord(s, fullUnicode, group.Index, group.Length)
			capturedValue = NewStringValue(GetMatchString(agent, s, captureI))
		}
		indices[i] = captureI
		A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(i), capturedValue)

		groupName := group.Name
		if _, err := strconv.Atoi(groupName); err != nil {
			MustGetObject(groups).CreateDataPropertyOrThrow(NewStringPropertyKey(groupName), capturedValue)
			groupNames[i-1] = groupName
		}

		i++
	}

	if hasIndices {
		indicesArray := MakeMatchIndicesIndexPairArray(agent, s, indices, groupNames, hasGroups)
		A.CreateDataPropertyOrThrow(NewStringPropertyKey("indices"), (indicesArray).ToValue())
	}
	co.value = A.ToValue()
	return
}

// 22.2.7.3
func AdvanceStringIndex(s string, index JSInt, unicode bool) JSInt {
	if !unicode {
		return index + 1
	}
	codeUnits := utf16.Encode([]rune(s))
	length := JSInt(len(codeUnits))
	if index+1 >= length {
		return index + 1
	}
	first := codeUnits[index]
	if first < 0xD800 || first > 0xDBFF {
		return index + 1
	}
	second := codeUnits[index+1]
	if second < 0xDC00 || second > 0xDFFF {
		return index + 1
	}
	return index + 2
}

// GetStringIndex converts a Unicode code point index into ECMAScript's
// UTF-16 code-unit index. Out-of-range code point indices map to the string
// length, as required by the RegExp abstract operation.
func GetStringIndex(s string, codePointIndex JSInt) JSInt {
	if codePointIndex <= 0 {
		return 0
	}
	codeUnitIndex := JSInt(0)
	for i, codePoint := range []rune(s) {
		if JSInt(i) == codePointIndex {
			return codeUnitIndex
		}
		if codePoint > 0xFFFF {
			codeUnitIndex += 2
		} else {
			codeUnitIndex++
		}
	}
	return codeUnitIndex
}

func regexpInputCharacters(codeUnits []uint16, fullUnicode bool) []rune {
	if fullUnicode {
		return utf16.Decode(codeUnits)
	}
	input := make([]rune, len(codeUnits))
	for i, codeUnit := range codeUnits {
		input[i] = rune(codeUnit)
	}
	return input
}

func codePointIndexForStringIndex(codeUnits []uint16, stringIndex JSInt) int {
	codePointIndex := 0
	for codeUnitIndex := JSInt(0); codeUnitIndex < JSInt(len(codeUnits)); codePointIndex++ {
		width := JSInt(1)
		if codeUnits[codeUnitIndex] >= 0xD800 && codeUnits[codeUnitIndex] <= 0xDBFF &&
			codeUnitIndex+1 < JSInt(len(codeUnits)) &&
			codeUnits[codeUnitIndex+1] >= 0xDC00 && codeUnits[codeUnitIndex+1] <= 0xDFFF {
			width = 2
		}
		if stringIndex < codeUnitIndex+width {
			return codePointIndex
		}
		codeUnitIndex += width
	}
	return codePointIndex
}

func regexpMatchRecord(s string, fullUnicode bool, index int, length int) *MatchRecord {
	start := JSInt(index)
	end := JSInt(index + length)
	if fullUnicode {
		start = GetStringIndex(s, start)
		end = GetStringIndex(s, end)
	}
	return &MatchRecord{StartIndex: start, EndIndex: end}
}

// 22.2.7.6
func GetMatchString(agent *Agent, s string, match *MatchRecord) string {
	Assert(match.StartIndex <= match.EndIndex)
	codeUnits := utf16.Encode([]rune(s))
	Assert(match.EndIndex <= JSInt(len(codeUnits)))
	return string(utf16.Decode(codeUnits[match.StartIndex:match.EndIndex]))
}

// 22.2.7.7
func GetMatchIndexPair(agent *Agent, s string, match *MatchRecord) ObjectType {
	Assert(match.StartIndex <= match.EndIndex)
	return CreateArrayFromList(agent, []Value{
		NewNumberValue(JSNumber(match.StartIndex)),
		NewNumberValue(JSNumber(match.EndIndex)),
	})
}

// 22.2.7.8
func MakeMatchIndicesIndexPairArray(agent *Agent, s string, indices []*MatchRecord, groupNames []string, hasGroups bool) ObjectType {
	n := len(indices)
	Assert(float64(n) < POW_2_32-1)
	Assert(len(groupNames) == n-1)
	A := ArrayCreate(agent, 0, nil)
	var groups Value
	if hasGroups {
		groups = (OrdinaryObjectCreate(agent, nil, nil)).ToValue()
	} else {
		groups = UndefinedValue
	}

	A.CreateDataPropertyOrThrow(NewStringPropertyKey("groups"), groups)
	i := 0
	for i < n {
		var matchIndexPair Value = UndefinedValue
		matchIndices := indices[i]
		if matchIndices != nil {
			matchIndexPair = (GetMatchIndexPair(agent, s, matchIndices)).ToValue()
		}
		A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(JSInt(i)), matchIndexPair)
		if i > 0 && groupNames[i-1] != "" {
			MustGetObject(groups).CreateDataPropertyOrThrow(NewStringPropertyKey(groupNames[i-1]), matchIndexPair)
		}
		i++
	}
	return A
}
