package token

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
