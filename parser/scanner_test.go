package parser

import (
	tt "github.com/Seeingu/coldmoon/token"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScannerTokens(t *testing.T) {
	s := NewScanner("print('hello world')")
	var tokens []tt.Token
	for s.HasNextToken() {
		tokens = append(tokens, s.Scan())
	}
	expected := []tt.Token{
		{TokenType: tt.Identifier, Literal: "print"},
		{TokenType: tt.LeftParenthesis, Literal: "("},
		{TokenType: tt.String, Literal: "hello world"},
		{TokenType: tt.RightParenthesis, Literal: ")"},
	}
	assert.Equal(t, expected, tokens)
}
