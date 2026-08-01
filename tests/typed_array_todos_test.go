package tests

import "testing"

// TestTypedArrayToLocaleStringFormatsElements verifies the typed-array-specific
// locale join operation for number and BigInt element types.
func TestTypedArrayToLocaleStringFormatsElements(t *testing.T) {
	testSource(t, `
const numbers = new Uint8Array(2);
const bigints = new BigInt64Array(2);
assertEqual(numbers.toLocaleString(), "0,0");
assertEqual(bigints.toLocaleString(), "0,0");
`)
}

// TestTypedArrayToLocaleStringPropagatesElementFailure verifies that element
// localization is observable and abrupt completions are not swallowed.
func TestTypedArrayToLocaleStringPropagatesElementFailure(t *testing.T) {
	testSource(t, `
const original = {};
Number.prototype.toLocaleString = function() { throw original; };
const values = new Uint8Array(1);
let caught;
try {
  values.toLocaleString();
} catch (error) {
  caught = error;
}
assert(caught === original);
`)
}

// TestTypedArrayToStringTagIsGetter verifies the descriptor shape and the
// specification-visible name returned for concrete typed arrays.
func TestTypedArrayToStringTagIsGetter(t *testing.T) {
	testSource(t, `
const descriptor = Object.getOwnPropertyDescriptor(Object.getPrototypeOf(Uint8Array.prototype), Symbol.toStringTag);
assert(typeof descriptor.get === "function");
assert(descriptor.set === undefined);
assertEqual(Object.prototype.toString.call(new Uint8Array()), "[object Uint8Array]");
`)
}

// TestFloatingTypedArrayStoresIEEE754Bits verifies that floating typed arrays
// encode Number values as IEEE-754 payloads, including the sign bit of zero,
// instead of applying the integer typed-array conversion.
func TestFloatingTypedArrayStoresIEEE754Bits(t *testing.T) {
	testSource(t, `
const float32 = new Float32Array(4);
float32[0] = 1.5;
float32[1] = -0;
float32[2] = Infinity;
float32[3] = NaN;
assertEqual(float32[0], 1.5, "Float32 finite value");
assert(1 / float32[1] === -Infinity, "Float32 negative zero");
assertEqual(float32[2], Infinity, "Float32 infinity");
assert(Number.isNaN(float32[3]), "Float32 NaN");

const float64 = new Float64Array(4);
float64[0] = -13.25;
float64[1] = -0;
float64[2] = Infinity;
float64[3] = NaN;
assertEqual(float64[0], -13.25, "Float64 finite value");
assert(1 / float64[1] === -Infinity, "Float64 negative zero");
assertEqual(float64[2], Infinity, "Float64 infinity");
assert(Number.isNaN(float64[3]), "Float64 NaN");

const copied = new Float64Array(4);
copied.set(float64, 0);
assertEqual(copied[0], -13.25, "Float64 set preserves raw payload");
assert(1 / copied[1] === -Infinity, "Float64 set preserves negative zero");

const sliced = float64.slice(0, 2);
assertEqual(sliced[0], -13.25, "Float64 slice preserves raw payload");
assert(1 / sliced[1] === -Infinity, "Float64 slice preserves negative zero");

const converted = new Float64Array(new Float32Array([1.5]));
assertEqual(converted[0], 1.5, "floating typed-array conversion uses numeric value");
`)
}
