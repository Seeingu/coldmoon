package coldmoon

// 10.4.7.1 SetPrototypeOf
func ImmutableSetPrototypeOf(obj ObjectType, proto ObjectType) bool {
	return SetImmutablePrototype(obj, proto)
}

// 10.4.7.2
func SetImmutablePrototype(obj ObjectType, proto ObjectType) bool {
	current := obj.InternalMethods().GetPrototypeOf(obj)

	if ObjectSameValue(proto, current) {
		return true
	}

	return false
}
