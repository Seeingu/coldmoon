package coldmoon

type ScriptRecord struct {
	Realm          *Realm
	ECMAScriptCode interface{}
	LoadedModules  interface{}
	HostDefined    interface{}
}

// ParseScript
// 16.1.5
func ParseScript(sourceText string, realm *Realm, hostDefined interface{}) *ScriptRecord {
	s := &ScriptRecord{}
	s.Realm = realm
	s.HostDefined = hostDefined
	return s
}
