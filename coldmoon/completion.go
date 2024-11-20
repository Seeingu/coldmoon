package coldmoon

type CompletionType int

const (
	CompletionTypeNormal CompletionType = iota
	CompletionTypeBreak
	CompletionTypeContinue
	CompletionTypeReturn
	CompletionTypeThrow
)

// CompletionRecord
// 6.2.4
type CompletionRecord struct {
	Type  CompletionType
	Value Value
	// Target is a string or empty
	Target string
}

func NewNormalCompletion(value Value) *CompletionRecord {
	return &CompletionRecord{Type: CompletionTypeNormal, Value: value}
}

func NewThrowCompletion(value Value) *CompletionRecord {
	return &CompletionRecord{Type: CompletionTypeThrow, Value: value}
}

var TypeErrorCompletion = NewThrowCompletion(NewStringValue("TypeError"))
