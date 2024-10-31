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

func (d *DeclarativeEnvironment) OuterEnv() EnvironmentRecord {
	return d.outerEnv
}
