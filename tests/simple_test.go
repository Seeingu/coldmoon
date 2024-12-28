package tests

import (
	"testing"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func testSource(t *testing.T, s string) {
	Debug.Enable()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	sourceText := s
	evaluate(`
function assert(a, msg) {
	if (a) {
		return;
	}
	if (msg) {
		throw new Error('assertion failed: ' + msg);
	} else {
		throw new Error('assertion failed');
	}
}
`, realm)
	evaluate(sourceText, realm)
	Debug.Disable()
}

func TestBaseline(t *testing.T) {
	sourceTexts := []string{
		`
const a = [1,2,3]
for (const i in a) {
}
`,
		`
const foo = function* () {
  yield 'a';
  yield 'b';
  yield 'c';
};

let str = '';
for (const val of foo()) {
  str = str + val;
}
assert(str === 'abc');

const gen = foo();
assert(gen.next().value === 'a');
assert(gen.next().value === 'b');
assert(gen.next().value === 'c');

const infinite = function*() {
  let index = 0;

  while (true) {
    yield index++;
  }
}

const generator = infinite(); // "Generator { }"

assert(generator.next().value === 0); // 0
assert(generator.next().value === 1); // 1
assert(generator.next().value === 2); // 2
`,
		`
const string1 = "A string primitive";
const string2 = 'Also a string primitive';
const string4 = new String("A String object");
assert("cat".charAt(1) === 'a');
assert("cat"[1] === 'a');
const strPrim = "foo"; // A literal is a string primitive
const strPrim2 = String(1); // Coerced into the string primitive "1"
const strPrim3 = String(true); // Coerced into the string primitive "true"
const strObj = new String(strPrim); // String with new returns a string wrapper object.

assert(typeof strPrim === "string", "typeof: expected string");
assert(typeof strPrim2 === "string", "typeof: expected string");
assert(typeof strPrim3 === "string", "typeof: expected string");
assert(typeof strObj === "object", "typeof: expected object"); 
`,
		// TODO: String locale compare
		"const string3 = `Yet another string primitive`;",
		`const a = {};
const b = a?.b ?? 1;
assert(b === 1, 'b should be 1')`,
		`
class ValidatorClass {
  get [Symbol.toStringTag]() {
    return 'Validator';
  }
}
const v = new ValidatorClass();
assert(Object.prototype.toString.call(v) === '[object Validator]', 'should be [object Validator]');
`,
		`const sab = new SharedArrayBuffer(1024);
const ta = new Uint8Array(sab);
ta[0] = 5; // 5
ta[123] = 12;
Atomics.add(ta, 0, 12); // 5
Atomics.and(ta, 0, 1); 
Atomics.compareExchange(ta, 0, 5, 12); // 1
Atomics.exchange(ta, 0, 12); // 1
Atomics.isLockFree(1); // true
Atomics.isLockFree(2); // true
Atomics.isLockFree(3); // false
Atomics.isLockFree(4); // true
Atomics.or(ta, 0, 1); // 12
Atomics.store(ta, 0, 12); // 12
Atomics.sub(ta, 0, 2); // 12
Atomics.xor(ta, 0, 1); // 10
Atomics.load(ta, 0); 
`,
		"Boolean(function() {}())",
		"typeof Boolean(void 0)",
		"void 0",
		`const desc = Object.getOwnPropertyDescriptor(BigInt64Array, 'BYTES_PER_ELEMENT');
let a = 0;
if (Object.prototype.hasOwnProperty.call(desc, 'enumerable')) {
	a = 1;
}
a;`,
		`const a = {
  b: 8,
  writable: false,
  enumerable: false,
  configurable: false
};
delete a.b;
let b = 2;
a.b = b;
a.b;`,
		`let a = 1;
for (var i = 0; i < 3; i++) {
	a += i;
}
a;`,
		`var a = [];
var a = [1,2,3];
a[4294967295] = "not an array element";
a[4294967295];`,
		`function toString(a) {
	return a;
}
let a = toString(10);
a + 'hello';`,
		`let a = 1;
let b = a + 1;
let c = a + b + b;
'hello' + a + "world";
c += a + b;
a ? b : c;
var d = 1 ?	2 : 3;`,
		"``",
		`class A {
	static sa = 'a';
}
const B = class {}`,
		`function a() {
}
function* b() {
}
async function* c() {
}
async function d() {
}
const a = async () => {
}`,
		`2 == 1;
2 > 1;
false || 1;
true && 1;
true ? 2 : 1;
2 ** 3;
() => 123;
Date.UTC(2012);`,
	}
	for _, sourceText := range sourceTexts {
		testSource(t, sourceText)
	}
}
