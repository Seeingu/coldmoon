package coldmoon

import "testing"

func TestModuleNamespaceFiltersResolutionsAndOrdersKeys(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	module := emptySourceTextModule(realm)
	module.LocalExportEntries = []ExportEntry{
		{ExportName: "zeta", LocalName: "zetaBinding"},
		{ExportName: "alpha", LocalName: "alphaBinding"},
	}
	env := NewModuleEnvironment(realm.GlobalEnv)
	env.CreateMutableBinding("zetaBinding", false)
	env.InitializeBinding("zetaBinding", NewNumberValue(2))
	env.CreateMutableBinding("alphaBinding", false)
	env.InitializeBinding("alphaBinding", NewNumberValue(1))
	module.Environment = env

	missing := emptySourceTextModule(realm)
	module.IndirectExportEntries = []ExportEntry{{
		ExportName:    "missing",
		ModuleRequest: "missing-module",
		ImportName:    "absent",
	}}
	module.LoadedModules["missing-module"] = missing

	left := emptySourceTextModule(realm)
	right := emptySourceTextModule(realm)
	left.LocalExportEntries = []ExportEntry{{ExportName: "ambiguous", LocalName: "left"}}
	right.LocalExportEntries = []ExportEntry{{ExportName: "ambiguous", LocalName: "right"}}
	module.StarExportEntries = []ExportEntry{
		{ModuleRequest: "left", ImportName: ImportNameAllButDefault},
		{ModuleRequest: "right", ImportName: ImportNameAllButDefault},
	}
	module.LoadedModules["left"] = left
	module.LoadedModules["right"] = right

	namespace := GetModuleNamespace(agent, module).(*ModuleNamespace)
	if len(namespace.Exports) != 2 || namespace.Exports[0] != "alpha" || namespace.Exports[1] != "zeta" {
		t.Fatalf("namespace exports = %v, want [alpha zeta]", namespace.Exports)
	}
	keys := namespace.internalMethods().OwnPropertyKeys(namespace)
	if len(keys) != 3 {
		t.Fatalf("namespace own keys = %v, want two strings and @@toStringTag", keys)
	}
	for i, want := range []string{"alpha", "zeta"} {
		key, ok := keys[i].(StringPropertyKey)
		if !ok || key.Value != want {
			t.Fatalf("namespace key %d = %#v, want %q", i, keys[i], want)
		}
	}
	if _, ok := keys[2].(SymbolPropertyKey); !ok {
		t.Fatalf("namespace final key = %#v, want symbol", keys[2])
	}

	alpha := namespace.Get(NewStringPropertyKey("alpha"))
	if number, ok := alpha.(*NumberValue); !ok || number.Data != 1 {
		t.Fatalf("namespace alpha = %v, want 1", alpha)
	}
	alphaDescriptor := namespace.internalMethods().GetOwnProperty(namespace, NewStringPropertyKey("alpha"))
	if alphaDescriptor == nil || !alphaDescriptor.IsFullyPopulated() ||
		!alphaDescriptor.Writable || !alphaDescriptor.Enumerable || alphaDescriptor.Configurable {
		t.Fatalf("namespace alpha descriptor = %#v", alphaDescriptor)
	}
	if deleted := namespace.internalMethods().Delete(namespace, NewStringPropertyKey("alpha")); deleted.Data() {
		t.Fatal("namespace allowed deletion of an exported binding")
	}
	if deleted := namespace.internalMethods().Delete(namespace, NewStringPropertyKey("unknown")); !deleted.Data() {
		t.Fatal("namespace rejected deletion of a missing property")
	}
	alphaKey := NewStringPropertyKey("alpha")
	for name, descriptor := range map[string]*PropertyDescriptor{
		"empty descriptor": {},
		"same value":       {Value: NewNumberValue(1)},
	} {
		if defined := namespace.internalMethods().DefineOwnProperty(namespace, alphaKey, descriptor); !defined.Data() {
			t.Fatalf("namespace rejected %s", name)
		}
	}
	for name, descriptor := range map[string]*PropertyDescriptor{
		"different value": {Value: NewNumberValue(2)},
		"non-writable":    {WritableSet: true, Writable: false},
		"non-enumerable":  {EnumerableSet: true, Enumerable: false},
		"configurable":    {ConfigurableSet: true, Configurable: true},
	} {
		if defined := namespace.internalMethods().DefineOwnProperty(namespace, alphaKey, descriptor); defined.Data() {
			t.Fatalf("namespace accepted %s", name)
		}
	}
}

func TestModuleNamespaceCreateSortsDirectInput(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	module := emptySourceTextModule(realm)
	namespace := ModuleNamespaceCreate(agent, module, []string{"zeta", "\uE000", "\U00010000", "alpha"}).(*ModuleNamespace)
	want := Exports{"alpha", "zeta", "\U00010000", "\uE000"}
	if len(namespace.Exports) != len(want) {
		t.Fatalf("namespace exports = %v, want %v", namespace.Exports, want)
	}
	for index := range want {
		if namespace.Exports[index] != want[index] {
			t.Fatalf("namespace exports = %v, want %v", namespace.Exports, want)
		}
	}
}
