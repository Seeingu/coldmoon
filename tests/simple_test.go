package tests

import (
	. "github.com/Seeingu/coldmoon/coldmoon"
	"testing"
)

func testSource(t *testing.T, s string) {
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	//sourceText := "\t{true; false\u2028;;;}\r\nnull;debugger\uFEFF"
	sourceText := s
	script := ParseScript(sourceText, realm, nil)
	_ = script.Evaluate()

}

func TestBaseline(t *testing.T) {
	sourceText := `
class A {}
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
