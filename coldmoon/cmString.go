package coldmoon

// CMString is an internal string type that can be converted
// - ToPropertyKey
// - ToValue
type CMString string

func (p CMString) ToPropertyKey() PropertyKey {
	return NewStringPropertyKey(string(p))
}

func (p CMString) ToValue() Value {
	return NewStringValue(string(p))
}
