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
func NewSymbolPropertyKey(value *Symbol) SymbolPropertyKey {
	return SymbolPropertyKey{Value: value}
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
	Value *Symbol
}

func (s SymbolPropertyKey) Hash() string {
	return fmt.Sprintf("%d", s.Value.Id)
}
