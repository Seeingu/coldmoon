package compiler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
		{"c", LocalScope, 0},
		{"d", LocalScope, 1},
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
