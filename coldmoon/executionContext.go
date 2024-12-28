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
}
