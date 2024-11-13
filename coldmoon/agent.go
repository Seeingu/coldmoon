package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type Agent struct {
	symbolId              uint64
	exception             Value
	ExecutionContextStack pkg.Stack[*ExecutionContext]
	HostHooks             *HostHooks
}

type HostHooks struct {
	HostEnsureCanCompileStrings func(realm *Realm)
	HostHasSourceTextAvailable  func(o ObjectType) bool
}

var WellKnownSymbols = map[WellKnownSymbolsKey]*SymbolValue{}

//go:generate stringer -type=ExceptionType
type ExceptionType int

const (
	EvalError ExceptionType = iota
	RangeError
	ReferenceError
	SyntaxError
	TypeError
	UriError
)

func NewAgent() *Agent {
	a := &Agent{}
	initWellKnownSymbols(a)
	a.HostHooks = &HostHooks{
		HostEnsureCanCompileStrings: HostEnsureCanCompileStrings,
		HostHasSourceTextAvailable:  HostHasSourceTextAvailable,
	}
	return a
}

func (a *Agent) runningExecutionContext() *ExecutionContext {
	Assert(a.ExecutionContextStack.Len() > 0)
	return a.ExecutionContextStack.Peek()
}

func (a *Agent) CurrentRealm() *Realm {
	return a.runningExecutionContext().Realm
}

func (a *Agent) ActiveFunctionObject() *Object {
	return a.runningExecutionContext().Function
}

// 9.4.1
func (a *Agent) GetActiveScriptOrModule() ScriptOrModule {
	if a.ExecutionContextStack.IsEmpty() {
		return ScriptOrModuleNull
	}
	var ec *ExecutionContext
	for i := a.ExecutionContextStack.Len() - 1; i >= 0; i-- {
		ec = a.ExecutionContextStack.Index(i)
		if ec.ScriptOrModule != ScriptOrModuleNull {
			return ec.ScriptOrModule
		}
	}
	return ScriptOrModuleNull
}

// 9.4.2
func (a *Agent) ResolveBinding(name string, env EnvironmentRecord, strict bool) *ReferenceRecord {
	if env == nil {
		env = a.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
	}
	return GetIdentifierReference(env, name, strict)
}

// 9.4.3
func (a *Agent) GetThisEnvironment() EnvironmentRecord {
	env := a.runningExecutionContext().ECMAScriptCode.LexicalEnvironment

	for {
		exists := env.HasThisBinding()
		if exists {
			return env
		}
		outer := env.OuterEnv()
		Assert(outer != nil)

		env = outer
	}
}

// 9.4.4
func (a *Agent) ResolveThisBinding() Value {
	envRec := a.GetThisEnvironment()
	return envRec.GetThisBinding()
}

// 9.4.6
func (a *Agent) GetGlobalObject() *Object {
	return a.CurrentRealm().GlobalObject
}

// MARK: - Well-known Symbols

type WellKnownSymbolsKey string

const (
	WellKnownSymbolsAsyncIterator      WellKnownSymbolsKey = "@@asyncIterator"
	WellKnownSymbolsHasInstance        WellKnownSymbolsKey = "@@hasInstance"
	WellKnownSymbolsIsConcatSpreadable WellKnownSymbolsKey = "@@isConcatSpreadable"
	WellKnownSymbolsIterator           WellKnownSymbolsKey = "@@iterator"
	WellKnownSymbolsMatch              WellKnownSymbolsKey = "@@match"
	WellKnownSymbolsMatchAll           WellKnownSymbolsKey = "@@matchAll"
	WellKnownSymbolsReplace            WellKnownSymbolsKey = "@@replace"
	WellKnownSymbolsSearch             WellKnownSymbolsKey = "@@search"
	WellKnownSymbolsSpecies            WellKnownSymbolsKey = "@@species"
	WellKnownSymbolsSplit              WellKnownSymbolsKey = "@@split"
	WellKnownSymbolsToPrimitive        WellKnownSymbolsKey = "@@toPrimitive"
	WellKnownSymbolsToStringTag        WellKnownSymbolsKey = "@@toStringTag"
	WellKnownSymbolsUnscopables        WellKnownSymbolsKey = "@@unscopables"
)

func initWellKnownSymbols(agent *Agent) {
	WellKnownSymbols[WellKnownSymbolsAsyncIterator] = agent.CreateSymbol("SymbolValue.asyncIterator")
	WellKnownSymbols[WellKnownSymbolsHasInstance] = agent.CreateSymbol("SymbolValue.hasInstance")
	WellKnownSymbols[WellKnownSymbolsIsConcatSpreadable] = agent.CreateSymbol("SymbolValue.isConcatSpreadable")
	WellKnownSymbols[WellKnownSymbolsIterator] = agent.CreateSymbol("SymbolValue.iterator")
	WellKnownSymbols[WellKnownSymbolsMatch] = agent.CreateSymbol("SymbolValue.match")
	WellKnownSymbols[WellKnownSymbolsMatchAll] = agent.CreateSymbol("SymbolValue.matchAll")
	WellKnownSymbols[WellKnownSymbolsReplace] = agent.CreateSymbol("SymbolValue.replace")
	WellKnownSymbols[WellKnownSymbolsSearch] = agent.CreateSymbol("SymbolValue.search")
	WellKnownSymbols[WellKnownSymbolsSpecies] = agent.CreateSymbol("SymbolValue.species")
	WellKnownSymbols[WellKnownSymbolsSplit] = agent.CreateSymbol("SymbolValue.split")
	WellKnownSymbols[WellKnownSymbolsToPrimitive] = agent.CreateSymbol("SymbolValue.toPrimitive")
	WellKnownSymbols[WellKnownSymbolsToStringTag] = agent.CreateSymbol("SymbolValue.toStringTag")
	WellKnownSymbols[WellKnownSymbolsUnscopables] = agent.CreateSymbol("SymbolValue.unscopables")
}

func (a *Agent) CreateSymbol(desc string) *SymbolValue {
	s := &SymbolValue{
		Id:          a.symbolId,
		Description: desc,
	}
	a.symbolId += 1
	return s
}

// 5.2.3.2
func (a *Agent) ThrowException(exceptionType ExceptionType, message string) ObjectType {
	m := exceptionType.String() + " " + message
	a.exception = NewStringValue(m)
	// TODO
	return &ErrorObject{}
}
