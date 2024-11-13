package coldmoon

type Intrinsics struct {
	// %Object.Prototype%
	ObjectPrototype ObjectType
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
	BooleanPrototype *BooleanObject
	// %Boolean%
	BooleanConstructor ObjectType
	// %String.Prototype%
	StringPrototype *StringObject
	// %String%
	StringConstructor ObjectType
	// %Number.Prototype%
	NumberPrototype *NumberObject
	// %Number%
	NumberConstructor ObjectType
	// &Symbol.Prototype%
	SymbolPrototype ObjectType
	// %Symbol%
	SymbolConstructor ObjectType
	// %BigInt.Prototype%
	BigIntPrototype ObjectType
	// %BigInt%
	BigIntConstructor ObjectType
	// %IsFinite%
	IsFinite ObjectType
	// %isNaN%
	IsNaN ObjectType
	// %eval%
	Eval ObjectType
	// %Error%
	ErrorConstructor ObjectType
	// %Error.prototype%
	ErrorPrototype ObjectType
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
