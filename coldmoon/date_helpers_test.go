package coldmoon

import (
	"math"
	"testing"
)

// TestMakeFullYearTruncatesBeforeApplyingLegacyOffset verifies the Date
// constructor's special handling for finite years from 0 through 99.
func TestMakeFullYearTruncatesBeforeApplyingLegacyOffset(t *testing.T) {
	tests := []struct {
		name string
		year JSNumber
		want JSNumber
	}{
		{name: "positive fractional legacy year", year: 1.9, want: 1901},
		{name: "upper legacy boundary", year: 99.9, want: 1999},
		{name: "outside legacy range", year: 100.9, want: 100},
		{name: "negative fractional year", year: -1.9, want: -1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := MakeFullYear(test.year); got != test.want {
				t.Fatalf("MakeFullYear(%v) = %v, want %v", test.year, got, test.want)
			}
		})
	}
	if got := MakeFullYear(JSNumber(math.NaN())); !got.IsNaN() {
		t.Fatalf("MakeFullYear(NaN) = %v, want NaN", got)
	}
}
