package coldmoon

type (
	ReferenceRecordBase      interface{}
	ReferenceRecordBaseValue struct {
		ReferenceRecordBase
		Value Value
	}
)

type ReferenceRecordBaseEnvironment struct {
	ReferenceRecordBase
	Environment EnvironmentRecord
}

type ReferenceRecordBaseUnresolvable struct {
	ReferenceRecordBase
}

// ReferencedName Enum
type ReferencedName struct {
	String      string
	Symbol      *SymbolValue
	PrivateName *PrivateName
}

type ReferenceRecord struct {
	Base           ReferenceRecordBase
	ReferencedName *ReferencedName
	Strict         bool
	ThisValue      Value
}

func NewReferenceRecord(base ReferenceRecordBase, referencedName *ReferencedName, strict bool, thisValue Value) *ReferenceRecord {
	return &ReferenceRecord{
		Base:           base,
		ReferencedName: referencedName,
		Strict:         strict,
		ThisValue:      thisValue,
	}
}

// 6.2.5.1
func (r *ReferenceRecord) IsPropertyReference() bool {
	switch r.Base.(type) {
	case *ReferenceRecordBaseUnresolvable, *ReferenceRecordBaseEnvironment:
		return false
	case *ReferenceRecordBaseValue:
		return true
	}
	panic("unreachable")
}

// 6.2.5.2
func (r *ReferenceRecord) IsUnresolvableReference() bool {
	_, ok := r.Base.(*ReferenceRecordBaseUnresolvable)
	return ok
}

// 6.2.5.3
func (r *ReferenceRecord) IsSuperReference() bool {
	return r.ThisValue != nil
}

// 6.2.5.4
func (r *ReferenceRecord) IsPrivateReference() bool {
	return r.ReferencedName.PrivateName != nil
}

// 6.2.5.5
func (r *ReferenceRecord) GetValue(agent *Agent) Value {
	if r.IsUnresolvableReference() {
		panic("ReferenceError")
	}
	if r.IsPropertyReference() {
		baseObj := MustGetObject(r.Base.(*ReferenceRecordBaseValue).Value)
		if r.IsPrivateReference() {
			return baseObj.PrivateGet(*r.ReferencedName.PrivateName)
		}

		var propKey PropertyKey
		if r.ReferencedName.PrivateName != nil {
			panic("unreachable")
		} else if r.ReferencedName.Symbol != nil {
			propKey = NewSymbolPropertyKey(r.ReferencedName.Symbol)
		} else {
			propKey = NewStringPropertyKey(r.ReferencedName.String)
		}
		return baseObj.InternalMethods().Get(baseObj, propKey, r.GetThisValue())
	} else {
		base := r.Base.(*ReferenceRecordBaseEnvironment)
		name := r.ReferencedName.String
		c := base.Environment.GetBindingValue(agent, name, r.Strict)
		return c.Data()
	}
}

// 6.2.5.6
func (r *ReferenceRecord) PutValue(agent *Agent, value Value) {
	if r.IsUnresolvableReference() {
		if r.Strict {
			panic("ReferenceError")
		}
		globalObj := agent.GetGlobalObject()
		globalObj.Set(NewStringPropertyKey(r.ReferencedName.String), value, setThrowTypeIgnore)
		return
	}

	if r.IsPropertyReference() {
		baseObj := ValueToObject(agent, r.Base.(*ReferenceRecordBaseValue).Value)

		if r.IsPrivateReference() {
			panic("implement me")
		}

		var referencedName PropertyKey
		if r.ReferencedName.PrivateName != nil {
			panic("unreachable")
		} else if r.ReferencedName.Symbol != nil {
			referencedName = NewSymbolPropertyKey(r.ReferencedName.Symbol)
		} else {
			referencedName = NewStringPropertyKey(r.ReferencedName.String)
		}

		succeeded := baseObj.InternalMethods().Set(
			baseObj,
			referencedName,
			value,
			r.GetThisValue())
		if !succeeded && r.Strict {
			panic("TypeError")
		}
		return
	}

	base := r.Base.(*ReferenceRecordBaseEnvironment)
	referencedName := r.ReferencedName.String
	base.Environment.SetMutableBinding(referencedName, value, r.Strict)
}

// 6.2.5.7
func (r *ReferenceRecord) GetThisValue() Value {
	Assert(r.IsPropertyReference())

	if r.IsSuperReference() {
		return r.ThisValue
	}
	return r.Base.(*ReferenceRecordBaseValue).Value
}

// 6.2.5.8
func (r *ReferenceRecord) InitializeReferencedBinding(value Value) {
	Assert(!r.IsUnresolvableReference())

	base := r.Base.(*ReferenceRecordBaseEnvironment).Environment
	base.InitializeBinding(r.ReferencedName.String, value)
}
