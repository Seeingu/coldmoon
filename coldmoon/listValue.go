package coldmoon

// ListValue is a internal value wrapper for list
type ListValue struct {
	Value
	Values []Value
}

func NewListValue(values []Value) *ListValue {
	l := &ListValue{
		Values: values,
	}
	l.Value = NewBaseValue(l)
	return l
}

func (l *ListValue) String() string {
	sb := "["
	for i, v := range l.Values {
		if i > 0 {
			sb += ", "
		}
		sb += v.String()
	}
	sb += "]"
	return sb
}
