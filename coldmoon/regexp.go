package coldmoon

import (
	"strings"

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

	dotAll := func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "s").Data()
	}
	global := func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "g").Data()
	}
	hasIndices := func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "d").Data()
	}
	ignoreCase := func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "i").Data()
	}
	multiline := func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "m").Data()
	}
	sticky := func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "y").Data()
	}
	unicode := func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "u").Data()
	}
	unicodeSets := func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "v").Data()
	}
	flags := func(this Value, arguments []Value, _ ObjectType) Value {
		if !ValueIsObject(this) {
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
		if r.Object.Get(NewStringPropertyKey("dotAll")).ToBoolean() {
			f += "s"
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
		if r.Object.Get(NewStringPropertyKey("sticky")).ToBoolean() {
			f += "y"
		}
		if r.Object.Get(NewStringPropertyKey("unicode")).ToBoolean() {
			f += "u"
		}
		if r.Object.Get(NewStringPropertyKey("unicodeSets")).ToBoolean() {
			f += "v"
		}
		return NewStringValue(f)
	}
	source := func(this Value, arguments []Value, _ ObjectType) Value {
		if !ValueIsObject(this) {
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
	toString := func(this Value, arguments []Value, _ ObjectType) Value {
		if !ValueIsObject(this) {
			return agent.ThrowException(TypeError, "RegExp.prototype.toString: 'this' is not an object")
		}
		r, ok := MustGetObject(this).(*RegExpObject)
		if !ok {
			return agent.ThrowException(TypeError, "RegExp.prototype.toString: 'this' is not a RegExp object")
		}
		src := ToString(agent, r.Get(NewStringPropertyKey("source")))
		_flags := ToString(agent, r.Get(NewStringPropertyKey("flags")))
		return NewStringValue("/" + EscapeRegExpPattern(src.String(), _flags.String()) + "/")
	}
	exec := func(this Value, arguments []Value, _ ObjectType) Value {
		if !ValueIsObject(this) {
			return agent.ThrowException(TypeError, "RegExp.prototype.exec: 'this' is not an object")
		}
		r, ok := MustGetObject(this).(*RegExpObject)
		if !ok {
			return agent.ThrowException(TypeError, "RegExp.prototype.exec: 'this' is not a RegExp object")
		}
		s := ToString(agent, arguments[0])
		result := RegExpBuiltinExec(agent, r, s.Data)
		if result.IsNull() {
			return NullValue
		} else if result.IsError() {
			panic("RegExp.prototype.exec: error")
		} else {
			return (result.Data()).ToValue()
		}
	}
	test := func(this Value, arguments []Value, _ ObjectType) Value {
		if !ValueIsObject(this) {
			return agent.ThrowException(TypeError, "RegExp.prototype.test: 'this' is not an object")
		}
		r, ok := MustGetObject(this).(*RegExpObject)
		if !ok {
			return agent.ThrowException(TypeError, "RegExp.prototype.test: 'this' is not a RegExp object")
		}
		s := ToString(agent, arguments[0])
		m := RegExpExec(agent, r, s.Data)
		if m.IsNull() {
			return FalseValue
		}
		return TrueValue
	}
	search := func(this Value, arguments []Value, _ ObjectType) Value {
		rx := this
		if !ValueIsObject(rx) {
			panic("TypeError")
		}
		rxObject := MustGetObject(rx)
		S := ToString(agent, arguments[0])
		previousLastIndex := rxObject.Get(NewStringPropertyKey("lastIndex"))
		if !SameValue(previousLastIndex, NewNumberValue(0)) {
			rxObject.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(0), setThrowTypeThrow)
		}
		result := RegExpExec(agent, rxObject.(*RegExpObject), S.Data)
		currentLastIndex := rxObject.Get(NewStringPropertyKey("lastIndex"))
		if !SameValue(currentLastIndex, previousLastIndex) {
			rxObject.Set(NewStringPropertyKey("lastIndex"), previousLastIndex, setThrowTypeThrow)
		}
		if result.IsNull() {
			return NewNumberValue(-1)
		}
		return result.Data().Get(NewStringPropertyKey("index"))
	}
	matchAll := func(this Value, arguments []Value, _ ObjectType) Value {
		if !ValueIsObject(this) {
			panic("TypeError")
		}
		rx, ok := this.(*ObjectValue)
		if !ok {
			panic("TypeError")
		}
		r := rx.Object
		s := ToString(agent, arguments[0])
		c := r.SpeciesConstructor(realm.Intrinsics.RegExpConstructor)
		_flags := ToString(agent, r.Get(NewStringPropertyKey("flags")))
		matcher := c.Data().Construct([]Value{this, _flags}, nil)
		lastIndex := ToLength(agent, r.Get(NewStringPropertyKey("lastIndex")))
		matcher.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(JSNumber(lastIndex)), setThrowTypeThrow)
		_global := strings.Contains(_flags.Data, "g")
		_fullUnicode := strings.Contains(_flags.Data, "u") || strings.Contains(_flags.Data, "v")
		return CreateRegExpStringIterator(agent, matcher.(*RegExpObject), s.Data, _global, _fullUnicode).ToValue()
	}

	DefineBuiltinFunction(object, "toString", toString, 0, realm)
	DefineBuiltinFunction(object, "exec", exec, 1, realm)
	DefineBuiltinFunction(object, "test", test, 1, realm)
	DefineBuiltinFunction(object, "@@search", search, 1, realm)
	DefineBuiltinFunction(object, "@@matchAll", matchAll, 1, realm)
	DefineBuiltinAccessor(realm, object, "dotAll", dotAll, nil)
	DefineBuiltinAccessor(realm, object, "global", global, nil)
	DefineBuiltinAccessor(realm, object, "hasIndices", hasIndices, nil)
	DefineBuiltinAccessor(realm, object, "ignoreCase", ignoreCase, nil)
	DefineBuiltinAccessor(realm, object, "multiline", multiline, nil)
	DefineBuiltinAccessor(realm, object, "sticky", sticky, nil)
	DefineBuiltinAccessor(realm, object, "unicode", unicode, nil)
	DefineBuiltinAccessor(realm, object, "unicodeSets", unicodeSets, nil)
	DefineBuiltinAccessor(realm, object, "flags", flags, nil)
	DefineBuiltinAccessor(realm, object, "source", source, nil)

	return object
}

func NewRegExpConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, target ObjectType) Value {
		pattern := argumentsList[0]
		flags := argumentsList[1]
		patternIsRegexp := IsRegExp(pattern)
		var newTarget ObjectType
		if target == nil {
			newTarget = agent.ActiveFunctionObject()
			if patternIsRegexp && flags == UndefinedValue {
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
		if ValueIsObject(pattern) {
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
		return RegExpInitialize(agent, o, p, f).Data().ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 2, "RegExp", builtinFunctionArgs{
		realm:         realm,
		isConstructor: true,
		prototype:     realm.Intrinsics.FunctionPrototype,
	})

	DefineBuiltinAccessorV2(realm, object, BuiltinAccessorParams{
		Getter: func(this Value, argumentsList []Value, newTarget ObjectType) Value {
			return this
		},
		WellKnownSymbolsKey: WellKnownSymbolsSpecies,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        (NewRegExpPrototype(realm)).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.RegExpPrototype, "constructor", (object).ToValue())

	return object
}

// 22.2.3.1
func RegExpCreate(agent *Agent, pattern Value, flags Value) CompletionObject {
	realm := agent.CurrentRealm()

	obj := RegExpAlloc(agent, realm.Intrinsics.RegExpConstructor)
	return RegExpInitialize(agent, obj, pattern, flags)
}

// 22.2.6.4.1
func RegExpHasFlag(agent *Agent, R Value, flag string) CompletionValue {
	if !ValueIsObject(R) {
		return agent.ThrowException(TypeError, "RegExpHasFlag: R is not an object").ToCompletion()
	}
	r, ok := MustGetObject(R).(*RegExpObject)
	if !ok {
		return agent.ThrowException(TypeError, "RegExpHasFlag: R is not a RegExp object").ToCompletion()
	}
	var f bool
	switch flag {
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
	// TODO
	if F == "" {
		return P
	}
	if F == "u" {
		return P
	}
	return P
}

type MatchRecord struct {
	StartIndex JSInt
	EndIndex   JSInt
}

// 22.2.7.1
func RegExpExec(agent *Agent, regExp *RegExpObject, s string) CompletionObject {
	exec := regExp.Get(NewStringPropertyKey("exec"))
	if IsCallable(exec) {
		result := exec.CallAssumeCallable((regExp).ToValue(), []Value{NewStringValue(s)})
		if !ValueIsObject(result) && result != nil {
			return NewCompletionObjectError(agent.ThrowException(TypeError, "RegExpExec: exec is not an object"))
		}
		return NewCompletionObject(MustGetObject(result))
	}
	return RegExpBuiltinExec(agent, regExp, s)
}

// 22.2.7.2
// return null, object, or throw
func RegExpBuiltinExec(agent *Agent, regExp *RegExpObject, s string) CompletionObject {
	length := len(s)
	lastIndex := int(ToLength(agent, regExp.Get(NewStringPropertyKey("lastIndex"))))

	flags := regExp.OriginalFlags
	global := strings.Contains(flags, "g")
	sticky := strings.Contains(flags, "y")
	hasIndices := strings.Contains(flags, "d")

	if !global && !sticky {
		lastIndex = 0
	}

	matcher := regExp.RegExpMatcher
	fullUnicode := strings.Contains(flags, "u") || strings.Contains(flags, "v")
	matchSucceeded := false
	input := s
	if fullUnicode {
		input = s
	}

	match, err := matcher.FindStringMatch(input)
	if err != nil {
		return NewCompletionObjectError(NewStringValue(err.Error()))
	}
	var matchRecord *MatchRecord
	for !matchSucceeded {
		if lastIndex > length {
			if global || sticky {
				regExp.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(0), setThrowTypeIgnore)
			}
			return NewCompletionObjectNull()
		}
		match, err = matcher.FindNextMatch(match)
		if err != nil {
			if sticky {
				regExp.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(0), setThrowTypeThrow)
				return NewCompletionObjectNull()
			}
			lastIndex++
			return NewCompletionObjectError(NewStringValue(err.Error()))
		} else {
			matchSucceeded = true
			matchRecord = &MatchRecord{
				StartIndex: JSInt(match.Index),
				EndIndex:   JSInt(match.Index + match.Length),
			}
		}
	}
	e := matchRecord.EndIndex
	if fullUnicode {
		// TODO: GetStringIndex
	}
	if global || sticky {
		regExp.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(JSNumber(e)), setThrowTypeIgnore)
	}
	n := JSInt(len(match.Captures))
	Assert(float64(n) < POW_2_32-1)

	A := ArrayCreate(agent, n+1, nil)
	A.CreateDataPropertyOrThrow(
		NewStringPropertyKey("index"),
		NewNumberValue(matchRecord.StartIndex.ToNumber()))
	A.CreateDataPropertyOrThrow(NewStringPropertyKey("input"), NewStringValue(s))

	indices := make([]*MatchRecord, n)
	indices = append(indices, matchRecord)

	matchedSubstr := NewStringValue(GetMatchString(agent, s, matchRecord))
	A.CreateDataPropertyOrThrow(NewStringPropertyKey("0"), matchedSubstr)

	var groups Value
	var hasGroups bool
	if regExp.RegExpRecord.CapturingGroupsCount > 0 {
		groups = (OrdinaryObjectCreate(agent, nil, nil)).ToValue()
		hasGroups = true
	} else {
		groups = UndefinedValue
		hasGroups = false
	}

	A.CreateDataPropertyOrThrow(NewStringPropertyKey("groups"), groups)
	i := JSInt(1)
	groupNames := make([]string, n-1)
	for i < n {
		var captureI *MatchRecord
		var capturedValue Value
		if i >= JSInt(len(match.Captures)) {
			indices = append(indices, nil)
			capturedValue = UndefinedValue
		} else {
			captureI = &MatchRecord{
				StartIndex: JSInt(match.Captures[i].Index),
				EndIndex:   JSInt(match.Captures[i].Index + match.Captures[i].Length),
			}
			capturedValue = NewStringValue(GetMatchString(agent, s, captureI))
			indices = append(indices, captureI)
		}
		A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(i), capturedValue)

		if hasGroups {
			groupName := match.Groups()[i].Name
			MustGetObject(groups).CreateDataPropertyOrThrow(NewStringPropertyKey(groupName), capturedValue)
			groupNames[i-1] = groupName
		} else {
			groupNames[i-1] = ""
		}

		i++
	}

	if hasIndices {
		indicesArray := MakeMatchIndicesIndexPairArray(agent, s, indices, groupNames, hasGroups)
		A.CreateDataPropertyOrThrow(NewStringPropertyKey("indices"), (indicesArray).ToValue())
	}
	return NewCompletionObject(A)
}

// 22.2.7.3
func AdvanceStringIndex(s string, index JSInt, unicode bool) JSInt {
	if !unicode {
		return index + 1
	}
	length := JSInt(len(s))
	if index+1 >= length {
		return index + 1
	}
	// TODO: code point at
	cp := s[index]
	return JSInt(cp) + index
}

// 22.2.7.6
func GetMatchString(agent *Agent, s string, match *MatchRecord) string {
	Assert(match.StartIndex <= match.EndIndex)
	return s[match.StartIndex:match.EndIndex]
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
		var matchIndices *MatchRecord
		if i < len(indices) {
			matchIndices = indices[i]
			matchIndexPair = (GetMatchIndexPair(agent, s, matchIndices)).ToValue()
		}
		A.CreateDataPropertyOrThrow(NewStringPropertyKey(groupNames[i]), matchIndexPair)
		if i > 0 && groupNames[i-1] != "" {
			MustGetObject(groups).CreateDataPropertyOrThrow(NewStringPropertyKey(groupNames[i-1]), matchIndexPair)
		}
		i++
	}
	return A
}
