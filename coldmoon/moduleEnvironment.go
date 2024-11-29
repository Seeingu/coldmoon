package coldmoon

type ModuleEnvironment struct {
	EnvironmentRecord
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
