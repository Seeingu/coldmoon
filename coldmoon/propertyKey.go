package coldmoon

import (
	"fmt"
)

type PropertyKey interface {
	Hash() string
}

func NewStringPropertyKey(value string) StringPropertyKey {
	return StringPropertyKey{Value: value}
}
func NewSymbolPropertyKey(value *SymbolValue) SymbolPropertyKey {
	return SymbolPropertyKey{Value: value}
}
func NewIntegerIndexPropertyKey(value int) IntegerIndexPropertyKey {
	return IntegerIndexPropertyKey{Value: value}
}

type StringPropertyKey struct {
	PropertyKey
	Value string
}

func (s StringPropertyKey) Hash() string {
	return s.Value
}

type SymbolPropertyKey struct {
	PropertyKey
	Value *SymbolValue
}

func (s SymbolPropertyKey) Hash() string {
	return fmt.Sprintf("%d", s.Value.Id)
}

type IntegerIndexPropertyKey struct {
	PropertyKey
	Value int
}

func (i IntegerIndexPropertyKey) Hash() string {
	return fmt.Sprintf("%d", i.Value)
}
