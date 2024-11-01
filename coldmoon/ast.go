package coldmoon

type node interface {
	String() string
	Bytecode(e *Executable)
}

// MARK: - IdentifierReference

type IdentifierReference struct {
	node
	Identifier IdentifierName
}

func (i *IdentifierReference) Bytecode(e *Executable) {
	e.AddInstruction(&IResolveBinding{Name: i.Identifier})
}

func (i *IdentifierReference) String() string {
	return string(i.Identifier)
}

type IdentifierName string

// MARK: - PrimaryExpression

type PrimaryExpression interface {
	node
}

type PrimaryExpressionIdentifierReference struct {
	PrimaryExpression
	IdentifierReference *IdentifierReference
}

func (p *PrimaryExpressionIdentifierReference) Bytecode(e *Executable) {
	p.IdentifierReference.Bytecode(e)
}
func (p *PrimaryExpressionIdentifierReference) String() string {
	return p.IdentifierReference.String()
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

type PrimaryExpressionThis struct {
	PrimaryExpression
}

func (p *PrimaryExpressionThis) Bytecode(e *Executable) {
	e.AddInstruction(&IResolveThisBinding{})
}
func (p *PrimaryExpressionThis) String() string {
	return "this"
}

type PrimaryExpressionParenthesizedExpression struct {
	PrimaryExpression
	Expression Expression
}

func (p *PrimaryExpressionParenthesizedExpression) Bytecode(e *Executable) {
	p.Expression.Bytecode(e)
}

func (p *PrimaryExpressionParenthesizedExpression) String() string {
	return "(" + p.Expression.String() + ")"
}

// MARK: - Literal

type Literal interface {
	node
	// 13.2.3.1
	// Bytecode
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

// MARK: - Condition

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

// MARK: - Statement
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

// MARK: - DebuggerStatement

type StatementDebugger struct {
	Statement
}

func (s *StatementDebugger) Bytecode(e *Executable) {
	// TODO: implement
}
func (s *StatementDebugger) String() string {
	return "debugger"
}

// MARK: - ExpressionStatement
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

// MARK: - BreakableStatement

type BreakableStatement struct {
	IterationStatement IterationStatement
}

func (b *BreakableStatement) Bytecode(e *Executable) {
	b.IterationStatement.Bytecode(e)
}

func (b *BreakableStatement) String() string {
	return b.IterationStatement.String()
}

// MARK: - IfStatement

type StatementIf struct {
	Statement
	Condition  Expression
	Consequent Statement
	Alternate  Statement
}

// 14.6.2
func (s *StatementIf) Bytecode(e *Executable) {
	s.Condition.Bytecode(e)

	e.AddInstruction(&ILoad{})
	jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
	e.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(e.Instructions)
	e.AddInstruction(&IStoreConstant{Value: UndefinedValue})
	s.Consequent.Bytecode(e)

	if s.Alternate != nil {
		jump := &IJump{Target: 0}
		e.AddInstruction(jump)

		e.AddInstruction(&IStoreConstant{Value: UndefinedValue})
		s.Alternate.Bytecode(e)
		jump.Target = len(e.Instructions)
	} else {
		jumpIfTrue.TargetElse = len(e.Instructions)
	}
}

func (s *StatementIf) String() string {
	sb := "If"
	sb += " " + s.Condition.String() + " \n"
	sb += s.Consequent.String()
	if s.Alternate != nil {
		sb += "Else\n"
		sb += s.Alternate.String()
	}
	return sb
}

// MARK: - IterationStatement

type IterationStatement interface {
	node
}

type StatementWhile struct {
	IterationStatement
	Condition Expression
	Body      Statement
}

func (s *StatementWhile) Bytecode(e *Executable) {
	e.AddInstruction(&ILoadConstant{Value: UndefinedValue})

	condition := len(e.Instructions)
	s.Condition.Bytecode(e)

	e.AddInstruction(&ILoad{})
	jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
	e.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(e.Instructions)
	e.AddInstruction(&IStore{})
	s.Body.Bytecode(e)
	e.AddInstruction(&ILoad{})

	e.AddInstruction(&IJump{Target: condition})

	jumpIfTrue.TargetElse = len(e.Instructions)
	e.AddInstruction(&IStore{})
}

func (s *StatementWhile) String() string {
	sb := "While"
	sb += " " + s.Condition.String() + " \n"
	sb += s.Body.String()
	return sb
}

// MARK: - Declaration

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

// 14.2.2
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

// 14.5.1
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
