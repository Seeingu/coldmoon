package runtime

import (
	"os"
	"path"

	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/pkg"
)

var test262Path string

func initPath() {
	if test262Path == "" {
		dir, _ := os.Getwd()
		test262Path = path.Join(dir, "..", "test262")
		if _, err := os.Stat(test262Path); os.IsNotExist(err) {
			test262Path = path.Join(dir, "test262")
		}
	}
}

func MakeTest262Path(p string) string {
	initPath()
	return path.Join(test262Path, p)
}

func GetTest262Path() string {
	initPath()
	return test262Path
}

func RegisterTest262Runtime(realm *coldmoon.Realm) {
	global := realm.GlobalObject
	console := CreateConsole(realm)
	global.CreateDataProperty(coldmoon.CMString("console").ToPropertyKey(), console.ToValue())

	test262 := coldmoon.OrdinaryObjectCreate(realm.Agent, realm.Intrinsics.ObjectPrototype, nil)
	var createRealm coldmoon.BehaviorFn = func(
		_ coldmoon.Value,
		_ []coldmoon.Value,
		_ coldmoon.ObjectType,
	) coldmoon.CompletionConvertable[coldmoon.Value] {
		otherRealm := coldmoon.CreateRealm(realm.Agent)
		otherGlobal := coldmoon.OrdinaryObjectCreate(
			realm.Agent,
			otherRealm.Intrinsics.ObjectPrototype,
			nil,
		)
		otherRealm.SetRealmGlobalObject(otherGlobal, nil)
		otherRealm.SetDefaultGlobalBindings()

		record := coldmoon.OrdinaryObjectCreate(
			realm.Agent,
			realm.Intrinsics.ObjectPrototype,
			nil,
		)
		record.CreateDataPropertyOrThrow(
			coldmoon.CMString("global").ToPropertyKey(),
			otherGlobal.ToValue(),
		)
		return record.ToValue()
	}
	coldmoon.DefineBuiltinFunction(
		realm,
		coldmoon.CMString("createRealm"),
		test262,
		createRealm,
		0,
	)
	global.CreateDataProperty(
		coldmoon.CMString("$262").ToPropertyKey(),
		test262.ToValue(),
	)

	files := []string{
		"sta.js",
		"assert.js",
		"isConstructor.js",
		"nans.js",
		"assertRelativeDateMs.js",
		"propertyHelper.js",
		"testTypedArray.js",
		"testAtomics.js",
		"compareArray.js",
	}
	for _, f := range files {
		println("Harness file: ", f)
		content := pkg.MustReadFile(MakeTest262Path("./harness/" + f))
		coldmoon.ParseScript(content, realm, nil).Evaluate()
	}
}
