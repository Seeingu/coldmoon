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
	script := NewParser(sourceText, ParserContext{FileName: "file.js"}).Parse()

	s := &ScriptRecord{
		Realm:          realm,
		HostDefined:    hostDefined,
		ECMAScriptCode: script,
	}

	fmt.Println("AST: ", s.ECMAScriptCode.String())
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
	defer agent.ExecutionContextStack.Pop()

	return GenerateAndRunBytecode(agent, s.ECMAScriptCode).Value
}
