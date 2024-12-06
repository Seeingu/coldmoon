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
func (d *DeclarativeEnvironment) GetBindingValue(agent *Agent, name string, strict bool) CompletionValue {
	binding, ok := d.Bindings[name]
	if !ok {
		return NewCompletionValueError(agent.ThrowException(ReferenceError, "Binding not found"))
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
	binding := d.Bindings[name]
	Assert(binding.Value == nil)

	binding.Value = value
}

// 9.1.1.1.5
func (d *DeclarativeEnvironment) SetMutableBinding(name string, value Value, strict bool) {
	binding, ok := d.Bindings[name]
	if !ok {
		if strict {
			panic("ReferenceError")
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
		panic("ReferenceError")
	}

	if binding.Mutable {
		binding.Value = value
	} else {
		if s {
			panic("ReferenceError")
		}
	}
}
