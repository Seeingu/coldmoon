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

func makeTest262Path(p string) string {
	initPath()
	return path.Join(test262Path, p)
}

func RegisterTest262Runtime(realm *coldmoon.Realm) {
	files := []string{
		"sta.js",
		"assert.js",
		"isConstructor.js",
		"nans.js",
		"assertRelativeDateMs.js",
		"propertyHelper.js",
		"testTypedArray.js",
		"testAtomics.js",
	}
	for _, f := range files {
		println("Harness file: ", f)
		content := pkg.MustReadFile(makeTest262Path("./harness/" + f))
		coldmoon.ParseScript(content, realm, nil).Evaluate()
	}
}
