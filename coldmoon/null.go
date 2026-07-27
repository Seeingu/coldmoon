package coldmoon

type nullValue struct {
	Value
}

var _ Value = (*nullValue)(nil)

func (n *nullValue) String() string {
	return "null"
}

var NullValue = &nullValue{}
