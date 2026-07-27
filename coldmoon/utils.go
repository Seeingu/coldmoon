package coldmoon

// TODO(BM): use object.defineBuiltinFunction
// Deprecated
func DefineBuiltinFunction(
	realm *Realm,
	name PropertyConvertable,
	object ObjectType,
	fn BehaviorFn,
	length JSInt,
) {
	functionName := name
	f := CreateBuiltinFunction(
		realm.Agent,
		fn,
		length,
		functionName,
		builtinFunctionArgs{realm: realm},
	)
	object.defineBuiltinProperty(name, &PropertyDescriptor{
		Value:        f.ToValue(),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
}

func IsUndefinedOrNil(v Value) bool {
	return v == nil || v == UndefinedValue
}

func IsUndefinedOrNull(v Value) bool {
	return IsUndefinedOrNil(v) || v == NullValue
}

// BindPrototypeAndConstructor set proto as the prototype of constructor
// and constructor as the constructor of proto
func BindPrototypeAndConstructor(proto ObjectType, constructor ObjectType) {
	constructor.defineBuiltinProperty(CMString("prototype"),
		&PropertyDescriptor{
			Value:        proto.ToValue(),
			Writable:     false,
			Enumerable:   false,
			Configurable: false,
		})
	proto.defineBuiltinProperty(CMString("constructor"), constructor.ToValue().ToBuiltinPropertyDescriptor())
}

// InitializeConstants initializes static constants
func InitializeConstants() {
	FalseValue.Value = NewBaseValue(FalseValue)
	TrueValue.Value = NewBaseValue(TrueValue)
	UndefinedValue.Value = NewBaseValue(UndefinedValue)
	NullValue.Value = NewBaseValue(NullValue)
}
