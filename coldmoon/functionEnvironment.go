package coldmoon

type ThisBindingStatus int

const (
	ThisBindingStatusLexical ThisBindingStatus = iota
	ThisBindingStatusInitialized
	ThisBindingStatusUninitialized
)

type FunctionEnvironment struct {
	*DeclarativeEnvironment
	thisBindingStatus ThisBindingStatus
	thisValue         Value
	FunctionObject    *ECMAScriptFunction
	NewTarget         ObjectType
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
		thisValue:              nil,
		DeclarativeEnvironment: NewDeclarativeEnvironment(function.Environment),
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

func (f *FunctionEnvironment) HasSuperBinding() bool {
	if f.thisBindingStatus == ThisBindingStatusLexical {
		return false
	}
	return f.FunctionObject.HomeObject != nil
}

// 9.1.1.3.4
func (f *FunctionEnvironment) GetThisBinding() Value {
	Assert(f.thisBindingStatus != ThisBindingStatusLexical)
	if f.thisBindingStatus == ThisBindingStatusUninitialized {
		panic("ReferenceError")
	}
	return f.thisValue
}

func (f *FunctionEnvironment) GetSuperBase() Value {
	home := f.FunctionObject.HomeObject
	if home == nil {
		return UndefinedValue
	}
	return home.InternalMethods().GetPrototypeOf(home).ToValue()
}
