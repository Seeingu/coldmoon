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
	token      t.Token
	identifier IdentifierExpression
	value      Expression
}

type AssignExpression struct {
	Expression
	left  Expression
	right Expression
}

type IdentifierExpression struct {
	Expression
	name t.Token
}

func (i IdentifierExpression) toString() string {
	return i.name.Literal
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
	then       BlockExpression
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

type BlockExpression struct {
	Expression
	expressions []Expression
}

type ArrowFunctionExpression struct {
	Expression
	args []Expression
	body BlockExpression
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
	body       BlockExpression
}

type WhileExpression struct {
	Expression
	condition Expression
	body      BlockExpression
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

type FunctionExpression struct {
	Expression
	name string
	args []Expression
	body BlockExpression
}

type ChainExpression struct {
	Expression
	identifier Expression
	properties []Expression
}

type CallExpression struct {
	Expression
	caller Expression
	args   []Expression
}

type NativeFunctionExpression struct {
	Expression
	fn func(...Object)
}

type ThrowExpression struct {
	Expression
	errorExpression Expression
}

type ClassInstantiateExpression struct {
	Expression
	caller Expression
	args   []Expression
}

type ClassExpression struct {
	Expression
	name        Expression
	constructor Expression
	args        []Expression
}

type SingleLineCommentExpression struct {
	Expression
	content string
}

type MultiLineCommentExpression struct {
	Expression
	content string
	// TODO: Doc comment
}

type ThisExpression struct {
	Expression
}
