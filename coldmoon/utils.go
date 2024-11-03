package coldmoon

func DefineBuiltinFunction(object ObjectType,
	name string,
	fn BehaviorFn,
	length float64,
	realm *Realm,
) {
	f := CreateBuiltinFunction(
		realm.Agent,
		fn,
		length,
		name,
		builtinFunctionArgs{realm: realm},
	)
	DefineBuiltinProperty(object, name, NewValueFromObject(f))
}

func DefineBuiltinProperty(object ObjectType, name string, value interface{}) {
	var descriptor *PropertyDescriptor
	switch v := value.(type) {
	case Value:
		descriptor = &PropertyDescriptor{
			Value:        v,
			Writable:     true,
			Enumerable:   false,
			Configurable: true,
		}
	case *PropertyDescriptor:
		descriptor = v
	default:
		panic("invalid value")
	}
	object.DefinePropertyOrThrow(NewStringPropertyKey(name), descriptor)
}
