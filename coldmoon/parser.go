package coldmoon

type Parser struct {
	SourceText string
	tokenizer  *Tokenizer
}

func NewParser(sourceText string) *Parser {
	return &Parser{
		SourceText: sourceText,
		tokenizer:  NewTokenizer(sourceText),
	}
}

func (p *Parser) Parse() *Script {
	list := p.statementList()
	return &Script{
		StatementList: list,
	}
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

func (p *Parser) statementListItem() StatementListItem {
	s := p.statement()
	if s == nil {
		p.automaticSemicolonInsertion()
		return nil
	}
	return &StatementListItemStatement{
		Statement: s,
	}
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

func (p *Parser) noLineTerminatorHere() {
	// TODO
}

func (p *Parser) breakableStatement() *BreakableStatement {
	return &BreakableStatement{
		IterationStatement: p.iterationStatement(),
	}
}

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

func (p *Parser) expressionStatement() *ExpressionStatement {
	expr := p.expression()
	t := p.tokenizer.CurrentToken
	if t.Type == TSemicolon {
		p.tokenizer.Next()
	}
	if expr == nil {
		panic("expressionStatement: expected expression")
	}

	return &ExpressionStatement{
		Expression: expr,
	}
}

func (p *Parser) unaryExpression() Expression {
	t := p.tokenizer.CurrentToken
	var operator UnaryOperator
	unaryMap := map[TokenType]UnaryOperator{
		TDelete: UnaryOperatorDelete,
		TVoid:   UnaryOperatorVoid,
		TTypeof: UnaryOperatorTypeof,
		TPlus:   UnaryOperatorAddition,
		TMinus:  UnaryOperatorSubtraction,
	}
	if op, ok := unaryMap[t.Type]; ok {
		p.tokenizer.Next()
		operator = op
	} else {
		panic("unimplemented")
	}
	expr := p.expression()
	return &UnaryExpression{
		Operator: operator,
		Operand:  expr,
	}

}

func (p *Parser) expression() Expression {
	t := p.tokenizer.CurrentToken
	if t.Type == TVoid {
		return p.unaryExpression()
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
	if t.Type != TIdentifier {
		panic("identifierReference: expected identifierOrKeyword")
	}
	name := t.Value
	p.tokenizer.Next()
	return &IdentifierReference{
		Identifier: IdentifierName(name),
	}
}

func (p *Parser) primaryExpression() PrimaryExpression {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TThis:
		p.tokenizer.Next()
		return &ExpressionPrimary{
			PrimaryExpression: &PrimaryExpressionThis{},
		}
	case TIdentifier:
		return p.identifierReference()
	case TLeftParen:
		return p.parenthesizedExpression()
	default:
		literal := p.literal()

		return &ExpressionPrimary{
			PrimaryExpression: &PrimaryExpressionLiteral{
				Literal: literal,
			},
		}
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
		return &LiteralNumeric{
			Value: t.Value,
		}
	case TString:
		return &LiteralString{
			Value: t.Value,
		}
	default:
		panic("literal: unhandled token")
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
