package coldmoon

// Package 16.2.1.4

type ModuleRecord struct {
	*SourceTextModule
}

type GraphLoadingState struct {
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

func GetImportedModule(referrer *SourceTextModule, specifier string) *ModuleRecord {
	// TODO
	return nil
}

func GetModuleNamespace(agent *Agent, module *ModuleRecord) ObjectType {
	namespace := module.SourceTextModule.Namespace
	if namespace == nil {
		exportedNames := module.GetExportedNames()

		unambiguousNames := make([]string, 0)
		for _, name := range exportedNames {
			resolution := module.ResolveExport(name, nil)
			if resolution != nil && !resolution.Ambiguous {
				unambiguousNames = append(unambiguousNames, name)
			}
		}
		namespace = ModuleNamespaceCreate(agent, module, unambiguousNames)
	}
	return namespace
}

func ContinueDynamicImport(agent *Agent, capability *PromiseCapability, moduleCompletion CompletionModule) {
	if moduleCompletion.IsAbrupt() {
		capability.Reject.ToValue().CallAssumeCallable(UndefinedValue, []Value{moduleCompletion.Error()})
		return
	}

	module := moduleCompletion.Data()
	loadPromise := module.LoadRequestedModules(agent, nil)

	var rejectedClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		capability.Reject.ToValue().CallAssumeCallable(UndefinedValue, []Value{argumentsList[0]})
		return nil
	}
	onRejected := CreateBuiltinFunction(agent, rejectedClosure, 1, "onRejected", builtinFunctionArgs{})
	var linkAndEvaluateClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		link := module.Link()
		if link.IsAbrupt() {
			capability.Reject.ToValue().CallAssumeCallable(UndefinedValue, []Value{link.Error})
			return nil
		}
		evaluatePromise := module.Evaluate()
		var fulfilledClosure BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
			namespace := GetModuleNamespace(agent, module)
			capability.Resolve.ToValue().CallAssumeCallable(UndefinedValue, []Value{namespace.ToValue()})
			return nil
		}
		onFulfilled := CreateBuiltinFunction(agent, fulfilledClosure, 0, "", builtinFunctionArgs{})
		PerformPromiseThen(agent, evaluatePromise, onFulfilled.ToValue(), onRejected.ToValue(), nil)
		return nil
	}

	linkAndEvaluate := CreateBuiltinFunction(agent, linkAndEvaluateClosure, 0, "", builtinFunctionArgs{})
	PerformPromiseThen(agent, loadPromise, linkAndEvaluate.ToValue(), onRejected.ToValue(), nil)
}

func FinishLoadingImportedModule(
	agent *Agent,
	referrer ImportedModuleReferrer,
	specifier string,
	payload ImportedModulePayload,
	result CompletionModule) {
	if result.Type == CompletionTypeNormal {
		module := result.Data()
		if referrer.Script != nil {
			_, ok := referrer.Script.LoadedModules[specifier]
			if !ok {
				referrer.Script.LoadedModules[specifier] = module
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
