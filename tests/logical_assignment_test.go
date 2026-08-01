package tests

import "testing"

// TestLogicalAssignmentShortCircuits verifies that RHS evaluation and writes
// occur only when the selected logical operator requires assignment.
func TestLogicalAssignmentShortCircuits(t *testing.T) {
	testSource(t, `
let calls = 0;
function rhs(value) { calls += 1; return value; }

let andValue = 0;
andValue &&= rhs(1);
assertEqual(andValue, 0);
assertEqual(calls, 0);
andValue = 2;
andValue &&= rhs(3);
assertEqual(andValue, 3);
assertEqual(calls, 1);

let orValue = "present";
orValue ||= rhs("replacement");
assertEqual(orValue, "present");
assertEqual(calls, 1);
orValue = "";
orValue ||= rhs("replacement");
assertEqual(orValue, "replacement");
assertEqual(calls, 2);

let nullishValue = false;
nullishValue ??= rhs(true);
assertEqual(nullishValue, false);
assertEqual(calls, 2);
nullishValue = null;
nullishValue ??= rhs("filled");
assertEqual(nullishValue, "filled");
assertEqual(calls, 3);
`)
}

// TestLogicalAssignmentEvaluatesReferenceOnce verifies observable getter/setter
// behavior and inferred names for an identifier assignment.
func TestLogicalAssignmentEvaluatesReferenceOnce(t *testing.T) {
	testSource(t, `
let baseEvaluations = 0;
const holder = { value: 0 };
function target() { baseEvaluations += 1; return holder; }
target().value &&= 1;
assertEqual(baseEvaluations, 1);
assertEqual(holder.value, 0);
target().value ||= 2;
assertEqual(baseEvaluations, 2);
assertEqual(holder.value, 2);

let callback;
callback ??= function() {};
assertEqual(callback.name, "callback");
`)
}
