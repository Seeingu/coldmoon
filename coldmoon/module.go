package coldmoon

// Package 16.2.1.4

type ModuleRecord struct {
	SourceTextModule *SourceTextModule
}

func GetImportedModule(referrer *SourceTextModule, specifier string) *ModuleRecord {
	// TODO
	return nil
}

func GetModuleNamespace(agent *Agent, module *ModuleRecord) ObjectType {
	// TODO
	return nil
}
