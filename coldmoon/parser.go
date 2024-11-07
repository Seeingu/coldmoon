package coldmoon

import "github.com/samber/lo"

type Parser struct {
	SourceText     string
	tokenizer      *Tokenizer
	inFunctionBody bool
	ctx            ParserContext
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

func (p *Parser) ParseNode() StatementList {
	return p.statementList()
}

func (p *Parser) statementList() (list StatementList) {
	for {
		item := p.statementListItem()
		if item == nil {
			break
		}
		if _, ok := item.(*StatementListItemStatement).Statement.(*StatementEmpty); ok {
			break
		}
		list = append(list, item)
	}

	return
}

func (p *Parser) functionBody() *FunctionBody {
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
	case TFunction:
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
	expr := p.expression()
	return &StatementThrow{
		Expression: expr,
	}
}

func (p *Parser) formalParameters() *FormalParameters {
	var items []FormalParametersItem
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightParen {
			break
		}
		identifier := p.bindingIdentifier()
		items = append(items, &FormalParameter{
			BindingElement: &BindingElement{
				Identifier: identifier,
			},
		})
		if p.tokenizer.CurrentToken.Type == TComma {
			p.tokenizer.Next()
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
	functionBody := p.functionBody()
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &FunctionDeclaration{
		Identifier:       identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             functionBody,
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
	functionBody := p.functionBody()
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

func (p *Parser) hoistableDeclaration() *DeclarationHoistable {
	t := p.tokenizer.CurrentToken
	if t.Type == TFunction {
		return &DeclarationHoistable{
			FunctionDeclaration: p.functionDeclaration(),
		}
	}
	panic("unimplemented")
}

func (p *Parser) declaration() Declaration {
	t := p.tokenizer.CurrentToken
	if t.Type == TFunction {
		return p.hoistableDeclaration()
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
	}
	return p.whileStatement()
}

func (p *Parser) doWhileStatement() *StatementDoWhile {
	p.tokenizer.MustMatch(TDo)
	body := p.statement()
	p.tokenizer.MustMatch(TWhile)
	p.tokenizer.MustMatch(TLeftParen)
	condition := p.expression()
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
	condition := p.expression()
	p.tokenizer.MustMatch(TRightParen)
	body := p.statement()

	return &StatementWhile{
		Condition: condition,
		Body:      body,
	}
}

// MARK: - Return

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
	expr := p.expression()
	p.tokenizer.MustMatch(TSemicolon)
	return &StatementReturn{
		Expression: expr,
	}
}

// MARK: - Condition

func (p *Parser) ifStatement() *StatementIf {
	p.tokenizer.MustMatch(TIf)
	p.tokenizer.MustMatch(TLeftParen)
	condition := p.expression()
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
	expr := p.expression()
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

// unaryExpression accept unary token
// if token is not unary, return nil
func (p *Parser) unaryExpression() Expression {
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
		return nil
	}
	expr := p.expression()
	return &UnaryExpression{
		Operator: operator,
		Operand:  expr,
	}

}

func (p *Parser) expression() Expression {
	unary := p.unaryExpression()
	if unary != nil {
		return unary
	}

	primary := p.primaryExpression()
	var expr Expression = primary
	for {
		secondary := p.secondaryExpression(primary)
		if secondary == nil {
			return expr
		}
		expr = secondary
	}
}

func (p *Parser) secondaryExpression(left PrimaryExpression) Expression {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TLeftParen:
		return p.callExpression(left)
	case TLeftBracket, TPeriod:
		return p.memberExpression(left)
	default:
		return nil
	}
}

func (p *Parser) memberExpression(left PrimaryExpression) *MemberExpression {
	token := p.tokenizer.CurrentToken
	var property ASTProperty
	if token.Type == TLeftBracket {
		p.tokenizer.Next()
		propertyExpression := p.expression()
		p.tokenizer.MustMatch(TRightBracket)
		property = &ASTPropertyExpression{
			Expression: propertyExpression,
		}
	} else if token.Type == TPeriod {
		p.tokenizer.Next()
		identifier := p.tokenizer.Next()
		if identifier.Type != TIdentifier {
			panic("memberExpression: expected identifier")
		}
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
		t := p.tokenizer.Peek()
		if t.Type == TRightParen {
			p.tokenizer.Next()
			break
		}
		expr := p.expression()
		list = append(list, expr)
	}
	p.tokenizer.MustMatch(TRightParen)
	return list
}

func (p *Parser) parenthesizedExpression() *PrimaryExpressionParenthesizedExpression {
	p.tokenizer.MustMatch(TLeftParen)
	expr := p.expression()
	p.tokenizer.MustMatch(TRightParen)
	return &PrimaryExpressionParenthesizedExpression{
		Expression: expr,
	}
}

func (p *Parser) identifierReference() *IdentifierReference {
	t := p.tokenizer.CurrentToken
	types := []TokenType{TIdentifier, TAwait, TYield}
	if !lo.Contains(types, t.Type) {
		panic("identifierReference: expected identifierOrKeyword")
	}
	name := t.Value
	p.tokenizer.Next()
	return &IdentifierReference{
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
		return p.parenthesizedExpression()
	case TFunction:
		return p.functionExpression()
	default:
		literal := p.literal()

		return &ExpressionPrimary{
			PrimaryExpression: &PrimaryExpressionLiteral{
				Literal: literal,
			},
		}
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
		prop := p.propertyDefinition()
		list = append(list, prop)
		if p.tokenizer.CurrentToken.Type == TComma {
			p.tokenizer.Next()
		}
	}
	return &PropertyDefinitionList{
		Items: list,
	}
}

func (p *Parser) propertyDefinition() PropertyDefinition {
	t := p.tokenizer.CurrentToken
	var propertyName PropertyName
	switch t.Type {
	case TIdentifier:
		identifierRef := p.identifierReference()
		if p.tokenizer.CurrentToken.Type != TColon {
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
		computedPropertyName := p.expression()
		propertyName = &PropertyNameComputed{
			Expression: computedPropertyName,
		}
		p.tokenizer.MustMatch(TRightBracket)
	default:
		panic("propertyDefinition: unexpected token")
	}
	p.tokenizer.MustMatch(TColon)
	value := p.expression()
	return &PropertyDefinitionNameAndExpression{
		Name:       propertyName,
		Expression: value,
	}
}

func (p *Parser) arrayLiteral() *PrimaryExpressionArrayLiteral {
	p.tokenizer.MustMatch(TLeftBracket)
	var list []ArrayElement
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightBracket {
			p.tokenizer.Next()
			break
		}
		if p.tokenizer.NextToken.Type == TComma {
			list = append(list, &ArrayElementElision{})
			p.tokenizer.Next()
		} else {
			expr := p.expression()
			list = append(list, &ArrayElementExpression{
				Expression: expr,
			})
			if p.tokenizer.CurrentToken.Type == TComma {
				p.tokenizer.Next()
			}
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
	case TNumber:
		return p.numericLiteral()
	case TString:
		return p.stringLiteral()
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
