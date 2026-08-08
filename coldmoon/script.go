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
	fileName := "file.js"
	baseDir := ""
	if hostDefined != nil {
		if hostDefined.FileName != "" {
			fileName = hostDefined.FileName
		}
		baseDir = hostDefined.BaseDir
	}
	return parseScript(sourceText, realm, hostDefined, ParserContext{
		FileName: fileName,
		BaseDir:  baseDir,
	})
}

func parseScript(sourceText string, realm *Realm, hostDefined *HostDefined, context ParserContext) *ScriptRecord {
	script := NewParser(sourceText, context).Parse()

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
	result := s.evaluateCompletion()
	if result.IsError() {
		return result.Error()
	}
	return result.Data()
}

func (s *ScriptRecord) evaluateCompletion() CompletionValue {
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

	return s.evaluateInCurrentContext()
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
	if globalEnv, ok := variableEnv.(*GlobalEnvironment); ok {
		for _, declaration := range s.Static.VarDeclarations {
			for _, name := range declaration.BoundNames() {
				if globalEnv.DeclarativeRecord.HasBinding(string(name)) {
					var completion CompletionValue
					return completion.ThrowError(agent, SyntaxError, fmt.Sprintf("Identifier %q has already been declared", name))
				}
			}
		}
		for _, declaration := range s.Static.HoistableDeclarations {
			for _, name := range declaration.BoundNames() {
				if globalEnv.DeclarativeRecord.HasBinding(string(name)) {
					var completion CompletionValue
					return completion.ThrowError(agent, SyntaxError, fmt.Sprintf("Identifier %q has already been declared", name))
				}
			}
		}
	}
	lexicalNames := make(map[IdentifierName]bool)
	for _, declaration := range s.Static.LexicalDeclarations {
		for _, name := range declaration.BoundNames() {
			hasExistingDeclaration := lexicalEnv.HasBinding(string(name))
			if globalEnv, ok := lexicalEnv.(*GlobalEnvironment); ok {
				hasExistingDeclaration = globalEnv.HasLexicalDeclaration(string(name)) || globalEnv.HasVarDeclaration(string(name))
			}
			if lexicalNames[name] || hasExistingDeclaration {
				var completion CompletionValue
				return completion.ThrowError(agent, SyntaxError, fmt.Sprintf("Identifier %q has already been declared", name))
			}
			lexicalNames[name] = true
		}
	}

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
		for _, varName := range decl.BoundNames() {
			if _, ok := seen[varName]; !ok {
				createVariableBinding(variableEnv, string(varName))
				seen[varName] = true
			}
		}
	}

	privateEnv := context.ECMAScriptCode.PrivateEnvironment
	for _, declaration := range s.Static.HoistableDeclarations {
		name, function := instantiateHoistableDeclaration(agent, declaration, lexicalEnv, privateEnv)
		createVariableBinding(variableEnv, name)
		variableEnv.SetMutableBinding(agent, name, function.ToValue(), false)
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
