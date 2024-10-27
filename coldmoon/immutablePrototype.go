package coldmoon

// 10.4.7.1 SetPrototypeOf
func ImmutableSetPrototypeOf(obj ObjectType, proto *Object) bool {
	return SetImmutablePrototype(obj.(*Object), proto)
}

// 10.4.7.2
func SetImmutablePrototype(obj *Object, proto *Object) bool {
	current := obj.InternalMethods().GetPrototypeOf(obj)

	if ObjectSameValue(proto, current) {
		return true
	}

	return false
}
