package coldmoon

import "strings"

func DefineBuiltinFunction(object ObjectType,
	name string,
	fn BehaviorFn,
	length float64,
	realm *Realm,
) {
	var functionName = name
	if strings.HasPrefix(name, "@@") {
		functionName = name[2:]
	}
	f := CreateBuiltinFunction(
		realm.Agent,
		fn,
		length,
		functionName,
		builtinFunctionArgs{realm: realm},
	)
	DefineBuiltinProperty(object, name, NewValueFromObject(f))
}

func DefineBuiltinFunctionWithAttributes(object ObjectType,
	name string,
	fn BehaviorFn,
	length float64,
	realm *Realm,
	attr PropertyDescriptorAttributes,
) {
	var functionName = name
	if strings.HasPrefix(name, "@@") {
		functionName = name[2:]
	}
	f := CreateBuiltinFunction(
		realm.Agent,
		fn,
		length,
		functionName,
		builtinFunctionArgs{realm: realm},
	)
	DefineBuiltinProperty(object, name, &PropertyDescriptor{
		Value:        NewValueFromObject(f),
		Writable:     attr.Writable,
		Configurable: attr.Configurable,
		Enumerable:   attr.Enumerable,
	})
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

func DefineBuiltinAccessor(realm *Realm, object ObjectType, name string, getter, setter BehaviorFn) {
	var get ObjectType
	if getter != nil {
		funName := "get " + name
		get = CreateBuiltinFunction(realm.Agent, getter, 0, funName, builtinFunctionArgs{realm: realm})
	}
	var set ObjectType
	if setter != nil {
		funName := "set " + name
		set = CreateBuiltinFunction(realm.Agent, setter, 1, funName, builtinFunctionArgs{realm: realm})
	}
	object.DefinePropertyOrThrow(NewStringPropertyKey(name), &PropertyDescriptor{
		Get:          get,
		Set:          set,
		Enumerable:   false,
		Configurable: true,
	})
}
