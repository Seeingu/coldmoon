package coldmoon

// 9.1.1.2
type ObjectEnvironment struct {
	EnvironmentRecord
	BindingObject     ObjectType
	IsWithEnvironment bool
	outerEnv          EnvironmentRecord
}

func (o *ObjectEnvironment) OuterEnv() EnvironmentRecord {
	return o.outerEnv
}

// 9.1.2.3
func NewObjectEnvironment(obj ObjectType, isWithEnvironment bool, outerEnv EnvironmentRecord) *ObjectEnvironment {
	return &ObjectEnvironment{
		BindingObject:     obj,
		IsWithEnvironment: isWithEnvironment,
		outerEnv:          outerEnv,
	}
}

// 9.1.1.2.8
func (o *ObjectEnvironment) HasThisBinding() bool {
	return false
}
