package coldmoon

// StaticSemanticsBoundNames
// spec: 8.2.1
type StaticSemanticsBoundNames interface {
	BoundNames() (l []IdentifierName)
}

// StaticSemanticsModuleRequests
// spec: 16.2.1.3
type StaticSemanticsModuleRequests interface {
	moduleRequests() []string
}

// StaticSemanticsImportEntries
// spec: 16.2.2.2
type StaticSemanticsImportEntries interface {
	importEntries() []ImportEntryRecord
}

// StaticSemanticsExportEntries
// spec: 16.2.3.4
type StaticSemanticsExportEntries interface {
	exportEntries() []ExportEntry
}

type StaticSemanticsIsAnonymousFunctionDefinition interface {
	IsAnonymousFunctionDefinition() bool
}

// StaticSemanticsIsComputedPropertyKey
// spec: 13.2.5.2
type StaticSemanticsIsComputedPropertyKey interface {
	IsComputedPropertyKey() bool
}

func IsComputedPropertyKeyDefault(expr PropertyName) bool {
	if _, ok := expr.(*ComputedPropertyName); ok {
		return true
	}
	return false
}

// StaticSemanticsIsConstantDeclaration
// spec: 8.2.3
type StaticSemanticsIsConstantDeclaration interface {
	IsConstantDeclaration() bool
}

func IsConstantDeclaration(d ASTNode) bool {
	if dd, ok := d.(StaticSemanticsIsConstantDeclaration); ok {
		return dd.IsConstantDeclaration()
	}
	return false
}

// StaticSemanticsLexicallyScopedDeclarations
// spec: 8.2.5
type StaticSemanticsLexicallyScopedDeclarations interface {
	LexicallyScopedDeclarations() []ASTNode
}
