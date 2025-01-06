package coldmoon

var WellKnownSymbols = map[WellKnownSymbolsKey]*SymbolValue{}

func initWellKnownSymbols(agent *Agent) {
	WellKnownSymbols[WellKnownSymbolsAsyncIterator] = agent.CreateSymbol("SymbolValue.asyncIterator")
	WellKnownSymbols[WellKnownSymbolsHasInstance] = agent.CreateSymbol("SymbolValue.hasInstance")
	WellKnownSymbols[WellKnownSymbolsIsConcatSpreadable] = agent.CreateSymbol("SymbolValue.isConcatSpreadable")
	WellKnownSymbols[WellKnownSymbolsIterator] = agent.CreateSymbol("SymbolValue.iterator")
	WellKnownSymbols[WellKnownSymbolsMatch] = agent.CreateSymbol("SymbolValue.match")
	WellKnownSymbols[WellKnownSymbolsMatchAll] = agent.CreateSymbol("SymbolValue.matchAll")
	WellKnownSymbols[WellKnownSymbolsReplace] = agent.CreateSymbol("SymbolValue.replace")
	WellKnownSymbols[WellKnownSymbolsSearch] = agent.CreateSymbol("SymbolValue.search")
	WellKnownSymbols[WellKnownSymbolsSpecies] = agent.CreateSymbol("SymbolValue.species")
	WellKnownSymbols[WellKnownSymbolsSplit] = agent.CreateSymbol("SymbolValue.split")
	WellKnownSymbols[WellKnownSymbolsToPrimitive] = agent.CreateSymbol("SymbolValue.toPrimitive")
	WellKnownSymbols[WellKnownSymbolsToStringTag] = agent.CreateSymbol("SymbolValue.toStringTag")
	WellKnownSymbols[WellKnownSymbolsUnscopables] = agent.CreateSymbol("SymbolValue.unscopables")
}

type WellKnownSymbolsKey string

func (w WellKnownSymbolsKey) Nil() bool {
	return w == ""
}

func (w WellKnownSymbolsKey) ToPropertyKey() PropertyKey {
	return NewSymbolPropertyKey(WellKnownSymbols[w])
}

func (w WellKnownSymbolsKey) ToName() string {
	switch w {
	case WellKnownSymbolsSpecies:
		return "[Symbol.species]"
	case WellKnownSymbolsToStringTag:
		return "[Symbol.toStringTag]"
	}
	// TODO: manage all string tags in one place
	panic("unimplemented")
}

const (
	WellKnownSymbolsAsyncIterator      WellKnownSymbolsKey = "@@asyncIterator"
	WellKnownSymbolsHasInstance        WellKnownSymbolsKey = "@@hasInstance"
	WellKnownSymbolsIsConcatSpreadable WellKnownSymbolsKey = "@@isConcatSpreadable"
	WellKnownSymbolsIterator           WellKnownSymbolsKey = "@@iterator"
	WellKnownSymbolsMatch              WellKnownSymbolsKey = "@@match"
	WellKnownSymbolsMatchAll           WellKnownSymbolsKey = "@@matchAll"
	WellKnownSymbolsReplace            WellKnownSymbolsKey = "@@replace"
	WellKnownSymbolsSearch             WellKnownSymbolsKey = "@@search"
	WellKnownSymbolsSpecies            WellKnownSymbolsKey = "@@species"
	WellKnownSymbolsSplit              WellKnownSymbolsKey = "@@split"
	WellKnownSymbolsToPrimitive        WellKnownSymbolsKey = "@@toPrimitive"
	WellKnownSymbolsToStringTag        WellKnownSymbolsKey = "@@toStringTag"
	WellKnownSymbolsUnscopables        WellKnownSymbolsKey = "@@unscopables"
)
