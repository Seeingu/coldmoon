package coldmoon

type RegExpStringIteratorObject struct {
	*Object
	RegExp      *RegExpObject
	String      string
	Global      bool
	FullUnicode bool
	Completed   bool
}

// 22.2.9.1
func CreateRegExpStringIterator(agent *Agent, R *RegExpObject, S string, global bool, fullUnicode bool) ObjectType {
	realm := agent.CurrentRealm()

	return &RegExpStringIteratorObject{
		Object:      NewObject(agent, realm.Intrinsics.RegExpStringIteratorPrototype),
		RegExp:      R,
		String:      S,
		Global:      global,
		FullUnicode: fullUnicode,
	}
}

func NewRegExpStringIteratorPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.IteratorPrototype)
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		iterator := MustGetObject(thisValue).(*RegExpStringIteratorObject)
		if iterator.Completed {
			return NewValueFromObject(CreateIterResultObject(agent, UndefinedValue, true))
		}

		match := RegExpExec(agent, iterator.RegExp, iterator.String)
		if match.IsNull() {
			iterator.Completed = true
			return NewValueFromObject(CreateIterResultObject(agent, UndefinedValue, true))
		}
		if !iterator.Global {
			iterator.Completed = true
			return NewValueFromObject(CreateIterResultObject(agent, match.Data().ToValue(), false))
		}

		matchStr := ToString(agent, match.Data().Get(NewStringPropertyKey("0")))
		if matchStr.Data == "" {
			thisIndex := ToLength(agent, iterator.RegExp.Get(NewStringPropertyKey("lastIndex")))
			nextIndex := AdvanceStringIndex(iterator.String, thisIndex, iterator.FullUnicode)
			iterator.RegExp.Set(NewStringPropertyKey("lastIndex"), NewNumberValue(float64(nextIndex)), setThrowTypeThrow)
		}
		return NewValueFromObject(CreateIterResultObject(agent, match.Data().ToValue(), false))
	}
	DefineBuiltinFunction(object, "next", next, 0, realm)
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("RegExp String Iterator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object

}
