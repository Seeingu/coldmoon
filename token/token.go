package token

import "github.com/samber/lo"

type Token struct {
	TokenType TokenType
	Literal   string
}

func NewToken(tokenType TokenType, literal string) Token {
	return Token{
		TokenType: tokenType,
		Literal:   literal,
	}
}

func (t Token) Is(tokenType TokenType) bool {
	return t.TokenType == tokenType
}

func (t Token) IsOneOf(tokenTypes []TokenType) bool {
	return lo.Contains(tokenTypes, t.TokenType)
}
