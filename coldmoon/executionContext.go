package coldmoon

// ExecutionContextScope owns one balanced entry on an Agent's execution
// context stack. Ordinary calls should defer Leave immediately after entering.
type ExecutionContextScope struct {
	agent   *Agent
	context *ExecutionContext
	left    bool
}

// Leave removes the context owned by this scope. It panics when contexts are
// left out of order or when the same scope is left twice.
func (s *ExecutionContextScope) Leave() {
	Assert(s != nil)
	Assert(!s.left)
	s.agent.suspendExecutionContext(s.context)
	s.left = true
}

// 9.4
type ScriptOrModule interface {
	_scriptOrModule()
	toReferrer() ImportedModuleReferrer
}

type ExecutionContextAdditionalState struct {
	LexicalEnvironment  EnvironmentRecord
	VariableEnvironment EnvironmentRecord
	PrivateEnvironment  *PrivateEnvironment
}

type ExecutionContext struct {
	Realm          *Realm
	ScriptOrModule ScriptOrModule
	Function       ObjectType
	ECMAScriptCode *ExecutionContextAdditionalState
	Generator      *GeneratorObject
	AsyncGenerator *AsyncGeneratorObject
	VM             *VM
	ch             chan struct{}
	yieldCh        chan struct{}
	generatorCh    chan Value
	// TODO: unify with yieldCh
	awaitCh     chan struct{}
	isSuspended bool
	Result      CompletionValue
}

func (e *ExecutionContext) Resume() {
	e.ch <- struct{}{}
}

func (e *ExecutionContext) Suspend() {
	<-e.ch
}

// enterExecutionContext makes context current and returns its lexical lifetime
// owner. VM allocation belongs to this transition rather than individual AST
// evaluations.
func (a *Agent) enterExecutionContext(context *ExecutionContext) *ExecutionContextScope {
	a.resumeExecutionContext(context)
	return &ExecutionContextScope{
		agent:   a,
		context: context,
	}
}

// resumeExecutionContext makes a suspended or newly created context current.
// Generator and async lifecycles use this when their stack lifetime is not
// lexical.
func (a *Agent) resumeExecutionContext(context *ExecutionContext) {
	Assert(context != nil)
	if context.ch == nil {
		context.ch = make(chan struct{})
	}
	if context.VM == nil {
		context.VM = NewVM2(a)
	} else {
		Assert(context.VM.agent == a)
	}
	a.ExecutionContextStack.Push(context)
}

// suspendExecutionContext removes the current context and verifies that the
// caller is suspending the expected frame.
func (a *Agent) suspendExecutionContext(context *ExecutionContext) {
	Assert(context != nil)
	Assert(!a.ExecutionContextStack.IsEmpty())
	Assert(a.ExecutionContextStack.Peek() == context)
	Assert(a.ExecutionContextStack.Pop() == context)
}
