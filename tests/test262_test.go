package tests

import (
	"os"
	"path"
	"testing"

	. "github.com/Seeingu/coldmoon/coldmoon"
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

func evaluate(fileName string, realm *Realm) {
	result := ParseScript(mustReadFile(fileName), realm, nil).Evaluate()
	if o, ok := ValueGetObject(result); ok {
		if e, ok := o.(*ErrorObject); ok {
			println("Return Error: ", e.Message)
			panic(e)
		}
	}
}

func testDataView(realm *Realm) {
	dataViewDir := "./test262/test/built-ins/DataView/"
	f := makeTest262Path(dataViewDir + "constructor.js")
	evaluate(f, realm)
}

func testTypedArray(realm *Realm) {
	typedArrayConstructorDir := "./test262/test/built-ins/TypedArrayConstructors"
	constructorNames := []string{
		"BigInt64Array",
		"Int8Array",
		"Float32Array",
	}
	fileNames := []string{
		"constructor.js",
		"BYTES_PER_ELEMENT.js",
		"is-a-constructor.js",
		"length.js",
		"name.js",
		"prototype.js",
		"prototype/BYTES_PER_ELEMENT.js",
		"prototype/constructor.js",
		"prototype/not-typedarray-object.js",
		// TODO:
		//"prop-desc.js",
		//"proto.js",
	}
	for _, name := range constructorNames {
		for _, fileName := range fileNames {
			d := typedArrayConstructorDir + "/" + name
			f := makeTest262Path(d + "/" + fileName)
			println("BigInt64Array: Testing file: ", fileName)
			evaluate(f, realm)
		}
	}
}

func testBoolean(realm *Realm) {
	boolDir := "./test262/test/built-ins/Boolean/"
	entries, err := os.ReadDir(makeTest262Path(boolDir))
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		f := makeTest262Path(boolDir + entry.Name())
		println("BOOLEAN: Testing file: ", entry.Name())
		evaluate(f, realm)
	}
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
		evaluate(f, realm)
	}
}

func TestHarness(t *testing.T) {
	files := []string{
		"sta.js",
		"assert.js",
		"isConstructor.js",
		"nans.js",
		"assertRelativeDateMs.js",
		"propertyHelper.js",
	}

	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	for _, f := range files {
		runTestHarness(realm, f, false)
	}

	// testArray(realm)
	testDataView(realm)
	testTypedArray(realm)

	Debug.Enable()
	// testBoolean(realm)
	Debug.Disable()
}
