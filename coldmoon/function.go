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

// 20.2.3.5
func (f *FunctionPrototype) ToString(thisValue Value) Value {
	f.Object.Agent()
	fun := thisValue
	o, ok := fun.(*ObjectValue)
	if ok {
		ecmascriptFunction, ok := o.Object.(*ECMAScriptFunction)
		if ok {
			return NewStringValue(ecmascriptFunction.SourceText)
		}

		builtinFunction, ok := o.Object.(*BuiltinFunction)
		if ok {
			name := builtinFunction.InitialName
			sourceText := "function " + name + "() { [native code] }"
			return NewStringValue(sourceText)
		}
	}
	if IsCallable(fun) {
		return NewStringValue("function () { [native code] }")
	}

	panic("TypeError")
}
