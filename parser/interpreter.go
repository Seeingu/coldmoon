package parser

type Interpreter struct {
}

var environment *Environment

func eval() {
	environment = NewEnvironment()
}
