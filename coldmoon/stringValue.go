package coldmoon

import "strings"

type StringValue struct {
	Value
	Data string
}

var _ Value = (*StringValue)(nil)

func (s *StringValue) String() string {
	return s.Data
}

func NewStringValue(value string) *StringValue {
	s := &StringValue{Data: value}
	s.Value = NewBaseValue(s)
	return s
}

// 22.1.3.32.1
func (s *StringValue) TrimString() string {
	return strings.TrimSpace(s.Data)
}
