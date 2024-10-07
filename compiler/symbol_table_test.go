package compiler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {
			Name:  "a",
			Scope: GlobalScope,
			Index: 0,
		},
		"b": {
			Name:  "b",
			Scope: GlobalScope,
			Index: 1,
		},
	}

	global := NewSymbolTable()
	a := global.Define("a")
	assert.Equal(t, expected["a"], a)
	b := global.Define("b")
	assert.Equal(t, expected["b"], b)
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
