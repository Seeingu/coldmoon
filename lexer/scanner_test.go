package lexer

import (
	"fmt"
	tt "github.com/Seeingu/coldmoon/token"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScannerTokens(t *testing.T) {
	s := NewScanner(`
(1 + 2) 
< 
"hello";
let o = {a: 1, 'b': 2, "c": 3}

	let sum = function(a, b) { a + b }
	sum(1, 2)

`)
	var tokens []tt.Token
	tokens = append(tokens, s.currentToken)
	for !s.currentToken.Is(tt.EOF) {
		tokens = append(tokens, s.Scan())
	}
	tokenTypes := []tt.TokenType{
		tt.LeftParenthesis,
		tt.Number,
		tt.Plus,
		tt.Number,
		tt.RightParenthesis,
		tt.Less,
		tt.String,
		tt.Semicolon,
		// let o = ...
		tt.Let,
		tt.Identifier,
		// index: 10
		tt.Equal,
		tt.LeftBracket,
		tt.String,
		tt.Colon,
		tt.Number,
		tt.Comma,
		tt.String,
		tt.Colon,
		tt.Number,
		tt.Comma,
		// index: 20
		tt.String,
		tt.Colon,
		tt.Number,
		tt.RightBracket,
		// sum = ...
		tt.Let,
		tt.Identifier,
		tt.Equal,
		tt.Function,
		tt.LeftParenthesis,
		tt.Identifier,
		tt.Comma,
		tt.Identifier,
		tt.RightParenthesis,
		tt.LeftBracket,
		tt.Identifier,
		tt.Plus,
		tt.Identifier,
		tt.RightBracket,
		// sum()
		tt.Identifier,
		tt.LeftParenthesis,
		tt.Number,
		tt.Comma,
		tt.Number,
		tt.RightParenthesis,
	}
	for i, tokenType := range tokenTypes {
		assert.Equal(t, tokenType, tokens[i].TokenType, fmt.Sprintf("index: %d, expected: %s, actual: %s", i, tokenType.String(), tokens[i].TokenType.String()))
	}
}
