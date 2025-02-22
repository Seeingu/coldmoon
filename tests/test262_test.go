package tests

import (
	"os"
	"path"
	"testing"

	"github.com/Seeingu/coldmoon/runtime"

	"github.com/Seeingu/coldmoon/pkg"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func makeTest262Path(p string) string {
	dir, _ := os.Getwd()
	return path.Join(dir, "..", p)
}

func runTestHarness(realm *Realm, f string, debug bool) {
	if debug {
		Debug.Enable()
	}
	content := pkg.MustReadFile(makeTest262Path("./test262/harness/" + f))
	v := ParseScript(content, realm, nil).Evaluate()
	defer func() {
		Debug.Disable()
	}()
	if debug {
		print(v.String())
	}
}

func evaluateFile(fileName string, realm *Realm) {
	Evaluate(pkg.MustReadFile(fileName), realm)
}

func readDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		files = append(files, path.Join(dir, entry.Name()))
	}
	return files
}

func testDataView(realm *Realm) {
	dataViewDir := "./test262/test/built-ins/DataView/"
	f := makeTest262Path(dataViewDir + "constructor.js")
	evaluateFile(f, realm)
}

func testTypedArray(realm *Realm) {
	typedArrayDir := "./test262/test/built-ins/TypedArray/"
	subdirs := []string{
		"Symbol.species",
	}

	for _, subdir := range subdirs {
		d := typedArrayDir + subdir
		for _, f := range readDir(makeTest262Path(d)) {
			println("Testing file: ", f)
			evaluateFile(f, realm)
		}
	}
}

func testSharedArrayBuffer(realm *Realm) {
	sharedArrayBufferDir := "./test262/test/built-ins/SharedArrayBuffer"
	files := []string{
		"is-a-constructor.js",
		"length.js",
		"prototype/constructor.js",
		"prototype/Symbol.toStringTag.js",
	}
	for _, f := range files {
		f = makeTest262Path(path.Join(sharedArrayBufferDir, f))
		println("Testing file: ", f)
		evaluateFile(f, realm)
	}
}

func testTypedArrayName(realm *Realm) {
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
			println("Testing file: ", f)
			evaluateFile(f, realm)
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
		evaluateFile(f, realm)
	}
}

func testArray(realm *Realm) {
	arrayDir := "./test262/test/built-ins/Array/"
	subdirs := []string{
		"Symbol.species",
	}
	for _, subdir := range subdirs {
		d := arrayDir + subdir
		for _, f := range readDir(makeTest262Path(d)) {
			println("Testing file: ", f)
			evaluateFile(f, realm)
		}
	}
}

func TestHarness(t *testing.T) {
	// FIXME: it is not working with new VM
	t.Skip("")
	agent := NewAgent()
	InitializeConstants()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	runtime.RegisterTest262Runtime(realm)

	testArray(realm)
	testDataView(realm)
	testTypedArrayName(realm)
	testTypedArray(realm)

	Debug.Enable()
	testSharedArrayBuffer(realm)
	// testBoolean(realm)
	Debug.Disable()
}
