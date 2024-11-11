package coldmoon

import "fmt"

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

// 20.4.3.3.1
func (s *SymbolValue) SymbolDescriptiveString() string {
	return fmt.Sprintf("Symbol(%s)", s.Description)
}
