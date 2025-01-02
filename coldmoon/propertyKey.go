package coldmoon

import (
	"fmt"
	"strconv"
)

type PropertyConvertable interface {
	ToPropertyKey() PropertyKey
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

// TODO: use as index key
func NewIntegerIndexPropertyKey(value JSInt) StringPropertyKey {
	return NewStringPropertyKey(fmt.Sprintf("%d", value))
}

type StringPropertyKey struct {
	PropertyKey
	Value string
}

var _ PropertyKey = StringPropertyKey{}

func (s StringPropertyKey) GetIndex() (JSInt, error) {
	if propertyKeyIndex, err := strconv.ParseFloat(s.Value, 64); err != nil {
		return JSInt(0), err
	} else {
		return JSInt(propertyKeyIndex), nil
	}
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
