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
	case *ReferenceRecordBaseUnresolvable:
		return false
	case *ReferenceRecordBaseValue, *ReferenceRecordBaseEnvironment:
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
func (r *ReferenceRecord) PutValue() {

}

// 6.2.5.7
func (r *ReferenceRecord) GetThisValue() Value {
	Assert(r.IsPropertyReference())

	if r.IsSuperReference() {
		return r.ThisValue
	}
	return r.Base.(*ReferenceRecordBaseValue).Value
}
