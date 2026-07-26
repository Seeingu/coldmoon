package coldmoon

import (
	"fmt"
)

type ScriptRecord struct {
	ScriptOrModule
	Realm          *Realm
	ECMAScriptCode *Script
	Static         ScriptStaticSemantics
	LoadedModules  map[string]ModuleRecord
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
		Static:         (StaticSemantics{}).AnalyzeScript(script),
		LoadedModules:  make(map[string]ModuleRecord),
	}

	if Debug.PrintAST {
		fmt.Println("AST: ", s.ECMAScriptCode.String())
	}
	return s
}

// 16.1.6
func (s *ScriptRecord) Evaluate() Value {
	agent := s.Realm.Agent

	globalEnv := s.Realm.GlobalEnv
	scriptContext := &ExecutionContext{
		Function:       nil,
		Realm:          s.Realm,
		ScriptOrModule: s,
		ch:             make(chan struct{}),
		ECMAScriptCode: &ExecutionContextAdditionalState{
			VariableEnvironment: globalEnv,
			LexicalEnvironment:  globalEnv,
			PrivateEnvironment:  nil,
		},
	}
	scope := agent.enterExecutionContext(scriptContext)
	defer scope.Leave()

	result := s.evaluateInCurrentContext()
	if result.IsError() {
		return result.Error()
	}
	return result.Data()
}

// evaluateInCurrentContext evaluates the script using the lexical and variable
// environments selected by its caller. PerformEval uses this entry point so it
// does not accidentally create a second, global script context.
func (s *ScriptRecord) evaluateInCurrentContext() CompletionValue {
	agent := s.Realm.Agent
	context := agent.RunningExecutionContext()
	Assert(context.ECMAScriptCode != nil)
	lexicalEnv := context.ECMAScriptCode.LexicalEnvironment
	variableEnv := context.ECMAScriptCode.VariableEnvironment
	script := s.ECMAScriptCode

	for _, declaration := range s.Static.LexicalDeclarations {
		switch declaration := declaration.(type) {
		case *LexicalDeclaration:
			for _, name := range declaration.BoundNames() {
				if declaration.IsConstantDeclaration() {
					lexicalEnv.CreateImmutableBinding(name, true)
				} else {
					lexicalEnv.CreateMutableBinding(name, false)
				}
			}
		case *ClassDeclaration:
			for _, name := range declaration.BoundNames() {
				lexicalEnv.CreateMutableBinding(name, false)
			}
		}
	}

	seen := make(map[IdentifierName]bool)
	for _, decl := range s.Static.VarDeclarations {
		varName := decl.BindingIdentifier
		if _, ok := seen[varName]; !ok {
			createVariableBinding(variableEnv, string(varName))
			seen[varName] = true
		}
	}

	for _, item := range script.StatementList {
		declarationItem, ok := item.(*StatementListItemDeclaration)
		if !ok {
			continue
		}
		hoistable, ok := declarationItem.Declaration.(*DeclarationHoistableFunction)
		if !ok {
			continue
		}
		functionDeclaration := hoistable.FunctionDeclaration
		name := string(functionDeclaration.Identifier)
		function := functionDeclaration.instantiateOrdinaryFunctionObject(agent, lexicalEnv, nil)
		createVariableBinding(variableEnv, name)
		variableEnv.SetMutableBinding(name, function.ToValue(), false)
	}

	return RunNode(agent, s.ECMAScriptCode)
}

func createVariableBinding(environment EnvironmentRecord, name string) {
	if global, ok := environment.(*GlobalEnvironment); ok {
		global.CreateGlobalVarBinding(name, true)
		return
	}
	if environment.HasBinding(name) {
		return
	}
	environment.CreateMutableBinding(name, false)
	environment.InitializeBinding(name, UndefinedValue)
}
