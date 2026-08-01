package coldmoon

import (
	"errors"
	"testing"
)

func TestCheckSourceReturnsSyntaxDiagnosticWithCodePointPosition(t *testing.T) {
	err := CheckSource(Source{
		Text: "const x = \"π\"; const y = }\n",
		Name: "unicode.js",
		Kind: SourceScript,
	}, nil)
	if err == nil {
		t.Fatal("CheckSource succeeded for invalid syntax")
	}

	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error type = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.ErrorName != "SyntaxError" {
		t.Fatalf("diagnostic kind = (%q, %q), want syntax SyntaxError", diagnostic.Category, diagnostic.ErrorName)
	}
	if diagnostic.Message != `unexpected token "}"` {
		t.Fatalf("diagnostic message = %q, want unexpected token", diagnostic.Message)
	}
	if diagnostic.SourceName != "unicode.js" || diagnostic.SourceLine != "const x = \"π\"; const y = }" {
		t.Fatalf("diagnostic source = (%q, %q)", diagnostic.SourceName, diagnostic.SourceLine)
	}
	if diagnostic.Span == nil {
		t.Fatal("diagnostic has no source span")
	}
	if got := diagnostic.Span.Start; got.Line != 1 || got.Column != 26 || got.Offset != 25 {
		t.Fatalf("diagnostic start = %#v, want line 1 column 26 offset 25 (message %q)", got, diagnostic.Message)
	}
}

func TestCheckSourceRejectsTrailingTokens(t *testing.T) {
	for _, kind := range []SourceKind{SourceScript, SourceModule} {
		t.Run(map[SourceKind]string{SourceScript: "script", SourceModule: "module"}[kind], func(t *testing.T) {
			err := CheckSource(Source{Text: "const x = 1; }", Name: "trailing.js", Kind: kind}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if diagnostic.Category != DiagnosticSyntax || diagnostic.Incomplete {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
			if diagnostic.Span == nil || diagnostic.Span.Start.Column != 14 {
				t.Fatalf("span = %#v, want trailing brace at column 14", diagnostic.Span)
			}
		})
	}
}

func TestCheckSourceRejectsInvalidChainedPostfixUpdate(t *testing.T) {
	err := CheckSource(Source{Text: "a++++", Name: "update.js", Kind: SourceScript}, nil)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.Incomplete {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestCheckSourceRejectsInvalidNewMetaProperty(t *testing.T) {
	err := CheckSource(Source{Text: "new.foo", Name: "new.js", Kind: SourceScript}, nil)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.Incomplete {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestCheckSourceRejectsInvalidAssignmentTarget(t *testing.T) {
	err := CheckSource(Source{Text: "1 = 2", Name: "assignment.js", Kind: SourceScript}, nil)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.Incomplete {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestCheckSourceRejectsPrefixOnlyOperatorAfterExpression(t *testing.T) {
	err := CheckSource(Source{Text: "3!", Name: "operator.js", Kind: SourceScript}, nil)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.Incomplete {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestCheckSourceRejectsScriptDeclarationCollisions(t *testing.T) {
	tests := map[string]string{
		"duplicate lexical":       "let value; const value = 1;",
		"lexical versus var":      "let value; var value;",
		"lexical versus function": "class value {}; function value() {}",
	}
	for name, sourceText := range tests {
		t.Run(name, func(t *testing.T) {
			err := CheckSource(Source{Text: sourceText, Name: "declarations.js", Kind: SourceScript}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if diagnostic.Category != DiagnosticSyntax || diagnostic.ErrorName != "SyntaxError" || diagnostic.Incomplete {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestCheckSourceRejectsModuleDeclarationCollisions(t *testing.T) {
	tests := map[string]string{
		"duplicate lexical":  `let value; const value = 1;`,
		"lexical versus var": `let value; var value;`,
		"import versus lexical": `import { value } from "./dep.js";
let value;`,
		"duplicate imported local": `import { value as local } from "./a.js";
import { other as local } from "./b.js";`,
	}
	for name, sourceText := range tests {
		t.Run(name, func(t *testing.T) {
			err := CheckSource(Source{Text: sourceText, Name: "declarations.mjs", Kind: SourceModule}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if diagnostic.Category != DiagnosticSyntax || diagnostic.ErrorName != "SyntaxError" || diagnostic.Incomplete {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestCheckSourceRejectsMissingConstInitializer(t *testing.T) {
	for _, sourceText := range []string{"const value;", "const first = 1, second;"} {
		t.Run(sourceText, func(t *testing.T) {
			err := CheckSource(Source{Text: sourceText, Name: "const.js", Kind: SourceScript}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if diagnostic.Category != DiagnosticSyntax || diagnostic.ErrorName != "SyntaxError" || diagnostic.Incomplete {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestCheckSourceRejectsMissingVariableDeclaration(t *testing.T) {
	for _, sourceText := range []string{"var;", "var value,;"} {
		t.Run(sourceText, func(t *testing.T) {
			err := CheckSource(Source{Text: sourceText, Name: "var.js", Kind: SourceScript}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if diagnostic.Category != DiagnosticSyntax || diagnostic.ErrorName != "SyntaxError" || diagnostic.Incomplete {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestCheckSourceRejectsInvalidObjectPropertyWithoutRecoveringInvariantPanic(t *testing.T) {
	err := CheckSource(Source{Text: `const value = {"name"};`, Name: "object.js", Kind: SourceScript}, nil)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.Incomplete {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestCheckSourceDoesNotRecoverProgrammingPanics(t *testing.T) {
	deferredPanic := false
	func() {
		defer func() {
			deferredPanic = recover() != nil
		}()
		_ = CheckSource(Source{Kind: SourceKind(255)}, nil)
	}()
	if !deferredPanic {
		t.Fatal("CheckSource recovered a non-parse programming panic")
	}
}

func TestParserSpeculativeRecoveryDoesNotRecoverInvariantPanic(t *testing.T) {
	parser := NewParser("value", ParserContext{FileName: "internal.js"})
	deferredPanic := false
	func() {
		defer func() {
			deferredPanic = recover() != nil
		}()
		_, _ = parserRecoverOk(parser, func() *ExpressionLogicalExpression {
			return parser.logicalExpression(nil, parser.acceptContextLowest())
		})
	}()
	if !deferredPanic {
		t.Fatal("parserRecoverOk recovered an internal parser invariant panic")
	}
}

func TestEvaluateSourceReturnsScriptValue(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	value, err := EvaluateSource(Source{
		Text: "1 + 2;",
		Name: "eval.js",
		Kind: SourceScript,
	}, realm)
	if err != nil {
		t.Fatalf("EvaluateSource returned error: %v", err)
	}
	number, ok := value.(*NumberValue)
	if !ok || number.Data != 3 {
		t.Fatalf("value = %#v, want number 3", value)
	}
}

func TestEvaluateSourceReturnsLexicalRedeclarationDiagnosticAcrossSubmissions(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	if _, err := EvaluateSource(Source{Text: "let x = 1;", Name: "repl", Kind: SourceScript}, realm); err != nil {
		t.Fatalf("first submission failed: %v", err)
	}
	_, err := EvaluateSource(Source{Text: "let x = 2;", Name: "repl", Kind: SourceScript}, realm)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.ErrorName != "SyntaxError" || diagnostic.Thrown == nil {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestEvaluateSourceReturnsLexicalVarCollisionDiagnosticAcrossSubmissions(t *testing.T) {
	tests := map[string]struct {
		first  string
		second string
	}{
		"lexical then var":      {first: "let value = 1;", second: "var value = 2;"},
		"var then lexical":      {first: "var value = 1;", second: "let value = 2;"},
		"lexical then function": {first: "const value = 1;", second: "function value() {}"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			InitializeConstants()
			agent := NewAgent()
			InitializeHostDefinedRealm(agent, nil)
			realm := agent.CurrentRealm()

			if _, err := EvaluateSource(Source{Text: test.first, Name: "repl", Kind: SourceScript}, realm); err != nil {
				t.Fatalf("first submission failed: %v", err)
			}
			_, err := EvaluateSource(Source{Text: test.second, Name: "repl", Kind: SourceScript}, realm)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if diagnostic.Category != DiagnosticRuntime || diagnostic.ErrorName != "SyntaxError" || diagnostic.Thrown == nil {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestEvaluateSourceAllowsLexicalDeclarationOverOrdinaryGlobalProperty(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	value, err := EvaluateSource(Source{Text: "let Array = 42; Array;", Name: "global.js", Kind: SourceScript}, realm)
	if err != nil {
		t.Fatalf("EvaluateSource returned error: %v", err)
	}
	number, ok := value.(*NumberValue)
	if !ok || number.Data != 42 {
		t.Fatalf("value = %#v, want number 42", value)
	}
}

func TestEvaluateSourceReturnsConstAssignmentDiagnosticAcrossSubmissions(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	if _, err := EvaluateSource(Source{Text: "const x = 1;", Name: "repl", Kind: SourceScript}, realm); err != nil {
		t.Fatalf("first submission failed: %v", err)
	}
	_, err := EvaluateSource(Source{Text: "x = 2;", Name: "repl", Kind: SourceScript}, realm)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.ErrorName != "TypeError" || diagnostic.Thrown == nil {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestEvaluateSourceReturnsThrownPrimitiveDiagnostic(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	value, err := EvaluateSource(Source{
		Text: "throw \"boom\";",
		Name: "throw.js",
		Kind: SourceScript,
	}, realm)
	if value != nil {
		t.Fatalf("value = %#v, want nil", value)
	}
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticRuntime || diagnostic.ErrorName != "Uncaught" || diagnostic.Message != "boom" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	if diagnostic.Thrown == nil || diagnostic.Thrown.String() != "boom" {
		t.Fatalf("thrown = %#v, want boom", diagnostic.Thrown)
	}
}

func TestEvaluateSourceExecutesModule(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	value, err := EvaluateSource(Source{
		Text: "const answer = 42;",
		Name: "/app/main.js",
		Kind: SourceModule,
	}, realm)
	if err != nil {
		t.Fatalf("EvaluateSource returned error: %v", err)
	}
	if value != UndefinedValue {
		t.Fatalf("value = %#v, want undefined", value)
	}
}

func TestEvaluateSourceUsesCanonicalNameForRelativeImportCycle(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	loader := &memoryModuleLoader{
		sources: map[string]string{
			"/app/dep.js": `import { root } from "./main.js";
export const dep = 41;`,
		},
		loadCount: make(map[string]int),
	}
	agent.HostHooks.HostLoadModule = loader

	_, err := EvaluateSource(Source{
		Text: `import { dep } from "./dep.js";
export const root = 1;
globalThis.cycleAnswer = dep + 1;`,
		Name:    "/app/main.js",
		BaseDir: "/app",
		Kind:    SourceModule,
	}, realm)
	if err != nil {
		t.Fatalf("EvaluateSource returned error: %v", err)
	}
	if loader.loadCount["/app/dep.js"] != 1 || loader.loadCount["/app/main.js"] != 0 {
		t.Fatalf("load counts = %#v, want dependency once and canonical entry zero times", loader.loadCount)
	}
	answer := realm.GlobalObject.Get(NewStringPropertyKey("cycleAnswer"))
	number, ok := answer.(*NumberValue)
	if !ok || number.Data != 42 {
		t.Fatalf("cycleAnswer = %#v, want 42", answer)
	}
}

func TestCheckSourceClassifiesIncompleteInput(t *testing.T) {
	for _, sourceText := range []string{
		"function value() {",
		"const value",
		"var value,",
		"const value = `open",
		"`head $",
		"`head ${value + `$",
		"const value = {",
		"class Value {",
		"/* open",
	} {
		t.Run(sourceText, func(t *testing.T) {
			err := CheckSource(Source{Text: sourceText, Name: "repl", Kind: SourceScript}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) || !diagnostic.Incomplete {
				t.Fatalf("error = %#v, want incomplete diagnostic", err)
			}
		})
	}

	err := CheckSource(Source{Text: "const value = };", Name: "repl", Kind: SourceScript}, nil)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) || diagnostic.Incomplete {
		t.Fatalf("error = %#v, want complete syntax diagnostic", err)
	}
}

func TestCheckSourceAcceptsDollarTextBeforeTemplateTail(t *testing.T) {
	for _, sourceText := range []string{"`$`", "`$value`"} {
		t.Run(sourceText, func(t *testing.T) {
			if err := CheckSource(Source{Text: sourceText, Name: "template.js", Kind: SourceScript}, nil); err != nil {
				t.Fatalf("CheckSource returned error: %v", err)
			}
		})
	}
}

func TestCheckSourceRejectsMissingLocalModuleExport(t *testing.T) {
	err := CheckSource(Source{
		Text: `export { missing };`,
		Name: "exports.mjs",
		Kind: SourceModule,
	}, nil)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.ErrorName != "SyntaxError" || diagnostic.Incomplete {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestCheckSourceDoesNotClassifyEarlyErrorsAsIncomplete(t *testing.T) {
	for _, sourceText := range []string{
		"return",
		"break",
		"continue",
		"import.meta",
	} {
		t.Run(sourceText, func(t *testing.T) {
			err := CheckSource(Source{Text: sourceText, Name: "repl", Kind: SourceScript}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if diagnostic.Incomplete {
				t.Fatalf("%q was classified incomplete", sourceText)
			}
		})
	}
}

func TestCheckSourceClassifiesIncompleteArrowHeaders(t *testing.T) {
	for _, sourceText := range []string{
		"a =>",
		"(a)=>",
		"const f=a=>",
		"(a)=>{",
		"async a=>",
		"async(a)=>",
	} {
		t.Run(sourceText, func(t *testing.T) {
			err := CheckSource(Source{Text: sourceText, Name: "repl", Kind: SourceScript}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if !diagnostic.Incomplete {
				t.Fatalf("diagnostic = %#v, want incomplete", diagnostic)
			}
		})
	}
}

func TestCheckSourceClassifiesMissingRequiredStatementsAsIncomplete(t *testing.T) {
	for _, sourceText := range []string{
		"if (true)",
		"if (true) {} else",
		"while (true)",
		"for (value in source)",
		"for (value of source)",
		"label:",
	} {
		t.Run(sourceText, func(t *testing.T) {
			err := CheckSource(Source{Text: sourceText, Name: "repl", Kind: SourceScript}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if !diagnostic.Incomplete {
				t.Fatalf("diagnostic = %#v, want incomplete", diagnostic)
			}
		})
	}
}

func TestCheckSourceKeepsNestedSpeculationRecoverableAfterArrowCommit(t *testing.T) {
	err := CheckSource(Source{
		Text: `const parse = () => { const pairs = {"]": "[", ")": "("}; return pairs; };`,
		Name: "nested-arrow.js",
		Kind: SourceScript,
	}, nil)
	if err != nil {
		t.Fatalf("CheckSource returned error for valid nested object literal: %v", err)
	}
}

func TestCheckSourceDoesNotLoadModuleDependencies(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	loader := &memoryModuleLoader{sources: map[string]string{}, loadCount: make(map[string]int)}
	agent.HostHooks.HostLoadModule = loader

	err := CheckSource(Source{
		Text:    `import {} from "./missing.js";`,
		Name:    "/app/main.js",
		BaseDir: "/app",
		Kind:    SourceModule,
	}, realm)
	if err != nil {
		t.Fatalf("CheckSource returned error: %v", err)
	}
	if len(loader.loadCount) != 0 {
		t.Fatalf("CheckSource loaded dependencies: %#v", loader.loadCount)
	}
}

func TestCheckSourceRejectsDuplicateModuleExports(t *testing.T) {
	tests := map[string]string{
		"default": `export default 1; export default 2;`,
		"named":   `const value = 1; export { value }; export { value };`,
	}
	for name, sourceText := range tests {
		t.Run(name, func(t *testing.T) {
			err := CheckSource(Source{Text: sourceText, Name: "exports.js", Kind: SourceModule}, nil)
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("error = %T, want *Diagnostic", err)
			}
			if diagnostic.Category != DiagnosticSyntax || diagnostic.ErrorName != "SyntaxError" || diagnostic.Incomplete {
				t.Fatalf("diagnostic = %#v", diagnostic)
			}
		})
	}
}

func TestEvaluateSourceReturnsDuplicateDefaultExportDiagnostic(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	_, err := EvaluateSource(Source{
		Text: `export default function () {} export default class {}`,
		Name: "exports.js",
		Kind: SourceModule,
	}, realm)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.ErrorName != "SyntaxError" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestEvaluateSourceReturnsModuleDeclarationCollisionDiagnostic(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	_, err := EvaluateSource(Source{
		Text: `let value; let value;`,
		Name: "declarations.mjs",
		Kind: SourceModule,
	}, realm)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.ErrorName != "SyntaxError" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestEvaluateSourceAttributesDependencySyntaxDiagnostic(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	agent.HostHooks.HostLoadModule = &memoryModuleLoader{
		sources: map[string]string{
			"/app/dep.js": `export const value = };`,
		},
		loadCount: make(map[string]int),
	}

	_, err := EvaluateSource(Source{
		Text:    `import { value } from "./dep.js";`,
		Name:    "/app/main.js",
		BaseDir: "/app",
		Kind:    SourceModule,
	}, realm)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticSyntax || diagnostic.SourceName != "dep.js" || diagnostic.SourceLine != `export const value = };` {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestEvaluateSourceReturnsAsynchronousCallbackDiagnostic(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	agent.Scheduler.ScheduleTimer(0, func() (completion CompletionValue) {
		return completion.ThrowTypeError(agent, "timer failed")
	})

	value, err := EvaluateSource(Source{Text: "1;", Name: "async.js", Kind: SourceScript}, realm)
	if value != nil {
		t.Fatalf("value = %#v, want nil after asynchronous failure", value)
	}
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticRuntime || diagnostic.ErrorName != "TypeError" || diagnostic.Message != "timer failed" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestEvaluateSourceDoesNotRecoverAsynchronousInvariantPanic(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	agent.Scheduler.ScheduleTimer(0, func() CompletionValue {
		panic("scheduler invariant")
	})

	deferredPanic := false
	func() {
		defer func() {
			deferredPanic = recover() != nil
		}()
		_, _ = EvaluateSource(Source{Text: "1;", Name: "async.js", Kind: SourceScript}, realm)
	}()
	if !deferredPanic {
		t.Fatal("EvaluateSource recovered a scheduler invariant panic")
	}
}

func TestEvaluateSourceReturnsErrorObjectDiagnostic(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	_, err := EvaluateSource(Source{
		Text: `throw new TypeError("broken");`,
		Name: "error.js",
		Kind: SourceScript,
	}, realm)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.ErrorName != "TypeError" || diagnostic.Message != "broken" || diagnostic.Thrown == nil {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestEvaluateSourceReturnsInvalidRegExpDiagnostic(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	_, err := EvaluateSource(Source{
		Text: `new RegExp("[");`,
		Name: "regexp.js",
		Kind: SourceScript,
	}, realm)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticRuntime || diagnostic.ErrorName != "SyntaxError" || diagnostic.Thrown == nil {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestEvaluateSourceReturnsModuleLoadingDiagnostic(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	agent.HostHooks.HostLoadModule = &memoryModuleLoader{
		sources:   map[string]string{},
		loadCount: make(map[string]int),
	}

	_, err := EvaluateSource(Source{
		Text:    `import {} from "./missing.js";`,
		Name:    "/app/main.js",
		BaseDir: "/app",
		Kind:    SourceModule,
	}, realm)
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error = %T, want *Diagnostic", err)
	}
	if diagnostic.Category != DiagnosticLink || diagnostic.ErrorName != "TypeError" || diagnostic.Thrown == nil {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	if diagnostic.Cause == nil || diagnostic.Unwrap() != diagnostic.Cause {
		t.Fatalf("cause = %#v, want original module loader error", diagnostic.Cause)
	}
}
