package coldmoon

// DefineBuiltinFunction installs a host-provided built-in function using the
// same descriptor attributes as engine intrinsics. This exported adapter is
// intentionally kept for host packages such as runtime; code inside coldmoon
// should call Object.defineBuiltinFunction directly.
func DefineBuiltinFunction(
	realm *Realm,
	name PropertyConvertable,
	object ObjectType,
	fn BehaviorFn,
	length JSInt,
) {
	object.Ref().defineBuiltinFunction(realm, name, fn, length)
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
