package parser

import (
	"errors"
	"fmt"
	t "github.com/Seeingu/coldmoon/token"
	"log"
	"strconv"
)

type Parser struct {
	scanner *Scanner
}

var precedenceMap = map[t.TokenType]int{}

func NewParser(scanner *Scanner) Parser {
	precedenceMap[t.Plus] = 10
	precedenceMap[t.Minus] = precedenceMap[t.Plus]
	precedenceMap[t.Star] = 15
	precedenceMap[t.Slash] = 15
	return Parser{
		scanner: scanner,
	}
}

// MARK: parser utils

func (p *Parser) currentToken() t.Token {
	return p.scanner.CurrentToken()
}

func (p *Parser) nextToken() t.Token {
	return p.scanner.NextToken()
}

func (p *Parser) isPrecedenceHigher(a, b t.Token) bool {
	return precedenceMap[a.TokenType] > precedenceMap[b.TokenType]
}

// MARK: parser errors

func (p *Parser) semanticError(m string) (Expression, error) {
	log.Printf("semantic error: %s\n", m)
	log.Printf("current token: %v\n", p.scanner.currentToken)
	return nil, errorSemanticError
}

func (p *Parser) syntaxError(m string) (Expression, error) {
	log.Printf("syntax error: %s\n", m)
	log.Printf("current token: %v\n", p.scanner.currentToken)
	return nil, errorGrammarNotValid
}

// matchToken will consume token if matched
func (p *Parser) matchToken(tokenType t.TokenType) (ok bool) {
	if !p.scanner.nextToken.Is(tokenType) {
		return false
	}
	p.scanner.Scan()
	return true
}

var (
	errorGrammarNotValid = errors.New("syntax is not valid")
	errorSemanticError   = errors.New("semantic error")
)

func (p *Parser) forExpression() (expr Expression, err error) {
	if !p.matchToken(t.LeftParenthesis) {
		return p.syntaxError("For, ( after for")
	}

	// init
	init, err := p.expression()
	if err != nil {
		return
	}
	if !p.matchToken(t.Semicolon) {
		return p.syntaxError("For, ; after init expression")
	}
	comparison, err := p.expression()
	if err != nil {
		return
	}
	if !p.matchToken(t.Semicolon) {
		return p.syntaxError("For, ; after comparison expression")
	}
	step, err := p.expression()
	if err != nil {
		return
	}

	if p.matchToken(t.RightParenthesis) {
		return p.syntaxError("For, )")
	}

	body, err := p.block()
	if err != nil {
		return
	}

	return ForExpression{
		init:       init,
		step:       step,
		comparison: comparison,
		body:       body,
	}, nil
}

func (p *Parser) whileExpression() (expr Expression, err error) {
	if ok := p.matchToken(t.LeftParenthesis); !ok {
		return p.syntaxError("While, ( after for")
	}

	condition, err := p.expression()
	if err != nil {
		return
	}

	if ok := p.matchToken(t.RightParenthesis); !ok {
		return p.syntaxError("While, condition )")
	}

	block, err := p.block()
	if err != nil {
		return
	}

	return WhileExpression{
		condition: condition,
		body:      block,
	}, nil

}

func (p *Parser) leftParenthesis() (expr Expression, err error) {
	// TODO: trinary, or anything else?

	// Arrow function
	args, err := p.functionArgs()
	if err != nil {
		return
	}

	if !p.matchToken(t.RightParenthesis) {
		return p.syntaxError("Arrow Function, )")
	}

	if !p.matchToken(t.EqualGreater) {
		return p.syntaxError("Arrow Function, bad syntax, not found =>")
	}

	block, err := p.block()
	if err != nil {
		return
	}
	return ArrowFunctionExpression{
		args: args,
		body: block,
	}, nil
}

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

func (p *Parser) stringExpression() (expr Expression, err error) {
	str := p.scanner.currentToken.Literal

	return StringExpression{
		value: str,
	}, nil
}

func (p *Parser) unaryExpression() (expr Expression, err error) {
	if p.scanner.nextToken.Is(t.Equal) {
		return p.syntaxError("unaryExpression: unexpected =")
	}

	value, err := p.expression()
	if err != nil {
		return
	}
	return UnaryExpression{
		unary: p.currentToken(),
		value: value,
	}, nil
}

func (p *Parser) expression() (Expression, error) {
	token := p.scanner.Scan()
	switch token.TokenType {
	case t.Function:
		return p.function()
	case t.If:
		return p.ifExpression()
	case t.Let:
		return p.defineExpression(token)
	case t.Var:
		return p.defineExpression(token)
	case t.Const:
		return p.defineExpression(token)
	case t.LeftSquareBracket:
		return p.arrayLiteralExpression()
	case t.LeftBracket:
		return p.objectLiteralExpression()
	case t.Return:
		return p.returnExpression()
	case t.Number:
		v, _ := strconv.Atoi(token.Literal)
		return NumberExpression{
			value: v,
		}, nil
	case t.Bang, t.Tilde:
		return p.unaryExpression()
	case t.Identifier:
		return p.identifier()
	case t.String:
		return p.stringExpression()
	case t.Boolean:
		return BooleanExpression{
			value: token.TokenType == t.True,
		}, nil
	case t.For:
		return p.forExpression()
	case t.While:
		return p.whileExpression()
	case t.LeftParenthesis:
		return p.leftParenthesis()
	default:
		return p.syntaxError(fmt.Sprintf("token %+v is not matched", token))
	}
}

func (p *Parser) block() (expr BlockExpression, err error) {
	isInBracket := p.matchToken(t.LeftBracket)

	var expressions []Expression
	// only parse one line when is not in bracket scope
	if !isInBracket {
		e, err := p.expression()
		if err != nil {
			return expr, err
		}
		expressions = append(expressions, e)
		return expr, err
	}

	for p.scanner.NextToken().TokenType != t.RightBracket {
		e, err := p.expression()
		if err != nil {
			return expr, err
		}
		expressions = append(expressions, e)
	}

	if ok := p.matchToken(t.RightBracket); !ok {
		_, err := p.syntaxError("block, bracket not matched")
		return expr, err
	}

	return
}

func (p *Parser) functionArgs() (args []Expression, err error) {

	if p.currentToken().Is(t.RightParenthesis) {
		return
	}
	arg, err := p.expression()
	if err != nil {
		return nil, err
	}
	args = append(args, arg)

	for p.matchToken(t.Comma) {
		arg, err := p.expression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args, nil
}

func (p *Parser) function() (expr Expression, err error) {
	name := p.scanner.Scan()
	if ok := p.matchToken(t.LeftParenthesis); !ok {
		return p.syntaxError("function, (")
	}

	args, err := p.functionArgs()
	if err != nil {
		return
	}

	if ok := p.matchToken(t.RightParenthesis); !ok {
		return p.syntaxError("function, )")
	}

	block, err := p.block()
	if err != nil {
		return
	}

	expr = FunctionExpression{
		name: name,
		args: args,
		body: block,
	}
	return

}

func (p *Parser) arrayLiteralExpression() (expr Expression, err error) {
	var elems []Expression

	if !p.matchToken(t.RightSquareBracket) {
		expr, err := p.expression()
		if err != nil {
			return expr, err
		}
		elems = append(elems, expr)
	}

	for !p.matchToken(t.RightSquareBracket) {
		if !p.matchToken(t.Comma) {
			return p.syntaxError("Array, should use , between tww elements in array")
		}
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		elems = append(elems, expr)
	}

	// TODO: Syntax: check has ], and report error somewhere

	return ArrayLiteralExpression{
		elements: elems,
	}, nil
}

func (p *Parser) objectLiteralExpression() (expr Expression, err error) {
	var pairs []PairExpression

	for !p.matchToken(t.RightBracket) {
		left, err := p.expression()
		if err != nil {
			return nil, err
		}
		if !p.matchToken(t.Colon) {
			return p.syntaxError("Object, should insert : between key and value")
		}
		right, err := p.expression()
		// optional
		p.matchToken(t.Comma)
		pairs = append(pairs, PairExpression{left, right})
	}

	return ObjectLiteralExpression{
		pairs: pairs,
	}, nil
}

func (p *Parser) defineExpression(token t.Token) (expr Expression, err error) {
	identifier := IdentifierExpression{
		name: p.scanner.Scan(),
	}
	var value Expression
	if p.matchToken(t.Equal) {
		value, err = p.expression()
		if err != nil {
			return
		}
	}

	switch token.TokenType {
	case t.Let:
		expr = LetExpression{
			identifier: identifier,
			value:      value,
		}
	case t.Const:
		expr = LetExpression{
			identifier: identifier,
			value:      value,
		}
	case t.Var:
		expr = VarExpression{
			identifier: identifier,
			value:      value,
		}
	default:
		break
	}
	return
}

func (p *Parser) identifier() (expr Expression, err error) {
	token := p.currentToken()
	expr = IdentifierExpression{
		name: token,
	}

	var keys []Expression
	// Chaining
	// TODO: ?.
	for p.matchToken(t.Dot) {
		id := p.currentToken()
		if !p.matchToken(t.Identifier) {
			return p.syntaxError("Identifier, unknown token after dot")
		}
		keys = append(keys, IdentifierExpression{
			name: id,
		})

		// TODO: check value type
		//return p.semanticError(fmt.Sprintf("Identifier, dot operation on %s is not allowed", reflect.TypeOf(value)))
	}
	if len(keys) > 0 {
		return ChainExpression{
			identifier: expr,
			keys:       keys,
		}, nil
	}

	// Call
	if p.matchToken(t.LeftParenthesis) {
		args, err := p.functionArgs()
		if err != nil {
			return nil, err
		}
		if !p.matchToken(t.RightParenthesis) {
			return p.syntaxError("Identifier, missing ) in call expression")

		}
		return CallExpression{
			caller: expr,
			args:   args,
		}, nil
	}

	return
}

func (p *Parser) binaryExpression() (expr Expression, err error) {
	left, err := p.expression()
	if err != nil {
		return
	}
	token := p.scanner.Scan()
	right, err := p.expression()
	if err != nil {
		return
	}

	aheadToken := p.nextToken()
	binaryTokenTypes := []t.TokenType{t.Plus, t.Minus, t.Star, t.Slash}
	if aheadToken.IsOneOf(binaryTokenTypes) {
		// already know what the token is
		p.scanner.Scan()
		if p.isPrecedenceHigher(aheadToken, token) {
			newRight, err := p.expression()
			if err != nil {
				return nil, err
			}
			right = BinaryExpression{
				left:     right,
				right:    newRight,
				operator: aheadToken,
			}
		} else {
			leftRight, err := p.expression()
			if err != nil {
				return expr, err
			}
			left = BinaryExpression{
				left:     left,
				right:    leftRight,
				operator: aheadToken,
			}
		}
	}

	expr = BinaryExpression{
		left:     left,
		right:    right,
		operator: token,
	}
	return

}

func (p *Parser) ifExpression() (expr Expression, err error) {
	if ok := p.matchToken(t.LeftParenthesis); !ok {
		return p.syntaxError("If, missing (")
	}

	condition, err := p.expression()
	if err != nil {
		return
	}

	if ok := p.matchToken(t.RightParenthesis); !ok {
		return p.syntaxError("If, missing )")
	}
	leftBracketMatched := p.matchToken(t.LeftBracket)

	block, err := p.block()

	if leftBracketMatched {
		if ok := p.matchToken(t.RightBracket); !ok {
			return p.syntaxError("If, RightBracket not matched")
		}
	}

	var elseBranch Expression
	if p.scanner.NextToken().TokenType != t.Else {
		p.scanner.Scan()
		if p.scanner.NextToken().TokenType == t.If {
			b, err := p.expression()
			if err != nil {
				return expr, err
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

func (p *Parser) Parse() ([]Expression, error) {
	var expressions []Expression
	s := p.scanner
	for s.HasNextToken() {
		expression, err := p.expression()
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expression)
	}
	return expressions, nil
}
