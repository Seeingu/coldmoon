package coldmoon

type ImportName interface {
	_importName()
}

type ImportNameString struct {
	ImportName
	String string
}

func (i ImportNameString) _importName() {}

type ImportNameNamespaceObject struct {
	ImportName
}

func (i ImportNameNamespaceObject) _importName() {}

type ImportNameAll struct {
	ImportName
}

func (i ImportNameAll) _importName() {}

type ImportNameAllButDefault struct {
	ImportName
}

func (i ImportNameAllButDefault) _importName() {

}

type ImportEntry struct {
	// [[ModuleRequest]]
	ModuleRequest string
	// [[ImportName]]
	ImportName ImportName
	// [[LocalName]]
	LocalName string
}

type ExportEntry struct {
	// [[ExportName]]
	ExportName string
	// [[ModuleRequest]]
	ModuleRequest string
	// [[ImportName]]
	ImportName ImportName
	// [[LocalName]]
	LocalName string
}

type SourceTextModule struct {
	ScriptOrModule
	// [[Realm]]
	Realm *Realm
	// [[Environment]]
	Environment EnvironmentRecord
	// [[Namespace]]
	Namespace ObjectType
	// [[ECMAScriptCode]]
	ECMAScriptCode *Module
	// [[Context]]
	Context *ExecutionContext
	// [[ImportMeta]]
	ImportMeta ObjectType
	// [[ImportEntries]]
	ImportEntries []ImportEntry
	// [[LocalExportEntries]]
	LocalExportEntries []ExportEntry
	// [[IndirectExportEntries]]
	IndirectExportEntries []ExportEntry
	// [[StarExportEntries]]
	StarExportEntries []ExportEntry
	// [[HostDefined]]
	HostDefined *HostDefined
	// [[HasTLA]]
	HasTLA bool
}

func (s *SourceTextModule) _scriptOrModule() {}
func (s *SourceTextModule) ToReferrer() ImportedModuleReferrer {
	return ImportedModuleReferrer{
		Script: nil,
		Module: s,
		Realm:  nil,
	}
}

func (s *SourceTextModule) Link() CompletionValue {
	// TODO
	return NewCompletionValue(UndefinedValue)
}

// 16.2.1.5.3
func (s *SourceTextModule) Evaluate() *PromiseObject {
	// TODO
	return nil
}

// [[BindingName]] union
type BindingName struct {
	String    string
	Namespace bool
}
type ResolvedBinding struct {
	// [[Module]]
	Module *ModuleRecord
	// [[BindingName]]
	BindingName *BindingName
	Ambiguous   bool
}

func (s *SourceTextModule) ResolveExport(exportName string, resolveSet []*ModuleRecord) *ResolvedBinding {
	// TODO
	return nil
}

func (s *SourceTextModule) GetExportedNames() Exports {
	// TODO
	return []string{}
}

// 16.2.1.6.1
func ParseModule(sourceText string, realm *Realm, hostDefined *HostDefined, ctx ParserContext) *SourceTextModule {
	body := NewParser(sourceText, ctx).ParseModule()

	return &SourceTextModule{
		Realm:          realm,
		HostDefined:    hostDefined,
		ECMAScriptCode: body,
	}
}

// 16.2.1.5.1.2
func ContinueModuleLoading(agent *Agent, state *GraphLoadingState, moduleCompletion CompletionModule) {
	// TODO
}

// 16.2.1.5.1
func (s *SourceTextModule) LoadRequestedModules(agent *Agent, hostDefined *HostDefined) ObjectType {
	realm := agent.CurrentRealm()
	pc := NewPromiseCapability(agent, realm.Intrinsics.Promise.ToValue())

	state := &GraphLoadingState{}

	InnerModuleLoading(agent, state, pc)
	return pc.Promise
}

// 16.2.1.6.4
func (s *SourceTextModule) InitializeEnvironment() CompletionValue {
	realm := s.Realm
	agent := realm.Agent
	env := NewModuleEnvironment(realm.GlobalEnv)
	s.Environment = env

	for _, importEntry := range s.ImportEntries {
		importedModule := GetImportedModule(s, importEntry.ModuleRequest)
		importName := importEntry.ImportName
		localName := importEntry.LocalName
		if _, ok := importName.(*ImportNameNamespaceObject); ok {
			namespace := GetModuleNamespace(agent, importedModule)
			env.CreateImmutableBinding(localName, true)
			env.InitializeBinding(localName, namespace.ToValue())
		} else {
			resolution := importedModule.SourceTextModule.ResolveExport(importName.(ImportNameString).String, []*ModuleRecord{})
			if resolution == nil {
				return NewCompletionValueError(
					agent.ThrowException(SyntaxError, "Failed to resolve export"))
			} else if resolution.Ambiguous {
				return NewCompletionValueError(
					agent.ThrowException(SyntaxError, "Ambiguous export"))
			}
			if resolution.BindingName.Namespace {
				namespace := GetModuleNamespace(agent, resolution.Module)
				env.CreateImmutableBinding(localName, true)
				env.InitializeBinding(localName, namespace.ToValue())
			} else {
				env.CreateImportBinding(localName, resolution.Module.SourceTextModule, resolution.BindingName.String)
			}
		}
	}

	moduleContext := &ExecutionContext{
		Realm:          realm,
		Function:       nil,
		ScriptOrModule: s,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  env,
			VariableEnvironment: env,
			PrivateEnvironment:  nil,
		},
	}
	s.Context = moduleContext
	agent.ExecutionContextStack.Push(moduleContext)
	code := s.ECMAScriptCode

	varDeclarations := code.ModuleItemList.VarScopedDeclarations()

	declaredVarNames := make(map[string]bool)

	for _, varDeclaration := range varDeclarations {
		varName := string(varDeclaration.BindingIdentifier)
		if _, ok := declaredVarNames[varName]; !ok {
			env.CreateMutableBinding(varName, false)
			env.InitializeBinding(varName, UndefinedValue)
			declaredVarNames[varName] = true
		}
	}

	// TODO:

	agent.ExecutionContextStack.Pop()
	return NewCompletionValue(UndefinedValue)
}

func (s *SourceTextModule) ExecuteModule(capability *PromiseCapability) {
	agent := s.Realm.Agent
	moduleContext := &ExecutionContext{
		Realm:          s.Realm,
		ScriptOrModule: s,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  s.Environment,
			VariableEnvironment: s.Environment,
		},
	}
	if !s.HasTLA {
		Assert(capability == nil)
		agent.ExecutionContextStack.Push(moduleContext)
		GenerateAndRunBytecode(agent, s.ECMAScriptCode)
		agent.ExecutionContextStack.Pop()
	} else {

	}

}

func InnerModuleLoading(agent *Agent, state *GraphLoadingState, capability *PromiseCapability) {
	// TODO
}
