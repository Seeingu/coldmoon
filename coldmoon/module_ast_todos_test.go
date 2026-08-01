package coldmoon

import (
	"reflect"
	"slices"
	"testing"
)

func moduleASTStringLiteral(value string) *StringLiteral {
	return &StringLiteral{Value: value}
}

func moduleASTIdentifierName(value string) *ModuleExportName {
	return &ModuleExportName{IdentifierName: IdentifierName(value)}
}

func moduleASTStringName(value string) *ModuleExportName {
	return &ModuleExportName{StringLiteral: moduleASTStringLiteral(value)}
}

func moduleASTImport(moduleRequest string, clause *ImportClause) *ModuleItemImportDeclaration {
	return &ModuleItemImportDeclaration{ImportDeclaration: &ImportDeclaration{
		ImportClause:    clause,
		ModuleSpecifier: moduleASTStringLiteral(moduleRequest),
	}}
}

func moduleASTExportFrom(moduleRequest string, clause *ExportFromClause) *ModuleItemExportDeclaration {
	return &ModuleItemExportDeclaration{ExportFrom: &ExportFrom{
		ExportFromClause: clause,
		ModuleSpecifier:  moduleASTStringLiteral(moduleRequest),
	}}
}

func TestModuleASTImportEntriesRequestsBoundNamesAndString(t *testing.T) {
	defaultAndNamespace := &ImportClause{
		ImportedDefaultBinding: "defaultValue",
		NamespaceImport:        "namespaceValue",
	}
	defaultAndNamed := &ImportClause{
		ImportedDefaultBinding: "otherDefault",
		NamedImports: &ImportsList{Items: []*ImportSpecifier{
			{ImportedBinding: "local"},
			{
				ImportedBinding:  "renamed",
				ModuleExportName: moduleASTStringName("external-name"),
			},
		}},
	}
	module := &Module{ModuleItemList: ModuleItemList{
		moduleASTImport("./side", nil),
		moduleASTImport("./combo", defaultAndNamespace),
		moduleASTImport("./named", defaultAndNamed),
		moduleASTExportFrom("./named", &ExportFromClause{Star: true}),
		moduleASTExportFrom("./fourth", &ExportFromClause{Star: true}),
	}}

	if got, want := module.moduleRequests(), []string{"./side", "./combo", "./named", "./fourth"}; !slices.Equal(got, want) {
		t.Fatalf("module requests = %v, want %v", got, want)
	}
	wantEntries := []ImportEntryRecord{
		{ModuleRequest: "./combo", ImportName: "default", LocalName: "defaultValue"},
		{ModuleRequest: "./combo", ImportName: ImportNameNamespaceObject, LocalName: "namespaceValue"},
		{ModuleRequest: "./named", ImportName: "default", LocalName: "otherDefault"},
		{ModuleRequest: "./named", ImportName: "local", LocalName: "local"},
		{ModuleRequest: "./named", ImportName: "external-name", LocalName: "renamed"},
	}
	if got := module.importEntries(); !reflect.DeepEqual(got, wantEntries) {
		t.Fatalf("import entries = %#v, want %#v", got, wantEntries)
	}
	if got, want := defaultAndNamespace.BoundNames(), []IdentifierName{"defaultValue", "namespaceValue"}; !slices.Equal(got, want) {
		t.Fatalf("default+namespace BoundNames = %v, want %v", got, want)
	}
	if got, want := defaultAndNamed.BoundNames(), []IdentifierName{"otherDefault", "local", "renamed"}; !slices.Equal(got, want) {
		t.Fatalf("default+named BoundNames = %v, want %v", got, want)
	}

	wantString := "import \"./side\";\n" +
		"import defaultValue, * as namespaceValue from \"./combo\";\n" +
		"import otherDefault, {local, \"external-name\" as renamed} from \"./named\";\n" +
		"export * from \"./named\";\n" +
		"export * from \"./fourth\";"
	if got := module.String(); got != wantString {
		t.Fatalf("module String() = %q, want %q", got, wantString)
	}
}

func TestModuleASTExportEntriesCoverEveryDeclarationForm(t *testing.T) {
	localNamed := &ModuleItemExportDeclaration{NamedExports: &NamedExports{
		ExportsList: &ExportsList{Items: []*ExportSpecifier{
			{Name: moduleASTIdentifierName("local")},
			{Name: moduleASTIdentifierName("source"), Alias: moduleASTIdentifierName("alias")},
		}},
	}}
	variable := &ModuleItemExportDeclaration{VariableStatement: &VariableStatement{
		DeclarationList: &VariableDeclarationList{Items: []*VariableDeclaration{
			{BindingIdentifier: "first"},
			{BindingIdentifier: "second"},
		}},
	}}
	lexical := &ModuleItemExportDeclaration{Declaration: &LexicalDeclaration{
		Type: LetOrConstConst,
		BindingList: &BindingList{Items: []*LexicalBinding{
			{Identifier: "lexicalA"},
			{Identifier: "lexicalB"},
		}},
	}}
	defaultFunction := &ModuleItemExportDeclaration{
		DefaultHoistableDeclaration: &DeclarationHoistableFunction{
			FunctionDeclaration: &FunctionDeclaration{Identifier: "namedDefault"},
		},
	}
	defaultClass := &ModuleItemExportDeclaration{
		DefaultClassDeclaration: &ClassDeclaration{},
	}
	defaultExpression := &ModuleItemExportDeclaration{
		DefaultExpression: &NumericLiteral{Value: "1"},
	}

	module := &Module{ModuleItemList: ModuleItemList{
		localNamed,
		moduleASTExportFrom("./star", &ExportFromClause{Star: true}),
		moduleASTExportFrom("./namespace", &ExportFromClause{StarAs: moduleASTStringName("ns-name")}),
		moduleASTExportFrom("./named", &ExportFromClause{NamedExports: &NamedExports{
			ExportsList: &ExportsList{Items: []*ExportSpecifier{
				{Name: moduleASTIdentifierName("remote"), Alias: moduleASTIdentifierName("renamed")},
				{Name: moduleASTStringName("strange-name"), Alias: moduleASTStringName("quoted-export")},
			}},
		}}),
		variable,
		lexical,
		defaultFunction,
		defaultClass,
		defaultExpression,
	}}

	want := []ExportEntry{
		{ExportName: "local", LocalName: "local"},
		{ExportName: "alias", LocalName: "source"},
		{ModuleRequest: "./star", ImportName: ImportNameAllButDefault},
		{ExportName: "ns-name", ModuleRequest: "./namespace", ImportName: ImportNameAll},
		{ExportName: "renamed", ModuleRequest: "./named", ImportName: "remote"},
		{ExportName: "quoted-export", ModuleRequest: "./named", ImportName: "strange-name"},
		{ExportName: "first", LocalName: "first"},
		{ExportName: "second", LocalName: "second"},
		{ExportName: "lexicalA", LocalName: "lexicalA"},
		{ExportName: "lexicalB", LocalName: "lexicalB"},
		{ExportName: "default", LocalName: "namedDefault"},
		{ExportName: "default", LocalName: "*default*"},
		{ExportName: "default", LocalName: "*default*"},
	}
	if got := module.exportEntries(); !reflect.DeepEqual(got, want) {
		t.Fatalf("export entries = %#v, want %#v", got, want)
	}
	if got, want := variable.BoundNames(), []IdentifierName{"first", "second"}; !slices.Equal(got, want) {
		t.Fatalf("variable export BoundNames = %v, want %v", got, want)
	}
	if got, want := defaultFunction.BoundNames(), []IdentifierName{"namedDefault", "*default*"}; !slices.Equal(got, want) {
		t.Fatalf("named default BoundNames = %v, want %v", got, want)
	}
	if got, want := defaultClass.BoundNames(), []IdentifierName{"*default*"}; !slices.Equal(got, want) {
		t.Fatalf("anonymous default class BoundNames = %v, want %v", got, want)
	}
	if got, want := defaultExpression.BoundNames(), []IdentifierName{"*default*"}; !slices.Equal(got, want) {
		t.Fatalf("default expression BoundNames = %v, want %v", got, want)
	}
}

func TestModuleASTDefaultHoistableVariants(t *testing.T) {
	variants := []struct {
		name      string
		named     DeclarationHoistable
		anonymous DeclarationHoistable
	}{
		{
			name:      "function",
			named:     &DeclarationHoistableFunction{FunctionDeclaration: &FunctionDeclaration{Identifier: "functionName"}},
			anonymous: &DeclarationHoistableFunction{FunctionDeclaration: &FunctionDeclaration{}},
		},
		{
			name:      "generator",
			named:     &DeclarationHoistableGenerator{GeneratorDeclaration: &GeneratorDeclaration{Identifier: "generatorName"}},
			anonymous: &DeclarationHoistableGenerator{GeneratorDeclaration: &GeneratorDeclaration{}},
		},
		{
			name:      "async function",
			named:     &DeclarationHoistableAsyncFunction{AsyncFunctionDeclaration: &AsyncFunctionDeclaration{Identifier: "asyncFunctionName"}},
			anonymous: &DeclarationHoistableAsyncFunction{AsyncFunctionDeclaration: &AsyncFunctionDeclaration{}},
		},
		{
			name:      "async generator",
			named:     &DeclarationHoistableAsyncGenerator{AsyncGeneratorDeclaration: &AsyncGeneratorDeclaration{Identifier: "asyncGeneratorName"}},
			anonymous: &DeclarationHoistableAsyncGenerator{AsyncGeneratorDeclaration: &AsyncGeneratorDeclaration{}},
		},
	}

	for _, variant := range variants {
		t.Run(variant.name+" named", func(t *testing.T) {
			export := &ModuleItemExportDeclaration{DefaultHoistableDeclaration: variant.named}
			declarationNames := moduleDeclarationBoundNames(variant.named)
			if len(declarationNames) != 1 {
				t.Fatalf("declaration names = %v, want one name", declarationNames)
			}
			name := declarationNames[0]
			if got, want := export.BoundNames(), []IdentifierName{name, "*default*"}; !slices.Equal(got, want) {
				t.Fatalf("BoundNames = %v, want %v", got, want)
			}
			if got, want := export.exportEntries(), []ExportEntry{{ExportName: "default", LocalName: string(name)}}; !reflect.DeepEqual(got, want) {
				t.Fatalf("export entries = %#v, want %#v", got, want)
			}
		})

		t.Run(variant.name+" anonymous", func(t *testing.T) {
			export := &ModuleItemExportDeclaration{DefaultHoistableDeclaration: variant.anonymous}
			if got, want := export.BoundNames(), []IdentifierName{"*default*"}; !slices.Equal(got, want) {
				t.Fatalf("BoundNames = %v, want %v", got, want)
			}
			if got, want := export.exportEntries(), []ExportEntry{{ExportName: "default", LocalName: "*default*"}}; !reflect.DeepEqual(got, want) {
				t.Fatalf("export entries = %#v, want %#v", got, want)
			}
		})
	}
}

func TestModuleASTEvaluationContinuesPastImports(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	module := &Module{ModuleItemList: ModuleItemList{
		moduleASTImport("./before", nil),
		&ModuleItemStatementListItem{StatementListItem: &StatementListItemStatement{
			Statement: &StatementExpression{Expression: &NumericLiteral{Value: "7"}},
		}},
		moduleASTImport("./after", nil),
		&ModuleItemExportDeclaration{NamedExports: &NamedExports{}},
	}}

	result := RunNode(agent, module)
	if result.IsAbrupt() {
		t.Fatalf("module evaluation was abrupt: %v", result.Error())
	}
	number, ok := result.Data().(*NumberValue)
	if !ok || number.Data != 7 {
		t.Fatalf("module evaluation result = %#v, want 7", result.Data())
	}
}

func TestModuleASTDefaultExpressionInitializesSyntheticBinding(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	env := NewModuleEnvironment(realm.GlobalEnv)
	env.CreateMutableBinding("*default*", false)
	context := &ExecutionContext{
		Realm: realm,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  env,
			VariableEnvironment: env,
		},
	}
	scope := agent.enterExecutionContext(context)
	defer scope.Leave()

	export := &ModuleItemExportDeclaration{DefaultExpression: &NumericLiteral{Value: "42"}}
	result := export.Evaluation(context.VM)
	if result.IsAbrupt() {
		t.Fatalf("default export evaluation was abrupt: %v", result.Error())
	}
	binding := env.GetBindingValue(agent, "*default*", true)
	if binding.IsAbrupt() {
		t.Fatalf("reading *default* was abrupt: %v", binding.Error())
	}
	number, ok := binding.Data().(*NumberValue)
	if !ok || number.Data != 42 {
		t.Fatalf("*default* = %#v, want 42", binding.Data())
	}
}
