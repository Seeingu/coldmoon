package tests

import "testing"

// TestRegExpSourceReturnsEscapedPattern verifies the observable source getter
// for empty patterns and slash characters.
func TestRegExpSourceReturnsEscapedPattern(t *testing.T) {
	testSource(t, `
assertEqual(new RegExp("").source, "(?:)");
assertEqual(new RegExp("a/b").source, "a\\/b");
`)
}
