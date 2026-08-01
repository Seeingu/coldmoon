package coldmoon

import "testing"

// TestEscapeRegExpPatternProducesLiteralSafeSource verifies empty patterns,
// slashes, and every ECMAScript line terminator.
func TestEscapeRegExpPatternProducesLiteralSafeSource(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{name: "empty", pattern: "", want: "(?:)"},
		{name: "slash", pattern: "a/b", want: `a\/b`},
		{name: "line terminators", pattern: "a\nb\rc\u2028d\u2029e", want: `a\nb\rc\u2028d\u2029e`},
		{name: "ordinary", pattern: "[a-z]+", want: "[a-z]+"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := EscapeRegExpPattern(test.pattern, "u"); got != test.want {
				t.Fatalf("EscapeRegExpPattern(%q) = %q, want %q", test.pattern, got, test.want)
			}
		})
	}
}
