package runtime

import "github.com/Seeingu/coldmoon/coldmoon"

// RegisterTerminalRuntime
// - console
func RegisterTerminalRuntime(realm *coldmoon.Realm) {
	global := realm.GlobalObject
	console := CreateConsole(realm)
	global.CreateDataProperty(coldmoon.PString("console").ToPropertyKey(), console.ToValue())

	// --- Non standard---
	definePrint(realm)
}
