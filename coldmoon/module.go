package coldmoon

// Package 16.2.1.4

type ModuleRecord struct {
	SourceTextModule *SourceTextModule
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
	// TODO
	return nil
}

func ContinueDynamicImport(agent *Agent, capability *PromiseCapability, moduleCompletion CompletionModule) {
	if moduleCompletion.IsAbrupt() {
		capability.Reject.ToValue().CallAssumeCallable(UndefinedValue, []Value{moduleCompletion.Error()})
		return
	}

	// TODO
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
