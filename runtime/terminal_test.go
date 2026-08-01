package runtime

import (
	"bytes"
	"os"
	"testing"

	"github.com/Seeingu/coldmoon/coldmoon"
)

func TestRegisterTerminalRuntimeWithOptionsRoutesOutput(t *testing.T) {
	realm := newConsoleTestRealm(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	RegisterTerminalRuntimeWithOptions(realm, TerminalOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	consoleValue := realm.GlobalObject.Get(coldmoon.NewStringPropertyKey("console"))
	console, ok := consoleValue.GetObject()
	if !ok {
		t.Fatal("global console is not an object")
	}
	callConsoleMethod(t, realm, console, "log", coldmoon.NewStringValue("hello"))
	callConsoleMethod(t, realm, console, "warn", coldmoon.NewStringValue("danger"))

	print := realm.GlobalObject.Get(coldmoon.NewStringPropertyKey("print"))
	result := print.Call(realm.Agent, coldmoon.UndefinedValue, []coldmoon.Value{
		coldmoon.NewStringValue("value"),
		coldmoon.NewNumberValue(7),
	})
	if result.IsAbrupt() {
		t.Fatalf("print returned abrupt completion: %v", result.Error())
	}

	if got, want := stdout.String(), "hello\nvalue 7\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got, want := stderr.String(), "[warn] danger\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestRegisterTerminalRuntimeWithOptionsExposesMutableArgvArray(t *testing.T) {
	realm := newConsoleTestRealm(t)
	want := []string{"/bin/coldmoon", "α.js", "--flag", ""}

	RegisterTerminalRuntimeWithOptions(realm, TerminalOptions{Argv: want})

	processValue := realm.GlobalObject.Get(coldmoon.NewStringPropertyKey("process"))
	process, ok := processValue.GetObject()
	if !ok {
		t.Fatal("global process is not an object")
	}
	argvValue := process.Get(coldmoon.NewStringPropertyKey("argv"))
	if !coldmoon.IsArray(argvValue) {
		t.Fatalf("process.argv is %s, want Array", argvValue.TypeString())
	}
	argv, ok := argvValue.GetObject()
	if !ok {
		t.Fatal("process.argv has no object value")
	}
	length := argv.LengthOfArrayLike()
	if length.IsAbrupt() {
		t.Fatalf("read process.argv.length: %v", length.Error())
	}
	if got := int(length.Data()); got != len(want) {
		t.Fatalf("process.argv.length = %d, want %d", got, len(want))
	}
	for index, expected := range want {
		value := argv.Get(coldmoon.NewIntegerIndexPropertyKey(coldmoon.JSInt(index)))
		if got := string(value.ToString()); got != expected {
			t.Errorf("process.argv[%d] = %q, want %q", index, got, expected)
		}
	}

	if ok := argv.CreateDataProperty(
		coldmoon.NewIntegerIndexPropertyKey(1),
		coldmoon.NewStringValue("changed.js"),
	); !ok {
		t.Fatal("process.argv element is not mutable")
	}
	if got := string(argv.Get(coldmoon.NewIntegerIndexPropertyKey(1)).ToString()); got != "changed.js" {
		t.Fatalf("mutated process.argv[1] = %q, want %q", got, "changed.js")
	}
}

func TestTerminalProcessOnlyProvidesArgv(t *testing.T) {
	realm := newConsoleTestRealm(t)
	RegisterTerminalRuntimeWithOptions(realm, TerminalOptions{Argv: []string{"coldmoon"}})

	processValue := realm.GlobalObject.Get(coldmoon.NewStringPropertyKey("process"))
	process, ok := processValue.GetObject()
	if !ok {
		t.Fatal("global process is not an object")
	}
	keys := coldmoon.InternalOwnPropertyKeys(process)
	if len(keys) != 1 || string(keys[0].ToValue().ToString()) != "argv" {
		t.Fatalf("process own properties = %v, want only argv", keys)
	}
	for _, name := range []string{"env", "cwd", "exit", "version", "versions"} {
		if value := process.Get(coldmoon.NewStringPropertyKey(name)); value != coldmoon.UndefinedValue {
			t.Errorf("process.%s = %s, want undefined", name, value.String())
		}
	}
}

func TestFormatValueMatchesTerminalRendering(t *testing.T) {
	realm := newConsoleTestRealm(t)
	object := coldmoon.OrdinaryObjectCreate(realm.Agent, realm.Intrinsics.ObjectPrototype, nil)
	object.CreateDataProperty(coldmoon.NewStringPropertyKey("answer"), coldmoon.NewNumberValue(42))

	tests := []struct {
		name  string
		value coldmoon.Value
		want  string
	}{
		{name: "missing", value: nil, want: "undefined"},
		{name: "undefined", value: coldmoon.UndefinedValue, want: "undefined"},
		{name: "string", value: coldmoon.NewStringValue("月"), want: "月"},
		{name: "number", value: coldmoon.NewNumberValue(7), want: "7"},
		{name: "boolean", value: coldmoon.FalseValue, want: "false"},
		{name: "object", value: object.ToValue(), want: object.String()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatValue(test.value); got != test.want {
				t.Fatalf("FormatValue() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRegisterTerminalRuntimeUsesOSDefaults(t *testing.T) {
	realm := newConsoleTestRealm(t)
	RegisterTerminalRuntime(realm)

	console := realm.GlobalObject.Get(coldmoon.NewStringPropertyKey("console"))
	if _, ok := console.GetObject(); !ok {
		t.Fatal("legacy registration did not install console")
	}
	for _, name := range []string{"print", "setTimeout", "assert"} {
		value := realm.GlobalObject.Get(coldmoon.NewStringPropertyKey(name))
		if !coldmoon.IsCallable(value) {
			t.Errorf("legacy registration did not install callable %s", name)
		}
	}

	process, ok := realm.GlobalObject.Get(coldmoon.NewStringPropertyKey("process")).GetObject()
	if !ok {
		t.Fatal("legacy registration did not install process")
	}
	argv, ok := process.Get(coldmoon.NewStringPropertyKey("argv")).GetObject()
	if !ok {
		t.Fatal("legacy process.argv is not an object")
	}
	length := argv.LengthOfArrayLike()
	if length.IsAbrupt() {
		t.Fatalf("read legacy process.argv.length: %v", length.Error())
	}
	if got := int(length.Data()); got != len(os.Args) {
		t.Fatalf("legacy process.argv.length = %d, want %d", got, len(os.Args))
	}
	for index, expected := range os.Args {
		value := argv.Get(coldmoon.NewIntegerIndexPropertyKey(coldmoon.JSInt(index)))
		if got := string(value.ToString()); got != expected {
			t.Errorf("legacy process.argv[%d] = %q, want %q", index, got, expected)
		}
	}
}
