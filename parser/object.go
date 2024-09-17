package parser

type Object interface {
	toString()
}

type NumberObject struct {
	Object
	value int
}

type StringObject struct {
	Object
	value string
}

type BooleanObject struct {
	Object
	value bool
}

type NullObject struct {
	Object
}

type UndefinedObject struct {
	Object
}

type ObjectPrototype struct {
	Object
	pairs map[string]Object
}

type FunctionObject struct {
	Object
	Prototype ObjectPrototype
	pairs     map[string]Object
}
