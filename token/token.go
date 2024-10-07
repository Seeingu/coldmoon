package token

import "github.com/samber/lo"

type Token struct {
	TokenType TokenType
	Literal   string
	Line      uint
	Col       uint
}

func NewToken(tokenType TokenType, literal string, line uint, col uint) Token {
	return Token{
		TokenType: tokenType,
		Literal:   literal,
		Line:      line,
		Col:       col,
	}
}

func (t Token) Is(tokenType TokenType) bool {
	return t.TokenType == tokenType
}

func (t Token) IsOneOf(tokenTypes []TokenType) bool {
	return lo.Contains(tokenTypes, t.TokenType)
}
