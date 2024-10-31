package coldmoon

type ThrowTypeError struct {
	*Object
}

func NewThrowTypeError(realm *Realm) ObjectType {
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return NewValueFromObject(realm.Agent.ThrowException(TypeError, ""))
	}

	object := CreateBuiltinFunction(
		realm.Agent,
		behavior,
		0,
		"",
		builtinFunctionArgs{
			realm: realm,
		},
	)
	object.(*BuiltinFunction).SetExtensible(false)

	DefineBuiltinProperty(object, "length", &PropertyDescriptor{
		Value:        NewNumberValue(0),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinProperty(object, "name", &PropertyDescriptor{
		Value:        NewStringValue(""),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
