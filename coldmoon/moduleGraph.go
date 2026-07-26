package coldmoon

import "fmt"

// ModuleResolution is the host-canonical identity and parse context for a
// module specifier.
type ModuleResolution struct {
	Identity    string
	HostDefined HostDefined
}

// ModuleLoader separates cheap identity resolution from source loading so a
// ModuleGraph cache hit never repeats host I/O.
type ModuleLoader interface {
	Resolve(referrer ImportedModuleReferrer, specifier string, hostDefined HostDefined) (ModuleResolution, error)
	Load(resolution ModuleResolution) (string, error)
}

type moduleCacheKey struct {
	realm    *Realm
	identity string
}

// ModuleGraph owns canonical SourceTextModule identities for one Agent.
type ModuleGraph struct {
	agent   *Agent
	modules map[moduleCacheKey]*SourceTextModule
}

// NewModuleGraph creates an empty graph for agent.
func NewModuleGraph(agent *Agent) *ModuleGraph {
	return &ModuleGraph{
		agent:   agent,
		modules: make(map[moduleCacheKey]*SourceTextModule),
	}
}

// Load asks the host to resolve source, then returns the Realm-local canonical
// module record. A record is cached before dependencies load, which makes
// cyclic graphs converge on the same identity.
func (g *ModuleGraph) Load(referrer ImportedModuleReferrer, specifier string, hostDefined HostDefined) (co CompletionModule) {
	loader := g.agent.HostHooks.HostLoadModule
	if loader == nil {
		return co.ThrowError(g.agent, TypeError, "module loading is not configured by the host")
	}

	resolution, err := loader.Resolve(referrer, specifier, hostDefined)
	if err != nil {
		return co.ThrowError(g.agent, TypeError, err.Error())
	}
	if resolution.Identity == "" {
		return co.ThrowError(g.agent, TypeError, fmt.Sprintf("module %q has no canonical identity", specifier))
	}

	realm := referrer.RealmRecord()
	if realm == nil {
		realm = g.agent.CurrentRealm()
	}
	key := moduleCacheKey{realm: realm, identity: resolution.Identity}
	if module, ok := g.modules[key]; ok {
		co.value = module
		return
	}

	sourceText, err := loader.Load(resolution)
	if err != nil {
		return co.ThrowError(g.agent, TypeError, err.Error())
	}
	module := ParseModule(sourceText, realm, resolution.HostDefined)
	module.Identity = resolution.Identity
	g.modules[key] = module
	co.value = module
	return
}

// Size returns the number of canonical Realm-local records in the graph.
func (g *ModuleGraph) Size() int {
	return len(g.modules)
}
