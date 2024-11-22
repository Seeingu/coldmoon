package coldmoon

import "github.com/dlclark/regexp2"

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
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)
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
				if SameValue(newTarget.ToValue(), patternConstructor) {
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
		return NewValueFromObject(
			RegExpInitialize(agent, o, p, f).Object,
		)

	}
	object := CreateBuiltinFunction(agent, behavior, 2, "RegExp", builtinFunctionArgs{
		realm:         realm,
		isConstructor: true,
		prototype:     realm.Intrinsics.FunctionPrototype,
	})

	var getter = func(this Value, arguments []Value, _ ObjectType) Value {
		return this
	}
	var dotAll = func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "s").Value
	}
	var global = func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "g").Value
	}
	var hasIndices = func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "d").Value
	}
	var ignoreCase = func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "i").Value
	}
	var multiline = func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "m").Value
	}
	var sticky = func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "y").Value
	}
	var unicode = func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "u").Value
	}
	var unicodeSets = func(this Value, arguments []Value, _ ObjectType) Value {
		return RegExpHasFlag(agent, this, "v").Value
	}
	var flags = func(this Value, arguments []Value, _ ObjectType) Value {
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
	var source = func(this Value, arguments []Value, _ ObjectType) Value {
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
	var toString = func(this Value, arguments []Value, _ ObjectType) Value {
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

	DefineBuiltinFunction(object, "toString", toString, 0, realm)

	DefineBuiltinAccessor(realm, object, "@@species", getter, nil)
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

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewRegExpPrototype(realm).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.RegExpPrototype, "constructor", object.ToValue())

	return object
}

// 22.2.6.4.1
func RegExpHasFlag(agent *Agent, R Value, flag string) *CompletionValue {
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
