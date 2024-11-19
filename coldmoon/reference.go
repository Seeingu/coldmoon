package coldmoon

type ReferenceRecordBase interface {
}
type ReferenceRecordBaseValue struct {
	ReferenceRecordBase
	Value Value
}

type ReferenceRecordBaseEnvironment struct {
	ReferenceRecordBase
	Environment EnvironmentRecord
}

type ReferenceRecordBaseUnresolvable struct {
	ReferenceRecordBase
}

type ReferencedName interface {
}

type ReferencedNameString struct {
	ReferencedName
	String string
}

type ReferencedNameSymbol struct {
	ReferencedName
	Symbol *SymbolValue
}

type ReferencedNamePrivateName struct {
	ReferencedName
}

type ReferenceRecord struct {
	Base           ReferenceRecordBase
	ReferencedName ReferencedName
	Strict         bool
	ThisValue      Value
}

func NewReferenceRecord(base ReferenceRecordBase, referencedName ReferencedName, strict bool, thisValue Value) *ReferenceRecord {
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
	_, ok := r.ReferencedName.(*ReferencedNamePrivateName)
	return ok
}

// 6.2.5.5
func (r *ReferenceRecord) GetValue() Value {
	if r.IsUnresolvableReference() {
		panic("ReferenceError")
	}
	if r.IsPropertyReference() {
		baseObj := r.Base.(*ReferenceRecordBaseValue).Value.(*ObjectValue).Object
		if r.IsPrivateReference() {
			panic("implement me")
		}

		switch referencedName := r.ReferencedName.(type) {
		case *ReferencedNameString:
			return baseObj.Get(NewStringPropertyKey(referencedName.String))
		case *ReferencedNameSymbol:
			return baseObj.Get(NewSymbolPropertyKey(referencedName.Symbol))
		case *ReferencedNamePrivateName:
			panic("unreachable")
		}
		panic("unreachable")
	} else {
		base := r.Base.(*ReferenceRecordBaseEnvironment)
		return base.Environment.GetBindingValue(r.ReferencedName.(*ReferencedNameString).String, r.Strict)
	}
}

// 6.2.5.6
func (r *ReferenceRecord) PutValue(agent *Agent, value Value) {
	if r.IsUnresolvableReference() {
		if r.Strict {
			panic("ReferenceError")
		}

		globalObj := agent.GetGlobalObject()

		globalObj.Set(NewStringPropertyKey(r.ReferencedName.(*ReferencedNameString).String), value, setThrowTypeIgnore)
		return
	}

	if r.IsPropertyReference() {
		baseObj := ValueToObject(agent, r.Base.(*ReferenceRecordBaseValue).Value)

		if r.IsPrivateReference() {
			panic("implement me")
		}

		var referencedName PropertyKey
		switch rn := r.ReferencedName.(type) {
		case *ReferencedNameString:
			referencedName = NewStringPropertyKey(rn.String)
		case *ReferencedNameSymbol:
			referencedName = NewSymbolPropertyKey(rn.Symbol)
		case *ReferencedNamePrivateName:
			panic("unreachable")
		}

		succeeded := baseObj.InternalMethods().Set(
			baseObj,
			referencedName,
			value,
			r.ThisValue)
		if !succeeded && r.Strict {
			panic("TypeError")
		}
		return
	}

	base := r.Base.(*ReferenceRecordBaseEnvironment)
	referencedName := r.ReferencedName.(*ReferencedNameString).String
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
	base.InitializeBinding(r.ReferencedName.(*ReferencedNameString).String, value)
}
