package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type Agent struct {
	symbolId              uint64
	exception             *Value
	ExecutionContextStack pkg.Stack[*ExecutionContext]
}

var WellKnownSymbols = map[WellKnownSymbolsKey]Symbol{}

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
	return a
}

func (a *Agent) runningExecutionContext() *ExecutionContext {
	Assert(a.ExecutionContextStack.Len() > 0)
	return a.ExecutionContextStack.Peek()
}

func (a *Agent) CurrentRealm() *Realm {
	return a.runningExecutionContext().Realm
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
	return NewValueFromObject(envRec.GetThisBinding())
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
	WellKnownSymbols[WellKnownSymbolsAsyncIterator] = agent.CreateSymbol("Symbol.asyncIterator")
	WellKnownSymbols[WellKnownSymbolsHasInstance] = agent.CreateSymbol("Symbol.hasInstance")
	WellKnownSymbols[WellKnownSymbolsIsConcatSpreadable] = agent.CreateSymbol("Symbol.isConcatSpreadable")
	WellKnownSymbols[WellKnownSymbolsIterator] = agent.CreateSymbol("Symbol.iterator")
	WellKnownSymbols[WellKnownSymbolsMatch] = agent.CreateSymbol("Symbol.match")
	WellKnownSymbols[WellKnownSymbolsMatchAll] = agent.CreateSymbol("Symbol.matchAll")
	WellKnownSymbols[WellKnownSymbolsReplace] = agent.CreateSymbol("Symbol.replace")
	WellKnownSymbols[WellKnownSymbolsSearch] = agent.CreateSymbol("Symbol.search")
	WellKnownSymbols[WellKnownSymbolsSpecies] = agent.CreateSymbol("Symbol.species")
	WellKnownSymbols[WellKnownSymbolsSplit] = agent.CreateSymbol("Symbol.split")
	WellKnownSymbols[WellKnownSymbolsToPrimitive] = agent.CreateSymbol("Symbol.toPrimitive")
	WellKnownSymbols[WellKnownSymbolsToStringTag] = agent.CreateSymbol("Symbol.toStringTag")
	WellKnownSymbols[WellKnownSymbolsUnscopables] = agent.CreateSymbol("Symbol.unscopables")
}

func (a *Agent) CreateSymbol(desc string) Symbol {
	s := Symbol{
		Id:          a.symbolId,
		Description: desc,
	}
	a.symbolId += 1
	return s
}

type ErrorObject struct {
	ObjectType
}

// 5.2.3.2
func (a *Agent) ThrowException(exceptionType ExceptionType, message string) ObjectType {
	m := exceptionType.String() + " " + message
	*a.exception = NewStringValue(m)
	// TODO
	return &ErrorObject{}
}
