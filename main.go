package main

import (
	"github.com/Seeingu/coldmoon/compiler"
	"github.com/Seeingu/coldmoon/lexer"
	"github.com/Seeingu/coldmoon/parser"
	VM "github.com/Seeingu/coldmoon/vm"
)

func main() {
	input := `
	function assert(a) {
		print("assert");
	}
	assert.v = function() {
		print("hello world");
	}
	assert('1')
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	parser.PrintProgram(program)

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		panic(err)
	}
	vm := VM.New(comp.Bytecode())
	comp.PrintBytecode()
	err = vm.Run()
	if err != nil {
		panic(err)
	}
}
