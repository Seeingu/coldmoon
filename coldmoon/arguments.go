package coldmoon

type ArgumentsObject struct {
	*Object
}

// 10.4.4.6
func CreateUnmappedArgumentsObject(agent *Agent, argumentsList []Value) ObjectType {
	realm := agent.CurrentRealm()

	length := len(argumentsList)

	obj := &ArgumentsObject{
		Object: NewObject(agent, realm.Intrinsics.ObjectPrototype),
	}

	obj.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(float64(length)),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})

	for index, value := range argumentsList {
		pk := NewIntegerIndexPropertyKey(index)
		obj.CreateDataPropertyOrThrow(pk, value)
	}

	obj.DefinePropertyOrThrow(NewStringPropertyKey("callee"), &PropertyDescriptor{
		Get:          realm.Intrinsics.ThrowTypeError,
		Set:          realm.Intrinsics.ThrowTypeError,
		Enumerable:   false,
		Configurable: false,
	})

	return obj
}

// TODO
func CreateMappedArgumentsObject(agent *Agent, formals *FormalParameters, argumentsList []Value, env EnvironmentRecord) ObjectType {
	panic("implement me")
}
