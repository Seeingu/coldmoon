package coldmoon

import "math/big"

type BigIntValue struct {
	Value
	Data *big.Int
}

var _ Value = (*BigIntValue)(nil)

func NewBigIntValue(v *big.Int) *BigIntValue {
	return &BigIntValue{
		Data: v,
	}
}

func NewBigIntFromBoolean(b bool) *BigIntValue {
	var i int64
	if b {
		i = 1
	}
	return &BigIntValue{
		Data: big.NewInt(i),
	}
}

func (b *BigIntValue) String() string {
	return b.Data.String()
}

func (b *BigIntValue) ToBoolean() bool {
	if b.Data.Int64() == 0 {
		return false
	}
	return true
}

func (b *BigIntValue) Equal(other BigIntValue) bool {
	return b.Data.Cmp(other.Data) == 0
}

// 6.1.6.2.1
func (b *BigIntValue) UnaryMinus() *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Neg(b.Data),
	}
}

// 6.1.6.2.2
func (b *BigIntValue) BitwiseNot() *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Not(b.Data),
	}
}

// 6.1.6.2.3
func (b *BigIntValue) Exponentiate(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Exp(b.Data, other.Data, nil),
	}
}

// 6.1.6.2.4
func (b *BigIntValue) Multiply(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Mul(b.Data, other.Data),
	}
}

// 6.1.6.2.5
func (b *BigIntValue) Divide(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Div(b.Data, other.Data),
	}
}

// 6.1.6.2.6
func (b *BigIntValue) Remainder(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Rem(b.Data, other.Data),
	}
}

// 6.1.6.2.7
func (b *BigIntValue) Add(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Add(b.Data, other.Data),
	}
}

// 6.1.6.2.8
func (b *BigIntValue) Subtract(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Sub(b.Data, other.Data),
	}
}

// 6.1.6.2.9
func (b *BigIntValue) LeftShift(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Lsh(b.Data, uint(other.Data.Int64())),
	}
}

// 6.1.6.2.10
func (b *BigIntValue) SignedRightShift(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Rsh(b.Data, uint(other.Data.Int64())),
	}
}

// 6.1.6.2.11
func (b *BigIntValue) UnsignedRightShift(other *BigIntValue) *BigIntValue {
	panic("TypeError")
}

// 6.1.6.2.12
func (b *BigIntValue) LessThan(other *BigIntValue) bool {
	return b.Data.Cmp(other.Data) == -1
}

// 6.1.6.2.18
func (b *BigIntValue) BitwiseAnd(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).And(b.Data, other.Data),
	}
}

// 6.1.6.2.19
func (b *BigIntValue) BitwiseXor(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Xor(b.Data, other.Data),
	}
}

// 6.1.6.2.20
func (b *BigIntValue) BitwiseOr(other *BigIntValue) *BigIntValue {
	return &BigIntValue{
		Data: new(big.Int).Or(b.Data, other.Data),
	}
}

// MARK: - BigInt Object

type BigIntObject struct {
	*Object
	Data *BigIntValue
}

func NewBigIntObject(agent *Agent, v *BigIntValue, prototype ObjectType) *BigIntObject {
	object := &BigIntObject{
		Object: NewObject(agent, prototype),
		Data:   v,
	}
	return object
}

func NewBigIntConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		value := argumentsList[0]

		if newTarget != nil {
			panic("TypeError")
		}

		prim := ToPrimitive(agent, value, PreferredTypeNumber)

		if num, ok := prim.(*NumberValue); ok {
			return NumberToBigInt(agent, num)
		}

		return ToBigInt(agent, prim)
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 1, "BigInt", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.BigIntPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinProperty(realm.Intrinsics.BigIntPrototype, "constructor",
		NewValueFromObject(object),
	)

	return object
}

func NewBigIntPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype)

	var toString = func(this Value, arguments []Value, newTarget ObjectType) Value {
		radix := arguments[0]

		x := thisBigIntValue(this)

		var radixMV float64
		if radix == nil {
			radixMV = 10
		} else {
			radixMV = ToIntegerOrInfinity(agent, radix)
		}

		if radixMV < 2 || radixMV > 36 {
			panic("RangeError")
		}

		return NewStringValue(x.String())
	}
	var valueOf = func(this Value, arguments []Value, newTarget ObjectType) Value {
		return thisBigIntValue(this)
	}
	var toLocaleString = func(this Value, arguments []Value, newTarget ObjectType) Value {
		return toString(thisBigIntValue(this), nil, nil)
	}

	DefineBuiltinFunction(object, "toString", toString, 0, realm)
	DefineBuiltinFunction(object, "valueOf", valueOf, 0, realm)
	DefineBuiltinFunction(object, "toLocaleString", toLocaleString, 0, realm)

	DefineBuiltinProperty(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("BigInt"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}

func thisBigIntValue(value Value) *BigIntValue {
	if bigint, ok := value.(*BigIntValue); ok {
		return bigint
	}
	if object, ok := value.(*ObjectValue); ok {
		bigInt, ok := object.Object.(*BigIntObject)
		if ok {
			return bigInt.Data
		}
	}

	panic("TypeError")
}

// 21.2.1.1.1
func NumberToBigInt(agent *Agent, number *NumberValue) *BigIntValue {
	if !IsIntegralNumber(number) {
		panic("RangeError")
	}

	return &BigIntValue{
		Data: big.NewInt(int64(number.Data)),
	}
}
