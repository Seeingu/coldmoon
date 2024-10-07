package compiler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNameShadow(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")

	expected := Symbol{"a", FunctionScope, 0}
	result, ok := global.Resolve(expected.Name)
	assert.True(t, ok)
	assert.Equal(t, expected, result)
}

func TestResolveFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewEnclosingSymbolTable(global)
	firstLocal.Define("b")

	secondLocal := NewEnclosingSymbolTable(firstLocal)
	secondLocal.Define("c")

	tests := []struct {
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{"a", GlobalScope, 0},
				{"b", LocalScope, 0},
			},
			[]Symbol{},
		},
		{
			secondLocal,
			[]Symbol{
				{"a", GlobalScope, 0},
				{"b", FreeScope, 0},
				{"c", LocalScope, 0},
			},
			[]Symbol{
				{"b", LocalScope, 0},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, ok := tt.table.Resolve(sym.Name)
			assert.True(t, ok)
			assert.Equal(t, sym, result)
		}

		assert.Equal(t, len(tt.table.FreeSymbols), len(tt.expectedFreeSymbols))

		for i, sym := range tt.expectedFreeSymbols {
			result := tt.table.FreeSymbols[i]
			assert.Equal(t, sym, result)
		}
	}
}

func TestResolveBuiltin(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosingSymbolTable(global)

	expected := []Symbol{
		{"a", BuiltinScope, 0},
		{"b", BuiltinScope, 1},
	}

	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}
	for _, table := range []*SymbolTable{global, firstLocal} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)
			assert.True(t, ok)
			assert.Equal(t, sym, result)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosingSymbolTable(global)
	local.Define("c")
	local.Define("d")

	local2 := NewEnclosingSymbolTable(local)
	local2.Define("e")

	expected := []Symbol{
		{"a", GlobalScope, 0},
		{"b", GlobalScope, 1},
		{"c", FreeScope, 0},
		{"d", FreeScope, 1},
		{"e", LocalScope, 0},
	}
	for _, s := range expected {
		result, ok := local2.Resolve(s.Name)
		assert.True(t, ok)
		assert.Equal(t, s, result)
	}
}

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {"a", GlobalScope, 0},
		"b": {"b", GlobalScope, 1},
		"c": {"c", LocalScope, 0},
		"d": {"d", LocalScope, 1},
	}

	global := NewSymbolTable()
	a := global.Define("a")
	assert.Equal(t, expected["a"], a)
	b := global.Define("b")
	assert.Equal(t, expected["b"], b)

	local := NewEnclosingSymbolTable(global)
	c := local.Define("c")
	assert.Equal(t, expected["c"], c)
	d := local.Define("d")
	assert.Equal(t, expected["d"], d)
}

func TestResolve(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, symbol := range expected {
		result, ok := global.Resolve(symbol.Name)
		assert.True(t, ok)
		assert.Equal(t, symbol, result)
	}
}
