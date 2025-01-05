package tests

import (
	"testing"

	"github.com/Seeingu/coldmoon/runtime"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func testSource(t *testing.T, s string) {
	Debug.Enable()
	agent := NewAgent()
	InitializeConstants()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	runtime.RegisterTerminalRuntime(realm)
	sourceText := s
	evaluate(sourceText, realm)
	Debug.Disable()
}

func TestBaseline(t *testing.T) {
	sourceTexts := []string{
		`
const str = 'table football';
const regex = new RegExp('foo*');
const globalRegex = new RegExp('foo*', 'g');
assertEqual(regex.test(str), true);
assertEqual(globalRegex.lastIndex, 0);
assertEqual(globalRegex.test(str), true);
assertEqual(globalRegex.lastIndex, 9);
assertEqual(globalRegex.test(str), false);
`,
		`
const utcDate1 = new Date(Date.UTC(96, 1, 2, 3, 4, 5));
const utcDate2 = new Date(Date.UTC(0, 0, 0, 0, 0, 0));

assertEqual(utcDate1.toUTCString(), "Fri, 02 Feb 1996 03:04:05 UTC");
assertEqual(utcDate2.toUTCString(), "Sun, 31 Dec 1899 00:00:00 UTC");

// check time zone
//assertEqual(new Date(8.64e15).toString(), "Sat Sep 13 275760 08:00:00 CST+0800"); 
assertEqual(new Date(8.64e15 + 1).toString(), "Invalid Date"); 

const date = new Date('December 17, 1995 03:24:00');

date[Symbol.toPrimitive]('string');
assertEqual(date[Symbol.toPrimitive]('number'), 819170640000);

`,
		`
const num1 = 42;
const num2 = 3.14;
const num3 = Number('123');
const num4 = parseInt('123', 10);
const num5 = parseFloat('3.14');
const num6 = 0b1010; // binary
const num7 = 0o52; // octal
const num8 = 0x2A; // hexadecimal
const num9 = 8.64e15;

assert(num1 === 42);
assert(num2 === 3.14);
assert(num3 === 123);
assert(num4 === 123);
assert(num5 === 3.14);
assert(num6 === 10);
assert(num7 === 42);
assert(num8 === 42);
assert(num9 === 8640000000000000);
`, `
const set1 = new Set();

set1.add(42);
set1.add('forty two');

const iterator1 = set1[Symbol.iterator]();

assertEqual(iterator1.next().value, 42);
assertEqual(iterator1.next().value, 'forty two');
`,
		`
const map1 = new Map();

map1.set('0', 'foo');
map1.set(1, 'bar');

const iterator1 = map1[Symbol.iterator]();

assertEqual(iterator1.next().value[0], 0);
assertEqual(iterator1.next().value[1], 'bar');
`,
		`
const map1 = new Map();
map1.set('a', 1);
map1.set('b', 2);
map1.set('c', 3);
assert(map1.get('a') === 1);
map1.set('a', 97);
assert(map1.get('a') === 97);
assert(map1.size === 3);
map1.delete('b');
assert(map1.size === 2);
`,
		`
const object1 = {
    [Symbol.toPrimitive](hint) {
        if (hint === 'number') {
            return 42;
        }
        return null;
    },
};

assert(+object1 === 42);
`,
		`
const a = [1,2,3]
for (const i in a) {
	assert(i === '0' || i === '1' || i === '2');
}

const array1 = ['a', 'b', 'c'];
const iterator1 = array1[Symbol.iterator]();

for (const value of iterator1) {
    assert(value === 'a' || value === 'b' || value === 'c');
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
`,
	}
	for _, sourceText := range sourceTexts {
		testSource(t, sourceText)
	}
}
