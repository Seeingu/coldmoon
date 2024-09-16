package parser

import t "github.com/Seeingu/coldmoon/token"

type Expression interface {
	toString() string
}

type LetExpression struct {
	Expression
	identifier IdentifierExpression
	value      Expression
}

type ConstExpression struct {
	Expression
	identifier IdentifierExpression
	value      Expression
}

type IdentifierExpression struct {
	Expression
	name t.Token
}

type VarExpression struct {
	Expression
	identifier IdentifierExpression
	value      Expression
}

type NumberExpression struct {
	Expression
	value int
}

type UnaryExpression struct {
	Expression
	unary t.Token
	value Expression
}

type BooleanExpression struct {
	Expression
	value bool
}
type StringExpression struct {
	Expression
	value string
}

type NullExpression struct {
	Expression
}

type UndefinedExpression struct {
	Expression
}

type IfExpression struct {
	Expression
	condition  Expression
	then       Block
	elseBranch Expression
}

func (i IfExpression) toString() string {
	return "if"
}

type BinaryExpression struct {
	Expression
	left     Expression
	right    Expression
	operator t.Token
}

type Block struct {
	Expression
	expressions []Expression
}

type FunctionExpression struct {
	Expression
	name t.Token
	args []Expression
	body Block
}

type ArrowFunctionExpression struct {
	Expression
	args []Expression
	body Block
}

type ReturnExpression struct {
	Expression
	value Expression
}

type ForExpression struct {
	Expression
	init       Expression
	step       Expression
	comparison Expression
	body       Block
}

type WhileExpression struct {
	Expression
	condition Expression
	body      Block
}

type ArrayLiteralExpression struct {
	Expression
	elements []Expression
}

type PairExpression struct {
	left  Expression
	right Expression
}
type ObjectLiteralExpression struct {
	Expression
	pairs []PairExpression
}
