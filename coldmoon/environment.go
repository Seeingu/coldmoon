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
	GetThisBinding() ObjectType
}
