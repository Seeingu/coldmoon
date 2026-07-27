package coldmoon

import (
	"math"
	"sort"
	"strconv"
)

type PropertyStorage struct {
	properties map[PropertyKey]*PropertyDescriptor
	order      []PropertyKey
}

func NewPropertyStorage() PropertyStorage {
	return PropertyStorage{
		properties: make(map[PropertyKey]*PropertyDescriptor),
	}
}

func (ps *PropertyStorage) Get(key PropertyKey) *PropertyDescriptor {
	return ps.properties[key]
}

func (ps *PropertyStorage) Set(key PropertyKey, value *PropertyDescriptor) {
	if _, exists := ps.properties[key]; !exists {
		ps.order = append(ps.order, key)
	}
	ps.properties[key] = value
}

func (ps *PropertyStorage) Delete(key PropertyKey) {
	if _, exists := ps.properties[key]; !exists {
		return
	}
	delete(ps.properties, key)
	for index, existing := range ps.order {
		if existing == key {
			ps.order = append(ps.order[:index], ps.order[index+1:]...)
			return
		}
	}
}

func (ps *PropertyStorage) Has(key PropertyKey) bool {
	_, ok := ps.properties[key]
	return ok
}

func (ps *PropertyStorage) Len() int {
	return len(ps.properties)
}

func (ps *PropertyStorage) Descriptors() []*PropertyDescriptor {
	descriptors := make([]*PropertyDescriptor, 0, len(ps.order))
	for _, key := range ps.order {
		if descriptor, ok := ps.properties[key]; ok {
			descriptors = append(descriptors, descriptor)
		}
	}
	return descriptors
}

// OrderedKeys implements the OrdinaryOwnPropertyKeys ordering contract:
// integer indices ascending, then strings by insertion, then Symbols by
// insertion.
func (ps *PropertyStorage) OrderedKeys() []PropertyKey {
	var indices []PropertyKey
	var stringsInOrder []PropertyKey
	var symbolsInOrder []PropertyKey
	for _, key := range ps.order {
		if _, exists := ps.properties[key]; !exists {
			continue
		}
		if _, isIndex := propertyKeyArrayIndex(key); isIndex {
			indices = append(indices, key)
			continue
		}
		switch key.(type) {
		case StringPropertyKey:
			stringsInOrder = append(stringsInOrder, key)
		case SymbolPropertyKey:
			symbolsInOrder = append(symbolsInOrder, key)
		default:
			stringsInOrder = append(stringsInOrder, key)
		}
	}
	sort.SliceStable(indices, func(left, right int) bool {
		leftIndex, _ := propertyKeyArrayIndex(indices[left])
		rightIndex, _ := propertyKeyArrayIndex(indices[right])
		return leftIndex < rightIndex
	})

	keys := make([]PropertyKey, 0, len(ps.properties))
	keys = append(keys, indices...)
	keys = append(keys, stringsInOrder...)
	keys = append(keys, symbolsInOrder...)
	return keys
}

func propertyKeyArrayIndex(key PropertyKey) (uint32, bool) {
	switch value := key.(type) {
	case IntegerIndexPropertyKey:
		if value.Value < 0 || uint64(value.Value) >= math.MaxUint32 {
			return 0, false
		}
		return uint32(value.Value), true
	case StringPropertyKey:
		index, err := strconv.ParseUint(value.Value, 10, 32)
		if err != nil || index >= math.MaxUint32 || strconv.FormatUint(index, 10) != value.Value {
			return 0, false
		}
		return uint32(index), true
	default:
		return 0, false
	}
}
