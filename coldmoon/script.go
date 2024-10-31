package coldmoon

import ast "github.com/Seeingu/coldmoon/coldmoon/parser"

type ScriptRecord struct {
	Realm          *Realm
	ECMAScriptCode *ast.Script
	LoadedModules  interface{}
	HostDefined    interface{}
}

// ParseScript
// 16.1.5
func ParseScript(sourceText string, realm *Realm, hostDefined interface{}) *ScriptRecord {
	script := ast.NewParser(sourceText).Parse()

	s := &ScriptRecord{
		Realm:          realm,
		HostDefined:    hostDefined,
		ECMAScriptCode: script,
	}
	return s
}
