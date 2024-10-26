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
