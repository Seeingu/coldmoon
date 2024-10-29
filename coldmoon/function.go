package coldmoon

type FunctionPrototype struct {
	*Object
}

func (f *FunctionPrototype) ToObject() *Object {
	return f.Object
}

func NewFunctionPrototype(realm *Realm) ObjectType {
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return UndefinedValue
	}
	f := CreateBuiltinFunction(realm.Agent, behavior, 0, "", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.ObjectPrototype,
		prefix:        "",
		isConstructor: false,
	})
	return f
}
