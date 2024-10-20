package ast

import (
	t "github.com/Seeingu/coldmoon/token"
	"strconv"
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
	var s string
	for _, stmt := range p.Statements {
		s += stmt.String() + ";\n"
	}
	return s
}

// MARK: Statement

type ExpressionStatement struct {
	Statement
	Expression Expression
}

var _ Expression = (*ExpressionStatement)(nil)

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Statement
	Token      t.Token
	Statements []Statement
}

var _ Statement = (*BlockStatement)(nil)

func (bs *BlockStatement) String() string {
	var s string
	s += "{"
	for _, stmt := range bs.Statements {
		s += stmt.String() + ";"
	}
	s += "}"
	return s
}

type LetStatement struct {
	Statement
	Token t.Token
	Name  *IdentifierExpression
	Value Expression
}

func (ls *LetStatement) String() string {
	var s string
	s += "let "
	s += ls.Name.String()

	if ls.Value != nil {
		s += " = "
		s += ls.Value.String()
	}
	return s
}

type ReturnStatement struct {
	Statement
	Token       t.Token
	ReturnValue Expression
}

type IntegerLiteral struct {
	Expression
	Token t.Token
	Value int64
}

var _ Expression = (*IntegerLiteral)(nil)

func (il *IntegerLiteral) String() string {
	return strconv.FormatInt(il.Value, 10)
}

type StringLiteral struct {
	Expression
	Token t.Token
	Value string
}

// String
func (sl *StringLiteral) String() string {
	return sl.Value
}

var _ Expression = (*StringLiteral)(nil)

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

func (ie *IdentifierExpression) String() string {
	return ie.Value
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

type PropertyAccessExpression struct {
	Expression
	Token    t.Token
	Left     Expression
	Property Expression
}

var _ Expression = (*PropertyAccessExpression)(nil)

func (pae *PropertyAccessExpression) String() string {
	return pae.Left.String() + "." + pae.Property.String()
}

type AssignmentExpression struct {
	Expression
	Token t.Token
	Left  Expression
	Value Expression
}

var _ Expression = (*AssignmentExpression)(nil)

func (ae *AssignmentExpression) String() string {
	return ae.Left.String() + " = " + ae.Value.String()
}

type ObjectLiteralExpression struct {
	Expression
	Token t.Token
	Pairs map[Expression]Expression
}

var _ Expression = (*ObjectLiteralExpression)(nil)

func (oe *ObjectLiteralExpression) String() string {
	var s string
	s += "{"
	for k, v := range oe.Pairs {
		s += k.String() + ":" + v.String() + ","
	}
	s += "}"
	return s
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

type funType int

const (
	FAnonymous funType = iota
	FLiteral
)

type FunctionLiteral struct {
	Expression
	Token t.Token
	// Name is optional, maybe not exist in anonymous function
	Name       *IdentifierExpression
	Parameters []*IdentifierExpression
	Body       *BlockStatement
	FunType    funType
}

var _ Expression = (*FunctionLiteral)(nil)

func (fl *FunctionLiteral) String() string {
	var s string
	s = "function"
	if fl.Name != nil {
		s += " " + fl.Name.String()
	}
	s += "("
	for i, p := range fl.Parameters {
		s += p.String()
		if i != len(fl.Parameters)-1 {
			s += ","
		}
	}
	s += ")"
	s += fl.Body.String()
	return s
}

type CallExpression struct {
	Expression
	Token        t.Token
	FunctionName Expression
	Arguments    []Expression
}

var _ Expression = (*CallExpression)(nil)

// String
func (ce *CallExpression) String() string {
	var s string
	s = ce.FunctionName.String() + "("
	for i, arg := range ce.Arguments {
		s += arg.String()
		if i != len(ce.Arguments)-1 {
			s += ","
		}
	}
	s += ")"
	return s
}
