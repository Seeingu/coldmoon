package coldmoon

type Symbol struct {
	Value
	Id          uint64
	Description string
}

var _ Value = (*Symbol)(nil)

func (s *Symbol) String() string {
	panic("TypeError")
}

func (s *Symbol) ToBoolean() bool {
	return true
}
