package coldmoon

import "math/big"

type BigInt struct {
	Value
	Data *big.Int
}

var _ Value = (*BigInt)(nil)

func NewBigIntValue(v *big.Int) *BigInt {
	return &BigInt{
		Data: v,
	}
}

func NewBigIntFromBoolean(b bool) *BigInt {
	var i int64
	if b {
		i = 1
	}
	return &BigInt{
		Data: big.NewInt(i),
	}
}

func (b *BigInt) String() string {
	return b.Data.String()
}

func (b *BigInt) ToBoolean() bool {
	if b.Data.Int64() == 0 {
		return false
	}
	return true
}

func (b *BigInt) Equal(other BigInt) bool {
	return b.Data.Cmp(other.Data) == 0
}
