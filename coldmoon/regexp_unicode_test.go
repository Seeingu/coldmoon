package coldmoon

import "testing"

func TestGetStringIndexUsesUTF16CodeUnits(t *testing.T) {
	t.Parallel()

	const input = "a😀b"
	tests := []struct {
		codePointIndex JSInt
		want           JSInt
	}{
		{codePointIndex: 0, want: 0},
		{codePointIndex: 1, want: 1},
		{codePointIndex: 2, want: 3},
		{codePointIndex: 3, want: 4},
		{codePointIndex: 4, want: 4},
	}
	for _, test := range tests {
		if got := GetStringIndex(input, test.codePointIndex); got != test.want {
			t.Fatalf("GetStringIndex(%q, %d) = %d, want %d", input, test.codePointIndex, got, test.want)
		}
	}
}

func TestAdvanceStringIndexUsesCodePointWidthInUnicodeMode(t *testing.T) {
	t.Parallel()

	const input = "😀a"
	if got := AdvanceStringIndex(input, 0, false); got != 1 {
		t.Fatalf("non-Unicode advance = %d, want 1", got)
	}
	if got := AdvanceStringIndex(input, 0, true); got != 2 {
		t.Fatalf("Unicode advance over astral code point = %d, want 2", got)
	}
	if got := AdvanceStringIndex(input, 1, true); got != 2 {
		t.Fatalf("Unicode advance from low surrogate = %d, want 2", got)
	}
	if got := AdvanceStringIndex(input, 2, true); got != 3 {
		t.Fatalf("Unicode advance over BMP code point = %d, want 3", got)
	}
}

func TestGetMatchStringSlicesByUTF16CodeUnits(t *testing.T) {
	t.Parallel()

	const input = "a😀b"
	match := &MatchRecord{StartIndex: 1, EndIndex: 3}
	if got := GetMatchString(nil, input, match); got != "😀" {
		t.Fatalf("GetMatchString() = %q, want emoji", got)
	}
}
