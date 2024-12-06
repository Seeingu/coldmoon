package coldmoon

import "fmt"

type ScriptRecord struct {
	ScriptOrModule
	Realm          *Realm
	ECMAScriptCode *Script
	LoadedModules  map[string]*ModuleRecord
	HostDefined    *HostDefined
}

func (s *ScriptRecord) _scriptOrModule() {}
func (s *ScriptRecord) ToReferrer() ImportedModuleReferrer {
	return ImportedModuleReferrer{
		Script: s,
		Module: nil,
		Realm:  nil,
	}
}

// ParseScript
// 16.1.5
func ParseScript(sourceText string, realm *Realm, hostDefined *HostDefined) *ScriptRecord {
	script := NewParser(sourceText, ParserContext{FileName: "file.js"}).Parse()

	s := &ScriptRecord{
		Realm:          realm,
		HostDefined:    hostDefined,
		ECMAScriptCode: script,
	}

	if Debug.PrintAST {
		fmt.Println("AST: ", s.ECMAScriptCode.String())
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
		ScriptOrModule: s,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			VariableEnvironment: globalEnv,
			LexicalEnvironment:  globalEnv,
			PrivateEnvironment:  nil,
		},
	}
	agent.ExecutionContextStack.Push(scriptContext)
	defer agent.ExecutionContextStack.Pop()

	script := s.ECMAScriptCode

	varScopedDeclarations := script.StatementList.VarScopedDeclarations()
	seen := make(map[IdentifierName]bool)
	for _, decl := range varScopedDeclarations {
		varName := decl.BindingIdentifier
		if _, ok := seen[varName]; !ok {
			globalEnv.CreateGlobalVarBinding(string(varName), true)
			seen[varName] = true
		}
	}

	return GenerateAndRunBytecode(agent, s.ECMAScriptCode).Data()
}
