package coldmoon

import "math"

type JSInt int64

func (i JSInt) IsNegInf() bool {
	return i == JSInt(math.Inf(-1))
}

func (i JSInt) IsPositiveInf() bool {
	return i == JSInt(math.Inf(1))
}

func (i JSInt) IsInf() bool {
	return i == JSInt(math.Inf(0))
}

func (i JSInt) Max(b JSInt) JSInt {
	return JSInt(math.Max(float64(i), float64(b)))
}

func (i JSInt) Min(b JSInt) JSInt {
	return JSInt(math.Min(float64(i), float64(b)))
}

func (i JSInt) ToNumber() JSNumber {
	return JSNumber(i)
}

func (i JSInt) ToValue() Value {
	return NewNumberValue(i.ToNumber())
}
