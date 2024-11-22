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
	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewRegExpPrototype(realm).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.RegExpPrototype, "constructor", object.ToValue())

	return object
}
