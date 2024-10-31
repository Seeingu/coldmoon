package coldmoon

type ScriptRecord struct {
	Realm          *Realm
	ECMAScriptCode *Script
	LoadedModules  interface{}
	HostDefined    interface{}
}

// ParseScript
// 16.1.5
func ParseScript(sourceText string, realm *Realm, hostDefined interface{}) *ScriptRecord {
	script := NewParser(sourceText).Parse()

	s := &ScriptRecord{
		Realm:          realm,
		HostDefined:    hostDefined,
		ECMAScriptCode: script,
	}
	return s
}
