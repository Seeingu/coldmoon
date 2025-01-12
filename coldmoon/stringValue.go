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

// TODO(WM): make it a method of BaseValue
// 22.1.3.35.1
func ThisStringValue(agent *Agent, v Value) string {
	switch v := v.(type) {
	case *StringValue:
		return v.Data
	case *ObjectValue:
		s, ok := v.Object.(*StringObject)
		if ok {
			return s.Data
		}

	}
	panic("TypeError")
}

// 22.1.3.32.1
func (s *StringValue) TrimString() string {
	return strings.TrimSpace(s.Data)
}
