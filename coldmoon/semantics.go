package coldmoon

// ScriptStaticSemantics is the declaration snapshot computed when a Script is
// parsed. Runtime evaluation consumes this record instead of rediscovering
// static facts through AST interface assertions.
type ScriptStaticSemantics struct {
	VarDeclarations       []*VariableDeclaration
	LexicalDeclarations   []Declaration
	HoistableDeclarations []DeclarationHoistable
}

// FunctionBodyStaticSemantics is the declaration snapshot for a function body.
type FunctionBodyStaticSemantics struct {
	ParameterNames        []IdentifierName
	VarDeclarations       []*VariableDeclaration
	HoistableDeclarations []DeclarationHoistable
	LexicalDeclarations   []Declaration
	LexicalNames          []IdentifierName
}

// ModuleStaticSemantics is the import/export snapshot for a parsed Module.
type ModuleStaticSemantics struct {
	RequestedModules []string
	ImportEntries    []ImportEntryRecord
	ExportEntries    []ExportEntry
	VarDeclarations  []*VariableDeclaration
}

// StaticSemantics owns whole-program declaration and module queries.
type StaticSemantics struct{}

// AnalyzeScript computes the static facts used by script instantiation.
func (StaticSemantics) AnalyzeScript(script *Script) ScriptStaticSemantics {
	result := ScriptStaticSemantics{
		VarDeclarations: script.StatementList.VarScopedDeclarations(),
	}
	for _, item := range script.StatementList {
		declarationItem, ok := item.(*StatementListItemDeclaration)
		if !ok {
			continue
		}
		switch declaration := declarationItem.Declaration.(type) {
		case DeclarationHoistable:
			result.HoistableDeclarations = append(result.HoistableDeclarations, declaration)
		case *LexicalDeclaration, *ClassDeclaration:
			result.LexicalDeclarations = append(result.LexicalDeclarations, declarationItem.Declaration)
		}
	}
	return result
}

// AnalyzeFunctionBody computes declaration facts used by function
// instantiation.
func (StaticSemantics) AnalyzeFunctionBody(
	body *FunctionBody,
	parameters *FormalParameters,
) FunctionBodyStaticSemantics {
	return FunctionBodyStaticSemantics{
		ParameterNames:        parameters.BoundNames(),
		VarDeclarations:       body.VarScopedDeclarations(),
		HoistableDeclarations: body.StatementList.HoistableDeclarations(),
		LexicalDeclarations:   body.StatementList.LexicalDeclarations(),
		LexicalNames:          body.LexicallyDeclaredNames(),
	}
}

// AnalyzeModule computes module requests and import/export entries once.
func (StaticSemantics) AnalyzeModule(module *Module) ModuleStaticSemantics {
	return ModuleStaticSemantics{
		RequestedModules: module.moduleRequests(),
		ImportEntries:    module.importEntries(),
		ExportEntries:    module.exportEntries(),
		VarDeclarations:  module.ModuleItemList.VarScopedDeclarations(),
	}
}

// RuntimeSemantics owns AST evaluation for one execution context.
type RuntimeSemantics struct {
	vm *VM
}

func newRuntimeSemantics(agent *Agent) RuntimeSemantics {
	context := agent.RunningExecutionContext()
	Assert(context.VM != nil)
	return RuntimeSemantics{vm: context.VM}
}

// Evaluate runs node with context-local interpreter state.
func (r RuntimeSemantics) Evaluate(node ASTNode) CompletionValue {
	return node.Evaluation(r.vm)
}
