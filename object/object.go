package object

import "hash/fnv"

//go:generate stringer -type Type -trimprefix type
type Type int

const (
	TypeInt Type = iota
	TypeBool
	TypeString
	TypeArray
	TypeObject
)

type Object interface {
	toString()
	Type() Type
}

type Integer struct {
	Object
	Value int64
}

func (i Integer) Type() Type { return TypeInt }

type NumberObject struct {
	Object
	value int
}

type StringObject struct {
	Object
	Hashable
	Value string
}

func (s StringObject) Type() Type { return TypeString }
func (s StringObject) HashKey() HashKey {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type BooleanObject struct {
	Object
	Value bool
}

func (b BooleanObject) Type() Type { return TypeBool }

type ArrayObject struct {
	Object
	Elements []Object
}

func (a ArrayObject) Type() Type { return TypeArray }

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
	args      map[string]Object
}

type HashKey struct {
	Type  Type
	Value uint64
}
type Hashable interface {
	HashKey() HashKey
}

type HashPair struct {
	Key   Object
	Value Object
}

type ObjectObject struct {
	Object
	Pairs map[HashKey]HashPair
}

func (o ObjectObject) Type() Type { return TypeObject }

type NativeFunctionObject struct {
	Object
	name string
	fn   func(...Object)
}
