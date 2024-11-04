package coldmoon

type ThisBindingStatus int

const (
	ThisBindingStatusLexical ThisBindingStatus = iota
	ThisBindingStatusInitialized
	ThisBindingStatusUninitialized
)

type FunctionEnvironment struct {
	EnvironmentRecord
	outerEnv               EnvironmentRecord
	thisBindingStatus      ThisBindingStatus
	thisValue              Value
	FunctionObject         *ECMAScriptFunction
	NewTarget              ObjectType
	DeclarativeEnvironment *DeclarativeEnvironment
}

func NewFunctionEnvironment(function *ECMAScriptFunction, newTarget ObjectType) *FunctionEnvironment {
	var thisBindingStatus ThisBindingStatus = ThisBindingStatusUninitialized
	if function.ThisMode == ThisModeLexical {
		thisBindingStatus = ThisBindingStatusLexical
	}
	return &FunctionEnvironment{
		FunctionObject:         function,
		thisBindingStatus:      thisBindingStatus,
		NewTarget:              newTarget,
		outerEnv:               function.Environment,
		thisValue:              nil,
		DeclarativeEnvironment: NewDeclarativeEnvironment(nil),
	}
}

// 9.1.1.3.1
func (f *FunctionEnvironment) BindThisValue(value Value) Value {
	Assert(f.thisBindingStatus != ThisBindingStatusLexical)

	if f.thisBindingStatus == ThisBindingStatusInitialized {
		panic("ReferenceError")
	}
	f.thisValue = value
	f.thisBindingStatus = ThisBindingStatusInitialized
	return value
}

// 9.1.1.3.2
func (f *FunctionEnvironment) HasThisBinding() bool {
	return f.thisBindingStatus != ThisBindingStatusLexical
}

// 9.1.1.3.4
func (f *FunctionEnvironment) GetThisBinding() ObjectType {
	Assert(f.thisBindingStatus != ThisBindingStatusLexical)
	if f.thisBindingStatus == ThisBindingStatusUninitialized {
		panic("ReferenceError")
	}
	return f.thisValue.(ObjectType)
}

func (f *FunctionEnvironment) OuterEnv() EnvironmentRecord {
	return f.outerEnv
}

func (f *FunctionEnvironment) HasBinding(name string) bool {
	return f.DeclarativeEnvironment.HasBinding(name)
}

func (f *FunctionEnvironment) GetBindingValue(name string, strict bool) Value {
	return f.DeclarativeEnvironment.GetBindingValue(name, strict)
}

func (f *FunctionEnvironment) WithBaseObject() ObjectType {
	return f.DeclarativeEnvironment.WithBaseObject()
}
