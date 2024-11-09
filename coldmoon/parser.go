package coldmoon

import "github.com/samber/lo"

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
	case TPeriod, TQuestionDot:
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
	case TEqualsGreaterThan:
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
	case precAssocUnaryPlus, precAssocUnaryMinus:
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

// unaryExpression accept unary token
// if token is not unary, return nil
func (p *Parser) unaryExpression() (Expression, bool) {
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

func (p *Parser) expression(accept *acceptContext) Expression {
	unary, ok := p.unaryExpression()
	if ok {
		return unary
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

		secondary := p.secondaryExpression(primary, newAcceptContext)
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
	case TLeftBracket, TPeriod:
		return p.memberExpression(left)
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
	default:
		panic("secondaryExpression: unexpected token")
	}
	return left
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
		accept := p.acceptContext(TLeftBracket)
		propertyExpression := p.expression(accept)
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
		t := p.tokenizer.Peek()
		if t.Type == TRightParen {
			p.tokenizer.Next()
			break
		}
		expr := p.expression(p.acceptContextLowest())
		list = append(list, expr)
	}
	p.tokenizer.MustMatch(TRightParen)
	return list
}

func (p *Parser) parenthesizedExpression() *PrimaryExpressionParenthesizedExpression {
	p.tokenizer.MustMatch(TLeftParen)
	expr := p.expression(p.acceptContextLowest())
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
		computedPropertyName := p.expression(
			p.acceptContext(TLeftBracket))
		propertyName = &PropertyNameComputed{
			Expression: computedPropertyName,
		}
		p.tokenizer.MustMatch(TRightBracket)
	default:
		panic("propertyDefinition: unexpected token")
	}
	p.tokenizer.MustMatch(TColon)
	value := p.expression(p.acceptContext(TComma))
	return &PropertyDefinitionNameAndExpression{
		Name:       propertyName,
		Expression: value,
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
		if p.tokenizer.NextToken.Type == TComma {
			list = append(list, &ArrayElementElision{})
			p.tokenizer.Next()
		} else {
			expr := p.expression(accept)
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
