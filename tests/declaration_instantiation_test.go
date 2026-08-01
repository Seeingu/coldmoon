package tests

import (
	"testing"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func TestFunctionBodyHoistsAllFunctionDeclarationKinds(t *testing.T) {
	realm := evaluateAsyncSource(t, `
function ordinaryOuter() {
  return ordinaryInner();
  function ordinaryInner() { return 41; }
}
assertEqual(ordinaryOuter(), 41, "ordinary declaration hoisted");

function generatorOuter() {
  return generatorInner();
  function* generatorInner() { yield 42; }
}
const generator = generatorOuter();
assertEqual(generator.next().value, 42, "generator declaration hoisted");

function asyncOuter() {
  return asyncInner();
  async function asyncInner() { return 43; }
}
let asyncValue;
asyncOuter().then(function (value) { asyncValue = value; });

function asyncGeneratorOuter() {
  return asyncGeneratorInner();
  async function* asyncGeneratorInner() { yield 44; }
}
let asyncGeneratorValue;
asyncGeneratorOuter().next().then(function (result) {
  asyncGeneratorValue = result.value;
});
`)
	Evaluate(`
assertEqual(asyncValue, 43, "async function declaration hoisted");
assertEqual(asyncGeneratorValue, 44, "async generator declaration hoisted");
`, realm)
}

func TestBlockDeclarationInstantiationCreatesLexicalScope(t *testing.T) {
	testSource(t, `
let value = 1;
{
  assertEqual(value, 1, "outer value visible before shadow declaration initialization");
}
{
  let value = 2;
  assertEqual(value, 2, "block lexical binding");
  assertEqual(blockFunction(), 3, "block function is initialized before evaluation");
  function blockFunction() { return 3; }
}
assertEqual(value, 1, "block lexical environment restored");
`)
}
