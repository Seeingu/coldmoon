package coldmoon

type BooleanValue struct {
	Value
	Data bool
}

var _ Value = (*BooleanValue)(nil)

func (b *BooleanValue) String() string {
	if b.Data {
		return "true"
	} else {
		return "false"
	}
}

func (b *BooleanValue) ToBoolean() bool {
	return b.Data
}
