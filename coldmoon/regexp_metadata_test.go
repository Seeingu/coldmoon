package coldmoon

import "testing"

// TestCountLeftCapturingParensWithinUsesCompiledMetadata verifies that plain,
// named, and non-capturing groups are distinguished after pattern parsing.
func TestCountLeftCapturingParensWithinUsesCompiledMetadata(t *testing.T) {
	tests := []struct {
		pattern string
		want    int
	}{
		{pattern: "plain", want: 0},
		{pattern: "(a)(b)", want: 2},
		{pattern: "(?:a)(b)", want: 1},
		{pattern: "(?<named>a)(?:b)", want: 1},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			regexp, err := ParsePattern(test.pattern, false, false)
			if err != nil {
				t.Fatalf("ParsePattern(%q): %v", test.pattern, err)
			}
			if got := CountLeftCapturingParensWithin(regexp); got != test.want {
				t.Fatalf("capturing group count = %d, want %d", got, test.want)
			}
		})
	}
}
