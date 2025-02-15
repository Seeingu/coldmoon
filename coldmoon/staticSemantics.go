package coldmoon

// 8.2.1
// TODO: remove boundName interface
type StaticSemanticsBoundNames interface {
	// TODO: lower case should be more appropriate
	BoundNames() (l []IdentifierName)
}

// 16.2.1.3
type StaticSemanticsModuleRequests interface {
	moduleRequests() []string
}

// 16.2.2.2
type StaticSemanticsImportEntries interface {
	importEntries() []ImportEntryRecord
}

// 16.2.3.4
type StaticSemanticsExportEntries interface {
	exportEntries() []ExportEntry
}

type StaticSemanticsIsAnonymousFunctionDefinition interface {
	IsAnonymousFunctionDefinition() bool
}

// 13.2.5.2
type StaticSemanticsIsComputedPropertyKey interface {
	IsComputedPropertyKey() bool
}

func IsComputedPropertyKeyDefault(expr PropertyName) bool {
	if _, ok := expr.(*ComputedPropertyName); ok {
		return true
	}
	return false
}

// TODO: add missing semantics
