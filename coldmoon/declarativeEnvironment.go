package coldmoon

type Binding struct {
	Value     Value
	Strict    bool
	Mutable   bool
	Deletable bool
}

type DeclarativeEnvironment struct {
	EnvironmentRecord
	outerEnv EnvironmentRecord
	Bindings map[string]*Binding
}

var _ EnvironmentRecord = (*DeclarativeEnvironment)(nil)

// 9.1.2.2
func NewDeclarativeEnvironment(outerEnv EnvironmentRecord) *DeclarativeEnvironment {
	return &DeclarativeEnvironment{
		outerEnv: outerEnv,
		Bindings: make(map[string]*Binding),
	}
}

// 9.1.1.1.8
func (d *DeclarativeEnvironment) HasThisBinding() bool {
	return false
}

// 9.1.1.1.6
func (d *DeclarativeEnvironment) GetBindingValue(agent *Agent, name string, strict bool) (co Completion[Value]) {
	binding, ok := d.Bindings[name]
	if !ok || binding.Value == nil {
		co.err = agent.ThrowException(ReferenceError, "Binding not found")
		return
	}
	return binding.Value.ToCompletion()
}

func (d *DeclarativeEnvironment) DeleteBinding(name string) bool {
	binding, ok := d.Bindings[name]
	Assert(ok)

	if !binding.Deletable {
		return false
	}

	delete(d.Bindings, name)
	return true
}

func (d *DeclarativeEnvironment) OuterEnv() EnvironmentRecord {
	return d.outerEnv
}

func (d *DeclarativeEnvironment) HasBinding(name string) bool {
	_, ok := d.Bindings[name]
	return ok
}

func (d *DeclarativeEnvironment) HasSuperBinding() bool {
	return false
}

// 9.1.1.1.10
func (d *DeclarativeEnvironment) WithBaseObject() ObjectType {
	return nil
}

// 9.1.1.1.3
func (d *DeclarativeEnvironment) CreateImmutableBinding(name string, strict bool) {
	d.Bindings[name] = &Binding{
		Strict:    strict,
		Mutable:   false,
		Deletable: false,
	}
}

// 9.1.1.2.2
func (d *DeclarativeEnvironment) CreateMutableBinding(name string, deletable bool) {
	d.Bindings[name] = &Binding{
		Mutable:   true,
		Deletable: deletable,
	}
}

// 9.1.1.1.4
func (d *DeclarativeEnvironment) InitializeBinding(name string, value Value) {
	binding, ok := d.Bindings[name]
	Assert(ok && IsUndefinedOrNil(binding.Value))

	binding.Value = value
}

// 9.1.1.1.5
func (d *DeclarativeEnvironment) SetMutableBinding(agent *Agent, name string, value Value, strict bool) {
	binding, ok := d.Bindings[name]
	if !ok {
		if strict {
			// 9.1.1.1.5 step 1.b: an assignment to an undeclared binding is a
			// ReferenceError in strict mode. The Value panic is recovered by the
			// nearest builtin or source boundary and becomes a JS throw; a bare
			// string panic would crash the host process instead.
			panic(agent.ThrowException(ReferenceError, "ReferenceError: assignment to undeclared binding"))
		}
		d.CreateMutableBinding(name, true)
		d.InitializeBinding(name, value)
		return
	}

	s := strict
	if binding.Strict {
		s = binding.Strict
	}

	if binding.Value == nil {
		panic(agent.ThrowException(ReferenceError, "Binding is not initialized"))
	}

	if binding.Mutable {
		binding.Value = value
	} else {
		if s {
			panic(agent.ThrowException(TypeError, "Assignment to constant variable"))
		}
	}
}
