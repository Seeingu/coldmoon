package parser

import (
	"fmt"
	"github.com/Seeingu/coldmoon/ast"
	"github.com/Seeingu/coldmoon/lexer"
	t "github.com/Seeingu/coldmoon/token"
	"strconv"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	scanner *lexer.Scanner
	errors  []string

	prefixParseFns map[t.TokenType]prefixParseFn
	infixParseFns  map[t.TokenType]infixParseFn
}

func New(l *lexer.Scanner) *Parser {
	p := &Parser{
		scanner: l,
	}
	p.prefixParseFns = make(map[t.TokenType]prefixParseFn)
	p.registerPrefix(t.Identifier, p.parseIdentifier)
	p.registerPrefix(t.Number, p.parseIntegerLiteral)
	p.registerPrefix(t.String, p.parseStringLiteral)
	p.registerPrefix(t.True, p.parseBoolean)
	p.registerPrefix(t.False, p.parseBoolean)
	p.registerPrefix(t.LeftParenthesis, p.parseGroupedExpression)
	p.registerPrefix(t.LeftSquareBracket, p.parseArrayLiteral)
	p.registerPrefix(t.LeftBracket, p.parseObjectLiteral)
	p.registerPrefix(t.Minus, p.parsePrefixExpression)
	p.registerPrefix(t.Bang, p.parsePrefixExpression)
	p.registerPrefix(t.If, p.parseIfExpression)
	p.registerPrefix(t.Function, p.parseFunctionLiteral)

	p.infixParseFns = make(map[t.TokenType]infixParseFn)
	p.registerInfix(t.Plus, p.parseInfixExpression)
	p.registerInfix(t.Minus, p.parseInfixExpression)
	p.registerInfix(t.Star, p.parseInfixExpression)
	p.registerInfix(t.Slash, p.parseInfixExpression)
	p.registerInfix(t.Less, p.parseInfixExpression)
	p.registerInfix(t.Greater, p.parseInfixExpression)
	p.registerInfix(t.EqualEqual, p.parseInfixExpression)
	p.registerInfix(t.GreaterEqual, p.parseInfixExpression)
	p.registerInfix(t.LessEqual, p.parseInfixExpression)
	p.registerInfix(t.BangEqual, p.parseInfixExpression)
	p.registerInfix(t.LeftSquareBracket, p.parseIndexExpression)
	p.registerInfix(t.LeftParenthesis, p.parseCallExpression)
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	for !p.currentToken().Is(t.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.scanner.Scan()
	}
	return program

}

// MARK: Private

type precedenceType int

const (
	_ precedenceType = iota
	PLowest
	PEquals
	PLessOrGreater
	PSum
	PProduct
	PPrefix
	PCall
	PIndex
)

var precedences = map[t.TokenType]precedenceType{
	t.EqualEqual:        PEquals,
	t.EqualEqualEqual:   PEquals,
	t.BangEqual:         PEquals,
	t.LessEqual:         PLessOrGreater,
	t.Less:              PLessOrGreater,
	t.Greater:           PLessOrGreater,
	t.Plus:              PSum,
	t.Minus:             PSum,
	t.Star:              PProduct,
	t.Slash:             PProduct,
	t.LeftParenthesis:   PCall,
	t.LeftSquareBracket: PIndex,
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken().TokenType {
	case t.Let:
		return p.parseLetStatement()
	case t.Return:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.currentToken()}

	if !p.expectNextToken(t.Identifier) {
		return nil
	}

	stmt.Name = &ast.IdentifierExpression{Token: p.currentToken(), Value: p.currentToken().Literal}

	if !p.expectNextToken(t.Equal) {
		return nil
	}
	p.scanner.Scan()

	stmt.Value = p.parseExpression(PLowest)

	if fn, ok := stmt.Value.(*ast.FunctionLiteral); ok {
		fn.Name = stmt.Name
	}

	if p.nextToken().Is(t.Semicolon) {
		p.scanner.Scan()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.currentToken()}
	// skip return
	p.scanner.Scan()

	stmt.ReturnValue = p.parseExpression(PLowest)
	if p.nextToken().Is(t.Semicolon) {
		p.scanner.Scan()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{}
	stmt.Expression = p.parseExpression(PLowest)

	if p.nextToken().Is(t.Semicolon) {
		p.scanner.Scan()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence precedenceType) ast.Expression {
	prefixFn := p.prefixParseFns[p.currentToken().TokenType]
	if prefixFn == nil {
		p.noPrefixParseFnError(p.currentToken().TokenType)
		return nil
	}
	leftExp := prefixFn()
	for !p.nextToken().Is(t.Semicolon) && precedence < p.nextTokenPrecedence() {
		infix := p.infixParseFns[p.nextToken().TokenType]
		if infix == nil {
			return leftExp
		}

		p.scanner.Scan()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.IdentifierExpression{
		Token: p.currentToken(),
		Value: p.currentToken().Literal,
	}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	literal := &ast.StringLiteral{Token: p.currentToken()}
	literal.Value = p.currentToken().Literal
	return literal
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{}
	value, err := strconv.ParseInt(p.currentToken().Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currentToken().Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	literal.Value = value

	return literal

}

func (p *Parser) parseArrayLiteral() ast.Expression {
	literal := &ast.ArrayLiteralExpression{Token: p.currentToken()}
	literal.Elements = p.parseExpressionList(t.RightSquareBracket)
	return literal
}

func (p *Parser) parseObjectLiteral() ast.Expression {
	o := &ast.ObjectLiteralExpression{Token: p.currentToken()}
	o.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.nextToken().Is(t.RightBracket) {
		// skip { or ,
		p.scanner.Scan()

		key := p.parseExpression(PLowest)

		if !p.expectNextToken(t.Colon) {
			return nil
		}
		// skip :
		p.scanner.Scan()

		value := p.parseExpression(PLowest)
		o.Pairs[key] = value

		if !p.nextToken().Is(t.RightBracket) && !p.expectNextToken(t.Comma) {
			return nil
		}
	}

	if !p.expectNextToken(t.RightBracket) {
		return nil
	}

	return o
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanExpression{
		Token: p.currentToken(),
		Value: p.currentToken().Is(t.True),
	}
}

func (p *Parser) parseExpressionList(end t.TokenType) []ast.Expression {
	var list []ast.Expression

	if p.nextToken().Is(end) {
		p.scanner.Scan()
		return list
	}

	p.scanner.Scan()
	list = append(list, p.parseExpression(PLowest))

	for p.nextToken().Is(t.Comma) {
		p.scanner.Scan()
		p.scanner.Scan()
		list = append(list, p.parseExpression(PLowest))
	}

	if !p.expectNextToken(end) {
		return nil
	}

	return list

}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.scanner.Scan()

	exp := p.parseExpression(PLowest)
	if !p.expectNextToken(t.RightParenthesis) {
		return nil
	}
	return exp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	e := &ast.PrefixExpression{
		Token:    p.currentToken(),
		Operator: p.currentToken().Literal,
	}

	p.scanner.Scan()

	e.Right = p.parseExpression(PPrefix)
	return e
}

func (p *Parser) parseIfExpression() ast.Expression {
	e := &ast.IfExpression{
		Token: p.currentToken(),
	}

	if !p.expectNextToken(t.LeftParenthesis) {
		return nil
	}

	p.scanner.Scan()
	e.Condition = p.parseExpression(PLowest)

	if !p.expectNextToken(t.RightParenthesis) {
		return nil
	}

	if !p.expectNextToken(t.LeftBracket) {
		return nil
	}

	e.Consequence = p.parseBlockStatement()

	if p.nextToken().Is(t.Else) {
		p.scanner.Scan()

		if !p.expectNextToken(t.LeftBracket) {
			return nil
		}

		e.Alternative = p.parseBlockStatement()
	}

	return e
}

// function <identifier> params block
func (p *Parser) parseFunctionLiteral() ast.Expression {
	f := &ast.FunctionLiteral{Token: p.currentToken()}

	if !p.nextToken().Is(t.LeftParenthesis) {
		p.scanner.Scan()
		f.Name = p.parseIdentifier().(*ast.IdentifierExpression)
	}

	if !p.expectNextToken(t.LeftParenthesis) {
		return nil
	}
	f.Parameters = p.parseFunctionParameters()

	if !p.expectNextToken(t.LeftBracket) {
		return nil
	}

	f.Body = p.parseBlockStatement()

	return f
}

// startToken: (
// endToken: after )
func (p *Parser) parseFunctionParameters() []*ast.IdentifierExpression {
	var params []*ast.IdentifierExpression

	if p.nextToken().Is(t.RightParenthesis) {
		p.scanner.Scan()
		return params
	}

	// Skip (
	p.scanner.Scan()

	param := &ast.IdentifierExpression{Token: p.currentToken(), Value: p.currentToken().Literal}
	params = append(params, param)

	for p.nextToken().Is(t.Comma) {
		// Skip current literal
		p.scanner.Scan()
		// Skip ,
		p.scanner.Scan()
		param := &ast.IdentifierExpression{Token: p.currentToken(), Value: p.currentToken().Literal}
		params = append(params, param)
	}

	if !p.expectNextToken(t.RightParenthesis) {
		return nil
	}

	return params
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	b := &ast.BlockStatement{Token: p.currentToken()}
	b.Statements = []ast.Statement{}

	p.scanner.Scan()

	for !p.currentToken().Is(t.RightBracket) && !p.currentToken().Is(t.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			b.Statements = append(b.Statements, stmt)
		}
		p.scanner.Scan()
	}

	return b
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	e := &ast.InfixExpression{
		Token:    p.currentToken(),
		Operator: p.currentToken().Literal,
		Left:     left,
	}

	precedence := p.currentPrecedence()
	p.scanner.Scan()

	e.Right = p.parseExpression(precedence)

	return e

}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	e := &ast.IndexExpression{Token: p.currentToken(), Left: left}
	p.scanner.Scan()

	e.Index = p.parseExpression(PLowest)

	if !p.expectNextToken(t.RightSquareBracket) {
		return nil
	}

	return e
}

func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	e := &ast.CallExpression{Token: p.currentToken()}
	e.FunctionName = fn
	e.Arguments = p.parseExpressionList(t.RightParenthesis)
	return e
}

func (p *Parser) currentToken() t.Token {
	return p.scanner.CurrentToken()
}

func (p *Parser) nextToken() t.Token {
	return p.scanner.NextToken()
}

func (p *Parser) registerPrefix(tokenType t.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType t.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t t.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t.String())
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextTokenPrecedence() precedenceType {
	if p, ok := precedences[p.nextToken().TokenType]; ok {
		return p
	}

	return PLowest
}

func (p *Parser) currentPrecedence() precedenceType {
	if p, ok := precedences[p.currentToken().TokenType]; ok {
		return p
	}

	return PLowest
}

func (p *Parser) matchToken(tokenType t.TokenType) (ok bool) {
	if !p.scanner.CurrentToken().Is(tokenType) {
		return false
	}
	p.scanner.Scan()
	return true
}

func (p *Parser) matchNextToken(tokenType t.TokenType) (ok bool) {
	if !p.scanner.NextToken().Is(tokenType) {
		return false
	}
	p.scanner.Scan()
	return true
}

func (p *Parser) expectNextToken(tokenType t.TokenType) bool {
	if !p.matchNextToken(tokenType) {
		p.tokenMatchError(p.nextToken(), tokenType)
		return false
	}
	return true
}

func (p *Parser) tokenMatchError(token t.Token, tokenType t.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf("expected match token %s, got %s", token.TokenType.String(), tokenType.String()))
}
