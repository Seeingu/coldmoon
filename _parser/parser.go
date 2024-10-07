package _parser

import (
	"errors"
	"fmt"
	"github.com/Seeingu/coldmoon/lexer"
	t "github.com/Seeingu/coldmoon/token"
	"log"
	"strconv"
)

type Parser struct {
	scanner *lexer.Scanner
}

var precedenceMap = map[t.TokenType]int{}

func New(scanner *lexer.Scanner) Parser {
	precedenceMap[t.BarBar] = 5
	precedenceMap[t.AmpersandAmpersand] = 5
	precedenceMap[t.Plus] = 10
	precedenceMap[t.Minus] = precedenceMap[t.Plus]
	precedenceMap[t.Star] = 15
	precedenceMap[t.Slash] = 15
	return Parser{
		scanner: scanner,
	}
}

// MARK: parser utils

// skipToNextToken same as scan,
// just means discard result
func (p *Parser) skipToNextToken() {
	p.scan()
}

// scan will return current token after move cursor
func (p *Parser) scan() t.Token {
	return p.scanner.Scan()
}

func (p *Parser) currentToken() t.Token {
	return p.scanner.CurrentToken()
}

func (p *Parser) nextToken() t.Token {
	return p.scanner.NextToken()
}

func (p *Parser) isPrecedenceHigher(a, b t.Token) bool {
	return precedenceMap[a.TokenType] > precedenceMap[b.TokenType]
}

func (p *Parser) skipPossibleSemicolon() {
	if p.currentToken().Is(t.Semicolon) {
		p.skipToNextToken()
	}
}

// MARK: parser errors

func (p *Parser) semanticError(m string) (Expression, error) {
	log.Printf("semantic error: %s\n", m)
	log.Printf("current token: %+v\n", p.scanner.CurrentToken())
	return nil, errorSemanticError
}

func (p *Parser) syntaxError(m string) (Expression, error) {
	log.Printf("syntax error: %s\n", m)
	log.Printf("current token: %+v\n", p.scanner.CurrentToken())
	return nil, errorGrammarNotValid
}

// matchToken will consume if current token matched
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
	// skip return
	p.skipToNextToken()
	value, err := p.maybeBinary()
	if err != nil {
		return
	}
	expr = ReturnExpression{
		value: value,
	}
	p.skipPossibleSemicolon()
	return
}

func (p *Parser) stringExpression() (expr Expression, err error) {
	str := p.scanner.CurrentToken().Literal
	p.scanner.Scan()

	return StringExpression{
		value: str,
	}, nil
}

// TODO: maybe split into two
func (p *Parser) primaryOrIdentifier() (expr Expression, err error) {
	token := p.currentToken()
	switch token.TokenType {
	case t.Number:
		v, _ := strconv.Atoi(token.Literal)
		p.scanner.Scan()
		return NumberExpression{
			value: v,
		}, nil
	case t.String:
		return p.stringExpression()
	case t.Identifier:
		return p.identifier()
	case t.Boolean:
		return BooleanExpression{
			value: token.TokenType == t.True,
		}, nil
	default:
		return p.syntaxError("Primary, unreachable ")
	}

}

func (p *Parser) maybeUnary() (expr Expression, err error) {
	token := p.currentToken()
	unaryTokenTypes := []t.TokenType{
		t.Tilde,
		t.Bang,
	}

	if token.Is(t.Function) {
		return p.function()
	}

	if !token.IsOneOf(unaryTokenTypes) {
		return p.primaryOrIdentifier()
	}
	p.scanner.Scan()
	value, err := p.primaryOrIdentifier()
	if err != nil {
		return
	}
	return UnaryExpression{
		unary: token,
		value: value,
	}, nil
}

func (p *Parser) instantiateClass() (expr Expression, err error) {
	identifier := p.scanner.Scan()
	name := IdentifierExpression{name: identifier}

	if ok := p.matchToken(t.LeftParenthesis); !ok {
		return p.syntaxError("Class, (")
	}

	args, err := p.functionArgs()
	if err != nil {
		return
	}

	if ok := p.matchToken(t.RightParenthesis); !ok {
		return p.syntaxError("Class, )")
	}

	return ClassInstantiateExpression{
		caller: name,
		args:   args,
	}, nil

}

func (p *Parser) throwExpression() (expr Expression, err error) {
	// skip throw
	p.skipToNextToken()

	var errorExpression Expression
	if p.matchToken(t.New) {
		errorExpression, err = p.instantiateClass()
		if err != nil {
			return nil, err
		}
	} else if p.currentToken().Is(t.String) {
		errorExpression, err = p.stringExpression()
	} else if p.currentToken().Is(t.Identifier) {
		errorExpression, err = p.identifier()
	}
	p.skipPossibleSemicolon()
	return ThrowExpression{
		errorExpression: errorExpression,
	}, err
}

func (p *Parser) expression() (Expression, error) {
	token := p.currentToken()
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
	case t.For:
		return p.forExpression()
	case t.While:
		return p.whileExpression()
	case t.LeftParenthesis:
		return p.leftParenthesis()
	case t.Throw:
		return p.throwExpression()
	case t.SlashSlash:
		p.skipToNextToken()
		return SingleLineCommentExpression{content: token.Literal}, nil
	case t.SlashStar:
		p.skipToNextToken()
		return MultiLineCommentExpression{content: token.Literal}, nil
	case t.Number, t.String, t.Boolean, t.Identifier, t.This:
		return p.maybeAssign()
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

	for !p.matchToken(t.RightBracket) {
		e, err := p.expression()
		if err != nil {
			return expr, err
		}
		expressions = append(expressions, e)
	}

	return
}

func (p *Parser) functionArgs() (args []Expression, err error) {
	if p.currentToken().Is(t.RightParenthesis) {
		return
	}
	arg, err := p.maybeBinary()
	if err != nil {
		return nil, err
	}
	args = append(args, arg)

	for p.matchToken(t.Comma) {
		arg, err := p.maybeBinary()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args, nil
}

// TODO: check is anonymous available
func (p *Parser) function() (expr Expression, err error) {
	// anonymous
	var name string
	if p.nextToken().Is(t.LeftParenthesis) {
		// skip function
		p.skipToNextToken()
	} else {
		name = p.scan().Literal
		// skip name
		p.skipToNextToken()
	}

	if !p.matchToken(t.LeftParenthesis) {
		return p.syntaxError("function, (")
	}

	args, err := p.functionArgs()
	if err != nil {
		return
	}

	if !p.matchToken(t.RightParenthesis) {
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

	for p.scanner.HasNextToken() && !p.matchToken(t.RightSquareBracket) {
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

func (p *Parser) maybeChain(target Expression) (expr Expression, err error) {
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
			identifier: target,
			properties: keys,
		}, nil
	}

	return target, nil
}

func (p *Parser) maybeAssign() (expr Expression, err error) {
	token := p.currentToken()

	if token.Is(t.This) {
		expr = ThisExpression{}
	} else if token.Is(t.Identifier) {
		expr = IdentifierExpression{
			name: token,
		}
	}

	// skip first key
	p.skipToNextToken()
	expr, err = p.maybeChain(expr)
	if err != nil {
		return nil, err
	}
	if p.matchToken(t.Equal) {
		return p.assignExpression(expr)
	}

	p.skipPossibleSemicolon()
	return

}

func (p *Parser) assignExpression(left Expression) (expr Expression, err error) {
	right, err := p.maybeBinary()
	if err != nil {
		return nil, err
	}
	expr = AssignExpression{
		left:  left,
		right: right,
	}
	p.skipPossibleSemicolon()
	return
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

	expr, err = p.maybeChain(expr)
	if err != nil {
		return nil, err
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

func (p *Parser) maybeBinary() (expr Expression, err error) {
	left, err := p.maybeUnary()
	if err != nil {
		return
	}
	operatorToken := p.scan()
	binaryTokenTypes := []t.TokenType{t.Plus, t.Minus, t.Star, t.Slash, t.BarBar, t.AmpersandAmpersand}
	if !operatorToken.IsOneOf(binaryTokenTypes) {
		return left, nil
	}
	// skip operator
	p.skipToNextToken()
	right, err := p.maybeUnary()
	if err != nil {
		return
	}

	aheadToken := p.nextToken()
	if aheadToken.IsOneOf(binaryTokenTypes) {
		// already know what the operatorToken is
		p.scanner.Scan()
		if p.isPrecedenceHigher(aheadToken, operatorToken) {
			newRight, err := p.maybeUnary()
			if err != nil {
				return nil, err
			}
			right = BinaryExpression{
				left:     right,
				right:    newRight,
				operator: aheadToken,
			}
		} else {
			leftRight, err := p.maybeUnary()
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
		operator: operatorToken,
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
