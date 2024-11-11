package coldmoon

type Intrinsics struct {
	// %Object.Prototype%
	ObjectPrototype *ObjectPrototype
	// %Object%
	ObjectConstructor ObjectType
	// %Function.Prototype%
	FunctionPrototype ObjectType
	// %Function%
	FunctionConstructor ObjectType
	// %Array%
	ArrayConstructor ObjectType
	// %Array.Prototype%
	ArrayPrototype *ArrayPrototype
	// %Boolean.Prototype%
	BooleanPrototype *BooleanPrototype
	// %Boolean%
	BooleanConstructor ObjectType
	// %String.Prototype%
	StringPrototype *StringObject
	// %String%
	StringConstructor ObjectType
	// %IsFinite%
	IsFinite ObjectType
	// %isNaN%
	IsNaN ObjectType
	// %eval%
	Eval ObjectType
	// %ThrowTypeError%
	ThrowTypeError ObjectType
}

func (i *Intrinsics) Get(key string) ObjectType {
	switch key {
	case "%Object.Prototype%":
		return i.ObjectPrototype
	case "%Object%":
		return i.ObjectConstructor
	case "%Function.Prototype%":
		return i.FunctionPrototype
	case "%Function%":
		return i.FunctionConstructor
	case "%Boolean.Prototype%":
		return i.BooleanPrototype
	case "%Boolean%":
		return i.BooleanConstructor
	case "%IsFinite%":
		return i.IsFinite
	case "%isNaN%":
		return i.IsNaN
	case "%eval%":
		return i.Eval
	case "%ThrowTypeError%":
		return i.ThrowTypeError
	}
	panic("unknown intrinsic")
}
