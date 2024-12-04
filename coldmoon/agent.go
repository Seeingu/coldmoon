package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type QueuedPromiseJob struct {
	job   *Job
	realm *Realm
}
type Agent struct {
	symbolId              uint64
	exception             Value
	ExecutionContextStack pkg.Stack[*ExecutionContext]
	HostHooks             *HostHooks
	GlobalSymbolRegistry  map[string]*SymbolValue
	QueuedPromiseJobs     pkg.Stack[*QueuedPromiseJob]
}

type HostHooks struct {
	HostEnsureCanCompileStrings    func(realm *Realm)
	HostHasSourceTextAvailable     func(o ObjectType) bool
	HostMakeJobCallback            func(callback ObjectType) *JobCallback
	HostCallJobCallback            func(callback *JobCallback, this Value, arguments []Value) Value
	HostEnqueuePromiseJob          func(agent *Agent, job *Job, realm *Realm)
	HostPromiseRejectionTracker    func(promise *PromiseObject, operation PromiseRejectionTrackerOperation)
	HostResizeArrayBuffer          func(buffer *ArrayBufferObject, newByteLength JSInt) ResizeArrayBufferHandled
	HostEnsureCanAddPrivateElement func()
	HostGetImportMetaProperties    func(module *SourceTextModule) ImportMetaProperties
	HostFinalizeImportMeta         func(meta ObjectType, module *SourceTextModule)
	HostLoadImportedModule         func(agent *Agent, referrer ImportedModuleReferrer, specifier string, hostDefined *HostDefined, payload ImportedModulePayload)
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
	URIError
	AggregateError
)

func NewAgent() *Agent {
	a := &Agent{}
	initWellKnownSymbols(a)
	a.HostHooks = &HostHooks{
		HostEnsureCanCompileStrings: HostEnsureCanCompileStrings,
		HostHasSourceTextAvailable:  HostHasSourceTextAvailable,
		HostMakeJobCallback:         HostMakeJobCallback,
		HostCallJobCallback:         HostCallJobCallback,
		HostEnqueuePromiseJob:       HostEnqueuePromiseJob,
		HostPromiseRejectionTracker: HostPromiseRejectionTracker,
		HostResizeArrayBuffer:       HostResizeArrayBuffer,
		HostEnsureCanAddPrivateElement: func() {
			// TODO
		},
		HostGetImportMetaProperties: HostGetImportMetaProperties,
		HostFinalizeImportMeta:      HostFinalizeImportMeta,
		HostLoadImportedModule:      HostLoadImportedModule,
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

func (a *Agent) ActiveFunctionObject() ObjectType {
	return a.runningExecutionContext().Function
}

// 9.4.1
// return nil if there is no active script or module
func (a *Agent) GetActiveScriptOrModule() ScriptOrModule {
	if a.ExecutionContextStack.IsEmpty() {
		return nil
	}
	var ec *ExecutionContext
	for i := a.ExecutionContextStack.Len() - 1; i >= 0; i-- {
		ec = a.ExecutionContextStack.Index(i)
		if ec.ScriptOrModule != nil {
			return ec.ScriptOrModule
		}
	}
	return nil
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

// 9.4.5
func (a *Agent) GetNewTarget() ObjectType {
	envRec := a.GetThisEnvironment()
	fun, ok := envRec.(*FunctionEnvironment)
	Assert(ok)
	return fun.NewTarget
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
func (a *Agent) ThrowException(exceptionType ExceptionType, message string) Value {
	realm := a.CurrentRealm()
	constructor := realm.Intrinsics.Get("%" + exceptionType.String() + "%")
	errorObject := constructor.Construct([]Value{NewStringValue(message)}, nil)
	a.exception = NewValueFromObject(errorObject)
	return a.exception
}
