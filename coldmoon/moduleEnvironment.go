package coldmoon

type IndirectBinding struct {
	Module      *SourceTextModule
	BindingName string
}
type ModuleEnvironment struct {
	*DeclarativeEnvironment
	IndirectBindings map[string]*IndirectBinding
}

var _ EnvironmentRecord = (*ModuleEnvironment)(nil)

func NewModuleEnvironment(outerEnv EnvironmentRecord) *ModuleEnvironment {
	return &ModuleEnvironment{
		DeclarativeEnvironment: NewDeclarativeEnvironment(outerEnv),
	}
}

func (m *ModuleEnvironment) CreateImportBinding(name string, module *SourceTextModule, bindingName string) {
	m.IndirectBindings[name] = &IndirectBinding{
		Module:      module,
		BindingName: bindingName,
	}
}

// 9.1.1.5.1
func (m *ModuleEnvironment) GetBindingValue(agent *Agent, name string, strict bool) *CompletionValue {
	Assert(strict)
	Assert(m.HasBinding(name))
	if binding, ok := m.IndirectBindings[name]; ok {
		m := binding.Module
		bindingName := binding.BindingName
		targetEnv := m.Environment
		if targetEnv == nil {
			return NewCompletionValueError(agent.ThrowException(ReferenceError, "Module is not initialized"))
		}
		return targetEnv.GetBindingValue(agent, bindingName, true)
	}
	return m.DeclarativeEnvironment.GetBindingValue(agent, name, true)
}

func (m *ModuleEnvironment) DeleteBinding(name string) bool {
	panic("unreachable")
}
func (m *ModuleEnvironment) HasThisBinding() bool {
	return false
}
func (m *ModuleEnvironment) GetThisBinding() Value {
	return UndefinedValue
}
