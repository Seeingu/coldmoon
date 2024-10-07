package _parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func parseSource(s string) ([]Expression, error) {
	ss := NewScanner(s)
	p := NewParser(ss)
	return p.Parse()
}

func TestParserError(t *testing.T) {
	source := "throw new Error('exception')"
	_, err := parseSource(source)
	assert.NoError(t, err)
}
