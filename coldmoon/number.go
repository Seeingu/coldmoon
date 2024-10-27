package coldmoon

import (
	"fmt"
	"math"
)

type NumberValue struct {
	Value
	Data float64
}

func NewNumberValue(v float64) *NumberValue {
	return &NumberValue{
		Data: v,
	}
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

func (n *NumberValue) IsPositiveZero() bool {
	return n.Data == 0 && math.Signbit(n.Data)
}

func (n *NumberValue) IsNegativeZero() bool {
	return n.Data == 0 && !math.Signbit(n.Data)
}

func (n *NumberValue) IsFinite() bool {
	return !math.IsInf(n.Data, 0)
}

func (n *NumberValue) Truncate() float64 {
	return n.Data
}

func (n *NumberValue) Round() float64 {
	return math.Round(n.Data)
}

func (n *NumberValue) Ceil() float64 {
	return math.Ceil(n.Data)
}

func (n *NumberValue) Floor() float64 {
	return math.Floor(n.Data)
}

// 6.1.6.1.14
func (n *NumberValue) SameValue(other NumberValue) bool {
	if n.IsNaN() && other.IsNaN() {
		return true
	}
	if n.IsPositiveZero() && other.IsNegativeZero() {
		return false
	}
	if n.IsNegativeZero() && other.IsPositiveZero() {
		return false
	}
	return n.Data == other.Data
}

// 6.1.6.1.15
func (n *NumberValue) SameValueZero(other NumberValue) bool {
	if n.IsNaN() && other.IsNaN() {
		return true
	}
	return n.Data == other.Data
}
