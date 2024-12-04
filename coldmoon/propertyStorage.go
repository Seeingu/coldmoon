package coldmoon

import "strconv"

type PropertyStorage struct {
	Properties map[PropertyKey]*PropertyDescriptor
}

func (ps *PropertyStorage) Keys() []string {
	keys := make([]string, 0, len(ps.Properties))
	for key := range ps.Properties {
		switch k := key.(type) {
		case StringPropertyKey:
			keys = append(keys, k.Value)
		case SymbolPropertyKey:
			keys = append(keys, "Symbol: "+k.Value.Description)
		case IntegerIndexPropertyKey:
			keys = append(keys, "Index: "+strconv.Itoa(k.Value))
		}
	}
	return keys
}

func NewPropertyStorage() PropertyStorage {
	return PropertyStorage{
		Properties: make(map[PropertyKey]*PropertyDescriptor),
	}
}

func (ps *PropertyStorage) Get(key PropertyKey) *PropertyDescriptor {
	return ps.Properties[key]
}

func (ps *PropertyStorage) Set(key PropertyKey, value *PropertyDescriptor) {
	ps.Properties[key] = value
}

func (ps *PropertyStorage) Delete(key PropertyKey) {
	delete(ps.Properties, key)
}

func (ps *PropertyStorage) Has(key PropertyKey) bool {
	_, ok := ps.Properties[key]
	return ok
}
