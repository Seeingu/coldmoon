package coldmoon

type RegExpStringIteratorObject struct {
	*Object
	RegExp      *RegExpObject
	string      string
	Global      bool
	FullUnicode bool
	Completed   bool
}

// 22.2.9.1
func CreateRegExpStringIterator(agent *Agent, R *RegExpObject, S string, global bool, fullUnicode bool) ObjectType {
	realm := agent.CurrentRealm()

	return &RegExpStringIteratorObject{
		Object:      NewObject(agent, realm.Intrinsics.RegExpStringIteratorPrototype, "RegExpStringIterator"),
		RegExp:      R,
		string:      S,
		Global:      global,
		FullUnicode: fullUnicode,
	}
}

func NewRegExpStringIteratorPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.IteratorPrototype, "RegExpStringIteratorPrototype")
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		iterator := MustGetObject(thisValue).(*RegExpStringIteratorObject)
		if iterator.Completed {
			return (CreateIterResultObject(agent, UndefinedValue, true)).ToValue()
		}

		match := RegExpExec(agent, iterator.RegExp, iterator.string)
		if match.IsNull() {
			iterator.Completed = true
			return (CreateIterResultObject(agent, UndefinedValue, true)).ToValue()
		}
		if !iterator.Global {
			iterator.Completed = true
			return (CreateIterResultObject(agent, (match.Data()).ToValue(), false)).ToValue()
		}

		matchStr := ToString(agent, match.Data().Get(NewStringPropertyKey("0")))
		if matchStr.Data == "" {
			thisIndex := ToLength(agent, iterator.RegExp.Get(NewStringPropertyKey("lastIndex")))
			nextIndex := AdvanceStringIndex(iterator.string, thisIndex, iterator.FullUnicode)
			iterator.RegExp.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(nextIndex.ToNumber()), setThrowTypeThrow)
		}
		return (CreateIterResultObject(agent, (match.Data()).ToValue(), false)).ToValue()
	}
	object.defineBuiltinFunction(realm, CMString("next"), next, 0)
	object.defineToStringTag("RegExp String Iterator")
	return object
}
