package coldmoon

import (
	"fmt"
	"path"
	"testing"
)

type memoryModuleLoader struct {
	sources   map[string]string
	loadCount map[string]int
}

func (l *memoryModuleLoader) Resolve(
	_ ImportedModuleReferrer,
	specifier string,
	hostDefined HostDefined,
) (ModuleResolution, error) {
	identity := specifier
	if !path.IsAbs(identity) {
		identity = path.Join(hostDefined.BaseDir, identity)
	}
	identity = path.Clean("/" + identity)
	return ModuleResolution{
		Identity: identity,
		HostDefined: HostDefined{
			FileName: path.Base(identity),
			BaseDir:  path.Dir(identity),
		},
	}, nil
}

func (l *memoryModuleLoader) Load(resolution ModuleResolution) (string, error) {
	source, ok := l.sources[resolution.Identity]
	if !ok {
		return "", fmt.Errorf("module not found: %s", resolution.Identity)
	}
	l.loadCount[resolution.Identity]++
	return source, nil
}

func TestModuleGraphCachesCanonicalIdentityAcrossCycle(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	loader := &memoryModuleLoader{
		sources: map[string]string{
			"/root.js": `import { dep } from "./dep.js";
export function root() { return dep(); }`,
			"/dep.js": `import { root } from "./root.js";
export function dep() { return 1; }`,
		},
		loadCount: make(map[string]int),
	}
	agent.HostHooks.HostLoadModule = loader

	rootResult := agent.ModuleGraph.Load(realm.ToReferrer(), "root.js", HostDefined{})
	if rootResult.IsAbrupt() {
		t.Fatalf("load root: %v", rootResult.Error())
	}
	root := rootResult.Data().(*SourceTextModule)
	promise := root.LoadRequestedModules()
	if promise.PromiseState != PromiseStateFulfilled {
		t.Fatalf("load graph promise state = %v, want fulfilled", promise.PromiseState)
	}

	again := agent.ModuleGraph.Load(realm.ToReferrer(), "./root.js", HostDefined{})
	if again.Data() != root {
		t.Fatal("equivalent root specifiers produced different module records")
	}
	if loader.loadCount["/root.js"] != 1 || loader.loadCount["/dep.js"] != 1 {
		t.Fatalf("module load counts = %#v, want each identity once", loader.loadCount)
	}
	if agent.ModuleGraph.Size() != 2 {
		t.Fatalf("module graph size = %d, want 2", agent.ModuleGraph.Size())
	}

	dependency := root.LoadedModules["./dep.js"].(*SourceTextModule)
	if dependency.LoadedModules["./root.js"] != root {
		t.Fatal("cyclic import did not reuse the canonical root record")
	}
}
