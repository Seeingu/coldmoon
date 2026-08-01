package tests

import "testing"

func TestParserRegressionCasesExecute(t *testing.T) {
	t.Run("BigInt literals", func(t *testing.T) {
		testSource(t, `
assertEqual(0x20000000000001n, 9007199254740993n);
assertEqual(0b1010_0001n + 1_000n, 1161n);
`)
	})

	t.Run("line terminators and ASI", func(t *testing.T) {
		testSource(t, `
function returnBeforeLineTerminator() {
  return  41;
}
assertEqual(returnBeforeLineTerminator(), undefined);

let counter = 1;
counter
++counter;
assertEqual(counter, 2);

const increment = value =>
  value + 1;
assertEqual(increment(4), 5);
`)
	})

	t.Run("cases after default", func(t *testing.T) {
		testSource(t, `
let trace = "";
switch (2) {
case 1:
  trace += "one";
  break;
default:
  trace += "default";
case 2:
  trace += "two";
  break;
}
assertEqual(trace, "two");

trace = "";
switch (9) {
default:
  trace += "default";
case 2:
  trace += "two";
  break;
}
assertEqual(trace, "defaulttwo");
`)
	})

	t.Run("anonymous class and cover initialized name", func(t *testing.T) {
		testSource(t, `
const Anonymous = class {};
const instance = new Anonymous();
assertEqual(instance instanceof Anonymous, true);

let cover;
({cover = 7} = {});
assertEqual(cover, 7);
`)
	})

	t.Run("callable source text excludes trailing trivia", func(t *testing.T) {
		testSource(t, `
function ordinary() {} /* trailing declaration comment */
assertEqual(ordinary.toString(), "function ordinary() {}");

function* generator() {} // trailing generator comment
assertEqual(generator.toString(), "function* generator() {}");

async function asynchronous() {} /* trailing async comment */
assertEqual(asynchronous.toString(), "async function asynchronous() {}");

async function* asyncGenerator() {} // trailing async generator comment
assertEqual(asyncGenerator.toString(), "async function* asyncGenerator() {}");

const functionExpression = function named() {}; /* trailing expression comment */
assertEqual(functionExpression.toString(), "function named() {}");

const generatorExpression = function* namedGenerator() {}; // trailing generator expression comment
assertEqual(generatorExpression.toString(), "function* namedGenerator() {}");

const asyncFunctionExpression = async function namedAsync() {}; /* trailing async expression comment */
assertEqual(asyncFunctionExpression.toString(), "async function namedAsync() {}");

const asyncGeneratorExpression = async function* namedAsyncGenerator() {}; // trailing async generator expression comment
assertEqual(asyncGeneratorExpression.toString(), "async function* namedAsyncGenerator() {}");

class ClassDeclaration {} /* trailing class declaration comment */
assertEqual(ClassDeclaration.toString(), "class ClassDeclaration {}");

const classExpression = class NamedClass {}; // trailing class expression comment
assertEqual(classExpression.toString(), "class NamedClass {}");

const arrow = value => value + 1; /* trailing arrow comment */
assertEqual(arrow.toString(), "value => value + 1");

const asyncArrow = async value => value + 1; // trailing async arrow comment
assertEqual(asyncArrow.toString(), "async value => value + 1");
`)
	})

	t.Run("var binding pattern", func(t *testing.T) {
		testSource(t, `
var [varFirst, varSecond] = [10, 20];
assertEqual(varFirst, 10);
assertEqual(varSecond, 20);
`)
	})

	t.Run("lexical binding pattern", func(t *testing.T) {
		testSource(t, `
const {kept, missing = 5, ...remaining} = {kept: 3, extra: 9};
assertEqual(kept, 3);
assertEqual(missing, 5);
assertEqual(remaining.extra, 9);
`)
	})

	t.Run("parameter binding pattern", func(t *testing.T) {
		testSource(t, `
function unpack({x}, [y, ...rest]) {
  return x + y + rest[0];
}
assertEqual(unpack({x: 1}, [2, 3]), 6);
`)
	})

	t.Run("for-of binding pattern", func(t *testing.T) {
		testSource(t, `
let loopTotal = 0;
for (const [left, right] of [[4, 5]]) {
  loopTotal = left + right;
}
assertEqual(loopTotal, 9);
`)
	})

	t.Run("try catch finally", func(t *testing.T) {
		testSource(t, `
let finalized = false;
try {
  finalized = false;
} finally {
  finalized = true;
}
assertEqual(finalized, true);

let caughtWithoutBinding = false;
try {
  throw 1;
} catch {
  caughtWithoutBinding = true;
}
assertEqual(caughtWithoutBinding, true);

let caughtMessage = "";
try {
  throw {message: "caught"};
} catch ({message}) {
  caughtMessage = message;
}
assertEqual(caughtMessage, "caught");
`)
	})
}

func TestTaggedTemplateEvaluationAndSiteIdentity(t *testing.T) {
	source := `
let firstTemplate;
let callCount = 0;
function tag(strings, value) {
  callCount++;
  assert(Object.isFrozen(strings));
  assert(Object.isFrozen(strings.raw));
  assertEqual(strings[0], "line\n");
  assertEqual(strings.raw[0], "line\\n");
  assertEqual(strings[1], "!");
  if (firstTemplate === undefined) firstTemplate = strings;
  else assert(strings === firstTemplate, "template object is stable per source site");
  return value;
}
function render(value) {
  return tag` + "`line\\n${value}!`" + `;
}
assertEqual(render(41), 41);
assertEqual(render(42), 42);
assertEqual(callCount, 2);

const receiver = {
  value: 7,
  tag(strings) { return this.value; }
};
assertEqual(receiver.tag` + "`receiver`" + `, 7);

let siteOne;
let siteTwo;
function capture(strings) { return strings; }
siteOne = capture` + "`same`" + `;
siteTwo = capture` + "`same`" + `;
assert(siteOne !== siteTwo, "different source sites use different template objects");
`
	testSource(t, source)
}

func TestTaggedTemplateEscapeSemantics(t *testing.T) {
	source := "function capture(strings) { return strings; }\n" +
		"const escaped = capture`\\${value}\\``;\n" +
		"assertEqual(escaped[0], \"${value}`\");\n" +
		"assertEqual(escaped.raw[0], \"\\\\${value}\\\\`\");\n" +
		"const invalid = capture`\\xZ1`;\n" +
		"assertEqual(invalid[0], undefined);\n" +
		"assertEqual(invalid.raw[0], \"\\\\xZ1\");\n"
	testSource(t, source)
}
