package tests

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeTest262Path(p string) string {
	dir, _ := os.Getwd()
	return path.Join(dir, "..", p)
}

func TestParseSta(t *testing.T) {
	t.Skip()
	filePath := makeTest262Path("./test262/harness/sta.js")
	_, err := os.ReadFile(filePath)
	assert.Nil(t, err)
}
