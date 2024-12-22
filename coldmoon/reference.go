package coldmoon

// ReferenceRecordBase Enum
type ReferenceRecordBase struct {
	value        Value
	env          EnvironmentRecord
	unresolvable bool
}

func NewReferenceRecordBaseUnresolvable() *ReferenceRecordBase {
	return &ReferenceRecordBase{unresolvable: true}
}

func NewReferenceRecordBaseValue(value Value) *ReferenceRecordBase {
	return &ReferenceRecordBase{value: value}
}

func NewReferenceRecordBaseEnv(env EnvironmentRecord) *ReferenceRecordBase {
	return &ReferenceRecordBase{env: env}
}

func (r *ReferenceRecordBase) Unresolvable() bool {
	return r.unresolvable
}

func (r *ReferenceRecordBase) Env() (EnvironmentRecord, bool) {
	return r.env, r.env != nil
}

func (r *ReferenceRecordBase) Value() (Value, bool) {
	return r.value, r.value != nil
}

// ReferencedName Enum
type ReferencedName struct {
	String      string
	Symbol      *SymbolValue
	PrivateName *PrivateName
}

type ReferenceRecord struct {
	// [[Base]]: Value, EnvironmentRecord, or UNRESOLVABLE
	Base           *ReferenceRecordBase
	ReferencedName *ReferencedName
	Strict         bool
	ThisValue      Value
}

func NewReferenceRecord(base *ReferenceRecordBase, referencedName *ReferencedName, strict bool, thisValue Value) *ReferenceRecord {
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
	return r.Base.Unresolvable()
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
		value, _ := r.Base.Value()
		baseObj := ValueToObject(agent, value)
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
		base, _ := r.Base.Env()
		name := r.ReferencedName.String
		c := base.GetBindingValue(agent, name, r.Strict)
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
		v, _ := r.Base.Value()
		baseObj := ValueToObject(agent, v)

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

	env, _ := r.Base.Env()
	referencedName := r.ReferencedName.String
	env.SetMutableBinding(referencedName, value, r.Strict)
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
