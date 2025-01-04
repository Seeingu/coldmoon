package coldmoon

type undefinedValue struct {
	Value
}

var _ Value = (*undefinedValue)(nil)

func (u *undefinedValue) String() string {
	return "undefined"
}

var UndefinedValue = &undefinedValue{}
