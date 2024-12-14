package coldmoon

import "strings"

func DefineBuiltinFunction(object ObjectType,
	name string,
	fn BehaviorFn,
	length JSInt,
	realm *Realm,
) {
	functionName := name
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
	DefineBuiltinPropertyV(object, name, f.ToValue())
}

func DefineBuiltinFunctionWithAttributes(object ObjectType,
	name string,
	fn BehaviorFn,
	length JSInt,
	realm *Realm,
	attr PropertyDescriptorAttributes,
) {
	functionName := name
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
	DefineBuiltinPropertyP(object, name, &PropertyDescriptor{
		Value:        f.ToValue(),
		Writable:     attr.Writable,
		Configurable: attr.Configurable,
		Enumerable:   attr.Enumerable,
	})
}

func DefineBuiltinPropertyP(object ObjectType, name string, p *PropertyDescriptor) {
	object.DefinePropertyOrThrow(NewStringPropertyKey(name), p)
}

func DefineBuiltinPropertyV(object ObjectType, name string, value Value) {
	descriptor := &PropertyDescriptor{
		Value:        value,
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
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

func IsUndefinedOrNil(v Value) bool {
	return v == nil || v == UndefinedValue
}
