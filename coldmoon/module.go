package coldmoon

// 16.2.1.4

type ModuleRecord interface {
	_moduleRecord()
}

type CompletionModule = Completion[ModuleRecord]

type ResolveSet struct {
	// [[Module]]
	Module ModuleRecord
	// [[ExportName]]
	ExportName string
}

// GraphLoadingState Record
type GraphLoadingState struct {
	IsLoading           bool
	PendingModulesCount int
	PromiseCapability   *PromiseCapability
	Visited             []ModuleRecord
	HostDefined         HostDefined
}

// ImportedModuleReferrer Enum
type ImportedModuleReferrer struct {
	Script *ScriptRecord
	Module *SourceTextModule
	Realm  *Realm
}

// ImportedModulePayload Enum
type ImportedModulePayload struct {
	GraphLoadingState *GraphLoadingState
	PromiseCapability *PromiseCapability
}

// 16.2.1.7
func GetImportedModule(referrer *SourceTextModule, specifier string) ModuleRecord {
	return referrer.LoadedModules[specifier]
}

func GetModuleNamespace(agent *Agent, module *SourceTextModule) ObjectType {
	namespace := module.Namespace
	if namespace == nil {
		exportedNames := module.GetExportedNames()

		unambiguousNames := make([]string, 0)
		for _, name := range exportedNames {
			resolution := module.ResolveExport(name, nil)
			if resolution != nil && !resolution.IsAmbiguous() {
				unambiguousNames = append(unambiguousNames, name)
			}
		}
		namespace = ModuleNamespaceCreate(agent, module, unambiguousNames)
	}
	return namespace
}

func ContinueDynamicImport(agent *Agent, capability *PromiseCapability, moduleCompletion CompletionModule) {
	if moduleCompletion.IsAbrupt() {
		capability.Reject.Call(UndefinedValue, []Value{moduleCompletion.Error()})
		return
	}

	module := moduleCompletion.Data().(*SourceTextModule)
	loadPromise := module.LoadRequestedModules(module.HostDefined)

	var rejectedClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		capability.Reject.Call(UndefinedValue, []Value{argumentsList[0]})
		return nil
	}
	onRejected := CreateBuiltinFunction(agent, rejectedClosure, 1, CMString("onRejected"), builtinFunctionArgs{})
	var linkAndEvaluateClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		err := module.Link()
		if err != nil {
			capability.Reject.Call(UndefinedValue, []Value{err})
			return nil
		}
		evaluatePromise := module.Evaluate()
		var fulfilledClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
			namespace := GetModuleNamespace(agent, module)
			capability.Resolve.Call(UndefinedValue, []Value{(namespace).ToValue()})
			return nil
		}
		onFulfilled := CreateBuiltinFunction(agent, fulfilledClosure, 0, CMString(""), builtinFunctionArgs{})
		PerformPromiseThen(agent, evaluatePromise, (onFulfilled).ToValue(), (onRejected).ToValue(), nil)
		return nil
	}

	linkAndEvaluate := CreateBuiltinFunction(agent, linkAndEvaluateClosure, 0, CMString(""), builtinFunctionArgs{})
	PerformPromiseThen(agent, loadPromise, (linkAndEvaluate).ToValue(), (onRejected).ToValue(), nil)
}

func FinishLoadingImportedModule(
	agent *Agent,
	referrer ImportedModuleReferrer,
	specifier string,
	payload ImportedModulePayload,
	result CompletionModule,
) {
	if result.t == CompletionTypeNormal {
		module := result.Data()
		if referrer.Script != nil {
			_, ok := referrer.Script.LoadedModules[specifier]
			if !ok {
				referrer.Script.LoadedModules[specifier] = module
			}
		} else if referrer.Module != nil {
			_, ok := referrer.Module.LoadedModules[specifier]
			if !ok {
				referrer.Module.LoadedModules[specifier] = module
			}
		}
	}
	if payload.GraphLoadingState != nil {
		ContinueModuleLoading(agent, payload.GraphLoadingState, result)
	} else if payload.PromiseCapability != nil {
		ContinueDynamicImport(agent, payload.PromiseCapability, result)
	} else {
		panic("unreachable")
	}
}
