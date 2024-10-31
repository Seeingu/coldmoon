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

	globalEnv := agent.CurrentRealm().GlobalEnv
	scriptContext := &ExecutionContext{
		Function:       nil,
		Realm:          s.Realm,
		ScriptOrModule: TScript,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			VariableEnvironment: globalEnv,
			LexicalEnvironment:  globalEnv,
			PrivateEnvironment:  nil,
		},
	}

	agent.ExecutionContextStack.Push(scriptContext)

	script := s.ECMAScriptCode

	exe := NewExecutable()
	vm := NewVM(agent)
	script.Bytecode(exe)
	fmt.Println("Executable: ", exe.String())
	result := vm.Run(exe)

	agent.ExecutionContextStack.Pop()
	Assert(!agent.ExecutionContextStack.IsEmpty())

	return result
}
