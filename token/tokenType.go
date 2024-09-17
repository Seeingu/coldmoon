package token

type TokenType int

const (
	Var TokenType = iota
	Const
	Let
	Number
	String
	Boolean
	Null
	Undefined
	True
	False
	Identifier
	If
	Else
	Return
	For
	While
	Object
	Function
	Comma
	Colon
	Semicolon
	Plus
	PlusPlus
	Minus
	MinusMinus
	PlusEqual
	MinusEqual
	Star
	StarEqual
	StarStar
	Slash
	SlashSlash
	SlashEqual
	SlashStar
	Question
	QuestionDot
	Tilde
	Dot
	DotDotDot
	Bang
	BangEqual
	Equal
	EqualEqual
	EqualEqualEqual
	// EqualGreater =>
	EqualGreater
	Greater
	GreaterEqual
	GreaterGreater
	GreaterGreaterGreater
	GreaterGreaterEqual
	Less
	LessLess
	LessLessLess
	LessEqual
	LessLessEqual
	LeftParenthesis
	RightParenthesis
	LeftBracket
	RightBracket
	LeftSquareBracket
	RightSquareBracket
	EOF
)
