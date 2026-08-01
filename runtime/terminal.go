package runtime

import (
	"io"
	"os"

	"github.com/Seeingu/coldmoon/coldmoon"
)

// TerminalOptions supplies the process-owned resources exposed by the terminal
// host bindings.
type TerminalOptions struct {
	Stdout io.Writer
	Stderr io.Writer
	Argv   []string
}

// RegisterTerminalRuntime installs the terminal host bindings for realm.
func RegisterTerminalRuntime(realm *coldmoon.Realm) {
	RegisterTerminalRuntimeWithOptions(realm, TerminalOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Argv:   os.Args,
	})
}

// RegisterTerminalRuntimeWithOptions installs terminal host bindings using
// resources supplied by the embedding host.
func RegisterTerminalRuntimeWithOptions(realm *coldmoon.Realm, options TerminalOptions) {
	RegisterFilesystemModuleLoader(realm)
	global := realm.GlobalObject
	console := createConsole(realm, options.Stdout, options.Stderr, nil)
	global.CreateDataProperty(coldmoon.CMString("console").ToPropertyKey(), console.ToValue())
	process := coldmoon.OrdinaryObjectCreate(realm.Agent, realm.Intrinsics.ObjectPrototype, nil)
	argvValues := make([]coldmoon.Value, len(options.Argv))
	for index, argument := range options.Argv {
		argvValues[index] = coldmoon.NewStringValue(argument)
	}
	process.CreateDataProperty(
		coldmoon.NewStringPropertyKey("argv"),
		coldmoon.CreateArrayFromList(realm.Agent, argvValues).ToValue(),
	)
	global.CreateDataProperty(coldmoon.NewStringPropertyKey("process"), process.ToValue())
	CreateSetTimeout(realm)

	// --- Non standard---
	defineAssert(realm)
	definePrint(realm, options.Stdout)
}
