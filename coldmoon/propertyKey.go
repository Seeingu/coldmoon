package coldmoon

import (
	"fmt"
)

type PropertyKey interface {
	PropertyKeyOrPrivateName
	Hash() string
	ToValue() Value
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

func (s StringPropertyKey) ToValue() Value {
	return NewStringValue(s.Value)
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
func (s SymbolPropertyKey) ToValue() Value {
	return s.Value
}

type IntegerIndexPropertyKey struct {
	PropertyKey
	Value int
}

func (i IntegerIndexPropertyKey) Hash() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i IntegerIndexPropertyKey) ToValue() Value {
	return NewStringValue(fmt.Sprintf("%d", i.Value))
}
