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
	HostDefined any
	// [[HasTLA]]
	HasTLA bool
}

func (s *SourceTextModule) _scriptOrModule() {}

// 16.2.1.6.1
func ParseModule(sourceText string, realm *Realm, hostDefined any, ctx ParserContext) *SourceTextModule {
	body := NewParser(sourceText, ctx).ParseModule()

	return &SourceTextModule{
		Realm:          realm,
		HostDefined:    hostDefined,
		ECMAScriptCode: body,
	}
}

// 16.2.1.6.4
func (s *SourceTextModule) InitializeEnvironment() {
	s.Environment = s.Realm.GlobalEnv
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
