package coldmoon

type Intrinsics struct {
	// %Object.Prototype%
	ObjectPrototype *ObjectPrototype
	// %Object%
	ObjectConstructor ObjectType
	// %Function.Prototype%
	FunctionPrototype ObjectType
	// %Boolean.Prototype%
	BooleanPrototype *BooleanPrototype
	// %Boolean%
	BooleanConstructor ObjectType
	// %IsFinite%
	IsFinite ObjectType
	// %isNaN%
	IsNaN ObjectType
	// %eval%
	Eval ObjectType
	// %ThrowTypeError%
	ThrowTypeError ObjectType
}

func (i *Intrinsics) Get(key string) interface{} {
	return nil
}
