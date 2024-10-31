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
	t := p.tokenizer.Peek()
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

	default:
		return p.expressionStatement()
	}
}

func (p *Parser) expressionStatement() *ExpressionStatement {
	expr := p.expression()
	t := p.tokenizer.Peek()
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

func (p *Parser) expression() Expression {
	return p.primaryExpression()
}

func (p *Parser) primaryExpression() *ExpressionPrimary {
	t := p.tokenizer.CurrentToken
	if t.Type == TThis {
		p.tokenizer.Next()
		return &ExpressionPrimary{
			PrimaryExpression: &PrimaryExpressionThis{},
		}
	}

	literal := p.literal()

	return &ExpressionPrimary{
		PrimaryExpression: &PrimaryExpressionLiteral{
			Literal: literal,
		},
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
	t := p.tokenizer.CurrentToken
	if t.Type != TLeftBrace {
		panic("block: expected {")
	}
	p.tokenizer.Next()

	list := p.statementList()

	t = p.tokenizer.CurrentToken
	if t.Type != TRightBrace {
		panic("block: expected }")
	}
	p.tokenizer.Next()

	return &Block{
		StatementList: list,
	}
}
