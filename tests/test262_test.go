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

func mustReadFile(filePath string) string {
	f, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return string(f)
}

func runTestHarness(realm *Realm, f string, debug bool) {
	if debug {
		Debug.Enable()
	}
	content := mustReadFile(makeTest262Path("./test262/harness/" + f))
	v := ParseScript(content, realm, nil).Evaluate()
	defer func() {
		Debug.Disable()
	}()
	if debug {
		print(v.String())
	}
}

func testDataView(realm *Realm) {
	dataViewDir := "./test262/test/built-ins/DataView/"
	f := makeTest262Path(dataViewDir + "constructor.js")
	ParseScript(mustReadFile(f), realm, nil).Evaluate()
}
func testArray(realm *Realm) {
	arrayDir := "./test262/test/built-ins/Array/"
	entries, err := os.ReadDir(makeTest262Path(arrayDir))
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		f := makeTest262Path(arrayDir + entry.Name())
		println("Testing file: ", entry.Name())
		ParseScript(mustReadFile(f), realm, nil).Evaluate()
	}

}

func TestHarness(t *testing.T) {
	files := []string{
		"sta.js",
		"assert.js",
		"isConstructor.js",
		"nans.js",
		"assertRelativeDateMs.js",
	}

	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	for _, f := range files {
		runTestHarness(realm, f, false)
	}

	testDataView(realm)

	Debug.Enable()
	//testArray(realm)
	Debug.Disable()
}
