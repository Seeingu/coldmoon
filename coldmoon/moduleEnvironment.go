package coldmoon

type ModuleEnvironment struct {
	*DeclarativeEnvironment
	outerEnv EnvironmentRecord
}

func NewModuleEnvironment(outerEnv EnvironmentRecord) *ModuleEnvironment {
	return &ModuleEnvironment{
		outerEnv: outerEnv,
	}
}

func (m *ModuleEnvironment) OuterEnv() EnvironmentRecord {
	return m.outerEnv
}

func (m *ModuleEnvironment) CreateImportBinding(name string, module *SourceTextModule, bindingName string) {
	// TODO
}
