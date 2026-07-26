package runtime

import "github.com/Seeingu/coldmoon/coldmoon"

// RegisterTerminalRuntime installs the terminal host bindings for realm.
func RegisterTerminalRuntime(realm *coldmoon.Realm) {
	RegisterFilesystemModuleLoader(realm)
	global := realm.GlobalObject
	console := CreateConsole(realm)
	global.CreateDataProperty(coldmoon.CMString("console").ToPropertyKey(), console.ToValue())
	CreateSetTimeout(realm)

	// --- Non standard---
	defineAssert(realm)
	definePrint(realm)
}
