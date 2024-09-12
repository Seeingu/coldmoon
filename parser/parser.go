package parser

import (
	"errors"
	t "github.com/Seeingu/coldmoon/token"
	"log"
	"strconv"
)

type Parser struct {
	scanner *Scanner
}

func NewParser(scanner *Scanner) Parser {
	return Parser{
		scanner: scanner,
	}
}

// matchToken will consume token and return if matched
func (p *Parser) matchToken(tokenType t.TokenType) (tt t.Token, ok bool) {
	token := p.scanner.Scan()
	if token.TokenType == tokenType {
		tt = token
		return
	} else {
		ok = false
		return
	}
}

var (
	errorGrammarNotValid = errors.New("syntax is not valid")
)

func (p *Parser) returnExpression() (expr Expression, err error) {
	value, err := p.expression()
	if err != nil {
		return
	}
	expr = ReturnExpression{
		value: value,
	}
	return
}

func (p *Parser) expression() (Expression, error) {
	token := p.scanner.Scan()
	switch token.TokenType {
	case t.Function:
		return p.function()
	case t.If:
		return p.ifExpression()
	case t.Var:
		return p.varExpression()
	case t.Const:
		return p.constExpression()
	case t.Return:
		return p.returnExpression()
	case t.Number:
		v, _ := strconv.Atoi(token.Literal)
		return NumberExpression{
			value: v,
		}, nil
	case t.String:
		return StringExpression{
			value: token.Literal,
		}, nil
	case t.Boolean:
		return BooleanExpression{
			value: token.TokenType == t.True,
		}, nil
	default:
		return nil, errorGrammarNotValid
	}
}

func (p *Parser) block() (expr Block, err error) {
	_, isInBracket := p.matchToken(t.LeftBracket)

	var expressions []Expression
	// only parse one line when is not in bracket scope
	if !isInBracket {
		e, err := p.expression()
		if err != nil {
			return
		}
		expressions = append(expressions, e)
		return
	}

	for p.scanner.NextToken().TokenType != t.RightBracket {
		e, err := p.expression()
		if err != nil {
			return
		}
		expressions = append(expressions, e)
	}

	if _, ok := p.matchToken(t.RightBracket); !ok {
		log.Println("block, bracket not matched")
		err = errorGrammarNotValid
		return
	}

	return
}

func (p *Parser) functionArgs() []Expression {
	var args []Expression

	return args
}

func (p *Parser) function() (expr Expression, err error) {
	name := p.scanner.Scan()
	if _, ok := p.matchToken(t.LeftParenthesis); !ok {
		log.Println("function, (")
		err = errorGrammarNotValid
		return
	}

	args := p.functionArgs()

	if _, ok := p.matchToken(t.RightParenthesis); !ok {
		log.Println("function, )")
		err = errorGrammarNotValid
		return
	}

	block, err := p.block()
	if err != nil {
		return
	}

	expr = FunctionExpression{
		name: name.Literal,
		args: args,
		body: block,
	}
	return

}

func (p *Parser) constExpression() (expr Expression, err error) {
	identifier, err := p.identifierExpression()
	if err != nil {
		return
	}
	// TODO: optional assign
	if _, ok := p.matchToken(t.Equal); !ok {
		log.Println("TEMP, const, assign value")
		err = errorGrammarNotValid
	}

	value, err := p.expression()
	if err != nil {
		return
	}
	expr = ConstExpression{
		identifier: identifier,
		value:      value,
	}
	return
}
func (p *Parser) varExpression() (expr Expression, err error) {
	identifier, err := p.identifierExpression()
	if err != nil {
		return
	}
	// TODO: optional assign
	if _, ok := p.matchToken(t.Equal); !ok {
		log.Println("TEMP, var, assign value")
		err = errorGrammarNotValid
	}

	value, err := p.expression()
	if err != nil {
		return
	}
	expr = VarExpression{
		identifier: identifier,
		value:      value,
	}
	return
}

func (p *Parser) identifierExpression() (expr IdentifierExpression, err error) {
	token := p.scanner.Scan()
	if token.TokenType != t.String {
		log.Println("var, should use identifier after var keyword")
		err = errorGrammarNotValid
		return
	}
	expr = IdentifierExpression{
		name: token,
	}
	return
}

func (p *Parser) conditionExpression() (expr Expression, err error) {
	left, err := p.expression()
	if err != nil {
		return
	}
	token := p.scanner.Scan()
	right, err := p.expression()
	if err != nil {
		return
	}

	expr = BinaryExpression{
		left:     left,
		right:    right,
		operator: token,
	}
	return

}

func (p *Parser) ifExpression() (expr Expression, err error) {
	if _, ok := p.matchToken(t.LeftParenthesis); !ok {
		err = errorGrammarNotValid
		return
	}

	condition, err := p.conditionExpression()
	if err != nil {
		return
	}

	if _, ok := p.matchToken(t.RightParenthesis); !ok {
		err = errorGrammarNotValid
		return
	}
	_, leftBracketMatched := p.matchToken(t.LeftBracket)

	block, err := p.block()

	if leftBracketMatched {
		if _, ok := p.matchToken(t.RightBracket); !ok {
			log.Println("If, RightBracket not matched")
			err = errorGrammarNotValid
			return
		}
	}

	var elseBranch Expression
	if p.scanner.NextToken().TokenType != t.Else {
		p.scanner.Scan()
		if p.scanner.NextToken().TokenType == t.If {
			b, err := p.expression()
			if err != nil {
				return
			}
			elseBranch = b
		}
	}

	expr = IfExpression{
		condition:  condition,
		then:       block,
		elseBranch: elseBranch,
	}
	return
}

func (p *Parser) Parse() {
	s := p.scanner
	for !s.IsAtEnd() {
		token := s.Scan()
		switch token.TokenType {

		}

	}

}
