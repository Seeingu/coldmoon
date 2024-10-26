package coldmoon

type Type int

const (
	Normal Type = iota
	Break
	Continue
	Return
	Throw
)

// CompletionRecord
// 6.2.4
type CompletionRecord struct {
	Type  Type
	Value Value
	// Target is a string or empty
	Target string
}

func NormalCompletion(value Value) *CompletionRecord {
	return &CompletionRecord{Type: Normal, Value: value}
}

func ThrowCompletion(value Value) *CompletionRecord {
	return &CompletionRecord{Type: Throw, Value: value}
}
