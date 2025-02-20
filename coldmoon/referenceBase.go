package coldmoon

// ReferenceRecordBase
// [[Base]]: Value, EnvironmentRecord, or UNRESOLVABLE
type ReferenceRecordBase interface {
	IsUnresolvable() bool
	Env() (EnvironmentRecord, bool)
	Value() (Value, bool)
}

type defaultReferenceRecordBase struct{}

func (d *defaultReferenceRecordBase) IsUnresolvable() bool {
	return false
}

func (d *defaultReferenceRecordBase) Env() (EnvironmentRecord, bool) {
	return nil, false
}

func (d *defaultReferenceRecordBase) Value() (Value, bool) {
	return nil, false
}

// UnresolvableReferenceRecordBase struct
type UnresolvableReferenceRecordBase struct {
	*defaultReferenceRecordBase
}

func (u *UnresolvableReferenceRecordBase) IsUnresolvable() bool {
	return true
}

type ValueReferenceRecordBase struct {
	*defaultReferenceRecordBase
	value Value
}

func (v *ValueReferenceRecordBase) Value() (Value, bool) {
	return v.value, v.value != nil
}

type EnvReferenceRecordBase struct {
	*defaultReferenceRecordBase
	env EnvironmentRecord
}

func (e *EnvReferenceRecordBase) Env() (EnvironmentRecord, bool) {
	return e.env, e.env != nil
}

func NewReferenceRecordBaseUnresolvable() ReferenceRecordBase {
	return &UnresolvableReferenceRecordBase{}
}

func NewReferenceRecordBaseValue(value Value) ReferenceRecordBase {
	return &ValueReferenceRecordBase{value: value}
}

func NewReferenceRecordBaseEnv(env EnvironmentRecord) ReferenceRecordBase {
	return &EnvReferenceRecordBase{env: env}
}
