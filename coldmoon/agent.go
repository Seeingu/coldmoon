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
	// [[IsLittleEndian]]
	IsLittleEndian bool
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
	HostLoadImportedModule         func(agent *Agent, referrer ImportedModuleReferrer, specifier string, hostDefined HostDefined, payload ImportedModulePayload)
}

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

func (a *Agent) RunningExecutionContext() *ExecutionContext {
	Assert(a.ExecutionContextStack.Len() > 0)
	return a.ExecutionContextStack.Peek()
}

func (a *Agent) RunJobs() {
	for _, job := range a.QueuedPromiseJobs.Data() {
		previousRealm := a.RunningExecutionContext().Realm
		a.RunningExecutionContext().Realm = job.realm
		job.job.Fun(job.job.Captures)
		a.RunningExecutionContext().Realm = previousRealm
	}
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
	superConstructor := activeFunction.InternalMethods().GetPrototypeOf(activeFunction)
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
