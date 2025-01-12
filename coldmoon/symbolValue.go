package coldmoon

import "fmt"

type SymbolValue struct {
	Value
	Id          uint64
	Description string
	IsPrivate   bool
}

func (a *Agent) CreateSymbol(desc string) *SymbolValue {
	s := &SymbolValue{
		Id:          a.symbolId,
		Description: desc,
	}
	a.symbolId += 1
	s.Value = NewBaseValue(s)
	return s
}

var _ Value = (*SymbolValue)(nil)

func (s *SymbolValue) String() string {
	return "Symbol: " + s.Description
}

// 20.4.3.3.1
func (s *SymbolValue) SymbolDescriptiveString() string {
	return fmt.Sprintf("Symbol(%s)", s.Description)
}

// 20.4.3.4.1
func ThisSymbolValue(v Value) *SymbolValue {
	if symbol, ok := v.(*SymbolValue); ok {
		return symbol
	}
	if object, ok := v.(*ObjectValue); ok {
		s, ok := object.Object.(*SymbolObject)
		if ok {
			return s.Data
		}
	}

	// TODO: throw TypeError by agent
	panic("TypeError")
}

// 20.4.5.1
func KeyForSymbol(agent *Agent, symbol *SymbolValue) string {
	for _, value := range agent.GlobalSymbolRegistry {
		if value.Id == symbol.Id {
			return value.Description
		}
	}
	return ""
}
