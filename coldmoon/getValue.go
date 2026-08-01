package coldmoon

// AOGetValue 6.2.5.5 is a type of abstract operation
type AOGetValue interface {
	// The agent supplies realm and exception state when resolving references.
	GetValue(agent *Agent) CompletionValue
}
