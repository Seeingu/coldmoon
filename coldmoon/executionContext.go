package coldmoon

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
	VM             *VM
	ch             chan struct{}
	yieldCh        chan struct{}
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
