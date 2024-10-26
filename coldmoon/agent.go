package coldmoon

type Agent struct {
	symbolId uint64
}

var WellKnownSymbols = map[WellKnownSymbolsKey]Symbol{}

func NewAgent() *Agent {
	a := &Agent{}
	initWellKnownSymbols(a)
	return a
}

type WellKnownSymbolsKey string

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

func initWellKnownSymbols(agent *Agent) {
	WellKnownSymbols[WellKnownSymbolsAsyncIterator] = agent.CreateSymbol("Symbol.asyncIterator")
	WellKnownSymbols[WellKnownSymbolsHasInstance] = agent.CreateSymbol("Symbol.hasInstance")
	WellKnownSymbols[WellKnownSymbolsIsConcatSpreadable] = agent.CreateSymbol("Symbol.isConcatSpreadable")
	WellKnownSymbols[WellKnownSymbolsIterator] = agent.CreateSymbol("Symbol.iterator")
	WellKnownSymbols[WellKnownSymbolsMatch] = agent.CreateSymbol("Symbol.match")
	WellKnownSymbols[WellKnownSymbolsMatchAll] = agent.CreateSymbol("Symbol.matchAll")
	WellKnownSymbols[WellKnownSymbolsReplace] = agent.CreateSymbol("Symbol.replace")
	WellKnownSymbols[WellKnownSymbolsSearch] = agent.CreateSymbol("Symbol.search")
	WellKnownSymbols[WellKnownSymbolsSpecies] = agent.CreateSymbol("Symbol.species")
	WellKnownSymbols[WellKnownSymbolsSplit] = agent.CreateSymbol("Symbol.split")
	WellKnownSymbols[WellKnownSymbolsToPrimitive] = agent.CreateSymbol("Symbol.toPrimitive")
	WellKnownSymbols[WellKnownSymbolsToStringTag] = agent.CreateSymbol("Symbol.toStringTag")
	WellKnownSymbols[WellKnownSymbolsUnscopables] = agent.CreateSymbol("Symbol.unscopables")
}

func (a *Agent) CreateSymbol(desc string) Symbol {
	s := Symbol{
		Id:          a.symbolId,
		Description: desc,
	}
	a.symbolId += 1
	return s
}
