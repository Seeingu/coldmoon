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
	script := ParseScript(sourceText, realm, nil)
	_ = script.Evaluate()
	Debug.Disable()
}

func TestBaseline(t *testing.T) {
	sourceTexts := []string{
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
		`const a = {};
const b = a?.b;`,
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
