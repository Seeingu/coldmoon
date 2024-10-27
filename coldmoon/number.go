package coldmoon

import (
	"fmt"
	"math"
)

type NumberValue struct {
	Value
	Data float64
}

var _ Value = (*NumberValue)(nil)

func (n *NumberValue) String() string {
	return fmt.Sprintf("%f", n.Data)
}

func (n *NumberValue) ToBoolean() bool {
	if n.Data == 0 || math.IsNaN(n.Data) {
		return false
	}
	return true
}

func (n *NumberValue) IsNaN() bool {
	return math.IsNaN(n.Data)
}

func (n *NumberValue) IsPositiveInf() bool {
	return math.IsInf(n.Data, 1)
}

func (n *NumberValue) IsNegativeInf() bool {
	return math.IsInf(n.Data, -1)
}

func (n *NumberValue) IsFinite() bool {
	return !math.IsInf(n.Data, 0)
}

func (n *NumberValue) Truncate() float64 {
	return n.Data
}
