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

// TODO: add missing semantics
