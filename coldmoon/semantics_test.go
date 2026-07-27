package coldmoon

import "testing"

func TestStaticSemanticsSnapshotsScriptDeclarations(t *testing.T) {
	script := NewParser(
		`let lexical = 1; var variable = 2;`,
		ParserContext{FileName: "script.js"},
	).Parse()
	static := (StaticSemantics{}).AnalyzeScript(script)
	script.StatementList = nil

	if len(static.LexicalDeclarations) != 1 {
		t.Fatalf("lexical declaration count = %d, want 1", len(static.LexicalDeclarations))
	}
	if len(static.VarDeclarations) != 2 {
		t.Fatalf("var-scoped declaration count = %d, want 2", len(static.VarDeclarations))
	}
}

func TestStaticSemanticsOwnsModuleEntries(t *testing.T) {
	module := NewParser(
		`import { value } from "./dependency.js";
export function result() { return value; }`,
		ParserContext{FileName: "module.js"},
	).ParseModule()
	static := (StaticSemantics{}).AnalyzeModule(module)

	if len(static.RequestedModules) != 1 || static.RequestedModules[0] != "./dependency.js" {
		t.Fatalf("requested modules = %v", static.RequestedModules)
	}
	if len(static.ImportEntries) != 1 {
		t.Fatalf("import entry count = %d, want 1", len(static.ImportEntries))
	}
	if len(static.ExportEntries) != 1 {
		t.Fatalf("export entry count = %d, want 1", len(static.ExportEntries))
	}
}

func TestRuntimeSemanticsUsesCurrentContextVM(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	var seen *VM

	result := newRuntimeSemantics(agent).Evaluate(&vmCaptureNode{seen: &seen})

	if result.IsAbrupt() {
		t.Fatalf("runtime evaluation was abrupt: %v", result.Error())
	}
	if seen != agent.RunningExecutionContext().VM {
		t.Fatal("runtime semantics did not use the current execution context VM")
	}
}
