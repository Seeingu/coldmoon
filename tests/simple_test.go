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
	// sourceText := "\t{true; false\u2028;;;}\r\nnull;debugger\uFEFF"
	sourceText := s
	script := ParseScript(sourceText, realm, nil)
	_ = script.Evaluate()
	Debug.Disable()
}

func TestBaseline(t *testing.T) {
	var sourceText string

	sourceText = `
const a = {
  b: 8,
  writable: false,
  enumerable: false,
  configurable: false
};
delete a.b;
let b = 2;
a.b = b;
a.b;
`
	testSource(t, sourceText)

	sourceText = `
let a = 1;
for (var i = 0; i < 3; i++) {
	a += i;
}
for (var i in [1, 2, 3]) {
	a += i;
}
a;
`
	testSource(t, sourceText)

	sourceText = `
var a = [];
a[4294967295] = "not an array element";
a[4294967295];
`
	testSource(t, sourceText)

	sourceText = `
function toString(a) {
	return a;
}
let a = toString(10);
a + 'hello';
`
	testSource(t, sourceText)

	sourceText = `
let a = 1;
let b = a + 1;
let c = a + b + b;
'hello' + a + "world";
c += a + b;
a ? b : c;
var d = 1 ?	2 : 3;
`
	testSource(t, sourceText)

	sourceText = `
function a() {}
a.b = 1;
a.b;
`
	testSource(t, sourceText)

	sourceText = `
const a = {};
const b = a?.b;
`
	testSource(t, sourceText)

	sourceText = "``"
	testSource(t, sourceText)

	sourceText = `
class A {
	static sa = 'a';
}
const B = class {}
`
	testSource(t, sourceText)

	sourceText = `
function a() {
}
function* b() {
}
async function* c() {
}
async function d() {
}
const a = async () => {
}
`
	testSource(t, sourceText)

	sourceText = `
2 == 1;
2 > 1;
false || 1;
true && 1;
true ? 2 : 1;
2 ** 3;
() => 123;
Date.UTC(2012);
`
	testSource(t, sourceText)
}
