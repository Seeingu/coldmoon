package coldmoon

type Intrinsics struct {
	// %Object.Prototype%
	ObjectPrototype *ObjectPrototype
	// %Function.Prototype%
	FunctionPrototype ObjectType
	// %Boolean.Prototype%
	BooleanPrototype *BooleanPrototype
	// %Boolean%
	BooleanConstructor ObjectType
	// %ThrowTypeError%
	ThrowTypeError ObjectType
}

func (i *Intrinsics) Get(key string) interface{} {
	return nil
}
