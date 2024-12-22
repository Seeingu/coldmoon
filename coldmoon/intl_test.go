package coldmoon

import (
	"testing"

	"github.com/Seeingu/icu4xgo"
)

// Ensure icu4xgo linked static library correctly
func TestLocale(t *testing.T) {
	locale := icu4xgo.NewLocale("en-US")
	if locale.Language() != "en" {
		t.Errorf("Language is not en")
	}
}
