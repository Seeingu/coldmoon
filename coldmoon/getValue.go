package coldmoon

// AOGetValue 6.2.5.5 is a type of abstract operation
type AOGetValue interface {
	// TODO: agent is not necessary
	// TODO(BM): GetValue return completion
	GetValue(agent *Agent) Value
}
