package coldmoon

import (
	"fmt"
	"github.com/samber/lo"
)

type Parser struct {
	SourceText              string
	tokenizer               *Tokenizer
	inFunctionBody          bool
	callExpressionForbidden bool
	ctx                     ParserContext
}

func NewParser(sourceText string, ctx ParserContext) *Parser {
	return &Parser{
		SourceText: sourceText,
		tokenizer:  NewTokenizer(sourceText),
		ctx:        ctx,
	}
}

func (p *Parser) Parse() *Script {
	return &Script{
		StatementList: p.ParseNode(),
	}
}

type ParserContext struct {
	FileName string
}

type precedence int
type precedenceAssociativityAlt int

const (
	precAssocNewArgs precedenceAssociativityAlt = iota
	precAssocFunctionArgs
	precAssocPostfixIncrement
	precAssocPostfixDecrement
	precAssocUnaryPlus
	precAssocUnaryMinus
	precAssocPrefixIncrement
	precAssocPrefixDecrement
)

type associativity int

const (
	associativeNone associativity = iota
	associativeLeft
	associativeRight
)

type acceptContext struct {
	precedence    precedence
	associativity associativity
}

// MARK: - Precedence

func (p *Parser) acceptContextLowest() *acceptContext {
	return &acceptContext{
		precedence: 0,
	}
}

func (p *Parser) acceptContext(t TokenType) *acceptContext {
	switch t {
	case TLeftParen:
		return &acceptContext{
			precedence: 18,
		}
	case TDot, TQuestionDot:
		return &acceptContext{
			precedence:    17,
			associativity: associativeLeft,
		}
	case TLeftBracket:
		return &acceptContext{
			precedence: 17,
		}
	case TNew:
		return &acceptContext{
			precedence: 16,
		}
	case TPlusPlus, TMinusMinus:
		return &acceptContext{
			precedence: 15,
		}
	case TWave, TTypeof, TDelete, TNot:
		return &acceptContext{
			precedence: 14,
		}
	case TStarStar:
		return &acceptContext{
			precedence:    13,
			associativity: associativeRight,
		}
	case TStar, TSlash, TPercent:
		return &acceptContext{
			precedence:    12,
			associativity: associativeLeft,
		}
	case TPlus, TMinus:
		return &acceptContext{
			precedence:    11,
			associativity: associativeLeft,
		}
	case TLeftShift, TRightShift, TUnsignedRightShift:
		return &acceptContext{
			precedence:    10,
			associativity: associativeLeft,
		}
	case TLessThan, TLessThanEquals, TGreaterThan, TGreaterThanEquals, TInstanceof, TIn:
		return &acceptContext{
			precedence:    9,
			associativity: associativeLeft,
		}
	case TEqualsEquals, TNotEquals, TStrictEquals, TStrictNotEquals:
		return &acceptContext{
			precedence:    8,
			associativity: associativeLeft,
		}
	case TAmpersand:
		return &acceptContext{
			precedence:    7,
			associativity: associativeLeft,
		}
	case TCaret:
		return &acceptContext{
			precedence:    6,
			associativity: associativeLeft,
		}
	case TPipe:
		return &acceptContext{
			precedence:    5,
			associativity: associativeLeft,
		}
	case TAmpersandAmpersand:
		return &acceptContext{
			precedence:    4,
			associativity: associativeLeft,
		}
	case TPipePipe, TQuestionQuestion:
		return &acceptContext{
			precedence:    3,
			associativity: associativeLeft,
		}
	case TEquals, TPlusEquals, TMinusEquals, TStarEquals, TStarStarEquals, TPercentEquals, TLeftShiftEquals, TRightShiftEquals, TUnsignedRightShiftEquals, TAmpersandEquals, TCaretEquals, TPipeEquals:
		return &acceptContext{
			precedence:    2,
			associativity: associativeRight,
		}
	case TQuestion:
		return &acceptContext{
			precedence:    2,
			associativity: associativeRight,
		}
	case TArrow:
		return &acceptContext{
			precedence:    2,
			associativity: associativeRight,
		}
	case TYield, TDotDotDot:
		return &acceptContext{
			precedence: 2,
		}
	case TComma:
		return &acceptContext{
			precedence:    1,
			associativity: associativeLeft,
		}
	default:
		return p.acceptContextLowest()
	}
}

func (p *Parser) precedence(t TokenType) precedence {
	return p.acceptContext(t).precedence
}

func (p *Parser) acceptContextAlt(flag precedenceAssociativityAlt) *acceptContext {
	switch flag {
	case precAssocNewArgs, precAssocFunctionArgs:
		return &acceptContext{
			precedence: 17,
		}
	case precAssocPostfixIncrement, precAssocPostfixDecrement:
		return &acceptContext{
			precedence: 15,
		}
	case precAssocPrefixIncrement, precAssocPrefixDecrement, precAssocUnaryPlus, precAssocUnaryMinus:
		return &acceptContext{
			precedence: 14,
		}
	}
	panic("unreachable")
}

func (p *Parser) precedenceAlt(flag precedenceAssociativityAlt) precedence {
	return p.acceptContextAlt(flag).precedence
}

func (p *Parser) associativity(t TokenType) associativity {
	return p.acceptContext(t).associativity
}

func (p *Parser) associativityAlt(flag precedenceAssociativityAlt) associativity {
	return p.acceptContextAlt(flag).associativity
}

// MARK: - Parse

func (p *Parser) ParseNode() StatementList {
	return p.statementList()
}

func (p *Parser) statementList() (list StatementList) {
	for {
		item := p.statementListItem()
		if item == nil {
			break
		}
		switch stmt := item.(type) {
		case *StatementListItemStatement:
			if _, ok := stmt.Statement.(*StatementEmpty); ok {
				break
			}
		}
		list = append(list, item)
	}

	return
}

func (p *Parser) functionBody(functionType FunctionType) *FunctionBody {
	inFunctionBodyBefore := p.inFunctionBody
	p.inFunctionBody = true
	defer func() {
		p.inFunctionBody = inFunctionBodyBefore
	}()

	list := p.statementList()
	return &FunctionBody{
		StatementList: list,
	}
}

func (p *Parser) statementListItem() StatementListItem {
	t := p.tokenizer.CurrentToken
	var stmt Statement
	switch t.Type {
	case TAsync, TFunction, TLet, TConst, TClass:
		d := p.declaration()
		stmt = &StatementListItemDeclaration{
			Declaration: d,
		}
	default:
		s := p.statement()
		if s == nil {
			p.automaticSemicolonInsertion()
			return nil
		}

		stmt = &StatementListItemStatement{
			Statement: s,
		}
	}
	return stmt
}

func (p *Parser) automaticSemicolonInsertion() {
	t := p.tokenizer.CurrentToken
	if t.Type == TSemicolon {
		p.tokenizer.Next()
	}
}

func (p *Parser) statement() Statement {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TVar:
		return p.variableStatement()
	case TSemicolon:
		p.tokenizer.Next()
		return &StatementEmpty{}
	case TEOF:
		return nil
	case TLeftBrace:
		return p.blockStatement()
	case TDebugger:
		p.tokenizer.Next()
		return &StatementDebugger{}
	case TIf:
		return p.ifStatement()
	case TWhile, TDo:
		return p.breakableStatement()
	case TThrow:
		return p.throwStatement()
	case TTry:
		return p.tryStatement()
	case TRightBrace:
		return nil
	case TReturn:
		return p.returnStatement()
	default:
		return p.expressionStatement()
	}
}

func (p *Parser) throwStatement() *StatementThrow {
	p.tokenizer.MustMatch(TThrow)
	p.noLineTerminatorHere()
	expr := p.expression(p.acceptContextLowest())
	return &StatementThrow{
		Expression: expr,
	}
}

func (p *Parser) tryStatement() *StatementTry {
	p.tokenizer.MustMatch(TTry)
	block := p.block()
	var catch *Block
	var finally *Block
	var catchParameter CatchParameter
	if p.tokenizer.CurrentToken.Type == TCatch {
		p.tokenizer.Next()
		if p.tokenizer.CurrentToken.Type == TLeftParen {
			p.tokenizer.Next()
			catchParameter = CatchParameter(p.bindingIdentifier())
			p.tokenizer.MustMatch(TRightParen)
		}
		catch = p.block()
	}
	if p.tokenizer.CurrentToken.Type == TFinally {
		finally = p.block()
	}
	if catch == nil && finally == nil {
		panic("tryStatement: expected catch or finally")
	}
	return &StatementTry{
		CatchParameter: catchParameter,
		TryBlock:       block,
		CatchBlock:     catch,
		FinallyBlock:   finally,
	}
}

func (p *Parser) formalParameters() *FormalParameters {
	var items []FormalParametersItem
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightParen {
			break
		}
		if p.tokenizer.Match(TDotDotDot) {
			identifier := p.bindingIdentifier()
			items = append(items, &FormalParameterFunctionRestParameter{
				BindingRestElement: &BindingElement{
					Identifier: identifier,
				},
			})
			p.tokenizer.Match(TComma)
			break
		} else {
			identifier := p.bindingIdentifier()
			items = append(items, &FormalParameter{
				BindingElement: &BindingElement{
					Identifier: identifier,
				},
			})
			p.tokenizer.Match(TComma)
		}
	}

	return &FormalParameters{
		Items: items,
	}
}

func (p *Parser) functionDeclaration() *FunctionDeclaration {
	startOffset := p.tokenizer.Index
	p.tokenizer.MustMatch(TFunction)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	functionBody := p.functionBody(FunctionTypeNormal)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &FunctionDeclaration{
		Identifier:       identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             functionBody,
	}
}

func (p *Parser) asyncArrowFunction() *PrimaryExpressionAsyncArrowFunction {
	startOffset := p.tokenizer.Index
	p.tokenizer.Match(TAsync)
	var params *FormalParameters
	if p.tokenizer.Match(TLeftParen) {
		params = p.formalParameters()
		p.tokenizer.MustMatch(TRightParen)
	} else {
		identifier := p.bindingIdentifier()
		params = &FormalParameters{
			Items: []FormalParametersItem{
				&FormalParameter{
					BindingElement: &BindingElement{
						Identifier: identifier,
					},
				},
			},
		}
	}
	p.tokenizer.MustMatch(TArrow)
	var body *FunctionBody
	if p.tokenizer.Match(TLeftBrace) {
		body = p.functionBody(FunctionTypeAsync)
		p.tokenizer.MustMatch(TRightBrace)
	} else {
		// prec: greater than ,
		e := p.expression(p.acceptContext(TYield))
		body = &FunctionBody{
			StatementList: StatementList{
				&StatementListItemStatement{
					Statement: &StatementExpression{
						Expression: e,
					},
				},
			},
		}
	}
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &PrimaryExpressionAsyncArrowFunction{
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) asyncFunctionExpression() *PrimaryExpressionAsyncFunctionExpression {
	startOffset := p.tokenizer.Index
	p.tokenizer.Match(TAsync)
	p.tokenizer.MustMatch(TFunction)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeAsync)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &PrimaryExpressionAsyncFunctionExpression{
		Identifier:       identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) asyncGeneratorExpression() *PrimaryExpressionAsyncGeneratorExpression {
	startOffset := p.tokenizer.Index
	p.tokenizer.Match(TAsync)
	p.tokenizer.MustMatch(TFunction)
	p.tokenizer.MustMatch(TStar)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeAsyncGenerator)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &PrimaryExpressionAsyncGeneratorExpression{
		IdentifierName:   identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) generatorExpression() *PrimaryExpressionGeneratorExpression {
	startOffset := p.tokenizer.Index
	p.tokenizer.MustMatch(TFunction)
	p.tokenizer.MustMatch(TStar)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeGenerator)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &PrimaryExpressionGeneratorExpression{
		IdentifierName:   identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) functionExpression() *PrimaryExpressionFunctionExpression {
	startOffset := p.tokenizer.Index
	p.tokenizer.MustMatch(TFunction)
	var identifier IdentifierName
	if p.tokenizer.CurrentToken.Type == TIdentifier {
		identifier = p.bindingIdentifier()
	}
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	functionBody := p.functionBody(FunctionTypeNormal)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &PrimaryExpressionFunctionExpression{
		Identifier:       identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             functionBody,
	}
}

func (p *Parser) noLineTerminatorHere() {
	// TODO
}

func (p *Parser) classDeclaration() *DeclarationClass {
	startIndex := p.tokenizer.Index
	p.tokenizer.MustMatch(TClass)
	identifier := p.bindingIdentifier()
	classTail := p.classTail()
	sourceText := p.SourceText[startIndex:p.tokenizer.Index]
	return &DeclarationClass{
		IdentifierName: identifier,
		ClassTail:      classTail,
		SourceText:     sourceText,
	}
}
func (p *Parser) classTail() *ClassTail {
	var classHeritage Expression
	if p.tokenizer.Match(TExtends) {
		classHeritage = p.expression(p.acceptContextLowest())
	}
	p.tokenizer.MustMatch(TLeftBrace)
	classBody := p.classBody()
	p.tokenizer.MustMatch(TRightBrace)
	return &ClassTail{
		ClassHeritage: classHeritage,
		ClassBody:     classBody,
	}
}

func (p *Parser) classBody() *ClassBody {
	var items []ClassElement
	for {
		if p.tokenizer.CurrentToken.Type == TRightBrace {
			break
		}
		items = append(items, p.classElement())
	}
	return &ClassBody{
		ClassElementList: &ClassElementList{
			Items: items,
		},
	}
}
func (p *Parser) classElement() ClassElement {
	if p.tokenizer.Match(TStatic) {
		def := p.propertyDefinition(MethodDefinitionTypeNil)
		return &ClassElementStaticMethodDefinition{
			MethodDefinition: def.(*PropertyDefinitionMethodDefinition),
		}
	}
	return &ClassElementEmpty{}
}

func (p *Parser) lexicalDeclaration() *DeclarationLexical {
	t := p.tokenizer.CurrentToken
	var lexicalType LexicalDeclarationType
	if t.Type == TLet {
		lexicalType = LexicalDeclarationTypeLet
	} else if t.Type == TConst {
		lexicalType = LexicalDeclarationTypeConst
	} else {
		panic("lexicalDeclaration: expected let or const")
	}
	p.tokenizer.Next()
	list := p.bindingList()
	p.automaticSemicolonInsertion()
	return &DeclarationLexical{
		Type:        lexicalType,
		BindingList: list,
	}
}
func (p *Parser) bindingList() *BindingList {
	var items []*LexicalBinding
	for {
		item := p.lexicalBinding()
		items = append(items, item)
		if p.tokenizer.CurrentToken.Type == TComma {
			p.tokenizer.Next()
		} else {
			break
		}
	}
	return &BindingList{
		Items: items,
	}
}
func (p *Parser) lexicalBinding() *LexicalBinding {
	identifier := p.bindingIdentifier()
	var init Expression
	if p.tokenizer.CurrentToken.Type == TEquals {
		p.tokenizer.Next()
		init = p.expression(p.acceptContext(TYield))
	}
	return &LexicalBinding{
		Identifier:  identifier,
		Initializer: init,
	}
}

func (p *Parser) hoistableDeclaration() DeclarationHoistable {
	t := p.tokenizer.CurrentToken
	if t.Type == TFunction && p.tokenizer.NextToken.Type == TStar {
		d := p.generatorDeclaration()
		return &DeclarationHoistableGenerator{
			GeneratorDeclaration: d,
		}
	} else if t.Type == TFunction {
		functionDeclaration := p.functionDeclaration()
		return &DeclarationHoistableFunction{
			FunctionDeclaration: functionDeclaration,
		}
	} else if t.Type == TAsync {
		startOffset := p.tokenizer.Index
		p.tokenizer.MustMatch(TAsync)
		p.tokenizer.MustMatch(TFunction)
		if p.tokenizer.CurrentToken.Type == TStar {
			d := p.asyncGeneratorDeclaration(startOffset)
			return &DeclarationHoistableAsyncGenerator{
				AsyncGeneratorDeclaration: d,
			}
		}
		return &DeclarationHoistableAsyncFunction{
			AsyncFunctionDeclaration: p.asyncFunctionDeclaration(startOffset),
		}

	}
	panic("unimplemented")
}

func (p *Parser) asyncFunctionDeclaration(startOffset int) *AsyncFunctionDeclaration {
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeAsync)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &AsyncFunctionDeclaration{
		Identifier:       identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) asyncGeneratorDeclaration(startOffset int) *AsyncGeneratorDeclaration {
	p.tokenizer.MustMatch(TStar)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	formalParams := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeAsyncGenerator)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &AsyncGeneratorDeclaration{
		Identifier:       identifier,
		FormalParameters: formalParams,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) generatorDeclaration() *GeneratorDeclaration {
	startOffset := p.tokenizer.Index
	p.tokenizer.MustMatch(TFunction)
	p.tokenizer.MustMatch(TStar)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	formalParams := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	functionBody := p.functionBody(FunctionTypeGenerator)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &GeneratorDeclaration{
		Identifier:       identifier,
		FormalParameters: formalParams,
		SourceText:       sourceText,
		Body:             functionBody,
	}
}

func (p *Parser) declaration() Declaration {
	t := p.tokenizer.CurrentToken
	if t.Type == TFunction || t.Type == TAsync {
		return p.hoistableDeclaration()
	}
	if t.Type == TClass {
		return p.classDeclaration()
	}
	if t.Type == TLet || t.Type == TConst {
		return p.lexicalDeclaration()
	}
	panic("unimplemented")
}

func (p *Parser) breakableStatement() *BreakableStatement {
	return &BreakableStatement{
		IterationStatement: p.iterationStatement(),
	}
}

// MARK: - Iteration

func (p *Parser) iterationStatement() IterationStatement {
	t := p.tokenizer.CurrentToken
	if t.Type == TDo {
		return p.doWhileStatement()
	} else if t.Type == TWhile {
		return p.whileStatement()
	}
	return p.forStatement()
}

func (p *Parser) forStatement() *StatementFor {
	p.tokenizer.MustMatch(TFor)
	p.tokenizer.MustMatch(TLeftParen)
	var init ForStatementInitializer
	t := p.tokenizer.CurrentToken
	if t.Type == TVar {
		init = &ForStatementInitializerVariable{
			VariableStatement: p.variableStatement(),
		}
	} else if t.Type == TLet || t.Type == TConst {
		init = &ForStatementInitializerLexicalDeclaration{
			LexicalDeclaration: p.lexicalDeclaration(),
		}
	} else {
		init = &ForStatementInitializerExpression{
			Expression: p.expression(p.acceptContextLowest()),
		}
	}
	p.tokenizer.MustMatch(TSemicolon)
	var condition Expression
	if p.tokenizer.CurrentToken.Type != TSemicolon {
		condition = p.expression(p.acceptContextLowest())
	}
	p.tokenizer.MustMatch(TSemicolon)
	var increment Expression
	if p.tokenizer.CurrentToken.Type != TRightParen {
		increment = p.expression(p.acceptContextLowest())
	}
	p.tokenizer.MustMatch(TRightParen)
	body := p.statement()

	return &StatementFor{
		Initializer: init,
		Condition:   condition,
		Increment:   increment,
		Body:        body,
	}

}

func (p *Parser) doWhileStatement() *StatementDoWhile {
	p.tokenizer.MustMatch(TDo)
	body := p.statement()
	p.tokenizer.MustMatch(TWhile)
	p.tokenizer.MustMatch(TLeftParen)
	condition := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TSemicolon)

	return &StatementDoWhile{
		Body:      body,
		Condition: condition,
	}
}

func (p *Parser) whileStatement() *StatementWhile {
	p.tokenizer.MustMatch(TWhile)
	p.tokenizer.MustMatch(TLeftParen)
	condition := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	body := p.statement()

	return &StatementWhile{
		Condition: condition,
		Body:      body,
	}
}

// MARK: - CompletionTypeReturn

func (p *Parser) returnStatement() *StatementReturn {
	p.tokenizer.MustMatch(TReturn)
	t := p.tokenizer.CurrentToken

	if !p.inFunctionBody {
		panic("returnStatement: not in function")
	}

	if t.Type == TSemicolon {
		p.tokenizer.Next()
		return &StatementReturn{}
	}
	expr := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TSemicolon)
	return &StatementReturn{
		Expression: expr,
	}
}

// MARK: - Condition

func (p *Parser) ifStatement() *StatementIf {
	p.tokenizer.MustMatch(TIf)
	p.tokenizer.MustMatch(TLeftParen)
	condition := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	consequent := p.statement()

	var alternate Statement
	if p.tokenizer.Peek().Type == TElse {
		p.tokenizer.Next()
		alternate = p.statement()
	}

	return &StatementIf{
		Condition:  condition,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

func (p *Parser) expressionStatement() *StatementExpression {
	expr := p.expression(p.acceptContextLowest())
	t := p.tokenizer.CurrentToken
	if t.Type == TSemicolon {
		p.tokenizer.Next()
	}
	if expr == nil {
		panic("expressionStatement: expected expression")
	}

	return &StatementExpression{
		Expression: expr,
	}
}

func (p *Parser) newExpression() (*ExpressionNewExpression, bool) {
	t := p.tokenizer.CurrentToken

	previous := p.callExpressionForbidden
	p.callExpressionForbidden = true
	defer func() {
		p.callExpressionForbidden = previous
	}()

	if t.Type != TNew {
		return nil, false
	}
	p.tokenizer.Next()
	accept := p.acceptContext(TNew)
	expr := p.expression(accept)
	p.callExpressionForbidden = previous
	args := p.arguments()
	return &ExpressionNewExpression{
		Callee:    expr,
		Arguments: args,
	}, true
}

// tryUnaryExpression accept unary token
// if token is not unary, return nil
func (p *Parser) tryUnaryExpression() (Expression, bool) {
	t := p.tokenizer.CurrentToken
	var operator UnaryOperator
	unaryMap := map[TokenType]UnaryOperator{
		TDelete: UnaryOperatorDelete,
		TVoid:   UnaryOperatorVoid,
		TTypeof: UnaryOperatorTypeof,
		TPlus:   UnaryOperatorAddition,
		TMinus:  UnaryOperatorSubtraction,
		TNot:    UnaryOperatorLogicalNot,
		TTilde:  UnaryOperatorBitwiseNot,
	}
	if op, ok := unaryMap[t.Type]; ok {
		p.tokenizer.Next()
		operator = op
	} else {
		return nil, false
	}
	var accept *acceptContext
	if t.Type == TPlus {
		accept = p.acceptContextAlt(precAssocUnaryPlus)
	} else if t.Type == TMinus {
		accept = p.acceptContextAlt(precAssocUnaryMinus)
	} else {
		accept = p.acceptContext(t.Type)
	}
	expr := p.expression(accept)
	return &UnaryExpression{
		Operator: operator,
		Operand:  expr,
	}, true
}

func (p *Parser) metaProperty() (MetaProperty, bool) {
	t := p.tokenizer.CurrentToken
	if t.Type != TNew || p.tokenizer.NextToken.Type == TDot {
		return nil, false
	}
	p.tokenizer.MustMatch(TNew)
	p.tokenizer.MustMatch(TDot)
	identifier := p.tokenizer.CurrentToken
	if identifier.Value != "target" {
		return nil, false
	}
	return &MetaPropertyNewTarget{}, true

}

func (p *Parser) updateExpression(primaryExpression Expression) (*ExpressionUpdate, bool) {
	t := p.tokenizer.CurrentToken
	var operator UpdateOperator
	if op, ok := UpdateOperatorMap[t.Type]; ok {
		operator = op
	} else {
		return nil, false
	}
	p.tokenizer.Next()
	var expr Expression
	var updateType UpdateExpressionType
	if primaryExpression == nil {
		expr = p.expression(p.acceptContextAlt(precAssocPrefixIncrement))
		updateType = UpdateExpressionTypePrefix
	} else {
		expr = primaryExpression
		updateType = UpdateExpressionTypePostfix
	}

	if updateType == UpdateExpressionTypePrefix && expr.AssignmentTargetType() != AssignmentTargetTypeSimple {
		panic("updateExpression: invalid assignment target for prefix")
	}
	if updateType == UpdateExpressionTypePrefix && expr.AssignmentTargetType() != AssignmentTargetTypeSimple {
		panic("updateExpression: invalid assignment target for postfix")
	}

	return &ExpressionUpdate{
		Operator: operator,
		Type:     updateType,
		Operand:  expr,
	}, true
}

func (p *Parser) expression(accept *acceptContext) Expression {
	unary, ok := p.tryUnaryExpression()
	if ok {
		return unary
	}
	meta, ok := p.metaProperty()
	if ok {
		return meta
	}
	update, ok := p.updateExpression(nil)
	if ok {
		return update
	}
	newExpression, ok := p.newExpression()
	if ok {
		return newExpression
	}

	primary := p.primaryExpression()
	var expr Expression = primary
	for {
		nextToken := p.tokenizer.CurrentToken
		newAcceptContext := p.acceptContext(nextToken.Type)
		if newAcceptContext.precedence <= accept.precedence {
			return expr
		}
		if newAcceptContext.precedence == accept.precedence && newAcceptContext.associativity == associativeLeft {
			return expr
		}

		secondary := p.secondaryExpression(expr, newAcceptContext)
		if secondary == nil {
			return expr
		}
		expr = secondary
	}
}

func (p *Parser) secondaryExpression(left PrimaryExpression, accept *acceptContext) Expression {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TLeftParen:
		return p.callExpression(left)
	case TLeftBracket, TDot:
		return p.memberExpression(left)
	case TPlusPlus, TMinusMinus:
		update, ok := p.updateExpression(left)
		if !ok {
			panic("secondaryExpression: expected update expression")
		}
		return update
	case TLessThan,
		TLessThanEquals,
		TGreaterThan,
		TGreaterThanEquals,
		TInstanceof,
		TIn:
		return p.relationalExpression(left, accept)
	case TEqualsEquals,
		TNotEquals,
		TStrictEquals,
		TStrictNotEquals:
		return p.equalityExpression(left, accept)
	case TAmpersandAmpersand, TPipePipe, TQuestionQuestion:
		return p.logicalExpression(left, accept)
	case TQuestion:
		return p.conditionalExpression(left, accept)
	case TComma:
		return p.sequenceExpression(left)
	case TStar,
		TStarStar,
		TSlash,
		TPercent,
		TPlus,
		TMinus,
		TLeftShift,
		TRightShift,
		TUnsignedRightShift,
		TAmpersand,
		TCaret,
		TPipe:
		return p.binaryExpression(left, accept)
	case TEquals,
		TPlusEquals,
		TMinusEquals,
		TStarEquals,
		TStarStarEquals,
		TPercentEquals,
		TLeftShiftEquals,
		TRightShiftEquals,
		TUnsignedRightShiftEquals,
		TAmpersandEquals,
		TCaretEquals,
		TPipeEquals,
		TAmpersandAmpersandEquals,
		TPipePipeEquals,
		TQuestionQuestionEquals:
		return p.assignmentExpression(left, accept)
	default:
		panic("secondaryExpression: unexpected token")
	}
	return left
}

func (p *Parser) assignmentExpression(left PrimaryExpression, accept *acceptContext) *ExpressionAssignmentExpression {
	t := p.tokenizer.CurrentToken
	p.tokenizer.Next()
	right := p.expression(accept)
	return &ExpressionAssignmentExpression{
		Operator: operatorAssignmentMap[t.Type],
		Left:     left,
		Right:    right,
	}
}

func (p *Parser) binaryExpression(left PrimaryExpression, accept *acceptContext) *ExpressionBinaryExpression {
	t := p.tokenizer.CurrentToken
	p.tokenizer.Next()
	right := p.expression(accept)
	return &ExpressionBinaryExpression{
		Operator: operatorBinaryMap[t.Type],
		Left:     left,
		Right:    right,
	}
}

func (p *Parser) sequenceExpression(left PrimaryExpression) *ExpressionSequenceExpression {
	list := []Expression{left}
	for {
		t := p.tokenizer.CurrentToken
		if t.Type != TComma {
			break
		}
		p.tokenizer.Next()
		accept := p.acceptContextLowest()
		expr := p.expression(accept)
		list = append(list, expr)
	}
	return &ExpressionSequenceExpression{
		Expressions: list,
	}
}

func (p *Parser) conditionalExpression(left PrimaryExpression, accept *acceptContext) *ExpressionConditionalExpression {
	p.tokenizer.MustMatch(TQuestion)
	consequent := p.expression(accept)
	p.tokenizer.MustMatch(TColon)
	alternate := p.expression(accept)
	return &ExpressionConditionalExpression{
		Test:       left,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

func (p *Parser) logicalExpression(left PrimaryExpression, accept *acceptContext) *ExpressionLogicalExpression {
	t := p.tokenizer.CurrentToken
	tokenTypes := []TokenType{
		TAmpersandAmpersand,
		TPipePipe,
		TQuestionQuestion,
	}
	if lo.Contains(tokenTypes, t.Type) {
		p.tokenizer.Next()
		right := p.expression(accept)
		return &ExpressionLogicalExpression{
			Operator: operatorLogicalMap[t.Type],
			Left:     left,
			Right:    right,
		}
	}
	panic("logicalExpression: unexpected token")
}

func (p *Parser) equalityExpression(left PrimaryExpression, accept *acceptContext) *ExpressionEqualityExpression {
	t := p.tokenizer.CurrentToken
	tokenTypes := []TokenType{
		TEqualsEquals,
		TNotEquals,
		TStrictEquals,
		TStrictNotEquals,
	}
	if lo.Contains(tokenTypes, t.Type) {
		p.tokenizer.Next()
		right := p.expression(accept)
		return &ExpressionEqualityExpression{
			Operator: operatorEqualityMap[t.Type],
			Left:     left,
			Right:    right,
		}
	}
	panic("equalityExpression: unexpected token")
}

func (p *Parser) relationalExpression(left PrimaryExpression, accept *acceptContext) *ExpressionRelationalExpression {
	t := p.tokenizer.CurrentToken
	tokenTypes := []TokenType{
		TLessThan,
		TLessThanEquals,
		TGreaterThan,
		TGreaterThanEquals,
		TInstanceof,
		TIn,
	}
	if lo.Contains(tokenTypes, t.Type) {
		p.tokenizer.Next()
		right := p.expression(accept)
		return &ExpressionRelationalExpression{
			Operator: operatorRelationMap[t.Type],
			Left:     left,
			Right:    right,
		}
	}
	panic("relationalExpression: unexpected token")
}

func (p *Parser) memberExpression(left PrimaryExpression) *MemberExpression {
	token := p.tokenizer.CurrentToken
	var property ASTProperty
	if token.Type == TLeftBracket {
		p.tokenizer.Next()
		propertyExpression := p.expression(p.acceptContextLowest())
		p.tokenizer.MustMatch(TRightBracket)
		property = &ASTPropertyExpression{
			Expression: propertyExpression,
		}
	} else if token.Type == TDot {
		p.tokenizer.Next()
		identifier := p.tokenizer.CurrentToken
		if identifier.Type != TIdentifier {
			panic("memberExpression: expected identifier")
		}
		p.tokenizer.Next()
		property = &ASTPropertyIdentifier{
			Identifier: IdentifierName(identifier.Value),
		}
	} else {
		panic("memberExpression: unexpected token")
	}

	return &MemberExpression{
		Member:   left,
		Property: property,
	}
}

func (p *Parser) callExpression(left PrimaryExpression) *CallExpression {
	if p.callExpressionForbidden {
		panic("callExpression: call expression forbidden")
	}
	args := p.arguments()
	return &CallExpression{
		Callee:    left,
		Arguments: args,
	}
}

func (p *Parser) arguments() Arguments {
	p.tokenizer.MustMatch(TLeftParen)
	var list Arguments
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightParen {
			break
		}
		// Precedence greater than TComma
		expr := p.expression(p.acceptContext(TYield))
		p.tokenizer.Match(TComma)
		list = append(list, expr)
	}
	p.tokenizer.MustMatch(TRightParen)
	return list
}

func (p *Parser) tryArrowFunction() *PrimaryExpressionArrowFunction {
	startOffset := p.tokenizer.Index
	p.tokenizer.store()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in tryArrowFunction", r)
			p.tokenizer.restore()
		}
	}()
	p.tokenizer.Match(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TArrow)
	p.noLineTerminatorHere()
	var body *FunctionBody
	if p.tokenizer.Match(TLeftBrace) {
		body = p.functionBody(FunctionTypeNormal)
		p.tokenizer.MustMatch(TRightBrace)
	} else {
		expression := p.expression(p.acceptContext(TComma))
		body = &FunctionBody{
			StatementList: StatementList{
				&StatementReturn{
					Expression: expression,
				},
			},
		}
	}
	return &PrimaryExpressionArrowFunction{
		FormalParameters: params,
		Body:             body,
		SourceText:       p.SourceText[startOffset:p.tokenizer.Index],
	}
}

func (p *Parser) parenthesizedExpression() *PrimaryExpressionParenthesizedExpression {
	p.tokenizer.MustMatch(TLeftParen)
	expr := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	return &PrimaryExpressionParenthesizedExpression{
		Expression: expr,
	}
}

func (p *Parser) identifierReference() *PrimaryExpressionIdentifierReference {
	t := p.tokenizer.CurrentToken
	types := []TokenType{TIdentifier, TAwait, TYield}
	if !lo.Contains(types, t.Type) {
		panic("identifierReference: expected identifierOrKeyword")
	}
	name := t.Value
	p.tokenizer.Next()
	return &PrimaryExpressionIdentifierReference{
		Identifier: IdentifierName(name),
	}
}

func (p *Parser) bindingIdentifier() IdentifierName {
	t := p.tokenizer.CurrentToken
	types := []TokenType{TIdentifier, TAwait, TYield}
	if !lo.Contains(types, t.Type) {
		panic("identifierReference: expected identifierOrKeyword")
	}
	name := t.Value
	p.tokenizer.Next()
	return IdentifierName(name)
}

func (p *Parser) primaryExpression() PrimaryExpression {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TThis:
		p.tokenizer.Next()
		return &ExpressionPrimary{
			PrimaryExpression: &PrimaryExpressionThis{},
		}
	case TLeftBracket:
		return p.arrayLiteral()
	case TLeftBrace:
		return p.objectLiteral()
	case TIdentifier:
		return p.identifierReference()
	case TLeftParen:
		e := p.tryArrowFunction()
		if e != nil {
			return e
		}
		return p.parenthesizedExpression()
	case TFunction:
		if p.tokenizer.NextToken.Type == TStar {
			return p.generatorExpression()
		}
		return p.functionExpression()
	case TAsync:
		p.tokenizer.MustMatch(TAsync)
		if p.tokenizer.NextToken.Type == TStar {
			return p.asyncGeneratorExpression()
		}
		if p.tokenizer.CurrentToken.Type == TLeftParen {
			return p.asyncArrowFunction()
		}
		return p.asyncFunctionExpression()
	case TRegularExpression:
		return p.regularExpressionLiteral()
	case TClass:
		return p.classExpression()
	default:
		literal := p.literal()

		return &ExpressionPrimary{
			PrimaryExpression: &PrimaryExpressionLiteral{
				Literal: literal,
			},
		}
	}
}

func (p *Parser) classExpression() *PrimaryExpressionClassExpression {
	startIndex := p.tokenizer.Index
	p.tokenizer.MustMatch(TClass)
	var identifier IdentifierName
	if p.tokenizer.Match(TIdentifier) {
		identifier = p.bindingIdentifier()
	}
	classTail := p.classTail()
	sourceText := p.SourceText[startIndex:p.tokenizer.Index]
	return &PrimaryExpressionClassExpression{
		IdentifierName: identifier,
		ClassTail:      classTail,
		SourceText:     sourceText,
	}
}

func (p *Parser) regularExpressionLiteral() *PrimaryExpressionRegularExpressionLiteral {
	t := p.tokenizer.CurrentToken
	p.tokenizer.MustMatch(TRegularExpression)
	var flags string
	if p.tokenizer.CurrentToken.Type == TIdentifier {
		flags = p.tokenizer.CurrentToken.Value
		p.tokenizer.Next()
	}
	return &PrimaryExpressionRegularExpressionLiteral{
		Pattern: t.Value,
		Flags:   flags,
	}
}

func (p *Parser) objectLiteral() *PrimaryExpressionObjectLiteral {
	p.tokenizer.MustMatch(TLeftBrace)
	list := p.propertyDefinitionList()
	p.tokenizer.MustMatch(TRightBrace)
	return &PrimaryExpressionObjectLiteral{
		PropertyList: list,
	}
}
func (p *Parser) propertyDefinitionList() *PropertyDefinitionList {
	var list []PropertyDefinition
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightBrace {
			break
		}
		if t.Type == TComma {
			p.tokenizer.Next()
			continue
		}
		prop := p.propertyDefinition(MethodDefinitionTypeNil)
		list = append(list, prop)
		if p.tokenizer.CurrentToken.Type == TComma {
			p.tokenizer.Next()
		}
	}
	return &PropertyDefinitionList{
		Items: list,
	}
}

func (p *Parser) propertyDefinition(methodType MethodDefinitionType) PropertyDefinition {
	t := p.tokenizer.CurrentToken
	var propertyName PropertyName

	// Prec: TComma + 1
	accept := p.acceptContext(TYield)
	switch t.Type {
	case TDotDotDot:
		p.tokenizer.Next()
		expr := p.expression(accept)
		return &PropertyDefinitionSpread{
			Spread: expr,
		}
	case TIdentifier:
		identifierRef := p.identifierReference()
		if methodType == MethodDefinitionTypeNil {
			isGet := identifierRef.Identifier == "get"
			isSet := identifierRef.Identifier == "set"
			if isGet {
				return p.propertyDefinition(MethodDefinitionTypeGet)
			}
			if isSet {
				return p.propertyDefinition(MethodDefinitionTypeSet)
			}
		}
		if p.tokenizer.CurrentToken.Type != TColon && p.tokenizer.CurrentToken.Type != TLeftParen {
			return &PropertyDefinitionIdentifierReference{
				IdentifierReference: identifierRef,
			}
		}
		propertyName = &PropertyNameLiteralIdentifier{
			Identifier: identifierRef.Identifier,
		}
	case TString:
		stringLiteral := p.stringLiteral()
		propertyName = &PropertyNameLiteralString{
			StringLiteral: stringLiteral,
		}
	case TNumber:
		numberLiteral := p.numericLiteral()
		propertyName = &PropertyNameLiteralNumeric{
			NumericLiteral: numberLiteral,
		}
	case TLeftBracket:
		p.tokenizer.Next()
		computedPropertyName := p.expression(p.acceptContextLowest())
		propertyName = &PropertyNameComputed{
			Expression: computedPropertyName,
		}
		p.tokenizer.MustMatch(TRightBracket)
	default:
		panic("propertyDefinition: unexpected token")
	}
	if p.tokenizer.Match(TColon) {
		value := p.expression(accept)
		return &PropertyDefinitionNameAndExpression{
			Name:       propertyName,
			Expression: value,
		}
	} else if p.tokenizer.Match(TLeftParen) {
		start := p.tokenizer.Index
		formalParameters := p.formalParameters()
		p.tokenizer.MustMatch(TRightParen)
		p.tokenizer.MustMatch(TLeftBrace)
		body := p.functionBody(FunctionTypeNormal)
		p.tokenizer.MustMatch(TRightBrace)
		sourceText := p.SourceText[start:p.tokenizer.Index]
		var m = MethodDefinitionTypeMethod
		if methodType != MethodDefinitionTypeNil {
			m = methodType
		}
		return &PropertyDefinitionMethodDefinition{
			Type:         m,
			PropertyName: propertyName,
			FunctionExpression: &PrimaryExpressionFunctionExpression{
				Identifier:       "",
				FormalParameters: formalParameters,
				SourceText:       sourceText,
				Body:             body,
			},
		}
	} else {
		panic("propertyDefinition: expected colon or left paren")
	}
}

func (p *Parser) arrayLiteral() *PrimaryExpressionArrayLiteral {
	p.tokenizer.MustMatch(TLeftBracket)
	var list []ArrayElement
	accept := p.acceptContext(TComma)
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightBracket {
			p.tokenizer.Next()
			break
		}
		if p.tokenizer.Match(TComma) {
			list = append(list, &ArrayElementElision{})
		} else if p.tokenizer.Match(TDotDotDot) {
			expr := p.expression(accept)
			list = append(list, &ArrayElementSpread{
				Spread: expr,
			})
			p.tokenizer.Match(TComma)
		} else {
			expr := p.expression(accept)
			list = append(list, &ArrayElementExpression{
				Expression: expr,
			})
			p.tokenizer.Match(TComma)
		}
	}
	return &PrimaryExpressionArrayLiteral{
		ElementList: list,
	}
}

func (p *Parser) literal() Literal {
	t := p.tokenizer.CurrentToken
	defer p.tokenizer.Next()
	switch t.Type {
	case TTrue, TFalse:
		return &LiteralBoolean{
			Bool: t.Type == TTrue,
		}
	case TNull:
		return &LiteralNull{}
	case TUndefined:
		return &LiteralUndefined{}
	case TNumber:
		return p.numericLiteral()
	case TString:
		return p.stringLiteral()
	case TComment:
		return &LiteralUndefined{}
	default:
		panic("literal: unhandled token")
	}
}

func (p *Parser) numericLiteral() *LiteralNumeric {
	t := p.tokenizer.CurrentToken
	if t.Type != TNumber {
		panic("numericLiteral: expected number")
	}
	return &LiteralNumeric{
		Value: t.Value,
	}
}

func (p *Parser) stringLiteral() *LiteralString {
	t := p.tokenizer.CurrentToken
	if t.Type != TString {
		panic("stringLiteral: expected string")
	}
	return &LiteralString{
		Value: t.Value,
	}
}

func (p *Parser) variableStatement() *StatementVariable {
	p.tokenizer.MustMatch(TVar)
	list := p.variableDeclarationList()
	return &StatementVariable{
		DeclarationList: list,
	}
}
func (p *Parser) variableDeclarationList() *VariableDeclarationList {
	var list []*VariableDeclaration
	for {
		declaration := p.variableDeclaration()
		list = append(list, declaration)
		if p.tokenizer.CurrentToken.Type == TComma {
			p.tokenizer.Next()
			continue
		}
		if p.tokenizer.CurrentToken.Type == TSemicolon {
			p.tokenizer.Next()
			break
		}
		if p.tokenizer.CurrentToken.Type == TEOF {
			break
		}
	}
	return &VariableDeclarationList{Items: list}

}

func (p *Parser) variableDeclaration() *VariableDeclaration {
	identifier := p.bindingIdentifier()
	var init Expression
	if p.tokenizer.Match(TEquals) {
		// Precedence greater than TComma
		init = p.expression(p.acceptContext(TYield))
	}
	return &VariableDeclaration{
		Identifier:  identifier,
		Initializer: init,
	}
}

func (p *Parser) blockStatement() *BlockStatementBlock {
	return &BlockStatementBlock{
		Block: p.block(),
	}
}

func (p *Parser) block() *Block {
	p.tokenizer.MustMatch(TLeftBrace)
	list := p.statementList()
	p.tokenizer.MustMatch(TRightBrace)

	return &Block{
		StatementList: list,
	}
}
