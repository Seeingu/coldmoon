package coldmoon

import (
	"math"
	"math/big"
	"testing"
)

func TestNumericToRawBytesUsesIEEE754Encoding(t *testing.T) {
	float32Values := []float64{1.5, math.Copysign(0, -1), math.Inf(1), math.NaN()}
	for _, littleEndian := range []bool{false, true} {
		for _, value := range float32Values {
			bytes := NumericToRawBytes(NewNumberValue(JSNumber(value)), TypedArrayNameFloat32, littleEndian)
			got := uint32(RawBytesToNumeric(4, bytes, littleEndian))
			want := math.Float32bits(float32(value))
			if got != want {
				t.Fatalf("Float32(%v), littleEndian=%v: bits = %#x, want %#x", value, littleEndian, got, want)
			}
		}

		float64Values := []float64{-13.25, math.Copysign(0, -1), math.Inf(1), math.NaN()}
		for _, value := range float64Values {
			bytes := NumericToRawBytes(NewNumberValue(JSNumber(value)), TypedArrayNameFloat64, littleEndian)
			got := RawBytesToNumeric(8, bytes, littleEndian)
			want := math.Float64bits(value)
			if got != want {
				t.Fatalf("Float64(%v), littleEndian=%v: bits = %#x, want %#x", value, littleEndian, got, want)
			}
		}
	}
}

func TestRawUint64ToBytesNeverRoutesThroughFloat64(t *testing.T) {
	values := []uint64{
		1<<53 + 1,
		math.MaxInt64,
		math.MaxInt64 + 1,
		math.MaxUint64,
	}
	for _, littleEndian := range []bool{false, true} {
		for _, value := range values {
			bytes := rawUint64ToBytes(value, 8, littleEndian)
			if got := RawBytesToNumeric(8, bytes, littleEndian); got != value {
				t.Fatalf("rawUint64ToBytes(%#x), littleEndian=%v = %#x", value, littleEndian, got)
			}
		}
	}
}

func TestNumericToRawBytesPreservesBigIntLow64Bits(t *testing.T) {
	values := []*big.Int{
		new(big.Int).SetUint64(1<<53 + 1),
		new(big.Int).SetUint64(math.MaxInt64),
		new(big.Int).SetUint64(math.MaxUint64),
		big.NewInt(-1),
	}
	for _, elementType := range []TypedArrayName{TypedArrayNameBigInt64, TypedArrayNameBigUint64} {
		for _, littleEndian := range []bool{false, true} {
			for _, value := range values {
				bytes := NumericToRawBytes(NewBigIntValue(value), elementType, littleEndian)
				got := RawBytesToNumeric(8, bytes, littleEndian)
				modulus := new(big.Int).Lsh(big.NewInt(1), 64)
				want := new(big.Int).Mod(new(big.Int).Set(value), modulus).Uint64()
				if got != want {
					t.Fatalf("NumericToRawBytes(%v, %v), littleEndian=%v = %#x, want %#x", value, elementType, littleEndian, got, want)
				}
			}
		}
	}
}
