package coldmoon

import (
	"sync"
	"sync/atomic"

	"github.com/Seeingu/coldmoon/pkg"
)

type QueuedPromiseJob struct {
	job   *Job
	realm *Realm
}
type Agent struct {
	symbolId                   uint64
	moduleAsyncEvaluationCount atomic.Uint64
	exception                  Value
	ExecutionContextStack      pkg.Stack[*ExecutionContext]
	executionContextMu         sync.Mutex
	executionContextReady      *sync.Cond
	asyncContinuationContexts  map[*ExecutionContext]struct{}
	ExecutionContextMap        map[uint64]*ExecutionContext
	HostHooks                  *HostHooks
	Scheduler                  *Scheduler
	ModuleGraph                *ModuleGraph
	GlobalSymbolRegistry       map[string]*SymbolValue
	// [[IsLittleEndian]]
	IsLittleEndian bool
	// CanBlock mirrors the agent's [[CanBlock]] field (ES2024 9.4.3): whether
	// this agent may suspend on Atomics.wait. The engine has no Worker agents,
	// so the main thread cannot block: nothing could ever notify it.
	CanBlock bool
}

// IncrementModuleAsyncEvaluationCount implements the agent-scoped counter
// used to preserve depth-first module evaluation order across TLA suspension.
func (a *Agent) IncrementModuleAsyncEvaluationCount() uint64 {
	return a.moduleAsyncEvaluationCount.Add(1) - 1
}

type HostHooks struct {
	HostEnsureCanCompileStrings    func(realm *Realm)
	HostHasSourceTextAvailable     func(o ObjectType) bool
	HostMakeJobCallback            func(callback ObjectType) *JobCallback
	HostCallJobCallback            func(callback *JobCallback, this Value, arguments []Value) CompletionValue
	HostEnqueuePromiseJob          func(agent *Agent, job *Job, realm *Realm)
	HostPromiseRejectionTracker    func(promise *PromiseObject, operation PromiseRejectionTrackerOperation)
	HostResizeArrayBuffer          func(buffer *ArrayBufferLike, newByteLength JSInt) ResizeArrayBufferHandled
	HostEnsureCanAddPrivateElement func()
	HostGetImportMetaProperties    func(module *SourceTextModule) ImportMetaProperties
	HostFinalizeImportMeta         func(meta ObjectType, module *SourceTextModule)
	HostLoadModule                 ModuleLoader
}

func NewAgent() *Agent {
	return NewAgentWithClock(nil)
}

// NewAgentWithClock creates an Agent whose scheduler uses clock. Tests use this
// constructor to control time without changing runtime semantics.
func NewAgentWithClock(clock Clock) *Agent {
	a := &Agent{
		ExecutionContextMap:       make(map[uint64]*ExecutionContext),
		asyncContinuationContexts: make(map[*ExecutionContext]struct{}),
		GlobalSymbolRegistry:      make(map[string]*SymbolValue),
	}
	a.executionContextReady = sync.NewCond(&a.executionContextMu)
	a.Scheduler = NewScheduler(a, clock)
	a.ModuleGraph = NewModuleGraph(a)
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
			// The default host accepts every private-element addition.
		},
		HostGetImportMetaProperties: HostGetImportMetaProperties,
		HostFinalizeImportMeta:      HostFinalizeImportMeta,
	}
	return a
}

func (a *Agent) RunningExecutionContext() *ExecutionContext {
	a.executionContextMu.Lock()
	defer a.executionContextMu.Unlock()
	Assert(a.ExecutionContextStack.Len() > 0)
	return a.ExecutionContextStack.Peek()
}

func (a *Agent) UniqueObjectId() uint64 {
	a.symbolId++
	return a.symbolId
}

func (a *Agent) FindExecutionContextById(id uint64) *ExecutionContext {
	return a.ExecutionContextMap[id]
}

func (a *Agent) CurrentRealm() *Realm {
	return a.RunningExecutionContext().Realm
}

func (a *Agent) ActiveFunctionObject() ObjectType {
	return a.RunningExecutionContext().Function.Ref()
}

// 9.4.1
// return nil if there is no active script or module
func (a *Agent) GetActiveScriptOrModule() ScriptOrModule {
	a.executionContextMu.Lock()
	defer a.executionContextMu.Unlock()
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
		Assert(a.RunningExecutionContext().ECMAScriptCode != nil)
		env = a.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment
	}
	return GetIdentifierReference(env, name, strict)
}

// 9.4.3
func (a *Agent) GetThisEnvironment() EnvironmentRecord {
	env := a.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment

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

// GetNewTarget
// spec: 9.4.5
func (a *Agent) GetNewTarget() ObjectType {
	envRec := a.GetThisEnvironment()
	fun, ok := envRec.(*FunctionEnvironment)
	Assert(ok)
	return fun.NewTarget
}

// GetSuperConstructor
// spec: 13.3.7.2
func (a *Agent) GetSuperConstructor() Value {
	envRec := a.GetThisEnvironment()
	funEnv := envRec.(*FunctionEnvironment)
	activeFunction := funEnv.FunctionObject
	superConstructor := activeFunction.internalMethods().GetPrototypeOf(activeFunction)
	return superConstructor.ToValue()
}

// 9.4.6
func (a *Agent) GetGlobalObject() *Object {
	return a.CurrentRealm().GlobalObject
}

// 5.2.3.2
func (a *Agent) ThrowException(exceptionType ExceptionType, message string) Value {
	realm := a.CurrentRealm()
	constructor := realm.Intrinsics.Get(exceptionType.ToIntrinsicName())
	errorObject := constructor.Construct([]Value{NewStringValue(message)}, nil).value
	a.exception = errorObject.ToValue()
	return a.exception
}

func (a *Agent) ThrowTypeError(message string) Value {
	return a.ThrowException(TypeError, message)
}

func (a *Agent) ThrowRangeError(message string) Value {
	return a.ThrowException(RangeError, message)
}

func (a *Agent) ThrowRangeExceptionObject(message string) ObjectType {
	realm := a.CurrentRealm()
	constructor := realm.Intrinsics.Get(RangeError.ToIntrinsicName())
	errorObject := constructor.Construct([]Value{NewStringValue(message)}, nil).value
	a.exception = errorObject.ToValue()
	return errorObject
}

func (a *Agent) ThrowTypeExceptionObject(message string) ObjectType {
	realm := a.CurrentRealm()
	constructor := realm.Intrinsics.Get(TypeError.ToIntrinsicName())
	errorObject := constructor.Construct([]Value{NewStringValue(message)}, nil).value
	a.exception = errorObject.ToValue()
	return errorObject
}
