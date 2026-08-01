package tests

import "testing"

// TestBindingsInferAnonymousFunctionNames verifies NamedEvaluation for var,
// lexical, async, generator, arrow, and class initializers.
func TestBindingsInferAnonymousFunctionNames(t *testing.T) {
	testSource(t, `
var ordinary = function() {};
let arrow = () => 1;
const asyncFunction = async function() {};
const generator = function*() {};
const asyncGenerator = async function*() {};
const classValue = class {};

assertEqual(ordinary.name, "ordinary");
assertEqual(arrow.name, "arrow");
assertEqual(asyncFunction.name, "asyncFunction");
assertEqual(generator.name, "generator");
assertEqual(asyncGenerator.name, "asyncGenerator");
assertEqual(classValue.name, "classValue");
`)
}

// TestAssignmentsInferAnonymousFunctionNames verifies that both identifier and
// property references contribute their referenced names.
func TestAssignmentsInferAnonymousFunctionNames(t *testing.T) {
	testSource(t, `
	let assigned;
	assigned = function() {};
	let parenthesized;
	parenthesized = (function() {});
	const holder = {};
	holder.member = () => 1;

	assertEqual(assigned.name, "assigned");
	assertEqual(parenthesized.name, "parenthesized");
	assertEqual(holder.member.name, "");
`)
}

// TestObjectPropertiesInferAnonymousFunctionNames verifies static and computed
// property keys, including symbol descriptions, in object initializers.
func TestObjectPropertiesInferAnonymousFunctionNames(t *testing.T) {
	testSource(t, `
const symbol = Symbol("callback");
const object = {
  ordinary: function() {},
  ["computed"]: async () => 1,
  [symbol]: class {}
};

assertEqual(object.ordinary.name, "ordinary");
assertEqual(object.computed.name, "computed");
assertEqual(object[symbol].name, "[callback]");
`)
}
