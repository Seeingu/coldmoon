package coldmoon

import "strconv"

type node interface {
	String() string
	Bytecode(e *Executable)
}

type AnalyzeQuery int

const (
	AnalyzeQueryIsReference AnalyzeQuery = iota
)

// MARK: - IdentifierReference

type IdentifierReference struct {
	node
	Identifier IdentifierName
}

func (i *IdentifierReference) Analyze(a AnalyzeQuery) bool {
	return a == AnalyzeQueryIsReference
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
	Expression
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

func (p *PrimaryExpressionIdentifierReference) Analyze(a AnalyzeQuery) bool {
	if a == AnalyzeQueryIsReference {
		return true
	}
	return false
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

func (p *PrimaryExpressionLiteral) Analyze(a AnalyzeQuery) bool {
	return false
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
func (p *PrimaryExpressionThis) Analyze(a AnalyzeQuery) bool {
	return false
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

func (p *PrimaryExpressionParenthesizedExpression) Analyze(a AnalyzeQuery) bool {
	return p.Expression.Analyze(a)
}

// MARK: - MemberExpression

type ASTProperty interface {
	String() string
}

type ASTPropertyIdentifier struct {
	ASTProperty
	Identifier IdentifierName
}

func (a *ASTPropertyIdentifier) String() string {
	return string(a.Identifier)
}

type ASTPropertyExpression struct {
	ASTProperty
	Expression Expression
}

func (a *ASTPropertyExpression) String() string {
	return a.Expression.String()
}

type MemberExpression struct {
	Expression
	Member   Expression
	Property ASTProperty
}

func (m *MemberExpression) Analyze(a AnalyzeQuery) bool {
	return a == AnalyzeQueryIsReference
}

func (m *MemberExpression) Bytecode(e *Executable) {
	m.Member.Bytecode(e)
	if m.Member.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
	strict := false

	switch prop := m.Property.(type) {
	case *ASTPropertyExpression:
		prop.Expression.Bytecode(e)
		e.AddInstruction(InsLoad)
		e.AddInstruction(&IEvaluatePropertyAccessWithExpressionKey{
			Strict: strict,
		})
	case *ASTPropertyIdentifier:
		e.AddInstruction(&IEvaluatePropertyAccessWithIdentifierKey{
			Strict: strict,
			Name:   prop.Identifier,
		})
	}
}

func (m *MemberExpression) String() string {
	return m.Member.String() + "." + m.Property.String()
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

// MARK: - LiteralNumeric

type NumericSystem int

const (
	NumericSystemDecimal NumericSystem = iota
	NumericSystemBinary
	NumericSystemOctal
	NumericSystemHex
)

type NumericProduction int

const (
	NumericProductionRegular NumericProduction = iota
	NumericProductionLegacyOctal
	NumericProductionNonOctalDecimalIntegerLiteral
)

type NumericType int

const (
	NumericTypeNumber NumericType = iota
	NumericTypeBigInt
)

type LiteralNumeric struct {
	Literal
	Value string
}

func (l *LiteralNumeric) NumericValue() (Value, error) {
	num, err := strconv.ParseFloat(l.Value, 64)
	if err != nil {
		return nil, err
	}
	return NewNumberValue(num), nil
}

func (l *LiteralNumeric) Bytecode(e *Executable) {
	v, err := l.NumericValue()
	if err != nil {
		panic(err)
	}
	e.AddInstruction(&IStoreConstant{Value: v})
}

func (l *LiteralNumeric) String() string {
	return l.Value
}

// MARK: - LiteralString

type LiteralString struct {
	Literal
	Value string
}

// 12.9.4.2
func (l *LiteralString) StringValue() Value {
	return NewStringValue(l.Value)
}

func (l *LiteralString) Bytecode(e *Executable) {
	e.AddInstruction(&IStoreConstant{Value: l.StringValue()})
}

func (l *LiteralString) String() string {
	return l.Value
}

// MARK: - Expression

type Expression interface {
	node
	Analyze(a AnalyzeQuery) bool
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

func (e *ExpressionPrimary) Analyze(a AnalyzeQuery) bool {
	return e.PrimaryExpression.Analyze(a)
}

// MARK: - UnaryExpression

type UnaryOperator int

const (
	UnaryOperatorDelete UnaryOperator = iota
	UnaryOperatorVoid
	UnaryOperatorTypeof
)

type UnaryExpression struct {
	Expression
	Operator UnaryOperator
	Operand  Expression
}

func (u *UnaryExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (u *UnaryExpression) Bytecode(e *Executable) {
	u.Operand.Bytecode(e)
	switch u.Operator {
	case UnaryOperatorDelete:
		panic("unimplemented")
	case UnaryOperatorVoid:
		u.Operand.Bytecode(e)
		if u.Operand.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
	case UnaryOperatorTypeof:
		u.Operand.Bytecode(e)
		e.AddInstruction(InsTypeof)
	}
}

func (u *UnaryExpression) String() string {
	return "UnaryExpression " + u.Operand.String()
}

// MARK: - CallExpression

type Arguments []Expression

type CallExpression struct {
	Expression
	Callee    Expression
	Arguments Arguments
}

func (c *CallExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (c *CallExpression) Bytecode(e *Executable) {
	c.Callee.Bytecode(e)

	e.AddInstruction(&ISetEvaluationContextReference{})
	isReference := c.Callee.Analyze(AnalyzeQueryIsReference)
	if isReference {
		e.AddInstruction(InsGetValue)
	}

	e.AddInstruction(InsLoadThisValue)
	for _, arg := range c.Arguments {
		arg.Bytecode(e)
		if arg.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(InsLoad)
	}

	e.AddInstruction(&ICall{ArgumentCount: len(c.Arguments)})
}

func (c *CallExpression) String() string {
	sb := c.Callee.String() + "("
	for i, arg := range c.Arguments {
		if i != 0 {
			sb += ", "
		}
		sb += arg.String()
	}
	sb += ")"
	return sb
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
	if s.Expression.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
}

func (s *StatementExpression) String() string {
	return "ExpressionStatement " + s.Expression.String()
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

// MARK: - ThrowStatement

type StatementThrow struct {
	Statement
	Expression Expression
}

func (s *StatementThrow) Bytecode(e *Executable) {
	s.Expression.Bytecode(e)
	if s.Expression.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(&ILoad{})
	e.AddInstruction(&IThrow{})
}

func (s *StatementThrow) String() string {
	return "Throw " + s.Expression.String()
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

	if s.Condition.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
	jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
	e.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(e.Instructions)
	e.AddInstruction(&IStoreConstant{Value: UndefinedValue})
	s.Consequent.Bytecode(e)
	endJump := &IJump{Target: 0}
	e.AddInstruction(endJump)

	// else
	jumpIfTrue.TargetElse = len(e.Instructions)
	e.AddInstruction(&IStoreConstant{Value: UndefinedValue})

	if s.Alternate != nil {
		s.Alternate.Bytecode(e)
	}
	endJump.Target = len(e.Instructions)
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

// MARK: - WhileStatement

type StatementWhile struct {
	IterationStatement
	Condition Expression
	Body      Statement
}

func (s *StatementWhile) Bytecode(e *Executable) {
	e.AddInstruction(&ILoadConstant{Value: UndefinedValue})

	condition := len(e.Instructions)
	s.Condition.Bytecode(e)
	if s.Condition.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}

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

// MARK: - DoWhileStatement
type StatementDoWhile struct {
	IterationStatement
	Condition Expression
	Body      Statement
}

func (s *StatementDoWhile) Bytecode(e *Executable) {
	e.AddInstruction(&ILoadConstant{Value: UndefinedValue})

	bodyIndex := len(e.Instructions)

	e.AddInstruction(&IStore{})
	s.Body.Bytecode(e)
	e.AddInstruction(&ILoad{})

	s.Condition.Bytecode(e)
	if s.Condition.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}

	e.AddInstruction(&ILoad{})
	jumpIfTrue := &IJumpIfTrue{Target: bodyIndex, TargetElse: 0}
	jumpIfTrue.TargetElse = len(e.Instructions)

	e.AddInstruction(&IStore{})
}

func (s *StatementDoWhile) String() string {
	sb := "DoWhile"
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
