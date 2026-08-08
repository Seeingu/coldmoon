package coldmoon

import (
	"fmt"
)

type PropertyConvertable interface {
	ToPropertyKey() PropertyKey
	ToName() string
}

type PropertyKey interface {
	PropertyKeyOrPrivateName
	Hash() string
	ToValue() Value
	ToReference() *ReferencedName
	GetIndex() (JSInt, error)
}

func NewStringPropertyKey(value string) StringPropertyKey {
	return StringPropertyKey{Value: value}
}

func NewSymbolPropertyKey(value *SymbolValue) SymbolPropertyKey {
	return SymbolPropertyKey{Value: value}
}

// ECMAScript index positions are represented by String property keys.
func NewIntegerIndexPropertyKey(value JSInt) StringPropertyKey {
	return NewStringPropertyKey(fmt.Sprintf("%d", value))
}

type StringPropertyKey struct {
	PropertyKey
	Value string
}

var _ PropertyKey = StringPropertyKey{}

func (s StringPropertyKey) GetIndex() (JSInt, error) {
	// Only canonical array indices (a decimal string of an integer in
	// [0, 2^32-2] with no leading zeros) index an array. ParseFloat accepted
	// "0.5", "1e2", "01", and "Infinity", corrupting Array length.
	// propertyKeyArrayIndex performs the canonical check; reuse it.
	index, ok := propertyKeyArrayIndex(s)
	if !ok {
		return JSInt(0), fmt.Errorf("StringPropertyKey %q is not an array index", s.Value)
	}
	return JSInt(index), nil
}

func (s StringPropertyKey) ToValue() Value {
	return NewStringValue(s.Value)
}

func (s StringPropertyKey) Hash() string {
	return s.Value
}

func (s StringPropertyKey) ToReference() *ReferencedName {
	return &ReferencedName{String: s.Value}
}

type SymbolPropertyKey struct {
	PropertyKey
	Value *SymbolValue
}

var _ PropertyKey = SymbolPropertyKey{}

func (s SymbolPropertyKey) GetIndex() (JSInt, error) {
	return 0, fmt.Errorf("SymbolPropertyKey cannot be used as index")
}

func (s SymbolPropertyKey) Hash() string {
	return fmt.Sprintf("%d", s.Value.Id)
}

func (s SymbolPropertyKey) ToValue() Value {
	return s.Value
}

func (s SymbolPropertyKey) ToReference() *ReferencedName {
	return &ReferencedName{Symbol: s.Value}
}

type IntegerIndexPropertyKey struct {
	PropertyKey
	Value JSInt
}

func (i IntegerIndexPropertyKey) GetIndex() (JSInt, error) {
	return i.Value, nil
}

func (i IntegerIndexPropertyKey) Hash() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i IntegerIndexPropertyKey) ToValue() Value {
	return NewStringValue(fmt.Sprintf("%d", i.Value))
}

func (i IntegerIndexPropertyKey) ToReference() *ReferencedName {
	return &ReferencedName{String: fmt.Sprintf("%d", i.Value)}
}
