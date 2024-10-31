package coldmoon

type node interface {
	String() string
	Bytecode(e *Executable)
}

type PrimaryExpression interface {
	node
}

type PrimaryExpressionLiteral struct {
	PrimaryExpression
	Literal Literal
}

func (p *PrimaryExpressionLiteral) Bytecode(e *Executable) {
	p.Literal.Bytecode(e)
}

func (p *PrimaryExpressionLiteral) String() string {
	return p.Literal.String()
}

type Literal interface {
	node
}

type LiteralNull struct {
	Literal
}

func (l *LiteralNull) Bytecode(e *Executable) {
	e.AddInstruction(&IStoreConstant{Value: NullValue})
}

func (l *LiteralNull) String() string {
	return "null"
}

type LiteralBoolean struct {
	Literal
	Bool bool
}

func (l *LiteralBoolean) Bytecode(e *Executable) {
	e.AddInstruction(&IStoreConstant{Value: NewBooleanValue(l.Bool)})
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

func (l *LiteralNumeric) Bytecode(e *Executable) {
	panic("TODO: LiteralNumeric")
}

func (l *LiteralNumeric) String() string {
	return "TODO: LiteralNumeric"
}

type LiteralString struct {
	Literal
}

func (l *LiteralString) Bytecode(e *Executable) {
	panic("TODO: LiteralString")
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

func (e *ExpressionPrimary) Bytecode(ex *Executable) {
	e.PrimaryExpression.Bytecode(ex)
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

func (s *StatementBlock) Bytecode(e *Executable) {
	s.BlockStatement.Bytecode(e)
}

func (s *StatementBlock) String() string {
	return s.BlockStatement.String()
}

type StatementEmpty struct {
	Statement
}

func (s *StatementEmpty) Bytecode(e *Executable) {
	// empty
}

func (s *StatementEmpty) String() string {
	return ""
}

type StatementDebugger struct {
	Statement
}

func (s *StatementDebugger) Bytecode(e *Executable) {
	panic("TODO: StatementDebugger")
}

func (s *StatementDebugger) String() string {
	return "debugger"
}

type StatementExpression struct {
	Statement
	Expression Expression
}

func (s *StatementExpression) Bytecode(e *Executable) {
	s.Expression.Bytecode(e)
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

func (b *BlockStatementBlock) Bytecode(e *Executable) {
	b.Block.Bytecode(e)
}

func (b *BlockStatementBlock) String() string {
	return b.Block.StatementList.String()
}

type Block struct {
	StatementList StatementList
}

func (b *Block) Bytecode(e *Executable) {
	b.StatementList.Bytecode(e)
}

func (b *Block) String() string {
	return b.StatementList.String()
}

type StatementList []StatementListItem

func (s StatementList) Bytecode(e *Executable) {
	for _, item := range s {
		item.Bytecode(e)
	}
}

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

var _ node = (*StatementListItemStatement)(nil)

func (s *StatementListItemStatement) Bytecode(e *Executable) {
	s.Statement.Bytecode(e)
}

func (s *StatementListItemStatement) String() string {
	return s.Statement.String()
}

type StatementListItemDeclaration struct {
	StatementListItem
	Declaration Declaration
}

func (s *StatementListItemDeclaration) Bytecode(e *Executable) {
	s.Declaration.Bytecode(e)
}

func (s *StatementListItemDeclaration) String() string {
	return s.Declaration.String()
}

type ExpressionStatement struct {
	Expression Expression
}

func (e *ExpressionStatement) Bytecode(ex *Executable) {
	e.Expression.Bytecode(ex)
}

func (e *ExpressionStatement) String() string {
	return e.Expression.String()
}

type Script struct {
	StatementList StatementList
}

func (s *Script) Bytecode(e *Executable) {
	s.StatementList.Bytecode(e)
}

func (s *Script) String() string {
	return s.StatementList.String()
}
