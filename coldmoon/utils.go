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
		builtinFunctionArgs{
			realm: realm,
		},
	)
	object.ToObject().CreateNonEnumerableDataProperty(NewStringPropertyKey(name), NewValueFromObject(f))
}

func DefineBuiltinPropertyValue(object ObjectType, name string, value Value) {
	object.ToObject().CreateNonEnumerableDataProperty(NewStringPropertyKey(name), value)
}

func DefineBuiltinPropertyDescriptor(object ObjectType, name string, value *PropertyDescriptor) {
	object.DefinePropertyOrThrow(NewStringPropertyKey(name), value)
}
