package coldmoon

type SymbolValue struct {
	Value
	Id          uint64
	Description string
}

var _ Value = (*SymbolValue)(nil)

func (s *SymbolValue) String() string {
	panic("TypeError")
}

func (s *SymbolValue) ToBoolean() bool {
	return true
}
