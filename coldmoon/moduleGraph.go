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
		return g.loadFailure(fmt.Errorf("module loading is not configured by the host"))
	}

	resolution, err := loader.Resolve(referrer, specifier, hostDefined)
	if err != nil {
		return g.loadFailure(err)
	}
	if resolution.Identity == "" {
		return g.loadFailure(fmt.Errorf("module %q has no canonical identity", specifier))
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
		return g.loadFailure(err)
	}
	module := ParseModule(sourceText, realm, resolution.HostDefined)
	module.Identity = resolution.Identity
	g.modules[key] = module
	co.value = module
	return
}

func (g *ModuleGraph) loadFailure(cause error) (co CompletionModule) {
	co = co.ThrowError(g.agent, TypeError, cause.Error())
	if object, ok := co.Error().GetObject(); ok {
		if exception, ok := object.(*ErrorObject); ok {
			exception.hostCause = cause
		}
	}
	return co
}

// Size returns the number of canonical Realm-local records in the graph.
func (g *ModuleGraph) Size() int {
	return len(g.modules)
}

// cachedSource returns the Realm-local entry for an explicit canonical identity.
// Empty identities are temporary inputs and therefore never have cache entries.
func (g *ModuleGraph) cachedSource(realm *Realm, identity string) *SourceTextModule {
	if identity == "" {
		return nil
	}
	return g.modules[moduleCacheKey{realm: realm, identity: identity}]
}

// cacheSource publishes a host-provided entry module with an explicit canonical
// identity before its dependencies load. An empty identity deliberately skips
// caching because display names and parse context are not module identities.
func (g *ModuleGraph) cacheSource(realm *Realm, identity string, module *SourceTextModule) *SourceTextModule {
	module.Identity = identity
	if identity == "" {
		return module
	}
	key := moduleCacheKey{realm: realm, identity: identity}
	if existing := g.cachedSource(realm, identity); existing != nil {
		return existing
	}
	g.modules[key] = module
	return module
}
