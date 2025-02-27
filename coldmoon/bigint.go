package coldmoon

import (
	"math/big"

	"github.com/Seeingu/coldmoon/pkg"
)

type BigIntValue struct {
	Value
	Data *big.Int
}

var bigZero = big.NewInt(0)

var _ Value = (*BigIntValue)(nil)

func NewBigIntValue(v *big.Int) *BigIntValue {
	b := &BigIntValue{
		Data: v,
	}
	b.Value = NewBaseValue(b)
	return b
}

func NewBigIntFromBoolean(b bool) *BigIntValue {
	var i int64
	if b {
		i = 1
	}
	return NewBigIntValue(big.NewInt(i))
}

func (b *BigIntValue) ToString() CMString {
	return CMString(b.Data.String() + "n")
}

func (b *BigIntValue) String() string {
	return b.Data.String()
}

func (b *BigIntValue) Equal(other BigIntValue) bool {
	return b.Data.Cmp(other.Data) == 0
}

// 6.1.6.2.1
func (b *BigIntValue) UnaryMinus() *BigIntValue {
	return NewBigIntValue(new(big.Int).Neg(b.Data))
}

// 6.1.6.2.2
func (b *BigIntValue) BitwiseNot() *BigIntValue {
	return NewBigIntValue(new(big.Int).Not(b.Data))
}

// 6.1.6.2.3
func (b *BigIntValue) Exponentiate(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Exp(b.Data, other.Data, nil))
}

// 6.1.6.2.4
func (b *BigIntValue) Multiply(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Mul(b.Data, other.Data))
}

// 6.1.6.2.5
func (b *BigIntValue) Divide(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Div(b.Data, other.Data))
}

// 6.1.6.2.6
func (b *BigIntValue) Remainder(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Rem(b.Data, other.Data))
}

// 6.1.6.2.7
func (b *BigIntValue) Add(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Add(b.Data, other.Data))
}

// 6.1.6.2.8
func (b *BigIntValue) Subtract(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Sub(b.Data, other.Data))
}

// 6.1.6.2.9
func (b *BigIntValue) LeftShift(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Lsh(b.Data, uint(other.Data.Int64())))
}

// 6.1.6.2.10
func (b *BigIntValue) SignedRightShift(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Rsh(b.Data, uint(other.Data.Int64())))
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
	return NewBigIntValue(new(big.Int).And(b.Data, other.Data))
}

// 6.1.6.2.19
func (b *BigIntValue) BitwiseXor(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Xor(b.Data, other.Data))
}

// 6.1.6.2.20
func (b *BigIntValue) BitwiseOr(other *BigIntValue) *BigIntValue {
	return NewBigIntValue(new(big.Int).Or(b.Data, other.Data))
}

// MARK: - BigInt Object

type BigIntObject struct {
	*Object
	Data *BigIntValue
}

func NewBigIntObject(agent *Agent, v *BigIntValue, prototype ObjectType) *BigIntObject {
	object := &BigIntObject{
		Object: NewObject(agent, prototype, "BigInt"),
		Data:   v,
	}
	object.ref = object
	return object
}

func NewBigIntConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		value := argumentsList[0]

		if newTarget != nil {
			return agent.ThrowTypeError("BigInt is not a constructor.")
		}

		var co CompletionValue
		prim, isAbrupt, rt := ReturnIfAbrupt(value.ToPrimitive(agent, PreferredTypeNumber), co)
		if isAbrupt {
			return rt
		}

		if num, ok := prim.(*NumberValue); ok {
			return NumberToBigInt(agent, num)
		}

		v, isAbrupt, rt := ReturnIfAbrupt(ToBigInt(agent, prim), co)
		if isAbrupt {
			return rt
		}
		return v
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 1, CMString("BigInt"), builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	BindPrototypeAndConstructor(realm.Intrinsics.BigIntPrototype, object)

	return object
}

func NewBigIntPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "BigIntPrototype")

	// 21.2.3.3
	toString := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		radix := pkg.SliceSafeGet(arguments, 0)

		x := thisBigIntValue(this)

		var radixMV JSInt
		if radix == nil {
			radixMV = 10
		} else {
			radixMV = ToIntegerOrInfinity(agent, radix)
		}

		if radixMV < 2 || radixMV > 36 {
			return agent.ThrowRangeError("Radix must be an integer between 2 and 36, inclusive.")
		}

		return x.ToString().ToValue()
	}
	valueOf := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return thisBigIntValue(this)
	}
	toLocaleString := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return toString(thisBigIntValue(this), nil, nil)
	}

	object.defineBuiltinFunction(realm, CMString("toString"), toString, 0)
	object.defineBuiltinFunction(realm, CMString("valueOf"), valueOf, 0)
	object.defineBuiltinFunction(realm, CMString("toLocaleString"), toLocaleString, 0)

	object.defineToStringTag("BigInt")
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

	return NewBigIntValue(big.NewInt(int64(number.Data)))
}
