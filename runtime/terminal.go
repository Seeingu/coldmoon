package runtime

import "github.com/Seeingu/coldmoon/coldmoon"

// RegisterTerminalRuntime
// - console
func RegisterTerminalRuntime(realm *coldmoon.Realm) {
	global := realm.GlobalObject
	console := CreateConsole(realm)
	global.CreateDataProperty(coldmoon.CMString("console").ToPropertyKey(), console.ToValue())

	// --- Non standard---
	defineAssert(realm)
	definePrint(realm)
}
