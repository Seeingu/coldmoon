package tests

import (
	. "github.com/Seeingu/coldmoon/coldmoon"
	"testing"
)

func testSource(t *testing.T, s string) {
	Debug.Enable()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	//sourceText := "\t{true; false\u2028;;;}\r\nnull;debugger\uFEFF"
	sourceText := s
	script := ParseScript(sourceText, realm, nil)
	_ = script.Evaluate()
	Debug.Disable()
}

func TestBaseline(t *testing.T) {
	var sourceText string

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
const a = {
	b: 1,
};
let b = 2;
`
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
