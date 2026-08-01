package tests

import "testing"

func TestClassFunctionToStringUsesClassSourceText(t *testing.T) {
	testSource(t, `
class Named {}
assertEqual(Named.toString(), "class Named {}");

const anonymous = class {};
assertEqual(anonymous.toString(), "class {}");

const explicit = class Inner {};
assertEqual(explicit.toString(), "class Inner {}");
`)
}

func TestUnaryBitwiseNotUsesInt32Semantics(t *testing.T) {
	testSource(t, `
assertEqual(~4294967296, -1);
assertEqual(~4294967295, 0);
assertEqual(~1.9, -2);
`)
}
