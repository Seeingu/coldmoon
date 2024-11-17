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

func TestParseSta(t *testing.T) {
}
