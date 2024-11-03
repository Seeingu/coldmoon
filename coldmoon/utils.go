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
	switch v := value.(type) {
	case Value:
		object.ToObject().CreateNonEnumerableDataProperty(NewStringPropertyKey(name), v)
	case *PropertyDescriptor:
		object.DefinePropertyOrThrow(NewStringPropertyKey(name), v)
	default:
		panic("invalid value")
	}
}
