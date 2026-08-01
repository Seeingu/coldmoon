package tests

import "testing"

func TestJSONStringifyEscapesControlsAndPreservesNonBMPCharacters(t *testing.T) {
	testSource(t, `
const serialized = JSON.stringify(
  String.fromCharCode(0, 31) + "regular" + String.fromCodePoint(0x1f600)
);

assertEqual(serialized, "\"\\u0000\\u001fregular😀\"");
`)
}
