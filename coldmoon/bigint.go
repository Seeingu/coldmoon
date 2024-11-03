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
