package main

import (
	p "github.com/Seeingu/coldmoon/parser"
)

// primitives1: number, string, boolean, null, undefined
// variable: var, let, const
// condition: if, else, return
// loop: for, while
// object
// function1: const
func main() {
	source := ""
	scanner := p.NewScanner(source)
	parser := p.NewParser(&scanner)
	parser.Parse()
}
