package coldmoon

import (
	"math"
	"reflect"
	"sort"

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

type AsyncEvaluationOrderState int

const (
	AsyncEvaluationOrderUnset AsyncEvaluationOrderState = iota
	AsyncEvaluationOrderInteger
	AsyncEvaluationOrderDone
)

// ModuleAsyncEvaluationOrder represents the [[AsyncEvaluationOrder]] union:
// either unset, an agent-assigned integer, or done.
type ModuleAsyncEvaluationOrder struct {
	State AsyncEvaluationOrderState
	Value uint64
}

func (o ModuleAsyncEvaluationOrder) IsInteger() bool {
	return o.State == AsyncEvaluationOrderInteger
}

type SourceTextModule struct {
	ScriptOrModule
	// Identity is the host-canonical module identifier used by ModuleGraph.
	Identity string
	// [[Realm]]
	Realm *Realm
	// [[Environment]]
	Environment EnvironmentRecord
	// [[Namespace]]
	Namespace ObjectType
	// [[ECMAScriptCode]]
	ECMAScriptCode *Module
	// Static contains parse-time module requests and declarations.
	Static ModuleStaticSemantics
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
	// [[AsyncEvaluationOrder]]
	AsyncEvaluationOrder ModuleAsyncEvaluationOrder
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
	Assert(module.Status == ModuleStatusUnlinked || module.Status == ModuleStatusLinked ||
		module.Status == ModuleStatusEvaluated || module.Status == ModuleStatusEvaluatingAsync)
	var stack pkg.Stack[ModuleRecord]
	_, err = s.InnerModuleLinking(&stack, 0)
	if err != nil {
		for _, item := range stack.Data() {
			i := item.(*SourceTextModule)
			Assert(i.Status == ModuleStatusLinking)
			i.Status = ModuleStatusUnlinked
		}
		return
	}
	Assert(module.Status == ModuleStatusLinked || module.Status == ModuleStatusEvaluatingAsync ||
		module.Status == ModuleStatusEvaluated)
	Assert(stack.Len() == 0)
	return
}

// 16.2.1.5.2.1
func (s *SourceTextModule) InnerModuleLinking(stack *pkg.Stack[ModuleRecord], i int) (index int, err Value) {
	index = i
	module := s
	if !module.isCyclic() {
		module.Link()
		return
	}

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
		requiredRecord := GetImportedModule(module, required)
		Assert(requiredRecord != nil)
		requiredModule, ok := requiredRecord.(*SourceTextModule)
		Assert(ok)
		index, err = requiredModule.InnerModuleLinking(stack, index)
		if err != nil {
			return
		}
		if requiredModule.isCyclic() {
			Assert(requiredModule.Status == ModuleStatusLinking ||
				requiredModule.Status == ModuleStatusLinked ||
				requiredModule.Status == ModuleStatusEvaluatingAsync ||
				requiredModule.Status == ModuleStatusEvaluated)
			if requiredModule.Status == ModuleStatusLinking {
				module.DFSAncestorIndex = int(math.Min(
					float64(module.DFSAncestorIndex),
					float64(requiredModule.DFSAncestorIndex),
				))
			}
		}
	}
	initializeResult := module.InitializeEnvironment()
	if initializeResult.IsAbrupt() {
		return index, initializeResult.Error()
	}
	Assert(lo.CountBy(stack.Data(), func(record ModuleRecord) bool {
		return record == ModuleRecord(module)
	}) == 1)
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
		if module.CycleRoot != nil {
			cycleRoot, ok := module.CycleRoot.(*SourceTextModule)
			Assert(ok)
			module = cycleRoot
		} else {
			// Abrupt evaluation can leave a module outside a completed SCC and
			// therefore without a cycle root. A later Evaluate call creates a
			// rejected capability for that module's retained error.
			Assert(module.Status == ModuleStatusEvaluated)
			Assert(module.EvaluationError != nil)
		}
	}
	Assert(module.Status == ModuleStatusLinked ||
		module.Status == ModuleStatusEvaluatingAsync ||
		module.Status == ModuleStatusEvaluated)

	if module.TopLevelCapability != nil {
		return module.TopLevelCapability.Promise
	}

	// Host callers may evaluate a linked module without an active script frame.
	// Promise construction still needs a current realm, so provide a lexical
	// host frame for the duration of synchronous module evaluation.
	var hostScope *ExecutionContextScope
	if !agent.hasExecutionContext() {
		hostScope = agent.enterExecutionContext(&ExecutionContext{
			Realm: realm,
			ch:    make(chan struct{}),
		})
		defer hostScope.Leave()
	}

	var stack pkg.Stack[ModuleRecord]
	capability := NewPromiseCapability(agent, realm.Intrinsics.Promise.ToValue())
	module.TopLevelCapability = capability
	_, err := module.InnerModuleEvaluation(&stack, 0)
	if err != nil {
		for _, item := range stack.Data() {
			i := item.(*SourceTextModule)
			Assert(i.Status == ModuleStatusEvaluating)
			i.Status = ModuleStatusEvaluated
			i.EvaluationError = err
		}

		Assert(module.Status == ModuleStatusEvaluated)
		Assert(module.EvaluationError == err)

		ReturnAssertNormal(capability.Reject.Call(UndefinedValue, []Value{err}))
	} else {
		Assert(module.Status == ModuleStatusEvaluated || module.Status == ModuleStatusEvaluatingAsync)
		if module.Status == ModuleStatusEvaluated {
			Assert(module.Status == ModuleStatusEvaluated)
			ReturnAssertNormal(capability.Resolve.Call(UndefinedValue, []Value{UndefinedValue}))
		}
		Assert(stack.Len() == 0)
	}

	return capability.Promise
}

// 16.2.1.5.3.1
func (s *SourceTextModule) InnerModuleEvaluation(stack *pkg.Stack[ModuleRecord], index int) (r int, err Value) {
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
		requiredRecord := GetImportedModule(m, required)
		Assert(requiredRecord != nil)
		requiredModule, ok := requiredRecord.(*SourceTextModule)
		Assert(ok)
		newIndex, err = requiredModule.InnerModuleEvaluation(stack, newIndex)
		if err != nil {
			return newIndex, err
		}

		Assert(requiredModule.Status == ModuleStatusEvaluated ||
			requiredModule.Status == ModuleStatusEvaluatingAsync ||
			requiredModule.Status == ModuleStatusEvaluating)
		if requiredModule.Status == ModuleStatusEvaluating {
			Assert(lo.Contains(stack.Data(), ModuleRecord(requiredModule)))
			m.DFSAncestorIndex = int(math.Min(
				float64(m.DFSAncestorIndex),
				float64(requiredModule.DFSAncestorIndex),
			))
		} else {
			Assert(requiredModule.CycleRoot != nil)
			cycleRoot, ok := requiredModule.CycleRoot.(*SourceTextModule)
			Assert(ok)
			requiredModule = cycleRoot
			Assert(requiredModule.Status == ModuleStatusEvaluated ||
				requiredModule.Status == ModuleStatusEvaluatingAsync)
			if requiredModule.EvaluationError != nil {
				return newIndex, requiredModule.EvaluationError
			}
		}

		if requiredModule.AsyncEvaluationOrder.IsInteger() {
			m.PendingAsyncDependencies++
			requiredModule.AsyncParentModules.Push(module)
		}
	}

	if m.PendingAsyncDependencies > 0 || m.HasTLA {
		Assert(m.AsyncEvaluationOrder.State == AsyncEvaluationOrderUnset)
		m.AsyncEvaluationOrder = ModuleAsyncEvaluationOrder{
			State: AsyncEvaluationOrderInteger,
			Value: m.agent().IncrementModuleAsyncEvaluationCount(),
		}
		if m.PendingAsyncDependencies == 0 {
			module.ExecuteAsyncModule()
		}
	} else {
		result := m.ExecuteModule(nil)
		if result.IsAbrupt() {
			return newIndex, result.Error()
		}
	}

	if m.DFSAncestorIndex == m.DFSIndex {
		for {
			requiredModule := stack.Pop().(*SourceTextModule)
			if requiredModule.AsyncEvaluationOrder.IsInteger() {
				requiredModule.Status = ModuleStatusEvaluatingAsync
			} else {
				Assert(requiredModule.AsyncEvaluationOrder.State == AsyncEvaluationOrderUnset)
				requiredModule.Status = ModuleStatusEvaluated
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
	Assert(s.Status == ModuleStatusEvaluating || s.Status == ModuleStatusEvaluatingAsync)
	Assert(s.HasTLA)
	Assert(s.AsyncEvaluationOrder.IsInteger())

	agent := s.agent()
	executionCapability := NewPromiseCapability(agent, s.Realm.Intrinsics.Promise.ToValue())
	var fulfilledClosure BehaviorFn = func(
		thisArgument Value,
		argumentsList []Value,
		newTarget ObjectType,
	) CompletionConvertable[Value] {
		s.asyncModuleExecutionFulfilled()
		return UndefinedValue
	}
	onFulfilled := CreateBuiltinFunction(
		agent,
		fulfilledClosure,
		0,
		CMString(""),
		builtinFunctionArgs{realm: s.Realm},
	)
	var rejectedClosure BehaviorFn = func(
		thisArgument Value,
		argumentsList []Value,
		newTarget ObjectType,
	) CompletionConvertable[Value] {
		s.asyncModuleExecutionRejected(argumentAt(argumentsList, 0))
		return UndefinedValue
	}
	onRejected := CreateBuiltinFunction(
		agent,
		rejectedClosure,
		0,
		CMString(""),
		builtinFunctionArgs{realm: s.Realm},
	)
	PerformPromiseThen(
		agent,
		executionCapability.Promise,
		onFulfilled.ToValue(),
		onRejected.ToValue(),
		nil,
	)

	result := s.ExecuteModule(executionCapability)
	Assert(!result.IsAbrupt())
}

// gatherAvailableAncestors gathers every synchronous ancestor made executable
// by this module's completion. The caller sorts the resulting list by DFS
// evaluation order before executing it.
func (s *SourceTextModule) gatherAvailableAncestors(
	execList *[]*SourceTextModule,
	included map[*SourceTextModule]struct{},
) {
	for _, parentRecord := range s.AsyncParentModules.Data() {
		parent, ok := parentRecord.(*SourceTextModule)
		Assert(ok)
		if _, exists := included[parent]; exists {
			continue
		}

		if parent.CycleRoot == nil {
			// Abrupt synchronous evaluation may have completed this parent
			// before the dependency's async execution promise settles.
			Assert(parent.Status == ModuleStatusEvaluated)
			Assert(parent.EvaluationError != nil)
			continue
		}
		cycleRoot, ok := parent.CycleRoot.(*SourceTextModule)
		Assert(ok)
		if cycleRoot.EvaluationError != nil {
			continue
		}

		Assert(parent.Status == ModuleStatusEvaluatingAsync)
		Assert(parent.EvaluationError == nil)
		Assert(parent.AsyncEvaluationOrder.IsInteger())
		Assert(parent.PendingAsyncDependencies > 0)
		parent.PendingAsyncDependencies--
		if parent.PendingAsyncDependencies != 0 {
			continue
		}

		*execList = append(*execList, parent)
		included[parent] = struct{}{}
		if !parent.HasTLA {
			parent.gatherAvailableAncestors(execList, included)
		}
	}
}

// asyncModuleExecutionFulfilled releases all newly available ancestors in
// their original DFS evaluation order.
func (s *SourceTextModule) asyncModuleExecutionFulfilled() {
	if s.Status == ModuleStatusEvaluated {
		// A sibling in the same async cycle may already have propagated a
		// rejection while this module's own execution promise was settling.
		Assert(s.EvaluationError != nil)
		return
	}
	Assert(s.Status == ModuleStatusEvaluatingAsync)
	Assert(s.AsyncEvaluationOrder.IsInteger())
	Assert(s.EvaluationError == nil)

	s.AsyncEvaluationOrder = ModuleAsyncEvaluationOrder{State: AsyncEvaluationOrderDone}
	s.Status = ModuleStatusEvaluated

	if s.TopLevelCapability != nil {
		Assert(s.CycleRoot == ModuleRecord(s))
		ReturnAssertNormal(s.TopLevelCapability.Resolve.Call(
			UndefinedValue,
			[]Value{UndefinedValue},
		))
	}

	execList := make([]*SourceTextModule, 0)
	s.gatherAvailableAncestors(&execList, make(map[*SourceTextModule]struct{}))
	sort.SliceStable(execList, func(i, j int) bool {
		Assert(execList[i].AsyncEvaluationOrder.IsInteger())
		Assert(execList[j].AsyncEvaluationOrder.IsInteger())
		return execList[i].AsyncEvaluationOrder.Value < execList[j].AsyncEvaluationOrder.Value
	})

	for _, parent := range execList {
		if parent.Status == ModuleStatusEvaluated {
			Assert(parent.EvaluationError != nil)
			continue
		}
		Assert(parent.Status == ModuleStatusEvaluatingAsync)
		Assert(parent.AsyncEvaluationOrder.IsInteger())
		Assert(parent.PendingAsyncDependencies == 0)
		Assert(parent.EvaluationError == nil)

		if parent.HasTLA {
			parent.ExecuteAsyncModule()
			continue
		}
		result := parent.ExecuteModule(nil)
		if result.IsAbrupt() {
			parent.asyncModuleExecutionRejected(result.Error())
			return
		}
		parent.AsyncEvaluationOrder = ModuleAsyncEvaluationOrder{State: AsyncEvaluationOrderDone}
		parent.Status = ModuleStatusEvaluated
		if parent.TopLevelCapability != nil {
			Assert(parent.CycleRoot == ModuleRecord(parent))
			ReturnAssertNormal(parent.TopLevelCapability.Resolve.Call(
				UndefinedValue,
				[]Value{UndefinedValue},
			))
		}
	}
}

// asyncModuleExecutionRejected records one shared evaluation failure across
// every ancestor waiting on this module and rejects the cycle root capability.
func (s *SourceTextModule) asyncModuleExecutionRejected(errorValue Value) {
	Assert(errorValue != nil)
	if s.Status == ModuleStatusEvaluated {
		Assert(s.EvaluationError != nil)
		return
	}
	Assert(s.Status == ModuleStatusEvaluatingAsync)
	Assert(s.AsyncEvaluationOrder.IsInteger())

	s.EvaluationError = errorValue
	s.AsyncEvaluationOrder = ModuleAsyncEvaluationOrder{State: AsyncEvaluationOrderDone}
	s.Status = ModuleStatusEvaluated
	parents := append([]ModuleRecord(nil), s.AsyncParentModules.Data()...)
	if s.TopLevelCapability != nil {
		Assert(s.CycleRoot == ModuleRecord(s))
		ReturnAssertNormal(s.TopLevelCapability.Reject.Call(
			UndefinedValue,
			[]Value{errorValue},
		))
	}
	for _, parentRecord := range parents {
		parent, ok := parentRecord.(*SourceTextModule)
		Assert(ok)
		parent.asyncModuleExecutionRejected(errorValue)
	}
}

// [[BindingName]] union
type BindingName struct {
	String    string
	Namespace bool
}

var bindingNameNamespace = &BindingName{Namespace: true}

func (b *BindingName) IsNamespace() bool {
	return b != nil && b.Namespace
}

func (b *BindingName) IsEqualTo(other *BindingName) bool {
	if b == nil || other == nil {
		return b == other
	}
	if b.IsNamespace() != other.IsNamespace() {
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
	sharedResolveSet := make(map[ResolveSet]bool, len(resolveSet))
	for _, resolution := range resolveSet {
		sharedResolveSet[resolution] = true
	}
	return s.resolveExport(exportName, sharedResolveSet)
}

func (s *SourceTextModule) resolveExport(exportName string, resolveSet map[ResolveSet]bool) *ResolvedBinding {
	Assert(s.Status != ModuleStatusNew)

	resolutionPair := ResolveSet{Module: s, ExportName: exportName}
	if resolveSet[resolutionPair] {
		return resolvedBindingNull
	}
	resolveSet[resolutionPair] = true

	for _, ee := range s.LocalExportEntries {
		if ee.ExportName == exportName {
			Assert(ee.ModuleRequest == "")
			Assert(ee.LocalName != "")
			return &ResolvedBinding{
				Module:      s,
				BindingName: &BindingName{String: ee.LocalName},
			}
		}
	}

	for _, e := range s.IndirectExportEntries {
		if e.ExportName == exportName {
			Assert(e.ModuleRequest != "")
			importedRecord := GetImportedModule(s, e.ModuleRequest)
			Assert(importedRecord != nil)
			importedModule, ok := importedRecord.(*SourceTextModule)
			Assert(ok)
			if e.ImportName == ImportNameAll {
				// Assert: module does not provide the direct binding for this export.
				return &ResolvedBinding{
					Module:      importedModule,
					BindingName: bindingNameNamespace,
				}
			} else {
				Assert(e.ImportName != "")
				Assert(e.ImportName != ImportNameAll)
				Assert(e.ImportName != ImportNameAllButDefault)
				Assert(e.ImportName != ImportNameNamespaceObject)
				return importedModule.resolveExport(string(e.ImportName), resolveSet)
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
		Assert(e.ImportName == ImportNameAllButDefault)
		importedRecord := GetImportedModule(s, e.ModuleRequest)
		Assert(importedRecord != nil)
		importedModule, ok := importedRecord.(*SourceTextModule)
		Assert(ok)
		resolution := importedModule.resolveExport(exportName, resolveSet)
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
	Assert(s.Status != ModuleStatusNew)
	return s.getExportedNames(make(map[ModuleRecord]bool))
}

// getExportedNames carries the export-star set used to terminate circular
// re-export graphs. Explicit exports are visited before star exports so the
// result follows the source module's declaration order.
func (s *SourceTextModule) getExportedNames(exportStarSet map[ModuleRecord]bool) Exports {
	if exportStarSet[s] {
		return nil
	}
	exportStarSet[s] = true

	exportedNames := make(Exports, 0, len(s.LocalExportEntries)+len(s.IndirectExportEntries))
	seenNames := make(map[string]bool)
	appendName := func(name string) {
		if !seenNames[name] {
			seenNames[name] = true
			exportedNames = append(exportedNames, name)
		}
	}
	for _, entry := range s.LocalExportEntries {
		appendName(entry.ExportName)
	}
	for _, entry := range s.IndirectExportEntries {
		appendName(entry.ExportName)
	}
	for _, entry := range s.StarExportEntries {
		Assert(entry.ModuleRequest != "")
		Assert(entry.ImportName == ImportNameAllButDefault)
		importedRecord := GetImportedModule(s, entry.ModuleRequest)
		Assert(importedRecord != nil)
		importedModule, ok := importedRecord.(*SourceTextModule)
		Assert(ok)
		for _, name := range importedModule.getExportedNames(exportStarSet) {
			if name != "default" {
				appendName(name)
			}
		}
	}
	return exportedNames
}

// 16.2.1.6.1
func ParseModule(sourceText string, realm *Realm, hostDefined HostDefined) *SourceTextModule {
	body := NewParser(sourceText, ParserContext{
		FileName: hostDefined.FileName,
		BaseDir:  hostDefined.BaseDir,
	}).ParseModule()

	static := (StaticSemantics{}).AnalyzeModule(body)
	requestedModules := static.RequestedModules
	importEntries := static.ImportEntries
	var indirectExportEntries []ExportEntry
	var localExportEntries []ExportEntry
	var starExportEntries []ExportEntry
	exportEntries := static.ExportEntries

	for _, ee := range exportEntries {
		if ee.ModuleRequest == "" {
			var importEntryBoundName *ImportEntryRecord
			for i := range importEntries {
				if importEntries[i].LocalName == ee.LocalName {
					importEntryBoundName = &importEntries[i]
					break
				}
			}
			if importEntryBoundName == nil {
				localExportEntries = append(localExportEntries, ee)
			} else {
				ie := importEntryBoundName
				if ie.ImportName == ImportNameNamespaceObject {
					indirectExportEntries = append(indirectExportEntries, ExportEntry{
						ModuleRequest: ie.ModuleRequest,
						ImportName:    ImportNameAll,
						ExportName:    ee.ExportName,
					})
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
		Static:                static,
		LoadedModules:         make(map[string]ModuleRecord),
		HasTLA:                moduleHasTopLevelAwait(body),
	}
}

// moduleHasTopLevelAwait performs the specification's syntax-directed
// Contains(AwaitExpression) query without crossing function-body boundaries.
// The AST currently has no generated visitor, so this small reflective walker
// keeps the query exhaustive as new expression containers are added.
func moduleHasTopLevelAwait(module *Module) bool {
	return syntaxContainsTopLevelAwait(reflect.ValueOf(module), make(map[uintptr]bool))
}

func syntaxContainsTopLevelAwait(node reflect.Value, visited map[uintptr]bool) bool {
	if !node.IsValid() {
		return false
	}
	if node.Kind() == reflect.Interface {
		if node.IsNil() {
			return false
		}
		return syntaxContainsTopLevelAwait(node.Elem(), visited)
	}
	if node.Kind() == reflect.Pointer {
		if node.IsNil() {
			return false
		}
		if node.CanInterface() {
			switch syntaxNode := node.Interface().(type) {
			case *AwaitExpression:
				return true
			case *ForInOfStatement:
				if syntaxNode.IsAwait {
					return true
				}
			case *ClassTail:
				return classTailContainsTopLevelAwait(syntaxNode, visited)
			case *FunctionBody:
				return false
			}
		}
		pointer := node.Pointer()
		if visited[pointer] {
			return false
		}
		visited[pointer] = true
		return syntaxContainsTopLevelAwait(node.Elem(), visited)
	}

	switch node.Kind() {
	case reflect.Struct:
		for i := 0; i < node.NumField(); i++ {
			if syntaxContainsTopLevelAwait(node.Field(i), visited) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < node.Len(); i++ {
			if syntaxContainsTopLevelAwait(node.Index(i), visited) {
				return true
			}
		}
	}
	return false
}

// classTailContainsTopLevelAwait follows ClassTail's specialized Contains
// semantics: heritage and computed names execute with the surrounding module,
// while method bodies, field initializers, and static blocks do not.
func classTailContainsTopLevelAwait(classTail *ClassTail, visited map[uintptr]bool) bool {
	if classTail == nil {
		return false
	}
	if syntaxContainsTopLevelAwait(reflect.ValueOf(classTail.ClassHeritage), visited) {
		return true
	}
	if classTail.ClassBody == nil || classTail.ClassBody.ClassElementList == nil {
		return false
	}
	for _, element := range classTail.ClassBody.ClassElementList.Items {
		var propertyName PropertyName
		switch element := element.(type) {
		case *ClassElementMethodDefinition:
			propertyName = element.MethodDefinition.PropertyName
		case *ClassElementFieldDefinition:
			propertyName = element.FieldDefinition.PropertyName
		case *ClassElementStaticBlock, *ClassElementEmpty:
			continue
		default:
			panic("unreachable")
		}
		if syntaxContainsTopLevelAwait(reflect.ValueOf(propertyName), visited) {
			return true
		}
	}
	return false
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
func (s *SourceTextModule) LoadRequestedModules() *PromiseObject {
	realm := s.Realm
	pc := NewPromiseCapability(s.agent(), realm.Intrinsics.Promise.ToValue())

	state := &GraphLoadingState{
		IsLoading:           true,
		PendingModulesCount: 1,
		Visited:             make([]ModuleRecord, 0),
		PromiseCapability:   pc,
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
	scope := agent.enterExecutionContext(moduleContext)
	defer scope.Leave()
	code := s.ECMAScriptCode

	lexicalDeclarations := moduleLexicallyScopedDeclarations(code)
	for _, declaration := range lexicalDeclarations {
		for _, boundName := range declaration.boundNames {
			name := string(boundName)
			Assert(!env.HasBinding(name))
			if declaration.constant {
				env.CreateImmutableBinding(name, true)
			} else {
				env.CreateMutableBinding(name, false)
			}
		}
	}
	for _, declaration := range lexicalDeclarations {
		name, function, ok := instantiateModuleFunctionDeclaration(agent, declaration, env)
		if ok {
			env.InitializeBinding(name, function.ToValue())
		}
	}

	varDeclarations := moduleVarScopedDeclarations(code)
	declaredVarNames := make(map[string]bool)

	for _, varDeclaration := range varDeclarations {
		for _, boundName := range varDeclaration.BoundNames() {
			varName := string(boundName)
			if _, ok := declaredVarNames[varName]; !ok {
				if !env.HasBinding(varName) {
					env.CreateMutableBinding(varName, false)
					env.InitializeBinding(varName, UndefinedValue)
				}
				declaredVarNames[varName] = true
			}
		}
	}

	co.value = UndefinedValue
	return
}

type moduleScopedDeclaration struct {
	declaration      Declaration
	boundNames       []IdentifierName
	constant         bool
	anonymousDefault bool
}

func newModuleScopedDeclaration(declaration Declaration) moduleScopedDeclaration {
	return moduleScopedDeclaration{
		declaration: declaration,
		boundNames:  sourceTextModuleDeclarationBoundNames(declaration),
		constant:    moduleDeclarationIsConstant(declaration),
	}
}

// moduleLexicallyScopedDeclarations extracts the declarations instantiated by
// ModuleDeclarationInstantiation without treating let/const declarations as
// var declarations. It retains the synthetic *default* binding that is not
// represented by anonymous declaration identifiers in the AST.
func moduleLexicallyScopedDeclarations(code *Module) (declarations []moduleScopedDeclaration) {
	for _, item := range code.ModuleItemList {
		switch item := item.(type) {
		case *ModuleItemStatementListItem:
			declarationItem, ok := item.StatementListItem.(*StatementListItemDeclaration)
			if ok {
				declarations = append(declarations, newModuleScopedDeclaration(declarationItem.Declaration))
			}
		case *ModuleItemExportDeclaration:
			switch {
			case item.Declaration != nil:
				declarations = append(declarations, newModuleScopedDeclaration(item.Declaration))
			case item.DefaultHoistableDeclaration != nil:
				declaration := newModuleScopedDeclaration(item.DefaultHoistableDeclaration)
				if len(declaration.boundNames) == 0 {
					declaration.boundNames = []IdentifierName{"*default*"}
					declaration.anonymousDefault = true
				}
				declarations = append(declarations, declaration)
			case item.DefaultClassDeclaration != nil:
				declaration := newModuleScopedDeclaration(item.DefaultClassDeclaration)
				if item.DefaultClassDeclaration.IdentifierName == "" {
					declaration.boundNames = []IdentifierName{"*default*"}
					declaration.anonymousDefault = true
				}
				declarations = append(declarations, declaration)
			case item.DefaultExpression != nil:
				declarations = append(declarations, moduleScopedDeclaration{
					boundNames: []IdentifierName{"*default*"},
				})
			}
		case *ModuleItemImportDeclaration:
			continue
		default:
			panic("unreachable")
		}
	}
	return declarations
}

func moduleVarScopedDeclarations(code *Module) (declarations []*VariableDeclaration) {
	for _, item := range code.ModuleItemList {
		switch item := item.(type) {
		case *ModuleItemStatementListItem:
			statementItem, ok := item.StatementListItem.(*StatementListItemStatement)
			if ok {
				declarations = append(declarations, statementItem.Statement.VarScopedDeclarations()...)
			}
		case *ModuleItemExportDeclaration:
			if item.VariableStatement != nil {
				declarations = append(declarations, item.VariableStatement.VarScopedDeclarations()...)
			}
		case *ModuleItemImportDeclaration:
			continue
		default:
			panic("unreachable")
		}
	}
	return declarations
}

func sourceTextModuleDeclarationBoundNames(declaration Declaration) []IdentifierName {
	switch declaration := declaration.(type) {
	case *DeclarationHoistableAsyncFunction:
		if declaration.AsyncFunctionDeclaration.Identifier == "" {
			return nil
		}
		return []IdentifierName{declaration.AsyncFunctionDeclaration.Identifier}
	case *DeclarationHoistableGenerator:
		if declaration.GeneratorDeclaration.Identifier == "" {
			return nil
		}
		return []IdentifierName{declaration.GeneratorDeclaration.Identifier}
	case *DeclarationHoistableAsyncGenerator:
		if declaration.AsyncGeneratorDeclaration.Identifier == "" {
			return nil
		}
		return []IdentifierName{declaration.AsyncGeneratorDeclaration.Identifier}
	default:
		return declaration.BoundNames()
	}
}

func moduleDeclarationIsConstant(declaration Declaration) bool {
	if _, ok := declaration.(*ClassDeclaration); ok {
		return true
	}
	return IsConstantDeclaration(declaration)
}

func instantiateModuleFunctionDeclaration(
	agent *Agent,
	scopedDeclaration moduleScopedDeclaration,
	env EnvironmentRecord,
) (name string, function ObjectType, ok bool) {
	declaration := scopedDeclaration.declaration
	switch declaration := declaration.(type) {
	case *DeclarationHoistableFunction:
		functionDeclaration := declaration.FunctionDeclaration
		name = functionDeclaration.Identifier
		function = functionDeclaration.instantiateOrdinaryFunctionObject(agent, env, nil)
	case *DeclarationHoistableAsyncFunction:
		functionDeclaration := declaration.AsyncFunctionDeclaration
		name = functionDeclaration.Identifier
		function = functionDeclaration.instantiateAsyncFunctionObject(agent, env, nil)
	case *DeclarationHoistableGenerator:
		functionDeclaration := declaration.GeneratorDeclaration
		name = functionDeclaration.Identifier
		function = functionDeclaration.instantiateGeneratorFunctionObject(agent, env, nil)
	case *DeclarationHoistableAsyncGenerator:
		functionDeclaration := declaration.AsyncGeneratorDeclaration
		name = functionDeclaration.Identifier
		function = functionDeclaration.instantiateAsyncGeneratorFunctionObject(agent, env, nil)
	default:
		return "", nil, false
	}
	if scopedDeclaration.anonymousDefault {
		name = "*default*"
		function.DefinePropertyOrThrow(NewStringPropertyKey("name"), &PropertyDescriptor{
			Value:           NewStringValue("default"),
			Writable:        false,
			WritableSet:     true,
			Enumerable:      false,
			EnumerableSet:   true,
			Configurable:    true,
			ConfigurableSet: true,
		})
	}
	return name, function, true
}

// ExecuteModule evaluates a module in its module environment. Synchronous
// abrupt completions are returned to InnerModuleEvaluation; top-level-await
// completions settle capability after the body resumes through the scheduler.
// spec: 16.2.1.6.5
func (s *SourceTextModule) ExecuteModule(capability *PromiseCapability) (co CompletionValue) {
	agent := s.Realm.Agent
	Assert(s.Environment != nil)
	moduleContext := &ExecutionContext{
		Realm:          s.Realm,
		ScriptOrModule: s,
		ch:             make(chan struct{}),
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  s.Environment,
			VariableEnvironment: s.Environment,
		},
	}
	s.Context = moduleContext
	if !s.HasTLA {
		Assert(capability == nil)
		scope := agent.enterExecutionContext(moduleContext)
		defer scope.Leave()
		result := RunNode(agent, s.ECMAScriptCode)
		if result.IsAbrupt() {
			return result
		}
		return UndefinedValue.ToCompletion()
	}

	Assert(capability != nil)
	moduleContext.awaitCh = make(chan struct{})
	callerContext := agent.prepareModuleAsyncContext(moduleContext)

	// The initial handoff runs synchronously from the caller's perspective: it
	// returns when the module either reaches its first await or completes. The
	// scheduler then owns all later resumptions and capability settlement.
	agent.Scheduler.StartTask(func() {
		agent.resumeExecutionContext(moduleContext)
		result := RunNode(agent, s.ECMAScriptCode)
		agent.finishModuleAsyncContext(moduleContext)
		agent.suspendExecutionContext(moduleContext)

		if result.IsAbrupt() {
			errorValue := result.Error()
			if errorValue == nil {
				errorValue = result.Data()
			}
			Assert(errorValue != nil)
			ReturnAssertNormal(capability.Reject.Call(
				UndefinedValue,
				[]Value{errorValue},
			))
		} else {
			ReturnAssertNormal(capability.Resolve.Call(
				UndefinedValue,
				[]Value{UndefinedValue},
			))
		}

		if !moduleContext.asyncCallerResumed {
			moduleContext.asyncCallerResumed = true
			moduleContext.Resume()
		}
	})
	moduleContext.Suspend()
	if callerContext == nil {
		Assert(!agent.hasExecutionContext())
	} else {
		Assert(agent.RunningExecutionContext() == callerContext)
	}
	return UndefinedValue.ToCompletion()
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
				result := s.agent().ModuleGraph.Load(module.ToReferrer(), required, module.HostDefined)
				FinishLoadingImportedModule(s.agent(), module.ToReferrer(), required, ImportedModulePayload{
					GraphLoadingState: state,
				}, result)
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
		}
		ReturnAssertNormal(state.PromiseCapability.Resolve.Call(
			UndefinedValue,
			[]Value{UndefinedValue},
		))
	}
}

// MARK: - Internal

func (s *SourceTextModule) _moduleRecord() {}

// SourceTextModule is a Cyclic Module Record in the specification's type
// hierarchy, regardless of whether its requested-module graph contains a cycle.
func (s *SourceTextModule) isCyclic() bool {
	return true
}

func (s *SourceTextModule) agent() *Agent {
	return s.Realm.Agent
}
