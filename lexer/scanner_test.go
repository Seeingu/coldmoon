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
		tt.Let,
		tt.Identifier,
		// index: 10
		tt.Equal,
		tt.LeftBracket,
		tt.String,
		tt.Colon,
		tt.Number,
		tt.Comma,
	}
	for i, tokenType := range tokenTypes {
		assert.Equal(t, tokenType, tokens[i].TokenType, fmt.Sprintf("index: %d, token: %s", i, tokens[i].TokenType.String()))
	}
}
