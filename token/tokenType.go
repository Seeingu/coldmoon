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
	If
	Else
	Return
	For
	While
	Object
	Function
	Comma
	Dot
	DotDotDot
	Equal
	EqualEqual
	EqualEqualEqual
	LeftParenthesis
	RightParenthesis
	LeftBracket
	RightBracket
	LeftSquareBracket
	RightSquareBracket
)
