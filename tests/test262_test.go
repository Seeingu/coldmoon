package tests

import (
	. "github.com/Seeingu/coldmoon/coldmoon"
	"os"
	"path"
	"testing"
)

func makeTest262Path(p string) string {
	dir, _ := os.Getwd()
	return path.Join(dir, "..", p)
}

func runTest(filePath string) {
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	sta, err := os.ReadFile(makeTest262Path("./test262/harness/sta.js"))
	if err != nil {
		panic(err)
	}
	assertJs, err := os.ReadFile(makeTest262Path("./test262/harness/assert.js"))

	ParseScript(string(sta), realm, nil).Evaluate()
	ParseScript(string(assertJs), realm, nil).Evaluate()
}

func mustReadFile(filePath string) string {
	f, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return string(f)
}

func runTestHarness(realm *Realm, f string) {
	content := mustReadFile(makeTest262Path("./test262/harness/" + f))
	v := ParseScript(content, realm, nil).Evaluate()
	print(v.String())
}

func TestHarness(t *testing.T) {
	files := []string{
		"isConstructor.js",
		"nans.js",
	}

	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	for _, f := range files {
		runTestHarness(realm, f)
	}
}
