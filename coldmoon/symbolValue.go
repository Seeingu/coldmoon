package coldmoon

import "fmt"

type SymbolValue struct {
	Value
	Id             uint64
	Description    string
	HasDescription bool
	IsPrivate      bool
}

func (a *Agent) CreateSymbol(desc string) *SymbolValue {
	s := &SymbolValue{
		Id:             a.symbolId,
		Description:    desc,
		HasDescription: true,
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
func ThisSymbolValue(agent *Agent, v Value) (co Completion[*SymbolValue]) {
	if symbol, ok := v.(*SymbolValue); ok {
		co.value = symbol
		return
	}
	if object, ok := v.(*ObjectValue); ok {
		s, ok := object.Object.(*SymbolObject)
		if ok {
			co.value = s.Data
			return
		}
	}

	return co.ThrowTypeError(agent, "Not a Symbol")
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
