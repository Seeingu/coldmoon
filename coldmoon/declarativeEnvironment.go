package coldmoon

type DeclarativeEnvironment struct {
	EnvironmentRecord
	outerEnv EnvironmentRecord
}

// 9.1.2.2
func NewDeclarativeEnvironment(outerEnv EnvironmentRecord) *DeclarativeEnvironment {
	return &DeclarativeEnvironment{
		outerEnv: outerEnv,
	}
}

// 9.1.1.1.8
func (d *DeclarativeEnvironment) HasThisBinding() bool {
	return false
}

// 9.1.1.1.6
func (d *DeclarativeEnvironment) GetBindingValue(name string, strict bool) Value {
	panic("implement me")
}

func (d *DeclarativeEnvironment) OuterEnv() EnvironmentRecord {
	return d.outerEnv
}

func (d *DeclarativeEnvironment) HasBinding(name string) bool {
	return false
}

// 9.1.1.1.10
func (d *DeclarativeEnvironment) WithBaseObject() ObjectType {
	return nil
}
