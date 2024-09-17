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
