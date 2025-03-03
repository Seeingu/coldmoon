package coldmoon

import (
	"math"

	"github.com/Seeingu/coldmoon/pkg"
	"github.com/samber/lo"
)

type ExportEntry struct {
	// [[ExportName]]
	ExportName string
	// [[ModuleRequest]]
	ModuleRequest string
	// [[ImportName]]
	ImportName ImportName
	// [[LocalName]]
	LocalName string
}

type ModuleStatus int

const (
	ModuleStatusNew ModuleStatus = iota
	ModuleStatusUnlinked
	ModuleStatusLinking
	ModuleStatusLinked
	ModuleStatusEvaluating
	ModuleStatusEvaluatingAsync
	ModuleStatusEvaluated
)

type SourceTextModule struct {
	ScriptOrModule
	// [[Realm]]
	Realm *Realm
	// [[Environment]]
	Environment EnvironmentRecord
	// [[Namespace]]
	Namespace ObjectType
	// [[ECMAScriptCode]]
	ECMAScriptCode *Module
	// [[Context]]
	Context *ExecutionContext
	// [[ImportMeta]]
	ImportMeta ObjectType
	// [[ImportEntries]]
	ImportEntries []ImportEntryRecord
	// [[RequestedModules]]
	RequestedModules []string
	// [[LoadedModules]]
	LoadedModules map[string]ModuleRecord
	// [[LocalExportEntries]]
	LocalExportEntries []ExportEntry
	// [[IndirectExportEntries]]
	IndirectExportEntries []ExportEntry
	// [[StarExportEntries]]
	StarExportEntries []ExportEntry
	// [[HostDefined]]
	HostDefined HostDefined
	// [[TopLevelCapability]]
	TopLevelCapability *PromiseCapability
	// [[HasTLA]]
	HasTLA bool
	// [[Status]]
	Status ModuleStatus
	// [[EvaluationError]]
	EvaluationError Value
	// [[AsyncEvaluation]]
	AsyncEvaluation bool
	// [[DFSIndex
	DFSIndex int
	// [[DFSAncestorIndex]]
	DFSAncestorIndex int
	// [[CycleRoot]]
	CycleRoot ModuleRecord
	// [[PendingAsyncDependencies]]
	PendingAsyncDependencies int
	// [[AsyncParentModules]]
	AsyncParentModules pkg.Stack[ModuleRecord]
}

func (s *SourceTextModule) ToReferrer() ImportedModuleReferrer {
	return ImportedModuleReferrer{
		Script: nil,
		Module: s,
		Realm:  nil,
	}
}

// 16.2.1.5.2
func (s *SourceTextModule) Link() (err Value) {
	module := s
	// TODO(BM): bitset
	Assert(module.Status == ModuleStatusUnlinked || module.Status == ModuleStatusLinked ||
		module.Status == ModuleStatusEvaluated || module.Status == ModuleStatusEvaluatingAsync)
	var stack pkg.Stack[ModuleRecord]
	_, err = s.InnerModuleLinking(stack, 0)
	if err != nil {
		for _, item := range stack.Data() {
			i := item.(*SourceTextModule)
			Assert(i.Status == ModuleStatusLinking)
			i.Status = ModuleStatusUnlinked
		}
		return
	}
	// TODO(BM): bitset
	Assert(module.Status == ModuleStatusLinked || module.Status == ModuleStatusEvaluatingAsync ||
		module.Status == ModuleStatusEvaluated)
	Assert(stack.Len() == 0)
	return
}

// 16.2.1.5.2.1
func (s *SourceTextModule) InnerModuleLinking(stack pkg.Stack[ModuleRecord], i int) (index int, err Value) {
	index = i
	module := s
	if !module.isCyclic() {
		module.Link()
		return
	}

	// TODO(BM): bitset
	if module.Status == ModuleStatusLinked ||
		module.Status == ModuleStatusEvaluated ||
		module.Status == ModuleStatusEvaluatingAsync ||
		module.Status == ModuleStatusLinking {
		return
	}
	Assert(module.Status == ModuleStatusUnlinked)
	module.Status = ModuleStatusLinking
	module.DFSIndex = index
	module.DFSAncestorIndex = index
	index++
	stack.Push(module)
	for _, required := range module.RequestedModules {
		requiredModule := GetImportedModule(module, required).(*SourceTextModule)
		index, _ = requiredModule.InnerModuleLinking(stack, index)
		if requiredModule.isCyclic() {
			// TODO(BM): bitset
			Assert(requiredModule.Status == ModuleStatusLinking ||
				requiredModule.Status == ModuleStatusLinked ||
				requiredModule.Status == ModuleStatusEvaluatingAsync ||
				requiredModule.Status == ModuleStatusEvaluated)
			module.DFSAncestorIndex = int(math.Min(float64(module.DFSAncestorIndex), float64(requiredModule.DFSIndex)))
		}
	}
	module.InitializeEnvironment()
	// TODO: Assert: module occurs exactly once in stack.
	Assert(module.DFSAncestorIndex <= module.DFSIndex)
	if module.DFSAncestorIndex == module.DFSIndex {
		for {
			requiredModule := stack.Pop().(*SourceTextModule)
			requiredModule.Status = ModuleStatusLinked
			if requiredModule == module {
				break
			}
		}
	}
	return
}

// 16.2.1.5.3
func (s *SourceTextModule) Evaluate() *PromiseObject {
	realm := s.Realm
	agent := realm.Agent
	module := s

	if module.Status == ModuleStatusEvaluatingAsync || module.Status == ModuleStatusEvaluated {
		panic("unimplemented")
	}

	if module.TopLevelCapability != nil {
		return module.TopLevelCapability.Promise
	}

	var stack pkg.Stack[ModuleRecord]
	capability := NewPromiseCapability(agent, realm.Intrinsics.Promise.ToValue())
	module.TopLevelCapability = capability
	_, err := s.InnerModuleEvaluation(stack, 0)
	if err != nil {
		for _, item := range stack.Data() {
			i := item.(*SourceTextModule)
			Assert(i.Status == ModuleStatusEvaluating)
			i.Status = ModuleStatusEvaluated
			i.EvaluationError = err
		}

		Assert(module.Status == ModuleStatusEvaluating)

		capability.Reject.Call(UndefinedValue, []Value{err})
	} else {
		// TODO(BM): bitset
		Assert(module.Status == ModuleStatusEvaluated || module.Status == ModuleStatusEvaluatingAsync)
		if !module.AsyncEvaluation {
			Assert(module.Status == ModuleStatusEvaluated)
			capability.Resolve.Call(UndefinedValue, []Value{UndefinedValue})
		}
		Assert(stack.Len() == 0)
	}

	return capability.Promise
}

// 16.2.1.5.3.1
func (s *SourceTextModule) InnerModuleEvaluation(stack pkg.Stack[ModuleRecord], index int) (r int, err Value) {
	module := s

	if !module.isCyclic() {
		promise := module.Evaluate()
		Assert(promise.PromiseState != PromiseStatePending)
		if promise.PromiseState == PromiseStateRejected {
			return index, promise.PromiseResult
		}
		return index, nil
	}
	m := module
	if m.Status == ModuleStatusEvaluatingAsync || m.Status == ModuleStatusEvaluated {
		if m.EvaluationError == nil {
			return index, nil
		} else {
			return index, m.EvaluationError
		}
	}

	if m.Status == ModuleStatusEvaluating {
		return index, nil
	}

	Assert(m.Status == ModuleStatusLinked)
	m.Status = ModuleStatusEvaluating
	m.DFSIndex = index
	m.DFSAncestorIndex = index
	m.PendingAsyncDependencies = 0
	newIndex := index + 1
	stack.Push(module)
	for _, required := range m.RequestedModules {
		requiredModule := GetImportedModule(m, required).(*SourceTextModule)
		newIndex, _ = requiredModule.InnerModuleEvaluation(stack, newIndex)
		if requiredModule != nil {
			stm := requiredModule
			Assert(stm.Status == ModuleStatusEvaluated || stm.Status == ModuleStatusEvaluatingAsync || stm.Status == ModuleStatusEvaluating)
			if stm.Status == ModuleStatusEvaluating {
				Assert(lo.Contains(stack.Data(), ModuleRecord(requiredModule)))
				m.DFSAncestorIndex = int(math.Min(float64(m.DFSAncestorIndex), float64(stm.DFSIndex)))
			} else {
				requiredModule = stm.CycleRoot.(*SourceTextModule)
				stm = requiredModule
				Assert(stm.Status == ModuleStatusEvaluated || stm.Status == ModuleStatusEvaluatingAsync)
				if stm.EvaluationError != nil {
					return newIndex, stm.EvaluationError
				}
				if stm.AsyncEvaluation {
					m.PendingAsyncDependencies++
					requiredModule.AsyncParentModules.Push(module)
				}
			}
		}
	}

	if m.PendingAsyncDependencies > 0 || m.HasTLA {
		m.AsyncEvaluation = true
		if m.PendingAsyncDependencies == 0 {
			module.ExecuteAsyncModule()
		}
	} else {
		m.ExecuteModule(nil)
	}

	if m.DFSAncestorIndex == m.DFSIndex {
		for {
			requiredModule := stack.Pop().(*SourceTextModule)
			if !requiredModule.AsyncEvaluation {
				requiredModule.Status = ModuleStatusEvaluated
			} else {
				requiredModule.Status = ModuleStatusEvaluatingAsync
			}

			requiredModule.CycleRoot = module
			if requiredModule == module {
				break
			}
		}
	}

	return newIndex, nil
}

// 16.2.1.5.3.2
func (s *SourceTextModule) ExecuteAsyncModule() {
	panic("unimplemented")
}

// [[BindingName]] union
type BindingName struct {
	String    string
	Namespace bool
}

var bindingNameNamespace = &BindingName{Namespace: true}

func (b *BindingName) IsNamespace() bool {
	return b == bindingNameNamespace
}

func (b *BindingName) IsEqualTo(other *BindingName) bool {
	if b.IsNamespace() && b.IsNamespace() != b.IsNamespace() {
		return false
	}
	if b.String != other.String {
		return false
	}
	return true
}

type ResolvedBinding struct {
	// [[Module]]
	Module ModuleRecord
	// [[BindingName]]
	BindingName *BindingName
}

var (
	resolvedBindingAmbiguous = &ResolvedBinding{}
	resolvedBindingNull      = &ResolvedBinding{}
)

func (r *ResolvedBinding) IsNull() bool {
	return r == resolvedBindingNull
}

func (r *ResolvedBinding) IsAmbiguous() bool {
	return r == resolvedBindingAmbiguous
}

func (s *SourceTextModule) ResolveExport(exportName string, resolveSet []ResolveSet) *ResolvedBinding {
	Assert(s.Status != ModuleStatusNew)

	for _, r := range resolveSet {
		if r.ExportName == exportName && r.Module == ModuleRecord(s) {
			// TODO: Assert: This is a circular import request.
			return resolvedBindingNull
		}
	}

	resolveSet = append(resolveSet, ResolveSet{
		Module:     s,
		ExportName: exportName,
	})

	for _, ee := range s.LocalExportEntries {
		if ee.ExportName == exportName {
			return &ResolvedBinding{
				Module:      s,
				BindingName: &BindingName{String: ee.LocalName},
			}
		}
	}

	for _, e := range s.IndirectExportEntries {
		if e.ExportName == exportName {
			Assert(e.ModuleRequest != "")
			importedModule := GetImportedModule(s, e.ModuleRequest).(*SourceTextModule)
			if e.ImportName == ImportNameAll {
				// Assert: module does not provide the direct binding for this export.
				return &ResolvedBinding{
					Module:      importedModule,
					BindingName: &BindingName{Namespace: true},
				}
			} else {
				// TODO: Assert: module imports a specific binding for this export.
				return importedModule.ResolveExport(string(e.ImportName), resolveSet)
			}
		}
	}

	if exportName == "default" {
		// Assert: A defaultexport was not explicitly defined by this module.
		return resolvedBindingNull
	}
	starResolution := resolvedBindingNull
	for _, e := range s.StarExportEntries {
		Assert(e.ModuleRequest != "")
		importedModule := GetImportedModule(s, e.ModuleRequest).(*SourceTextModule)
		resolution := importedModule.ResolveExport(exportName, resolveSet)
		if resolution.IsAmbiguous() {
			return resolvedBindingAmbiguous
		}
		if !resolution.IsNull() {
			if starResolution.IsNull() {
				starResolution = resolution
			} else {
				// Assert: There is more than one *import that includes the requested name.
				if resolution.Module != starResolution.Module {
					return resolvedBindingAmbiguous
				}
				if !resolution.BindingName.IsEqualTo(starResolution.BindingName) {
					return resolvedBindingAmbiguous
				}
			}
		}
	}
	return starResolution
}

func (s *SourceTextModule) GetExportedNames() Exports {
	panic("unimplemented")
}

// 16.2.1.6.1
func ParseModule(sourceText string, realm *Realm, hostDefined HostDefined) *SourceTextModule {
	body := NewParser(sourceText, ParserContext{
		FileName: hostDefined.FileName,
		BaseDir:  hostDefined.BaseDir,
	}).ParseModule()

	requestedModules := body.moduleRequests()
	importEntries := body.importEntries()
	var indirectExportEntries []ExportEntry
	var localExportEntries []ExportEntry
	var starExportEntries []ExportEntry
	exportEntries := body.exportEntries()

	for _, ee := range exportEntries {
		if ee.ModuleRequest == "" {
			var importEntryBoundName *ImportEntryRecord
			for _, ie := range importEntries {
				if ie.ImportName == ee.ImportName {
					importEntryBoundName = &ie
					break
				}
			}
			if importEntryBoundName == nil {
				localExportEntries = append(localExportEntries, ee)
			} else {
				ie := importEntryBoundName
				if ie.ImportName == ImportNameNamespaceObject {
					localExportEntries = append(localExportEntries, ee)
				} else {
					indirectExportEntries = append(indirectExportEntries, ExportEntry{
						ModuleRequest: ie.ModuleRequest,
						ImportName:    ie.ImportName,
						LocalName:     "",
						ExportName:    ee.ExportName,
					})
				}
			}
		} else if ee.ImportName == ImportNameAllButDefault {
			starExportEntries = append(starExportEntries, ee)
		} else {
			indirectExportEntries = append(indirectExportEntries, ee)
		}
	}

	return &SourceTextModule{
		Realm:                 realm,
		HostDefined:           hostDefined,
		RequestedModules:      requestedModules,
		ImportEntries:         importEntries,
		LocalExportEntries:    localExportEntries,
		IndirectExportEntries: indirectExportEntries,
		StarExportEntries:     starExportEntries,
		ECMAScriptCode:        body,
		LoadedModules:         make(map[string]ModuleRecord),
		// TODO
		// HasTLA: body.HasTLA()
	}
}

// 16.2.1.5.1.2
func ContinueModuleLoading(agent *Agent, state *GraphLoadingState, moduleCompletion CompletionModule) {
	if !state.IsLoading {
		return
	}
	if moduleCompletion.Data() != nil {
		module := moduleCompletion.Data().(*SourceTextModule)
		module.InnerModuleLoading(state)
	} else {
		state.IsLoading = false
		state.PromiseCapability.Reject.Call(UndefinedValue, []Value{moduleCompletion.Error()})
	}
}

// 16.2.1.5.1
func (s *SourceTextModule) LoadRequestedModules(hostDefined HostDefined) *PromiseObject {
	realm := s.Realm
	pc := NewPromiseCapability(s.agent(), realm.Intrinsics.Promise.ToValue())

	state := &GraphLoadingState{
		IsLoading:           true,
		PendingModulesCount: 1,
		Visited:             make([]ModuleRecord, 0),
		PromiseCapability:   pc,
		HostDefined:         hostDefined,
	}

	s.InnerModuleLoading(state)
	return pc.Promise
}

// 16.2.1.6.4
func (s *SourceTextModule) InitializeEnvironment() (co CompletionValue) {
	realm := s.Realm
	agent := realm.Agent
	env := NewModuleEnvironment(realm.GlobalEnv)
	s.Environment = env

	for _, importEntry := range s.ImportEntries {
		importedModule := GetImportedModule(s, importEntry.ModuleRequest).(*SourceTextModule)
		importName := importEntry.ImportName
		localName := importEntry.LocalName
		if importName == ImportNameNamespaceObject {
			namespace := GetModuleNamespace(agent, importedModule)
			env.CreateImmutableBinding(localName, true)
			env.InitializeBinding(localName, namespace.ToValue())
		} else {
			resolution := importedModule.ResolveExport(string(importName), []ResolveSet{})
			if resolution.IsNull() {
				co.err = agent.ThrowException(SyntaxError, "Failed to resolve export")
				return
			} else if resolution.IsAmbiguous() {
				co.err = agent.ThrowException(SyntaxError, "Ambiguous export")
				return
			}
			if resolution.BindingName.Namespace {
				namespace := GetModuleNamespace(agent, resolution.Module.(*SourceTextModule))
				env.CreateImmutableBinding(localName, true)
				env.InitializeBinding(localName, namespace.ToValue())
			} else {
				env.CreateImportBinding(localName, resolution.Module.(*SourceTextModule), resolution.BindingName.String)
			}
		}
	}

	moduleContext := &ExecutionContext{
		Realm:          realm,
		Function:       nil,
		ScriptOrModule: s,
		ch:             make(chan struct{}),
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  env,
			VariableEnvironment: env,
			PrivateEnvironment:  nil,
		},
	}
	s.Context = moduleContext
	agent.ExecutionContextStack.Push(moduleContext)
	code := s.ECMAScriptCode

	varDeclarations := code.ModuleItemList.VarScopedDeclarations()

	declaredVarNames := make(map[string]bool)

	for _, varDeclaration := range varDeclarations {
		varName := string(varDeclaration.BindingIdentifier)
		if _, ok := declaredVarNames[varName]; !ok {
			env.CreateMutableBinding(varName, false)
			env.InitializeBinding(varName, UndefinedValue)
			declaredVarNames[varName] = true
		}
	}

	// TODO:

	agent.ExecutionContextStack.Pop()
	co.value = UndefinedValue
	return
}

// ExecuteModule
// spec: 16.2.1.6.5
func (s *SourceTextModule) ExecuteModule(capability *PromiseCapability) {
	agent := s.Realm.Agent
	moduleContext := &ExecutionContext{
		Realm:          s.Realm,
		ScriptOrModule: s,
		ch:             make(chan struct{}),
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  s.Environment,
			VariableEnvironment: s.Environment,
		},
	}
	if !s.HasTLA {
		Assert(capability == nil)
		agent.ExecutionContextStack.Push(moduleContext)
		result := RunNode(agent, s.ECMAScriptCode)
		// TODO: standardize
		if result.IsAbrupt() {
			err := result.err
			if obj, ok := err.GetObject(); ok {
				msgKey := CMString("message").ToPropertyKey()
				if obj.HasProperty(msgKey) {
					msg := obj.Get(msgKey).String()
					panic("execute module failed: " + msg)
				}
			}
			panic("execute module failed: " + result.err.String())
		}
		agent.ExecutionContextStack.Pop()
	} else {
		panic("unimplemented")
	}
}

// 16.2.1.5.1.1
func (s *SourceTextModule) InnerModuleLoading(state *GraphLoadingState) {
	module := s
	Assert(state.IsLoading)
	if module.isCyclic() && module.Status == ModuleStatusNew &&
		!lo.Contains(state.Visited, ModuleRecord(module)) {
		state.Visited = append(state.Visited, module)
		requestedModulesCount := len(module.RequestedModules)
		state.PendingModulesCount += requestedModulesCount
		for _, required := range module.RequestedModules {
			if record, ok := module.LoadedModules[required]; ok {
				record.(*SourceTextModule).InnerModuleLoading(state)
			} else {
				HostLoadImportedModule(s.agent(), module.ToReferrer(), required, state.HostDefined, ImportedModulePayload{
					GraphLoadingState: state,
				})
			}
		}
		if !state.IsLoading {
			return
		}
	}
	Assert(state.PendingModulesCount > 0)
	state.PendingModulesCount--
	if state.PendingModulesCount == 0 {
		state.IsLoading = false
		for _, record := range state.Visited {
			loaded := record.(*SourceTextModule)
			if loaded.Status == ModuleStatusNew {
				loaded.Status = ModuleStatusUnlinked
			}
			state.PromiseCapability.Resolve.Call(UndefinedValue, []Value{UndefinedValue})
		}
	}
}

// MARK: - Internal

func (s *SourceTextModule) _moduleRecord() {}

// TODO: check if the module is cyclic
func (s *SourceTextModule) isCyclic() bool {
	return true
}

func (s *SourceTextModule) agent() *Agent {
	return s.Realm.Agent
}
