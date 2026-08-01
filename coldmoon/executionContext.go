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
	// yieldCh resumes an async generator after it yielded control.
	yieldCh          chan struct{}
	generatorCh      chan CompletionValue
	asyncGeneratorCh chan CompletionValue
	// awaitCh wakes an async function when its awaited promise settles. It is
	// separate from yieldCh because the two suspension protocols have distinct
	// owners and lifetimes.
	awaitCh     chan struct{}
	isSuspended bool
	// asyncCallerResumed records the one-time handoff from an async function
	// body back to its synchronous caller.
	asyncCallerResumed bool
	// Await-driven bodies run on finite scheduler tasks. These fields serialize
	// resumed segments with the promise jobs that released them; the module-only
	// fields additionally preserve the Agent stack's LIFO ownership for TLA.
	moduleAsync             bool
	moduleAsyncFinished     bool
	moduleResumeParent      *ExecutionContext
	moduleResumeBase        *ExecutionContext
	moduleResumeAllowsEmpty bool
	moduleAwaiting          bool
	asyncContinuationDone   chan struct{}
	Result                  CompletionValue
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
	a.executionContextMu.Lock()
	defer a.executionContextMu.Unlock()
	if context.moduleAsync {
		for !a.moduleResumeParentReady(context) {
			a.executionContextReady.Wait()
		}
		context.moduleAwaiting = false
	}
	if context.ch == nil {
		context.ch = make(chan struct{})
	}
	if context.VM == nil {
		context.VM = NewVM2(a)
	} else {
		Assert(context.VM.agent == a)
	}
	if context.awaitCh != nil {
		a.asyncContinuationContexts[context] = struct{}{}
		if context.asyncContinuationDone == nil {
			context.asyncContinuationDone = make(chan struct{})
		}
	}
	a.ExecutionContextStack.Push(context)
}

// suspendExecutionContext removes the current context and verifies that the
// caller is suspending the expected frame.
func (a *Agent) suspendExecutionContext(context *ExecutionContext) {
	Assert(context != nil)
	a.executionContextMu.Lock()
	defer a.executionContextMu.Unlock()
	Assert(!a.ExecutionContextStack.IsEmpty())
	Assert(a.ExecutionContextStack.Peek() == context)
	Assert(a.ExecutionContextStack.Pop() == context)
	if context.awaitCh != nil {
		if context.asyncContinuationDone != nil {
			close(context.asyncContinuationDone)
			context.asyncContinuationDone = nil
		}
		if context.moduleAsync && context.moduleAsyncFinished {
			delete(a.asyncContinuationContexts, context)
		} else {
			context.Result = CompletionValue{}
			context.asyncContinuationDone = make(chan struct{})
		}
	}
	if context.moduleAsync {
		if !context.moduleAsyncFinished {
			context.moduleAwaiting = true
			context.moduleResumeParent = context.moduleResumeBase
			context.moduleResumeAllowsEmpty = true
		}
	}
	a.executionContextReady.Broadcast()
}

func (a *Agent) moduleResumeParentReady(context *ExecutionContext) bool {
	if a.ExecutionContextStack.IsEmpty() {
		return context.moduleResumeAllowsEmpty
	}
	return a.ExecutionContextStack.Peek() == context.moduleResumeParent
}

// prepareModuleAsyncContext constrains a module body to resume only after the
// promise reaction that woke it has restored the surrounding execution stack.
func (a *Agent) prepareModuleAsyncContext(context *ExecutionContext) *ExecutionContext {
	a.executionContextMu.Lock()
	defer a.executionContextMu.Unlock()
	Assert(context != nil)
	Assert(!context.moduleAsync)

	var caller *ExecutionContext
	var base *ExecutionContext
	if !a.ExecutionContextStack.IsEmpty() {
		caller = a.ExecutionContextStack.Peek()
		base = a.ExecutionContextStack.Index(0)
	}
	context.moduleAsync = true
	context.moduleResumeParent = caller
	context.moduleResumeBase = base
	context.moduleResumeAllowsEmpty = caller == nil
	a.asyncContinuationContexts[context] = struct{}{}
	return caller
}

// finishModuleAsyncContext marks the current body segment as terminal. Its
// following suspension releases the promise job without arming another await.
func (a *Agent) finishModuleAsyncContext(context *ExecutionContext) {
	a.executionContextMu.Lock()
	defer a.executionContextMu.Unlock()
	Assert(context != nil && context.moduleAsync)
	context.moduleAsyncFinished = true
}

// releaseAsyncContinuationContext drops a completed non-module task from the
// continuation registry. Its final suspension has already closed any channel
// observed by a promise job.
func (a *Agent) releaseAsyncContinuationContext(context *ExecutionContext) {
	a.executionContextMu.Lock()
	defer a.executionContextMu.Unlock()
	Assert(context != nil && !context.moduleAsync)
	delete(a.asyncContinuationContexts, context)
	context.asyncContinuationDone = nil
}

// waitForAsyncContinuations keeps a promise job sequenced with every async
// continuation it released, through that continuation's next suspension or
// terminal completion. This prevents later jobs from observing intermediate
// async-function, async-generator, or top-level-await state.
func (a *Agent) waitForAsyncContinuations() {
	a.executionContextMu.Lock()
	waits := make([]chan struct{}, 0)
	for context := range a.asyncContinuationContexts {
		resultReady := context.Result.value != nil ||
			context.Result.err != nil ||
			context.Result.t != CompletionTypeNormal
		if resultReady && context.asyncContinuationDone != nil {
			waits = append(waits, context.asyncContinuationDone)
		}
	}
	a.executionContextMu.Unlock()

	for _, done := range waits {
		<-done
	}
}

func (a *Agent) hasExecutionContext() bool {
	a.executionContextMu.Lock()
	defer a.executionContextMu.Unlock()
	return !a.ExecutionContextStack.IsEmpty()
}
