package coldmoon

type PropertyStorage struct {
	Properties map[PropertyKey]*PropertyDescriptor
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
