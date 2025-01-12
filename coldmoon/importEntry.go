package coldmoon

type ImportName string

const (
	ImportNameNamespaceObject ImportName = "NAMESPACE-OBJECT"
	// ImportNameAll is `export * as ns from "mod"`
	ImportNameAll ImportName = "ALL"
	// ImportNameAllButDefault is `export * from "mod"`
	ImportNameAllButDefault ImportName = "ALL-BUT-DEFAULT"
)

type ImportEntryRecord struct {
	// [[ModuleRequest]]
	ModuleRequest string
	// [[ImportName]]
	ImportName ImportName
	// [[LocalName]]
	LocalName string
}
