package main

import (
	p "github.com/Seeingu/coldmoon/parser"
)

func main() {
	source := "print('hello world')"
	i := p.NewInterpreterWithSource(source)
	i.Eval()
}
