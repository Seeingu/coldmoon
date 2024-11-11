package coldmoon

// 9.1
type EnvironmentRecord interface {
	OuterEnv() EnvironmentRecord
	HasBinding(name string) bool
	CreateMutableBinding(name string, deletable bool)
	CreateImmutableBinding(name string)
	InitializeBinding(name string, value Value)
	SetMutableBinding(name string, value Value, strict bool)
	GetBindingValue(name string, strict bool) Value
	DeleteBinding(name string) bool
	HasThisBinding() bool
	HasSuperBinding() bool
	WithBaseObject() ObjectType
	GetThisBinding() Value
}

// 9.1.2.1
func GetIdentifierReference(env EnvironmentRecord, name string, strict bool) *ReferenceRecord {
	if env == nil {
		return NewReferenceRecord(
			&ReferenceRecordBaseUnresolvable{},
			&ReferencedNameString{String: name},
			strict,
			nil,
		)
	}

	exists := env.HasBinding(name)
	if exists {
		return NewReferenceRecord(
			&ReferenceRecordBaseEnvironment{Environment: env},
			&ReferencedNameString{String: name},
			strict,
			nil,
		)
	} else {
		outer := env.OuterEnv()
		return GetIdentifierReference(outer, name, strict)
	}
}
