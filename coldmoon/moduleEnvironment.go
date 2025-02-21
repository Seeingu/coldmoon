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
		IndirectBindings:       make(map[string]*IndirectBinding),
	}
}

// InitializeBinding is not standard
// When the name is not found in the declarative environment, it will be initialized in the declarative environment
// TODO: check for non-class name
func (m *ModuleEnvironment) InitializeBinding(name string, value Value) {
	if !m.HasBinding(name) {
		m.DeclarativeEnvironment.CreateImmutableBinding(name, true)
	}
	m.DeclarativeEnvironment.InitializeBinding(name, value)
}

func (m *ModuleEnvironment) CreateImportBinding(name string, module *SourceTextModule, bindingName string) {
	m.IndirectBindings[name] = &IndirectBinding{
		Module:      module,
		BindingName: bindingName,
	}
}

// 9.1.1.5.1
func (m *ModuleEnvironment) GetBindingValue(agent *Agent, name string, strict bool) (co CompletionValue) {
	// TODO: should be strict
	// Assert(strict)
	Assert(m.HasBinding(name))
	if binding, ok := m.IndirectBindings[name]; ok {
		m := binding.Module
		bindingName := binding.BindingName
		targetEnv := m.Environment
		if targetEnv == nil {
			co.err = agent.ThrowException(ReferenceError, "Module is not initialized")
			return
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
