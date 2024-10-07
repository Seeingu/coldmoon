package tests

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func makeTest262Path(p string) string {
	dir, _ := os.Getwd()
	return path.Join(dir, "..", p)
}

func TestParseSta(t *testing.T) {
	filePath := makeTest262Path("./test262/harness/sta.js")
	_, err := os.ReadFile(filePath)
	assert.Nil(t, err)
}
