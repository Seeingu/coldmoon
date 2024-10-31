package parser

type node interface {
	String() string
}

type PrimaryExpression interface {
	node
}

type PrimaryExpressionLiteral struct {
	PrimaryExpression
	Literal Literal
}

func (p *PrimaryExpressionLiteral) String() string {
	return p.Literal.String()
}

type Literal interface {
	String() string
}

type LiteralNull struct {
	Literal
}

func (l *LiteralNull) String() string {
	return "null"
}

type LiteralBoolean struct {
	Literal
	Bool bool
}

func (l *LiteralBoolean) String() string {
	if l.Bool {
		return "true"
	}
	return "false"
}

type LiteralNumeric struct {
	Literal
}

func (l *LiteralNumeric) String() string {
	return "TODO: LiteralNumeric"
}

type LiteralString struct {
	Literal
}

func (l *LiteralString) String() string {
	return "TODO: LiteralString"
}

type Expression interface {
	node
}

type ExpressionPrimary struct {
	Expression
	PrimaryExpression PrimaryExpression
}

func (e *ExpressionPrimary) String() string {
	return e.PrimaryExpression.String()
}

type Statement interface {
	node
}

type StatementBlock struct {
	Statement
	BlockStatement BlockStatement
}

func (s *StatementBlock) String() string {
	return s.BlockStatement.String()
}

type StatementEmpty struct {
	Statement
}

func (s *StatementEmpty) String() string {
	return ""
}

type StatementExpression struct {
	Statement
}

func (s *StatementExpression) String() string {
	return "TODO: StatementExpression"
}

type Declaration interface {
	node
}

type BlockStatement interface {
	node
}

type BlockStatementBlock struct {
	BlockStatement
	Block *Block
}

func (b *BlockStatementBlock) String() string {
	return b.Block.StatementList.String()
}

type Block struct {
	StatementList StatementList
}

func (b *Block) String() string {
	return b.StatementList.String()
}

type StatementList []StatementListItem

func (s StatementList) String() string {
	var str string
	for _, item := range s {
		switch t := item.(type) {
		case *StatementListItemStatement:
			str += t.Statement.String()
		case *StatementListItemDeclaration:
			str += t.Declaration.String()
		}
	}
	return str
}

type StatementListItem interface {
	node
}
type StatementListItemStatement struct {
	StatementListItem
	Statement Statement
}

func (s *StatementListItemStatement) String() string {
	return s.Statement.String()
}

type StatementListItemDeclaration struct {
	StatementListItem
	Declaration Declaration
}

func (s *StatementListItemDeclaration) String() string {
	return s.Declaration.String()
}

type ExpressionStatement struct {
	Expression Expression
}

func (e *ExpressionStatement) String() string {
	return e.Expression.String()
}

type Script struct {
	StatementList StatementList
}

func (s *Script) String() string {
	return s.StatementList.String()
}
