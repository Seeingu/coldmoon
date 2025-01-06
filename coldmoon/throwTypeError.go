package coldmoon

type ThrowTypeError struct {
	*Object
}

func NewThrowTypeError(realm *Realm) ObjectType {
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return realm.Agent.ThrowException(TypeError, "")
	}

	object := CreateBuiltinFunction(
		realm.Agent,
		behavior,
		0,
		CMString(""),
		builtinFunctionArgs{
			realm: realm,
		},
	)
	object.(*BuiltinFunction).SetExtensible(false)

	object.defineBuiltinProperty(CMString("length"), &PropertyDescriptor{
		Value:        NewNumberValue(0),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	object.defineBuiltinProperty(CMString("name"), &PropertyDescriptor{
		Value:        NewStringValue(""),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
