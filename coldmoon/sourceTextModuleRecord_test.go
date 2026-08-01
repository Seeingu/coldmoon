package coldmoon

import (
	"slices"
	"testing"
)

func newSourceTextModuleTestRealm(t *testing.T) (*Agent, *Realm) {
	t.Helper()
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	return agent, agent.CurrentRealm()
}

func emptySourceTextModule(realm *Realm) *SourceTextModule {
	return &SourceTextModule{
		Realm:          realm,
		ECMAScriptCode: &Module{},
		LoadedModules:  make(map[string]ModuleRecord),
		Status:         ModuleStatusUnlinked,
	}
}

func requireNumberBinding(t *testing.T, agent *Agent, module *SourceTextModule, name string, want JSNumber) {
	t.Helper()
	result := module.Environment.GetBindingValue(agent, name, true)
	if result.IsAbrupt() {
		t.Fatalf("binding %q lookup failed: %v", name, result.Error())
	}
	number, ok := result.Data().(*NumberValue)
	if !ok || number.Data != want {
		t.Fatalf("binding %q = %v, want %v", name, result.Data(), want)
	}
}

func requirePanic(t *testing.T, run func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("operation did not panic")
		}
	}()
	run()
}

func TestParseModuleDetectsTopLevelAwait(t *testing.T) {
	_, realm := newSourceTextModuleTestRealm(t)
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{name: "await expression", source: `await Promise.resolve(1);`, want: true},
		{name: "for await", source: `for await (const value of []) { value; }`, want: true},
		{name: "nested function", source: `async function nested() { await Promise.resolve(1); }`},
		{name: "class computed name", source: `class Example { [await 1]() {} }`, want: true},
		{name: "class field initializer", source: `class Example { field = await 1; }`},
		{name: "class static block", source: `class Example { static { await 1; } }`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module := ParseModule(test.source, realm, HostDefined{})
			if module.HasTLA != test.want {
				t.Fatalf("HasTLA = %v, want %v", module.HasTLA, test.want)
			}
		})
	}
}

func TestEvaluateSynchronousModuleWithoutExecutionContext(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	module := ParseModule(`var answer = 42;`, realm, HostDefined{})
	module.Status = ModuleStatusUnlinked
	if err := module.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}
	agent.ExecutionContextStack.Clear()

	promise := module.Evaluate()
	if promise.PromiseState != PromiseStateFulfilled {
		t.Fatalf("evaluation promise state = %v, want fulfilled", promise.PromiseState)
	}
	if !agent.ExecutionContextStack.IsEmpty() {
		t.Fatal("host evaluation context leaked after synchronous evaluation")
	}
	requireNumberBinding(t, agent, module, "answer", 42)
	if again := module.Evaluate(); again != promise {
		t.Fatal("repeated evaluation did not return the cycle root promise")
	}
}

func TestLoadRequestedModulesResolvesForAlreadyLoadedModule(t *testing.T) {
	_, realm := newSourceTextModuleTestRealm(t)
	module := ParseModule(`var ready = true;`, realm, HostDefined{})
	module.Status = ModuleStatusUnlinked

	promise := module.LoadRequestedModules()
	if promise.PromiseState != PromiseStateFulfilled {
		t.Fatalf("load promise state = %v, want fulfilled", promise.PromiseState)
	}
}

func TestEvaluateRejectsInsteadOfPanickingOnModuleThrow(t *testing.T) {
	_, realm := newSourceTextModuleTestRealm(t)
	module := ParseModule(`throw "boom";`, realm, HostDefined{})
	module.Status = ModuleStatusUnlinked
	if err := module.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	promise := module.Evaluate()
	if promise.PromiseState != PromiseStateRejected {
		t.Fatalf("evaluation promise state = %v, want rejected", promise.PromiseState)
	}
	if got := promise.PromiseResult.String(); got != "boom" {
		t.Fatalf("rejection reason = %q, want boom", got)
	}
	if module.Status != ModuleStatusEvaluated || module.EvaluationError != promise.PromiseResult {
		t.Fatal("module did not retain its abrupt evaluation result")
	}
	if again := module.Evaluate(); again != promise {
		t.Fatal("failed module evaluation was executed more than once")
	}
}

func TestFailedDependencyCanCreateItsOwnRejectedEvaluationPromise(t *testing.T) {
	_, realm := newSourceTextModuleTestRealm(t)
	dependency := ParseModule(`throw "dependency failed";`, realm, HostDefined{})
	root := ParseModule(`var skipped = 1;`, realm, HostDefined{})
	dependency.Status = ModuleStatusUnlinked
	root.Status = ModuleStatusUnlinked
	root.RequestedModules = []string{"dependency"}
	root.LoadedModules["dependency"] = dependency
	if err := root.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	rootPromise := root.Evaluate()
	if rootPromise.PromiseState != PromiseStateRejected {
		t.Fatalf("root promise state = %v, want rejected", rootPromise.PromiseState)
	}
	if dependency.CycleRoot != nil || dependency.TopLevelCapability != nil {
		t.Fatal("abrupt dependency was incorrectly folded into the entry module's cycle")
	}
	dependencyPromise := dependency.Evaluate()
	if dependencyPromise == rootPromise {
		t.Fatal("abrupt dependency reused an unrelated entry module promise")
	}
	if dependencyPromise.PromiseState != PromiseStateRejected ||
		dependencyPromise.PromiseResult.String() != "dependency failed" {
		t.Fatalf("dependency promise = (%v, %v)", dependencyPromise.PromiseState, dependencyPromise.PromiseResult)
	}
}

func TestLinkRollsBackEntireActiveCycleOnInstantiationError(t *testing.T) {
	_, realm := newSourceTextModuleTestRealm(t)
	left := emptySourceTextModule(realm)
	right := emptySourceTextModule(realm)
	left.RequestedModules = []string{"right"}
	left.LoadedModules["right"] = right
	right.RequestedModules = []string{"left"}
	right.LoadedModules["left"] = left
	left.ImportEntries = []ImportEntryRecord{{
		ModuleRequest: "right",
		ImportName:    "missing",
		LocalName:     "missing",
	}}

	if err := left.Link(); err == nil {
		t.Fatal("link unexpectedly succeeded with an unresolved import")
	}
	if left.Status != ModuleStatusUnlinked || right.Status != ModuleStatusUnlinked {
		t.Fatalf("cycle status after rollback = (%v, %v), want both unlinked", left.Status, right.Status)
	}
}

func TestModuleDeclarationInstantiationCreatesStaticBindings(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	module := ParseModule(`
let mutable = 1;
const fixed = 2;
class Example {}
var { destructured } = { destructured: 3 };
function callable() { return 4; }
`, realm, HostDefined{})
	module.Status = ModuleStatusUnlinked
	if err := module.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}
	env := module.Environment.(*ModuleEnvironment)
	for _, name := range []string{"mutable", "fixed", "Example", "destructured", "callable"} {
		if !env.HasBinding(name) {
			t.Fatalf("module binding %q was not instantiated", name)
		}
	}
	if env.Bindings["mutable"].Value != nil || !env.Bindings["mutable"].Mutable {
		t.Fatal("let binding was not created as an uninitialized mutable binding")
	}
	if env.Bindings["fixed"].Value != nil || env.Bindings["fixed"].Mutable {
		t.Fatal("const binding was not created as an uninitialized immutable binding")
	}
	if env.Bindings["Example"].Value != nil || env.Bindings["Example"].Mutable {
		t.Fatal("class binding was not created as an uninitialized immutable binding")
	}
	if env.Bindings["destructured"].Value != UndefinedValue {
		t.Fatal("destructuring var binding was not initialized to undefined")
	}
	if _, ok := env.Bindings["callable"].Value.GetObject(); !ok {
		t.Fatal("function declaration was not instantiated during linking")
	}

	promise := module.Evaluate()
	if promise.PromiseState != PromiseStateFulfilled {
		t.Fatalf("evaluation promise state = %v, want fulfilled", promise.PromiseState)
	}
	requireNumberBinding(t, agent, module, "mutable", 1)
	requireNumberBinding(t, agent, module, "fixed", 2)
	requireNumberBinding(t, agent, module, "destructured", 3)
}

func TestAnonymousDefaultExportsUseSyntheticModuleBinding(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		callable bool
	}{
		{name: "expression", source: `export default 41;`},
		{name: "function", source: `export default function () {}`, callable: true},
		{name: "async function", source: `export default async function () {}`, callable: true},
		{name: "generator", source: `export default function* () {}`, callable: true},
		{name: "async generator", source: `export default async function* () {}`, callable: true},
		{name: "class", source: `export default class {}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent, realm := newSourceTextModuleTestRealm(t)
			module := ParseModule(test.source, realm, HostDefined{})
			module.Status = ModuleStatusUnlinked
			if err := module.Link(); err != nil {
				t.Fatalf("link failed: %v", err)
			}

			env := module.Environment.(*ModuleEnvironment)
			if !env.HasBinding("*default*") {
				t.Fatal("synthetic *default* binding was not instantiated")
			}
			resolution := module.ResolveExport("default", nil)
			if resolution.IsNull() || resolution.IsAmbiguous() || resolution.Module != module ||
				resolution.BindingName.String != "*default*" {
				t.Fatalf("default resolution = %#v", resolution)
			}
			promise := module.Evaluate()
			if promise.PromiseState != PromiseStateFulfilled {
				t.Fatalf("evaluation promise state = %v, want fulfilled", promise.PromiseState)
			}
			value := ReturnAssertNormal(env.GetBindingValue(agent, "*default*", true))
			if test.name == "expression" {
				number, ok := value.(*NumberValue)
				if !ok || number.Data != 41 {
					t.Fatalf("default expression value = %v, want 41", value)
				}
			} else if _, ok := value.GetObject(); !ok {
				t.Fatalf("default declaration value = %v, want object", value)
			}
			if test.callable {
				object := MustGetObject(value)
				if got := object.Get(NewStringPropertyKey("name")).String(); got != "default" {
					t.Fatalf("anonymous default function name = %q, want default", got)
				}
			}
		})
	}
}

func TestAnonymousDefaultHoistableDeclarationsInitializeDuringLink(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{name: "function", source: `function temporary() {}`},
		{name: "async function", source: `async function temporary() {}`},
		{name: "generator", source: `function* temporary() {}`},
		{name: "async generator", source: `async function* temporary() {}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent, realm := newSourceTextModuleTestRealm(t)
			parsed := ParseModule(test.source, realm, HostDefined{})
			statementItem := parsed.ECMAScriptCode.ModuleItemList[0].(*ModuleItemStatementListItem)
			declarationItem := statementItem.StatementListItem.(*StatementListItemDeclaration)
			hoistable := declarationItem.Declaration.(DeclarationHoistable)
			switch declaration := hoistable.(type) {
			case *DeclarationHoistableFunction:
				declaration.FunctionDeclaration.Identifier = ""
			case *DeclarationHoistableAsyncFunction:
				declaration.AsyncFunctionDeclaration.Identifier = ""
			case *DeclarationHoistableGenerator:
				declaration.GeneratorDeclaration.Identifier = ""
			case *DeclarationHoistableAsyncGenerator:
				declaration.AsyncGeneratorDeclaration.Identifier = ""
			default:
				t.Fatalf("declaration = %T, want hoistable declaration", declaration)
			}

			module := emptySourceTextModule(realm)
			module.ECMAScriptCode = &Module{ModuleItemList: ModuleItemList{
				&ModuleItemExportDeclaration{DefaultHoistableDeclaration: hoistable},
			}}
			module.LocalExportEntries = []ExportEntry{{ExportName: "default", LocalName: "*default*"}}
			if err := module.Link(); err != nil {
				t.Fatalf("link failed: %v", err)
			}

			env := module.Environment.(*ModuleEnvironment)
			value := ReturnAssertNormal(env.GetBindingValue(agent, "*default*", true))
			object, ok := value.GetObject()
			if !ok {
				t.Fatalf("linked default value = %v, want function object", value)
			}
			if got := object.Get(NewStringPropertyKey("name")).String(); got != "default" {
				t.Fatalf("anonymous default function name = %q, want default", got)
			}
			if promise := module.Evaluate(); promise.PromiseState != PromiseStateFulfilled {
				t.Fatalf("evaluation promise state = %v, want fulfilled", promise.PromiseState)
			}
		})
	}
}

func TestNamedDefaultDeclarationUsesDeclaredBinding(t *testing.T) {
	_, realm := newSourceTextModuleTestRealm(t)
	module := ParseModule(`export default function named() {}`, realm, HostDefined{})
	module.Status = ModuleStatusUnlinked
	if err := module.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}
	env := module.Environment.(*ModuleEnvironment)
	if !env.HasBinding("named") || env.HasBinding("*default*") {
		t.Fatalf("module bindings = %v, want named without *default*", env.Bindings)
	}
	resolution := module.ResolveExport("default", nil)
	if resolution.IsNull() || resolution.BindingName.String != "named" {
		t.Fatalf("default resolution = %#v, want named", resolution)
	}
}

func TestModuleEnvironmentRejectsDynamicOrNonStrictBindingAccess(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	env := NewModuleEnvironment(realm.GlobalEnv)
	env.CreateMutableBinding("known", false)
	env.InitializeBinding("known", UndefinedValue)
	requirePanic(t, func() {
		env.InitializeBinding("missing", UndefinedValue)
	})
	requirePanic(t, func() {
		env.GetBindingValue(agent, "known", false)
	})
}

func TestGetExportedNamesTerminatesCircularStarGraph(t *testing.T) {
	_, realm := newSourceTextModuleTestRealm(t)
	left := emptySourceTextModule(realm)
	right := emptySourceTextModule(realm)
	left.LocalExportEntries = []ExportEntry{
		{ExportName: "left", LocalName: "left"},
		{ExportName: "default", LocalName: "default"},
	}
	right.LocalExportEntries = []ExportEntry{{ExportName: "right", LocalName: "right"}}
	left.StarExportEntries = []ExportEntry{{
		ModuleRequest: "right",
		ImportName:    ImportNameAllButDefault,
	}}
	right.StarExportEntries = []ExportEntry{{
		ModuleRequest: "left",
		ImportName:    ImportNameAllButDefault,
	}}
	left.LoadedModules["right"] = right
	right.LoadedModules["left"] = left

	if got, want := left.GetExportedNames(), (Exports{"left", "default", "right"}); !slices.Equal(got, want) {
		t.Fatalf("exported names = %v, want %v", got, want)
	}
}

func TestResolveExportCoversIndirectNamespaceAndStarAmbiguity(t *testing.T) {
	_, realm := newSourceTextModuleTestRealm(t)
	emptyName := emptySourceTextModule(realm)
	emptyName.LocalExportEntries = []ExportEntry{{ExportName: "", LocalName: "emptyBinding"}}
	if names := emptyName.GetExportedNames(); len(names) != 1 || names[0] != "" {
		t.Fatalf("empty-string exported names = %v", names)
	}
	if got := emptyName.ResolveExport("", nil); got.IsNull() || got.BindingName.String != "emptyBinding" {
		t.Fatalf("empty-string resolution = %#v", got)
	}

	target := emptySourceTextModule(realm)
	target.LocalExportEntries = []ExportEntry{{ExportName: "value", LocalName: "localValue"}}

	indirect := emptySourceTextModule(realm)
	indirect.LoadedModules["target"] = target
	indirect.IndirectExportEntries = []ExportEntry{
		{ExportName: "renamed", ModuleRequest: "target", ImportName: "value"},
		{ExportName: "namespace", ModuleRequest: "target", ImportName: ImportNameAll},
	}
	resolution := indirect.ResolveExport("renamed", nil)
	if resolution.Module != target || resolution.BindingName.String != "localValue" {
		t.Fatalf("indirect resolution = %#v", resolution)
	}
	namespace := indirect.ResolveExport("namespace", nil)
	if namespace.Module != target || !namespace.BindingName.IsNamespace() {
		t.Fatalf("namespace resolution = %#v", namespace)
	}

	other := emptySourceTextModule(realm)
	other.LocalExportEntries = []ExportEntry{{ExportName: "value", LocalName: "otherValue"}}
	ambiguous := emptySourceTextModule(realm)
	ambiguous.StarExportEntries = []ExportEntry{
		{ModuleRequest: "target", ImportName: ImportNameAllButDefault},
		{ModuleRequest: "other", ImportName: ImportNameAllButDefault},
	}
	ambiguous.LoadedModules["target"] = target
	ambiguous.LoadedModules["other"] = other
	if got := ambiguous.ResolveExport("value", nil); !got.IsAmbiguous() {
		t.Fatalf("resolution = %#v, want ambiguous", got)
	}
	if got := ambiguous.ResolveExport("default", nil); !got.IsNull() {
		t.Fatalf("default star resolution = %#v, want null", got)
	}

	leftPath := emptySourceTextModule(realm)
	rightPath := emptySourceTextModule(realm)
	for _, path := range []*SourceTextModule{leftPath, rightPath} {
		path.StarExportEntries = []ExportEntry{{
			ModuleRequest: "target",
			ImportName:    ImportNameAllButDefault,
		}}
		path.LoadedModules["target"] = target
	}
	shared := emptySourceTextModule(realm)
	shared.StarExportEntries = []ExportEntry{
		{ModuleRequest: "left", ImportName: ImportNameAllButDefault},
		{ModuleRequest: "right", ImportName: ImportNameAllButDefault},
	}
	shared.LoadedModules["left"] = leftPath
	shared.LoadedModules["right"] = rightPath
	if got := shared.ResolveExport("value", nil); got.IsAmbiguous() || got.Module != target {
		t.Fatalf("shared binding resolution = %#v, want target binding", got)
	}
}

func TestTopLevelAwaitDelaysAndThenExecutesAsyncAncestors(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	dependency := ParseModule(`await 1;`, realm, HostDefined{})
	root := ParseModule(`var completed = 1;`, realm, HostDefined{})
	dependency.Status = ModuleStatusUnlinked
	root.Status = ModuleStatusUnlinked
	root.RequestedModules = []string{"dependency"}
	root.LoadedModules["dependency"] = dependency
	if err := root.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	promise := root.Evaluate()
	if promise.PromiseState != PromiseStatePending {
		t.Fatalf("evaluation promise state = %v, want pending", promise.PromiseState)
	}
	before := root.Environment.GetBindingValue(agent, "completed", true)
	if before.IsAbrupt() || before.Data() != UndefinedValue {
		t.Fatalf("ancestor ran before async dependency settled: %v", before.Data())
	}

	agent.Scheduler.RunUntilIdle()
	if promise.PromiseState != PromiseStateFulfilled {
		t.Fatalf("evaluation promise state = %v, want fulfilled", promise.PromiseState)
	}
	if root.Status != ModuleStatusEvaluated || dependency.Status != ModuleStatusEvaluated {
		t.Fatalf("module statuses = (%v, %v), want evaluated", root.Status, dependency.Status)
	}
	requireNumberBinding(t, agent, root, "completed", 1)
}

func TestTopLevelAwaitPreservesDFSEvaluationOrderForAvailableAncestors(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	asyncDependency := ParseModule(`await 0; globalThis.tlaDFSOrder = "async";`, realm, HostDefined{})
	direct1 := ParseModule(`globalThis.tlaDFSOrder += ":direct-1";`, realm, HostDefined{})
	direct2 := ParseModule(`globalThis.tlaDFSOrder += ":direct-2";`, realm, HostDefined{})
	indirect := ParseModule(`globalThis.tlaDFSOrder += ":indirect";`, realm, HostDefined{})
	root := ParseModule(``, realm, HostDefined{})

	direct1.RequestedModules = []string{"async"}
	direct1.LoadedModules["async"] = asyncDependency
	direct2.RequestedModules = []string{"async"}
	direct2.LoadedModules["async"] = asyncDependency
	indirect.RequestedModules = []string{"direct-1"}
	indirect.LoadedModules["direct-1"] = direct1
	root.RequestedModules = []string{"direct-1", "direct-2", "indirect"}
	root.LoadedModules["direct-1"] = direct1
	root.LoadedModules["direct-2"] = direct2
	root.LoadedModules["indirect"] = indirect

	for _, module := range []*SourceTextModule{asyncDependency, direct1, direct2, indirect, root} {
		module.Status = ModuleStatusUnlinked
	}
	if err := root.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	promise := root.Evaluate()
	if promise.PromiseState != PromiseStatePending {
		t.Fatalf("evaluation promise state = %v, want pending", promise.PromiseState)
	}
	for want, module := range []*SourceTextModule{asyncDependency, direct1, direct2, indirect, root} {
		if !module.AsyncEvaluationOrder.IsInteger() || module.AsyncEvaluationOrder.Value != uint64(want) {
			t.Fatalf(
				"module %d async evaluation order = (%v, %d), want (integer, %d)",
				want,
				module.AsyncEvaluationOrder.State,
				module.AsyncEvaluationOrder.Value,
				want,
			)
		}
	}
	agent.Scheduler.RunUntilIdle()
	if promise.PromiseState != PromiseStateFulfilled {
		t.Fatalf("evaluation promise state = %v, want fulfilled", promise.PromiseState)
	}
	if got := realm.GlobalObject.Get(NewStringPropertyKey("tlaDFSOrder")).String(); got != "async:direct-1:direct-2:indirect" {
		t.Fatalf("TLA evaluation order = %q, want async:direct-1:direct-2:indirect", got)
	}
	for i, module := range []*SourceTextModule{asyncDependency, direct1, direct2, indirect, root} {
		if module.Status != ModuleStatusEvaluated ||
			module.AsyncEvaluationOrder.State != AsyncEvaluationOrderDone {
			t.Fatalf(
				"module %d completion state = (%v, %v), want (evaluated, done)",
				i,
				module.Status,
				module.AsyncEvaluationOrder.State,
			)
		}
	}
}

func TestTopLevelAwaitFailureRejectsWaitingCycleRoot(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	dependency := ParseModule(`await 1; throw "dependency failed";`, realm, HostDefined{})
	root := ParseModule(`var completed = 1;`, realm, HostDefined{})
	dependency.Status = ModuleStatusUnlinked
	root.Status = ModuleStatusUnlinked
	root.RequestedModules = []string{"dependency"}
	root.LoadedModules["dependency"] = dependency
	if err := root.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	promise := root.Evaluate()
	agent.Scheduler.RunUntilIdle()
	if promise.PromiseState != PromiseStateRejected {
		t.Fatalf("evaluation promise state = %v, want rejected", promise.PromiseState)
	}
	if got := promise.PromiseResult.String(); got != "dependency failed" {
		t.Fatalf("rejection reason = %q", got)
	}
	if root.EvaluationError != promise.PromiseResult || dependency.EvaluationError != promise.PromiseResult {
		t.Fatal("async rejection was not recorded on dependency ancestors")
	}
	completed := root.Environment.GetBindingValue(agent, "completed", true)
	if completed.IsAbrupt() || completed.Data() != UndefinedValue {
		t.Fatalf("rejected ancestor unexpectedly executed: %v", completed.Data())
	}
}

func TestRejectedAwaitRejectsModuleEvaluation(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	module := ParseModule(`await Promise.reject("awaited failure"); var skipped = 1;`, realm, HostDefined{})
	module.Status = ModuleStatusUnlinked
	if err := module.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	promise := module.Evaluate()
	agent.Scheduler.RunUntilIdle()
	if promise.PromiseState != PromiseStateRejected {
		t.Fatalf("evaluation promise state = %v, want rejected", promise.PromiseState)
	}
	if got := promise.PromiseResult.String(); got != "awaited failure" {
		t.Fatalf("rejection reason = %q", got)
	}
	skipped := module.Environment.GetBindingValue(agent, "skipped", true)
	if skipped.IsAbrupt() || skipped.Data() != UndefinedValue {
		t.Fatalf("module continued after rejected await: %v", skipped.Data())
	}
}

func TestTopLevelAwaitRunsWithoutPreexistingExecutionContext(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	module := ParseModule(`await 1; await 2; var completed = 3;`, realm, HostDefined{})
	module.Status = ModuleStatusUnlinked
	if err := module.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}
	agent.ExecutionContextStack.Clear()

	promise := module.Evaluate()
	if promise.PromiseState != PromiseStatePending {
		t.Fatalf("evaluation promise state = %v, want pending", promise.PromiseState)
	}
	if !agent.ExecutionContextStack.IsEmpty() {
		t.Fatal("host evaluation context leaked while top-level await was pending")
	}
	agent.Scheduler.RunUntilIdle()
	if promise.PromiseState != PromiseStateFulfilled {
		t.Fatalf("evaluation promise state = %v, want fulfilled", promise.PromiseState)
	}
	if !agent.ExecutionContextStack.IsEmpty() {
		t.Fatal("scheduler job context leaked after top-level await completed")
	}
	requireNumberBinding(t, agent, module, "completed", 3)
}
