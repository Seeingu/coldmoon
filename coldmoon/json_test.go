package coldmoon

import "testing"

func TestUnicodeEscapeUsesFourLowercaseHexadecimalDigits(t *testing.T) {
	tests := []struct {
		name string
		code uint16
		want string
	}{
		{name: "nul", code: 0x0000, want: `\u0000`},
		{name: "unit separator", code: 0x001f, want: `\u001f`},
		{name: "four digit value", code: 0xabcd, want: `\uabcd`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := UnicodeEscape(test.code); got != test.want {
				t.Fatalf("UnicodeEscape(%#04x) = %q, want %q", test.code, got, test.want)
			}
		})
	}
}

func TestQuoteJSONStringEscapesControlsAndPreservesOtherRunes(t *testing.T) {
	const input = "\x00\x1fregular😀"
	const want = `"\u0000\u001fregular😀"`

	if got := QuoteJSONString(nil, input); got != want {
		t.Fatalf("QuoteJSONString(%q) = %q, want %q", input, got, want)
	}
}
