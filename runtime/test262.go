package runtime

import (
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/pkg"
)

var test262Path string
var test262IncludesPattern = regexp.MustCompile(`(?m)^includes:\s*\[([^\]]*)\]`)
var test262OnlyStrictPattern = regexp.MustCompile(`(?m)^flags:\s*\[[^\]]*\bonlyStrict\b`)
var loadedTest262Includes = struct {
	sync.Mutex
	byRealm map[*coldmoon.Realm]map[string]bool
}{
	byRealm: make(map[*coldmoon.Realm]map[string]bool),
}

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
	loadedTest262Includes.Lock()
	loaded := make(map[string]bool, len(files))
	for _, file := range files {
		loaded[file] = true
	}
	loadedTest262Includes.byRealm[realm] = loaded
	loadedTest262Includes.Unlock()
}

// ReleaseTest262Runtime removes per-realm harness bookkeeping after a test.
func ReleaseTest262Runtime(realm *coldmoon.Realm) {
	loadedTest262Includes.Lock()
	delete(loadedTest262Includes.byRealm, realm)
	loadedTest262Includes.Unlock()
}

// RegisterTest262Includes evaluates the optional harness files declared by a
// test's frontmatter. Harnesses are loaded once per realm because many of them
// declare top-level lexical bindings.
func RegisterTest262Includes(realm *coldmoon.Realm, source string) {
	match := test262IncludesPattern.FindStringSubmatch(source)
	if len(match) != 2 {
		return
	}
	for _, entry := range strings.Split(match[1], ",") {
		file := strings.Trim(strings.TrimSpace(entry), `"'`)
		if file == "" {
			continue
		}

		loadedTest262Includes.Lock()
		loaded := loadedTest262Includes.byRealm[realm]
		if loaded == nil {
			loaded = make(map[string]bool)
			loadedTest262Includes.byRealm[realm] = loaded
		}
		alreadyLoaded := loaded[file]
		loadedTest262Includes.Unlock()
		if alreadyLoaded {
			continue
		}

		println("Harness file: ", file)
		content := pkg.MustReadFile(MakeTest262Path("./harness/" + file))
		coldmoon.ParseScript(content, realm, nil).Evaluate()
		loadedTest262Includes.Lock()
		loadedTest262Includes.byRealm[realm][file] = true
		loadedTest262Includes.Unlock()
	}
}

// PrepareTest262Source applies execution-mode flags from test262 frontmatter.
func PrepareTest262Source(source string) string {
	if test262OnlyStrictPattern.MatchString(source) {
		return "\"use strict\";\n" + source
	}
	return source
}
