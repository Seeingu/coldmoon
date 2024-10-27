package coldmoon

type Intrinsics struct {
	// %Object.Prototype%
	ObjectPrototype *ObjectPrototype
	// %Function.Prototype%
	FunctionPrototype ObjectType
}

func (i *Intrinsics) Get(key string) interface{} {
	return nil
}
