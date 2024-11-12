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
