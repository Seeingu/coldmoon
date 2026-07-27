package coldmoon

type PrivateName struct {
	Symbol *SymbolValue
}

func (p PrivateName) Equal(other PrivateName) bool {
	return p.Symbol.Id == other.Symbol.Id
}
