package ast

import (
	t "github.com/Seeingu/coldmoon/token"
)

// MARK: Interface

type JSNode interface {
	String() string
}

type Statement interface {
	JSNode
}

type Expression interface {
	JSNode
}
type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	return ""
}

// MARK: Statement

type ExpressionStatement struct {
	Statement
	Expression Expression
}

type BlockStatement struct {
	Statement
	Token      t.Token
	Statements []Statement
}

type LetStatement struct {
	Statement
	Token t.Token
	Name  *IdentifierExpression
	Value Expression
}

type IntegerLiteral struct {
	Expression
	Token t.Token
	Value int64
}

type StringLiteral struct {
	Expression
	Token t.Token
	Value string
}

type BooleanExpression struct {
	Expression
	Token t.Token
	Value bool
}

type InfixExpression struct {
	Expression
	Token    t.Token
	Left     Expression
	Operator string
	Right    Expression
}

type IdentifierExpression struct {
	Expression
	Token t.Token
	Value string
}

type ArrayLiteralExpression struct {
	Expression
	Token    t.Token
	Elements []Expression
}

type IndexExpression struct {
	Expression
	Token t.Token
	Left  Expression
	Index Expression
}

type ObjectLiteralExpression struct {
	Expression
	Token t.Token
	Pairs map[Expression]Expression
}

type PrefixExpression struct {
	Expression
	Token    t.Token
	Operator string
	Right    Expression
}

type IfExpression struct {
	Expression
	Token       t.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}
