package tests

import "testing"

func TestDestructuringBindingsAndAssignments(t *testing.T) {
	testSource(t, `
let first, second, tail;
[first, second = 20, ...tail] = [1, undefined, 3, 4];
assertEqual(first, 1, "array assignment first");
assertEqual(second, 20, "array assignment default");
assertEqual(tail.length, 2, "array assignment rest length");
assertEqual(tail[0], 3, "array assignment rest value");

let nested;
[[nested]] = [[7]];
assertEqual(nested, 7, "nested array assignment");

let sourceName, renamed, objectRest;
({ sourceName, renamed: renamed = 5, ...objectRest } = {
  sourceName: 2,
  renamed: undefined,
  kept: 9
});
assertEqual(sourceName, 2, "object assignment shorthand");
assertEqual(renamed, 5, "object assignment renamed default");
assertEqual(objectRest.kept, 9, "object assignment rest");
assertEqual(objectRest.sourceName, undefined, "object rest excludes consumed keys");
`)
}

func TestDestructuringDeclarationsParametersAndCatch(t *testing.T) {
	t.Run("variable and lexical declarations", func(t *testing.T) {
		testSource(t, `
var { varValue, nested: { deepValue }, ...varRest } = {
  varValue: 1,
  nested: { deepValue: 2 },
  extra: 3
};
assertEqual(varValue, 1, "var object binding");
assertEqual(deepValue, 2, "var nested binding");
assertEqual(varRest.extra, 3, "var rest binding");

const [constantValue = 4, ...constantRest] = [undefined, 5, 6];
assertEqual(constantValue, 4, "lexical default binding");
assertEqual(constantRest.length, 2, "lexical rest binding");
`)
	})

	t.Run("parameters", func(t *testing.T) {
		testSource(t, `
function unpack({ left, right: renamedRight = 8 }, [item = 9], ...remaining) {
  return left + renamedRight + item + remaining.length;
}
assertEqual(unpack({ left: 1 }, [], "a", "b"), 20, "parameter patterns");
`)
	})

	t.Run("catch pattern", func(t *testing.T) {
		testSource(t, `
let caughtCode;
let caughtRest;
try {
  throw { code: 7, detail: 8 };
} catch ({ code, ...remainingError }) {
  caughtCode = code;
  caughtRest = remainingError;
}
assertEqual(caughtCode, 7, "catch pattern");
assertEqual(caughtRest.detail, 8, "catch rest pattern");
`)
	})

	t.Run("optional catch binding", func(t *testing.T) {
		testSource(t, `
let catchWithoutParameter = false;
try {
  throw 1;
} catch {
  catchWithoutParameter = true;
}
assertEqual(catchWithoutParameter, true, "optional catch binding");
`)
	})
}

func TestDestructuringForOfBindings(t *testing.T) {
	testSource(t, `
let lexicalTotal = 0;
for (const [left, right] of [[1, 2], [3, 4]]) {
  lexicalTotal += left + right;
}
assertEqual(lexicalTotal, 10, "lexical for-of pattern");

let varTotal = 0;
for (var { value } of [{ value: 2 }, { value: 5 }]) {
  varTotal += value;
}
assertEqual(varTotal, 7, "var for-of pattern");
`)
}
