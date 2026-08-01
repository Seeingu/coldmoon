package tests

import "testing"

func TestOptionalChainKeepsFollowingSegmentsInTheChain(t *testing.T) {
	testSource(t, `
const missing = null;
assertEqual(missing?.value.child, undefined, "missing property chain");
assertEqual(missing?.method(), undefined, "missing call chain");

let keyEvaluations = 0;
assertEqual(missing?.[keyEvaluations++].child, undefined, "computed chain");
assertEqual(keyEvaluations, 0, "computed key skipped");

const object = {
    nested: { value: 7 },
    method: function() { return this.nested; }
};
assertEqual(object?.nested.value, 7, "present property chain");
assertEqual(object.method?.().value, 7, "optional method call");

let callable;
assertEqual(callable?.(), undefined, "missing optional call");

let threw = false;
try {
	const result = ({ value: undefined })?.value.child;
} catch (error) {
    threw = true;
}
assertEqual(threw, true, "non-optional tail throws");
`)
}
