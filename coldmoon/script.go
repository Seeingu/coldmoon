package coldmoon

import "fmt"

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

// 16.1.6
func (s *ScriptRecord) Evaluate() Value {
	agent := s.Realm.Agent
	exe := NewExecutable()
	vm := NewVM(agent)
	s.ECMAScriptCode.Bytecode(exe)
	fmt.Println("Executable: ", exe.String())

	return vm.Run(exe)
}
