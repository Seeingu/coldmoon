package coldmoon

import (
	"strconv"
)

type ASTNode interface {
	String() string
	Bytecode(e *Executable, c *BytecodeContext)
}

type BytecodeContext struct {
	agent                 *Agent
	containedInStrictCode bool
}

type AnalyzeQuery int

const (
	AnalyzeQueryIsReference AnalyzeQuery = iota
	AnalyzeQueryIsStringLiteral
)

// MARK: - IdentifierReference

type IdentifierReference struct {
	ASTNode
	Identifier IdentifierName
}

func (i *IdentifierReference) Analyze(a AnalyzeQuery) bool {
	return a == AnalyzeQueryIsReference
}

func (i *IdentifierReference) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IResolveBinding{Name: i.Identifier, Strict: c.containedInStrictCode})
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

func (p *PrimaryExpressionIdentifierReference) Bytecode(e *Executable, c *BytecodeContext) {
	p.IdentifierReference.Bytecode(e, c)
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

func (p *PrimaryExpressionLiteral) Bytecode(e *Executable, c *BytecodeContext) {
	p.Literal.Bytecode(e, c)
}
func (p *PrimaryExpressionLiteral) String() string {
	return p.Literal.String()
}

func (p *PrimaryExpressionLiteral) Analyze(a AnalyzeQuery) bool {
	switch a {
	case AnalyzeQueryIsReference:
		return false
	case AnalyzeQueryIsStringLiteral:
		return p.Literal.Analyze(a)
	default:
		panic("unreachable")
	}
}

type PrimaryExpressionThis struct {
	PrimaryExpression
}

func (p *PrimaryExpressionThis) Bytecode(e *Executable, c *BytecodeContext) {
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

func (p *PrimaryExpressionParenthesizedExpression) Bytecode(e *Executable, c *BytecodeContext) {
	p.Expression.Bytecode(e, c)
}

func (p *PrimaryExpressionParenthesizedExpression) String() string {
	return "(" + p.Expression.String() + ")"
}

func (p *PrimaryExpressionParenthesizedExpression) Analyze(a AnalyzeQuery) bool {
	switch a {
	case AnalyzeQueryIsReference:
		return p.Expression.Analyze(a)
	case AnalyzeQueryIsStringLiteral:
		return false
	default:
		panic("unreachable")
	}
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

func (m *MemberExpression) Bytecode(e *Executable, c *BytecodeContext) {
	m.Member.Bytecode(e, c)
	if m.Member.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
	strict := c.containedInStrictCode

	switch prop := m.Property.(type) {
	case *ASTPropertyExpression:
		prop.Expression.Bytecode(e, c)
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
	ASTNode
	Analyze(a AnalyzeQuery) bool
	// 13.2.3.1
	// Bytecode
}

type LiteralNull struct {
	Literal
}

var _ Literal = (*LiteralNull)(nil)

func (l *LiteralNull) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IStoreConstant{Value: NullValue})
}

func (l *LiteralNull) String() string {
	return "null"
}

func (l *LiteralNull) Analyze(a AnalyzeQuery) bool {
	return false
}

type LiteralBoolean struct {
	Literal
	Bool bool
}

var _ Literal = (*LiteralBoolean)(nil)

func (l *LiteralBoolean) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IStoreConstant{Value: NewBooleanValue(l.Bool)})
}

func (l *LiteralBoolean) String() string {
	if l.Bool {
		return "true"
	}
	return "false"
}

func (l *LiteralBoolean) Analyze(a AnalyzeQuery) bool {
	return false
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

var _ Literal = (*LiteralNumeric)(nil)

func (l *LiteralNumeric) NumericValue() (Value, error) {
	num, err := strconv.ParseFloat(l.Value, 64)
	if err != nil {
		return nil, err
	}
	return NewNumberValue(num), nil
}

func (l *LiteralNumeric) Bytecode(e *Executable, c *BytecodeContext) {
	v, err := l.NumericValue()
	if err != nil {
		panic(err)
	}
	e.AddInstruction(&IStoreConstant{Value: v})
}

func (l *LiteralNumeric) Analyze(a AnalyzeQuery) bool {
	return false
}

func (l *LiteralNumeric) String() string {
	return l.Value
}

// MARK: - LiteralString

type LiteralString struct {
	Literal
	Value string
}

var _ Literal = (*LiteralString)(nil)

// 12.9.4.2
func (l *LiteralString) StringValue() Value {
	return NewStringValue(l.Value)
}

func (l *LiteralString) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IStoreConstant{Value: l.StringValue()})
}

func (l *LiteralString) String() string {
	return l.Value
}

func (l *LiteralString) Analyze(a AnalyzeQuery) bool {
	return a == AnalyzeQueryIsStringLiteral
}

// MARK: - Expression

type Expression interface {
	ASTNode
	Analyze(a AnalyzeQuery) bool
}

type ExpressionPrimary struct {
	Expression
	PrimaryExpression PrimaryExpression
}

func (e *ExpressionPrimary) Bytecode(ex *Executable, c *BytecodeContext) {
	e.PrimaryExpression.Bytecode(ex, c)
}

func (e *ExpressionPrimary) String() string {
	return e.PrimaryExpression.String()
}

func (e *ExpressionPrimary) Analyze(a AnalyzeQuery) bool {
	switch a {
	case AnalyzeQueryIsReference:
		return e.PrimaryExpression.Analyze(a)
	case AnalyzeQueryIsStringLiteral:
		return e.PrimaryExpression.Analyze(a)
	default:
		panic("unreachable")
	}
}

// MARK: - UnaryExpression

type UnaryOperator int

const (
	UnaryOperatorDelete UnaryOperator = iota
	UnaryOperatorVoid
	UnaryOperatorTypeof
	UnaryOperatorAddition
	UnaryOperatorSubtraction
	UnaryOperatorLogicalNot
	UnaryOperatorBitwiseNot
)

type UnaryExpression struct {
	Expression
	Operator UnaryOperator
	Operand  Expression
}

func (u *UnaryExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (u *UnaryExpression) Bytecode(e *Executable, c *BytecodeContext) {
	u.Operand.Bytecode(e, c)
	switch u.Operator {
	case UnaryOperatorDelete:
		panic("unimplemented")
	case UnaryOperatorVoid:
		u.Operand.Bytecode(e, c)
		if u.Operand.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
	case UnaryOperatorTypeof:
		u.Operand.Bytecode(e, c)
		e.AddInstruction(InsTypeof)
	case UnaryOperatorAddition:
		u.Operand.Bytecode(e, c)

		if u.Operand.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(InsToNumber)
	case UnaryOperatorSubtraction:
		u.Operand.Bytecode(e, c)

		if u.Operand.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(InsToNumeric)
		e.AddInstruction(InsUnaryMinus)
	case UnaryOperatorLogicalNot:
		u.Operand.Bytecode(e, c)

		if u.Operand.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(&ILogicalNot{})
	case UnaryOperatorBitwiseNot:
		u.Operand.Bytecode(e, c)

		if u.Operand.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(&IBitwiseNot{})
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

var _ Expression = (*CallExpression)(nil)

func (c *CallExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (c *CallExpression) Bytecode(e *Executable, bc *BytecodeContext) {
	c.Callee.Bytecode(e, bc)

	e.AddInstruction(&ISetEvaluationContextReference{})
	isReference := c.Callee.Analyze(AnalyzeQueryIsReference)
	if isReference {
		e.AddInstruction(InsGetValue)
	}

	e.AddInstruction(InsLoadThisValue)
	for _, arg := range c.Arguments {
		arg.Bytecode(e, bc)
		if arg.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(InsLoad)
	}

	strict := bc.containedInStrictCode

	e.AddInstruction(&ICall{ArgumentCount: len(c.Arguments), Strict: strict})
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
	ASTNode
	Analyze(a AnalyzeQuery) bool
}

type StatementBlock struct {
	Statement
	BlockStatement BlockStatement
}

var _ Statement = (*StatementBlock)(nil)

func (s *StatementBlock) Analyze(a AnalyzeQuery) bool {
	return s.BlockStatement.Analyze(a)
}

func (s *StatementBlock) Bytecode(e *Executable, c *BytecodeContext) {
	s.BlockStatement.Bytecode(e, c)
}

func (s *StatementBlock) String() string {
	return s.BlockStatement.String()
}

type StatementEmpty struct {
	Statement
}

var _ Statement = (*StatementEmpty)(nil)

func (s *StatementEmpty) Bytecode(e *Executable, c *BytecodeContext) {
	// empty
}
func (s *StatementEmpty) String() string {
	return ""
}

// MARK: - DebuggerStatement

type StatementDebugger struct {
	Statement
}

var _ Statement = (*StatementDebugger)(nil)

func (s *StatementDebugger) Analyze(a AnalyzeQuery) bool {
	return false
}

func (s *StatementDebugger) Bytecode(e *Executable, c *BytecodeContext) {
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

var _ Statement = (*StatementExpression)(nil)

func (s *StatementExpression) Analyze(a AnalyzeQuery) bool {
	return s.Expression.Analyze(a)
}

func (s *StatementExpression) Bytecode(e *Executable, c *BytecodeContext) {
	s.Expression.Bytecode(e, c)
	if s.Expression.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
}

func (s *StatementExpression) String() string {
	return "ExpressionStatement " + s.Expression.String()
}

// MARK: - BreakableStatement

type BreakableStatement struct {
	Statement
	IterationStatement IterationStatement
}

func (b *BreakableStatement) Bytecode(e *Executable, c *BytecodeContext) {
	b.IterationStatement.Bytecode(e, c)
}

func (b *BreakableStatement) String() string {
	return b.IterationStatement.String()
}

// MARK: - ThrowStatement

type StatementThrow struct {
	Statement
	Expression Expression
}

func (s *StatementThrow) Bytecode(e *Executable, c *BytecodeContext) {
	s.Expression.Bytecode(e, c)
	if s.Expression.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsThrow)
}

func (s *StatementThrow) String() string {
	return "Throw " + s.Expression.String()
}

// MARK: - Function

type FunctionBody struct {
	ASTNode
	StatementList StatementList
}

func (f *FunctionBody) Bytecode(e *Executable, c *BytecodeContext) {
	f.StatementList.Bytecode(e, c)
}

func (f *FunctionBody) String() string {
	return f.StatementList.String()
}

// MARK: - FormalParameters

type FormalParameters struct {
	Items []FormalParametersItem
}

// 15.1.5
func (f *FormalParameters) ExpectedArgumentCount() int {
	return len(f.Items)
}

func (f *FormalParameters) String() string {
	var sb string
	for i, item := range f.Items {
		if i != 0 {
			sb += ", "
		}
		sb += item.String()
	}
	return sb
}

type FormalParametersItem interface {
	ASTNode
}

type FormalParameter struct {
	FormalParametersItem
	BindingElement *BindingElement
}

func (f *FormalParameter) String() string {
	return f.BindingElement.String()
}

type BindingElement struct {
	Identifier IdentifierName
}

func (b *BindingElement) String() string {
	return string(b.Identifier)
}

// MARK: - IfStatement

type StatementIf struct {
	Statement
	Condition  Expression
	Consequent Statement
	Alternate  Statement
}

func (s *StatementIf) Analyze(a AnalyzeQuery) bool {
	return false
}

// 14.6.2
func (s *StatementIf) Bytecode(e *Executable, c *BytecodeContext) {
	s.Condition.Bytecode(e, c)

	if s.Condition.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
	e.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(e.Instructions)
	e.AddInstruction(&IStoreConstant{Value: UndefinedValue})
	s.Consequent.Bytecode(e, c)
	endJump := &IJump{Target: 0}
	e.AddInstruction(endJump)

	// else
	jumpIfTrue.TargetElse = len(e.Instructions)
	e.AddInstruction(&IStoreConstant{Value: UndefinedValue})

	if s.Alternate != nil {
		s.Alternate.Bytecode(e, c)
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
	ASTNode
}

// MARK: - WhileStatement

type StatementWhile struct {
	IterationStatement
	Condition Expression
	Body      Statement
}

func (s *StatementWhile) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&ILoadConstant{Value: UndefinedValue})

	condition := len(e.Instructions)
	s.Condition.Bytecode(e, c)
	if s.Condition.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}

	jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
	e.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(e.Instructions)
	e.AddInstruction(&IStore{})
	s.Body.Bytecode(e, c)
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

func (s *StatementDoWhile) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&ILoadConstant{Value: UndefinedValue})

	bodyIndex := len(e.Instructions)

	e.AddInstruction(&IStore{})
	s.Body.Bytecode(e, c)
	e.AddInstruction(&ILoad{})

	s.Condition.Bytecode(e, c)
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
	ASTNode
	Analyze(a AnalyzeQuery) bool
}

type DeclarationHoistable struct {
	Declaration
	FunctionDeclaration *FunctionDeclaration
}

func (d *DeclarationHoistable) Analyze(a AnalyzeQuery) bool {
	return false
}

func (d *DeclarationHoistable) Bytecode(e *Executable, c *BytecodeContext) {
	d.FunctionDeclaration.Bytecode(e, c)
}

func (d *DeclarationHoistable) String() string {
	return d.FunctionDeclaration.String()
}

type FunctionDeclaration struct {
	ASTNode
	Identifier       IdentifierName
	Body             *FunctionBody
	FormalParameters *FormalParameters
}

// 15.2.2
func (f *FunctionDeclaration) functionBodyContainsUseStrict() bool {
	return f.Body.StatementList.ContainsDirective("use strict")
}

// 15.2.4
func (f *FunctionDeclaration) instantiateOrdinaryFunctionObject(agent *Agent, env EnvironmentRecord, privateEnv *PrivateEnvironment) ObjectType {
	realm := agent.CurrentRealm()
	name := f.Identifier
	sourceText := ""
	function := OrdinaryFunctionCreate(
		agent,
		realm.Intrinsics.FunctionPrototype,
		sourceText,
		f.FormalParameters,
		f.Body,
		functionCreateThisModeNonLexical,
		env,
		privateEnv,
		f.functionBodyContainsUseStrict(),
	)

	SetFunctionName(function.Object, NewStringPropertyKey(string(name)), "")
	return function
}

func (f *FunctionDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	realm := c.agent.CurrentRealm()
	env := realm.GlobalEnv
	function := f.instantiateOrdinaryFunctionObject(c.agent, env, nil)
	realm.GlobalEnv.ObjectRecord.BindingObject.Set(NewStringPropertyKey(string(f.Identifier)), NewValueFromObject(function), setThrowTypeIgnore)
}

func (f *FunctionDeclaration) String() string {
	return "FunctionDeclaration " + string(f.Identifier)
}

// MARK: - BlockStatement

type BlockStatement interface {
	Statement
}

type BlockStatementBlock struct {
	BlockStatement
	Block *Block
}

func (b *BlockStatementBlock) Bytecode(e *Executable, c *BytecodeContext) {
	b.Block.Bytecode(e, c)
}

func (b *BlockStatementBlock) String() string {
	return b.Block.StatementList.String()
}

type Block struct {
	StatementList StatementList
}

// 14.2.2
func (b *Block) Bytecode(e *Executable, c *BytecodeContext) {
	b.StatementList.Bytecode(e, c)
}

func (b *Block) String() string {
	return b.StatementList.String()
}

type StatementList []StatementListItem

func (s StatementList) ContainsDirective(directive string) bool {
	for _, item := range s {
		if !item.Analyze(AnalyzeQueryIsStringLiteral) {
			break
		}
		statementItem := item.(*StatementListItemStatement).Statement
		statementExpression := statementItem.(*StatementExpression).Expression
		primary := statementExpression.(*ExpressionPrimary).PrimaryExpression
		literal := primary.(*PrimaryExpressionLiteral).Literal.(*LiteralString)
		if literal.Value == directive {
			return true
		}
	}
	return false
}

func (s StatementList) Bytecode(e *Executable, c *BytecodeContext) {
	for _, item := range s {
		item.Bytecode(e, c)
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
	ASTNode
	Analyze(a AnalyzeQuery) bool
}
type StatementListItemStatement struct {
	StatementListItem
	Statement Statement
}

var _ ASTNode = (*StatementListItemStatement)(nil)

func (s *StatementListItemStatement) Analyze(a AnalyzeQuery) bool {
	return s.Statement.Analyze(a)
}

func (s *StatementListItemStatement) Bytecode(e *Executable, c *BytecodeContext) {
	s.Statement.Bytecode(e, c)
}

func (s *StatementListItemStatement) String() string {
	return s.Statement.String()
}

type StatementListItemDeclaration struct {
	StatementListItem
	Declaration Declaration
}

var _ ASTNode = (*StatementListItemDeclaration)(nil)

func (s *StatementListItemDeclaration) Analyze(a AnalyzeQuery) bool {
	return s.Declaration.Analyze(a)
}

func (s *StatementListItemDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	s.Declaration.Bytecode(e, c)
}

func (s *StatementListItemDeclaration) String() string {
	return s.Declaration.String()
}

type ExpressionStatement struct {
	Expression Expression
}

func (e *ExpressionStatement) Analyze(a AnalyzeQuery) bool {
	return e.Expression.Analyze(a)
}

// 14.5.1
func (e *ExpressionStatement) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Expression.Bytecode(ex, c)
}

func (e *ExpressionStatement) String() string {
	return e.Expression.String()
}

type Script struct {
	ASTNode
	StatementList StatementList
}

func (s *Script) Bytecode(e *Executable, c *BytecodeContext) {
	s.StatementList.Bytecode(e, c)
}

func (s *Script) String() string {
	return s.StatementList.String()
}

func (s *Script) IsStrict() bool {
	return s.StatementList.ContainsDirective("use strict")
}
