package coldmoon

// 9.4
type ScriptOrModule int

const (
	ScriptOrModuleNull ScriptOrModule = iota
	Script
	Module
)

type ExecutionContext struct {
	Realm          *Realm
	ScriptOrModule ScriptOrModule
	Function       *Object
}
