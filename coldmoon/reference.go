package coldmoon

// ReferencedName Enum
type ReferencedName struct {
	String      string
	Symbol      *SymbolValue
	PrivateName *PrivateName
}

type ReferenceRecord struct {
	// [[Base]]
	Base ReferenceRecordBase
	// [[ReferencedName]]
	ReferencedName *ReferencedName
	// [[Strict]]
	Strict bool
	// [[ThisValue]]
	ThisValue Value
}

type ReferenceRecordValue struct {
	Value
	record *ReferenceRecord
}

func NewReferenceRecordValue(referenceRecord *ReferenceRecord) *ReferenceRecordValue {
	r := &ReferenceRecordValue{
		record: referenceRecord,
	}
	r.Value = NewBaseValue(r)
	return r
}

func (r *ReferenceRecordValue) String() string {
	return "ReferenceRecordValue"
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
	_, ok := r.Base.Value()
	return ok
}

// 6.2.5.2
func (r *ReferenceRecord) IsUnresolvableReference() bool {
	return r.Base.IsUnresolvable()
}

// 6.2.5.3
func (r *ReferenceRecord) IsSuperReference() bool {
	return !IsUndefinedOrNil(r.ThisValue)
}

// 6.2.5.4
func (r *ReferenceRecord) IsPrivateReference() bool {
	return r.ReferencedName.PrivateName != nil
}

// GetValue
// spec: 6.2.5.5
func (r *ReferenceRecord) GetValue(agent *Agent) (co CompletionValue) {
	if r.IsUnresolvableReference() {
		return co.ThrowError(agent, ReferenceError, "Unresolvable reference: "+r.ReferencedName.String)
	}
	if r.IsPropertyReference() {
		value, _ := r.Base.Value()
		baseObj, isAbrupt, rt := ReturnIfAbrupt(value.ToObject(agent), co)
		if isAbrupt {
			return rt
		}
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
		return baseObj.
			InternalMethods().
			Get(baseObj, propKey, r.GetThisValue())
	} else {
		base, _ := r.Base.Env()
		name := r.ReferencedName.String
		c := base.GetBindingValue(agent, name, r.Strict)
		return c
	}
}

// PutValue
// spec: 6.2.5.6
// returns: UNUSED or abrupt completion
func (r *ReferenceRecord) PutValue(agent *Agent, value Value) (co CompletionValue) {
	if r.IsUnresolvableReference() {
		if r.Strict {
			return co.ThrowError(agent, ReferenceError, "Unresolvable reference")
		}
		globalObj := agent.GetGlobalObject()
		globalObj.Set(NewStringPropertyKey(r.ReferencedName.String), value, setThrowTypeIgnore)
		return
	}

	if r.IsPropertyReference() {
		v, _ := r.Base.Value()
		baseObj := v.ToObject(agent).value

		if r.IsPrivateReference() {
			panic("unimplemented")
		}

		var referencedName PropertyKey
		if r.ReferencedName.PrivateName != nil {
			panic("unreachable")
		} else if r.ReferencedName.Symbol != nil {
			referencedName = NewSymbolPropertyKey(r.ReferencedName.Symbol)
		} else {
			referencedName = NewStringPropertyKey(r.ReferencedName.String)
		}

		succeeded, isAbrupt, rt := ReturnIfAbrupt(
			baseObj.InternalMethods().Set(
				baseObj,
				referencedName,
				value,
				r.GetThisValue(),
			),
			co,
		)
		if isAbrupt {
			return rt
		}
		if !succeeded && r.Strict {
			return co.ThrowTypeError(agent, "Failed to set value")
		}
		return
	}

	env, _ := r.Base.Env()
	referencedName := r.ReferencedName.String
	env.SetMutableBinding(referencedName, value, r.Strict)
	return
}

// 6.2.5.7
func (r *ReferenceRecord) GetThisValue() Value {
	Assert(r.IsPropertyReference())

	if r.IsSuperReference() {
		return r.ThisValue
	}
	v, _ := r.Base.Value()
	return v
}

// 6.2.5.8
func (r *ReferenceRecord) InitializeReferencedBinding(value Value) {
	Assert(!r.IsUnresolvableReference())

	env, _ := r.Base.Env()
	env.InitializeBinding(r.ReferencedName.String, value)
}
