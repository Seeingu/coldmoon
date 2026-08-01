package coldmoon

import "testing"

func expectParserPanic(t *testing.T, source string) {
	t.Helper()
	deferredPanic := false
	func() {
		defer func() {
			deferredPanic = recover() != nil
		}()
		NewParser(source, ParserContext{}).Parse()
	}()
	if !deferredPanic {
		t.Fatalf("parsing %q did not panic", source)
	}
}

func TestTokenizerParsesBigIntLiteralForms(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{source: "123456789012345678901234567890n", want: "123456789012345678901234567890"},
		{source: "0x20000000000001n", want: "9007199254740993"},
		{source: "0b1010_0001n", want: "161"},
		{source: "0o7_7n", want: "63"},
		{source: "1_000_000n", want: "1000000"},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			token := NewTokenizer(test.source).CurrentToken
			if token.Type != TBigInt {
				t.Fatalf("token type = %v, want TBigInt", token.Type)
			}
			if token.Value != test.want {
				t.Fatalf("token value = %q, want %q", token.Value, test.want)
			}
			if token.StartIndex != 0 || token.EndIndex != len([]rune(test.source)) {
				t.Fatalf("token range = [%d,%d), want [0,%d)", token.StartIndex, token.EndIndex, len([]rune(test.source)))
			}
		})
	}
}

func TestTokenizerRejectsInvalidBigIntLiteralForms(t *testing.T) {
	for _, source := range []string{"1.5n", "1e3n", "01n", "1N", "0x_n", "1_n"} {
		t.Run(source, func(t *testing.T) {
			deferredPanic := false
			func() {
				defer func() {
					deferredPanic = recover() != nil
				}()
				NewTokenizer(source)
			}()
			if !deferredPanic {
				t.Fatalf("tokenizing %q did not panic", source)
			}
		})
	}
}

func TestTokenizerUsesExpressionContextForSlash(t *testing.T) {
	division := NewTokenizer("1n / 2n")
	if division.NextToken.Type != TSlash {
		t.Fatalf("slash after BigInt tokenized as %v, want division", division.NextToken.Type)
	}

	regexp := NewTokenizer("return /value/")
	if regexp.NextToken.Type != TRegularExpression {
		t.Fatalf("slash after return tokenized as %v, want regular expression", regexp.NextToken.Type)
	}

	comment := NewTokenizer("left/* comment */+right")
	if comment.CurrentToken.Type != TIdentifier || comment.NextToken.Type != TPlus {
		t.Fatalf("token after block comment = %v, want plus", comment.NextToken.Type)
	}

	NewParser("const quotient = {} / 2 / 3;", ParserContext{}).Parse()
	NewParser("const quotient = function() {} / 2 / 3;", ParserContext{}).Parse()
	NewParser("{} /value/.test('value');", ParserContext{}).Parse()
	NewParser("function value() {} /value/.test('value');", ParserContext{}).Parse()
}

func TestParserValidatesLabelTargetsAndFunctionBoundaries(t *testing.T) {
	NewParser("first: second: for (let i = 0; i < 1; i++) { continue first; }", ParserContext{}).Parse()
	expectParserPanic(t, "block: { continue block; }")
	expectParserPanic(t, "outer: { function nested() { break outer; } }")
	expectParserPanic(t, "outer: for (let i = 0; i < 1; i++) { function nested() { continue outer; } }")
}

func TestParserRejectsInvalidEscapeInUntaggedTemplate(t *testing.T) {
	expectParserPanic(t, "const invalid = `\\xZ1`;")
	expectParserPanic(t, "const invalid = `\\u{110000}`;")
}

func TestParserBuildsExportFromDeclarations(t *testing.T) {
	module := NewParser(`
export * from "all";
export * as namespace from "namespace";
export { value as renamed } from "named";
`, ParserContext{}).ParseModule()
	entries := module.exportEntries()
	if len(entries) != 3 {
		t.Fatalf("export entries = %d, want 3", len(entries))
	}
	if entries[0].ModuleRequest != "all" || entries[0].ImportName != ImportNameAllButDefault {
		t.Fatalf("star export = %#v", entries[0])
	}
	if entries[1].ModuleRequest != "namespace" || entries[1].ImportName != ImportNameAll || entries[1].ExportName != "namespace" {
		t.Fatalf("namespace export = %#v", entries[1])
	}
	if entries[2].ModuleRequest != "named" || entries[2].ImportName != "value" || entries[2].ExportName != "renamed" {
		t.Fatalf("named export = %#v", entries[2])
	}
}

func TestParserHonoursLineTerminatorsAndASI(t *testing.T) {
	script := NewParser("function value() { return\u2028 41; }", ParserContext{}).Parse()
	declaration := script.StatementList[0].(*StatementListItemDeclaration).Declaration.(*DeclarationHoistableFunction)
	returnStatement := declaration.FunctionDeclaration.Body.StatementList[0].(*StatementListItemStatement).Statement.(*ReturnStatement)
	if returnStatement.Expression != nil {
		t.Fatal("return followed by U+2028 unexpectedly retained its expression")
	}

	NewParser("const arrow = value =>\n value + 1;", ParserContext{}).Parse()
	NewParser("async\nfunction separated() {}", ParserContext{}).Parse()
	expectParserPanic(t, "function value() { throw /* line\nterminator */ 41; }")
	expectParserPanic(t, "const arrow = value\n=> value;")
	expectParserPanic(t, "let first = 1 let second = 2;")
}

func TestParserPreservesCasesAfterDefault(t *testing.T) {
	script := NewParser("switch (value) { case 0: zero(); default: fallback(); case 1: one(); }", ParserContext{}).Parse()
	statement := script.StatementList[0].(*StatementListItemStatement).Statement.(*BreakableStatement)
	caseBlock := statement.SwitchStatement.CaseBlock
	if len(caseBlock.CaseClauses) != 1 {
		t.Fatalf("cases before default = %d, want 1", len(caseBlock.CaseClauses))
	}
	if caseBlock.DefaultClause == nil {
		t.Fatal("default clause is nil")
	}
	if len(caseBlock.CaseClausesAfterDefault) != 1 {
		t.Fatalf("cases after default = %d, want 1", len(caseBlock.CaseClausesAfterDefault))
	}
	expectParserPanic(t, "switch (value) { default: first(); default: second(); }")
	expectParserPanic(t, "switch (value) { case 1:")
	expectParserPanic(t, "switch (value) { default:")
}

func TestParserBuildsBindingPatterns(t *testing.T) {
	script := NewParser("let {x: renamed, y = 3, nested: [head, ...tail], ...other} = source;", ParserContext{}).Parse()
	declaration := script.StatementList[0].(*StatementListItemDeclaration).Declaration.(*LexicalDeclaration)
	binding := declaration.BindingList.Items[0]
	objectPattern := binding.BindingPattern.ObjectBindingPattern
	if objectPattern == nil || len(objectPattern.Properties) != 1 {
		t.Fatal("object binding pattern was not preserved")
	}
	properties := objectPattern.Properties[0]
	if len(properties.BindingPropertyList) != 3 {
		t.Fatalf("binding properties = %d, want 3", len(properties.BindingPropertyList))
	}
	if properties.BindingRestProperty == nil || properties.BindingRestProperty.BindingIdentifier != "other" {
		t.Fatal("object rest binding was not preserved")
	}

	nested := properties.BindingPropertyList[2].PropertyNameAndBindingElement.BindingElement.BindingPattern.ArrayBindingPattern
	if nested == nil || len(nested.Elements) != 2 {
		t.Fatal("nested array binding pattern was not preserved")
	}
	if _, ok := nested.Elements[1].BindingRestElement.(*BindingRestElementIdentifier); !ok {
		t.Fatal("nested array rest element has the wrong AST type")
	}

	withInitializer := NewParser("let {value: [first] = []} = source;", ParserContext{}).Parse()
	initializedDeclaration := withInitializer.StatementList[0].(*StatementListItemDeclaration).Declaration.(*LexicalDeclaration)
	initializedProperty := initializedDeclaration.BindingList.Items[0].BindingPattern.ObjectBindingPattern.Properties[0].BindingPropertyList[0]
	if initializedProperty.PropertyNameAndBindingElement.BindingElement.Initializer == nil {
		t.Fatal("nested binding-pattern initializer was not preserved")
	}

	restPattern := NewParser("function collect(...[first, ...rest]) {}", ParserContext{}).Parse()
	function := restPattern.StatementList[0].(*StatementListItemDeclaration).Declaration.(*DeclarationHoistableFunction).FunctionDeclaration
	restParameter := function.FormalParameters.Items[0].(*FormalParameterFunctionRestParameter)
	if _, ok := restParameter.BindingRestElement.(*BindingRestElementPattern); !ok {
		t.Fatal("rest binding pattern has the wrong AST type")
	}

	variableScript := NewParser("var [first, second] = pair;", ParserContext{}).Parse()
	variableStatement := variableScript.StatementList[0].(*StatementListItemStatement).Statement.(*VariableStatement)
	if variableStatement.DeclarationList.Items[0].BindingPattern.ArrayBindingPattern == nil {
		t.Fatal("var binding pattern was not preserved")
	}

	NewParser("({cover = 7} = source);", ParserContext{}).Parse()
}

func TestParserSupportsFinallyOnlyAndAnonymousClassExpression(t *testing.T) {
	tryScript := NewParser("try {} finally {}", ParserContext{}).Parse()
	tryStatement := tryScript.StatementList[0].(*StatementListItemStatement).Statement.(*TryStatement)
	if tryStatement.Catch != nil || tryStatement.FinallyBlock == nil {
		t.Fatal("finally-only try statement has an incorrect AST shape")
	}
	catchScript := NewParser("try {} catch {}", ParserContext{}).Parse()
	catchStatement := catchScript.StatementList[0].(*StatementListItemStatement).Statement.(*TryStatement)
	if catchStatement.Catch == nil || catchStatement.Catch.CatchParameter != nil {
		t.Fatal("binding-less catch has an incorrect AST shape")
	}

	classScript := NewParser("const Anonymous = class {};", ParserContext{}).Parse()
	classDeclaration := classScript.StatementList[0].(*StatementListItemDeclaration).Declaration.(*LexicalDeclaration)
	classExpression := classDeclaration.BindingList.Items[0].Initializer.(*ClassExpression)
	if classExpression.IdentifierName != "" {
		t.Fatalf("anonymous class identifier = %q, want empty", classExpression.IdentifierName)
	}

	contextualClassScript := NewParser("const Contextual = class async {};", ParserContext{}).Parse()
	contextualClassDeclaration := contextualClassScript.StatementList[0].(*StatementListItemDeclaration).Declaration.(*LexicalDeclaration)
	contextualClassExpression := contextualClassDeclaration.BindingList.Items[0].Initializer.(*ClassExpression)
	if contextualClassExpression.IdentifierName != "async" {
		t.Fatalf("contextual class identifier = %q, want async", contextualClassExpression.IdentifierName)
	}
}
