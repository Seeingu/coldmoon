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

// MARK: - PrimaryExpression

type PrimaryExpression interface {
	Expression
	AssignmentTargetType() AssignmentTargetType
}

// MARK: - IdentifierReference

type IdentifierName string
type PrimaryExpressionIdentifierReference struct {
	PrimaryExpression
	Identifier IdentifierName
}

func (p *PrimaryExpressionIdentifierReference) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
}

func (p *PrimaryExpressionIdentifierReference) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IResolveBinding{Name: p.Identifier, Strict: c.containedInStrictCode})
}
func (p *PrimaryExpressionIdentifierReference) String() string {
	return string(p.Identifier)
}

func (p *PrimaryExpressionIdentifierReference) Analyze(a AnalyzeQuery) bool {
	if a == AnalyzeQueryIsReference {
		return true
	}
	return false
}

// MARK: - Literal

type PrimaryExpressionLiteral struct {
	PrimaryExpression
	Literal Literal
}

func (p *PrimaryExpressionLiteral) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
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

// MARK: - GeneratorExpression

type PrimaryExpressionGeneratorExpression struct {
	PrimaryExpression
	IdentifierName   IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (p *PrimaryExpressionGeneratorExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (p *PrimaryExpressionGeneratorExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (p *PrimaryExpressionGeneratorExpression) Bytecode(e *Executable, c *BytecodeContext) {
	strict := c.containedInStrictCode || p.Body.FunctionBodyContainsUseStrict()
	p.Body.Strict = strict
	e.AddInstruction(&IInstantiateGeneratorFunctionExpression{FunctionExpression: p})
}

func (p *PrimaryExpressionGeneratorExpression) String() string {
	return "GeneratorExpression"
}

// MARK: - AsyncGeneratorExpression

type PrimaryExpressionAsyncGeneratorExpression struct {
	PrimaryExpression
	IdentifierName   IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (p *PrimaryExpressionAsyncGeneratorExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}
func (p *PrimaryExpressionAsyncGeneratorExpression) Analyze(a AnalyzeQuery) bool {
	return false
}
func (p *PrimaryExpressionAsyncGeneratorExpression) Bytecode(e *Executable, c *BytecodeContext) {
	strict := c.containedInStrictCode || p.Body.FunctionBodyContainsUseStrict()
	p.Body.Strict = strict
	e.AddInstruction(&IInstantiateAsyncGeneratorFunctionExpression{FunctionExpression: p})
}
func (p *PrimaryExpressionAsyncGeneratorExpression) String() string {
	return "AsyncGeneratorExpression"
}

// MARK: - ThisExpression

type PrimaryExpressionThis struct {
	PrimaryExpression
}

func (p *PrimaryExpressionThis) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
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

// MARK: - ParenthesizedExpression

type PrimaryExpressionParenthesizedExpression struct {
	PrimaryExpression
	Expression Expression
}

func (p *PrimaryExpressionParenthesizedExpression) AssignmentTargetType() AssignmentTargetType {
	return p.Expression.AssignmentTargetType()
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

// MARK: - ArrayLiteral

type ArrayElement interface {
}
type ArrayElementElision struct {
	ArrayElement
}
type ArrayElementExpression struct {
	ArrayElement
	Expression Expression
}
type PrimaryExpressionArrayLiteral struct {
	PrimaryExpression
	ElementList []ArrayElement
}

func (p *PrimaryExpressionArrayLiteral) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IArrayCreate{})
	e.AddInstruction(InsLoad)
	for i, element := range p.ElementList {
		switch element := element.(type) {
		case *ArrayElementExpression:
			element.Expression.Bytecode(e, c)
			if element.Expression.Analyze(AnalyzeQueryIsReference) {
				e.AddInstruction(InsGetValue)
			}
			e.AddInstruction(InsLoad)
			e.AddInstruction(&IArraySetValue{Index: i})
			e.AddInstruction(InsLoad)
		case *ArrayElementElision:
			e.AddInstruction(InsStore)
			e.AddInstruction(&IArraySetLength{
				Length: i + 1,
			})
			e.AddInstruction(InsLoad)
		}
	}
	e.AddInstruction(InsStore)
}

func (p *PrimaryExpressionArrayLiteral) String() string {
	sb := "["
	for i, element := range p.ElementList {
		if i != 0 {
			sb += ", "
		}
		switch element := element.(type) {
		case *ArrayElementExpression:
			sb += element.Expression.String()
		case *ArrayElementElision:
			sb += ","
		}
	}
	sb += "]"
	return sb
}

func (p *PrimaryExpressionArrayLiteral) Analyze(a AnalyzeQuery) bool {
	return false
}

// MARK: - ObjectLiteral

type PrimaryExpressionObjectLiteral struct {
	PrimaryExpression
	PropertyList *PropertyDefinitionList
}

func (p *PrimaryExpressionObjectLiteral) Bytecode(e *Executable, c *BytecodeContext) {
	if len(p.PropertyList.Items) == 0 {
		e.AddInstruction(&IObjectCreate{})
		return
	}

	e.AddInstruction(&IObjectCreate{})

	p.PropertyList.Bytecode(e, c)
}

func (p *PrimaryExpressionObjectLiteral) String() string {
	sb := "{"
	sb += p.PropertyList.String()
	sb += "}"
	return sb
}

type PropertyDefinitionList struct {
	ASTNode
	Items []PropertyDefinition
}

func (p *PropertyDefinitionList) Bytecode(e *Executable, c *BytecodeContext) {
	for _, item := range p.Items {
		e.AddInstruction(InsLoad)
		item.Bytecode(e, c)
	}
}

func (p *PropertyDefinitionList) String() string {
	var sb string
	for i, item := range p.Items {
		if i != 0 {
			sb += ", "
		}
		sb += item.String()
	}
	return sb
}

// MARK: - PropertyDefinition

type PropertyDefinition interface {
	ASTNode
}

type PropertyDefinitionIdentifierReference struct {
	PropertyDefinition
	IdentifierReference *PrimaryExpressionIdentifierReference
}

func (p *PropertyDefinitionIdentifierReference) Bytecode(e *Executable, c *BytecodeContext) {
	propName := p.IdentifierReference.Identifier
	e.AddInstruction(&ILoadConstant{
		Value: NewStringValue(string(propName)),
	})

	e.AddInstruction(InsGetValue)
	e.AddInstruction(InsLoad)

	e.AddInstruction(&IObjectSetProperty{})
	e.AddInstruction(InsLoad)
}
func (p *PropertyDefinitionIdentifierReference) String() string {
	return p.IdentifierReference.String()
}

type PropertyDefinitionNameAndExpression struct {
	PropertyDefinition
	Name       PropertyName
	Expression Expression
}

func (p *PropertyDefinitionNameAndExpression) Bytecode(e *Executable, c *BytecodeContext) {
	p.Name.Bytecode(e, c)

	p.Expression.Bytecode(e, c)

	if p.Expression.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
	e.AddInstruction(&IObjectSetProperty{})
	e.AddInstruction(InsLoad)
}

func (p *PropertyDefinitionNameAndExpression) String() string {
	return p.Name.String() + ": " + p.Expression.String()
}

// MARK: - Method Definition

type MethodDefinitionType int

const (
	MethodDefinitionTypeNil MethodDefinitionType = iota
	MethodDefinitionTypeMethod
	MethodDefinitionTypeGet
	MethodDefinitionTypeSet
)

type PropertyDefinitionMethodDefinition struct {
	PropertyDefinition
	Type               MethodDefinitionType
	Name               PropertyName
	FunctionExpression *PrimaryExpressionFunctionExpression
}

func (p *PropertyDefinitionMethodDefinition) Bytecode(e *Executable, c *BytecodeContext) {
	strict := c.containedInStrictCode || p.FunctionExpression.Body.FunctionBodyContainsUseStrict()
	p.FunctionExpression.Body.Strict = strict
	p.Name.Bytecode(e, c)
	e.AddInstruction(InsLoad)
	e.AddInstruction(&IObjectDefineMethod{
		FunctionExpression: p.FunctionExpression,
		MethodType:         p.Type,
	})
}

// MARK: - PropertyName

type PropertyName interface {
	ASTNode
}
type PropertyNameLiteral interface {
	PropertyName
}
type PropertyNameLiteralIdentifier struct {
	PropertyNameLiteral
	Identifier IdentifierName
}

func (p *PropertyNameLiteralIdentifier) String() string {
	return string(p.Identifier)
}
func (p *PropertyNameLiteralIdentifier) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&ILoadConstant{Value: NewStringValue(string(p.Identifier))})
}

type PropertyNameLiteralString struct {
	PropertyNameLiteral
	StringLiteral *LiteralString
}

func (p *PropertyNameLiteralString) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&ILoadConstant{Value: p.StringLiteral.StringValue()})
}

func (p *PropertyNameLiteralString) String() string {
	return p.StringLiteral.String()
}

type PropertyNameLiteralNumeric struct {
	PropertyNameLiteral
	NumericLiteral *LiteralNumeric
}

func (p *PropertyNameLiteralNumeric) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&ILoadConstant{Value: NewStringValue(p.NumericLiteral.Value)})
}

func (p *PropertyNameLiteralNumeric) String() string {
	return p.NumericLiteral.String()
}

type PropertyNameComputed struct {
	PropertyName
	Expression Expression
}

func (p *PropertyNameComputed) String() string {
	return p.Expression.String()
}
func (p *PropertyNameComputed) Bytecode(e *Executable, c *BytecodeContext) {
	p.Expression.Bytecode(e, c)
	if p.Expression.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
}

// MARK: - FunctionExpression

type PrimaryExpressionFunctionExpression struct {
	PrimaryExpression
	Identifier       IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (p *PrimaryExpressionFunctionExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (p *PrimaryExpressionFunctionExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (p *PrimaryExpressionFunctionExpression) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IInstantiateOrdinaryFunctionExpression{FunctionExpression: p})
}

func (p *PrimaryExpressionFunctionExpression) String() string {
	return "FunctionExpression " + string(p.Identifier)
}

// MARK: - ArrowFunction

type PrimaryExpressionArrowFunction struct {
	PrimaryExpression
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (p *PrimaryExpressionArrowFunction) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (p *PrimaryExpressionArrowFunction) Analyze(a AnalyzeQuery) bool {
	return false
}

func (p *PrimaryExpressionArrowFunction) Bytecode(e *Executable, c *BytecodeContext) {
	strict := c.containedInStrictCode || p.Body.FunctionBodyContainsUseStrict()
	p.Body.Strict = strict
	e.AddInstruction(&IInstantiateArrowFunctionExpression{FunctionExpression: p})
}

func (p *PrimaryExpressionArrowFunction) String() string {
	return "ArrowFunction"
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

func (m *MemberExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
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
		if prop.Expression.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
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

type AssignmentTargetType int

const (
	AssignmentTargetTypeSimple AssignmentTargetType = iota
	AssignmentTargetTypeInvalid
)

type Expression interface {
	ASTNode
	Analyze(a AnalyzeQuery) bool
	AssignmentTargetType() AssignmentTargetType
}

// MARK: - PrimaryExpression

type ExpressionPrimary struct {
	Expression
	PrimaryExpression PrimaryExpression
}

func (e *ExpressionPrimary) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
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

// MARK: - UpdateExpression

type UpdateOperator int

const (
	UpdateOperatorIncrement UpdateOperator = iota
	UpdateOperatorDecrement
)

func (u UpdateOperator) String() string {
	switch u {
	case UpdateOperatorIncrement:
		return "++"
	case UpdateOperatorDecrement:
		return "--"
	}
	return ""
}

var UpdateOperatorMap = map[TokenType]UpdateOperator{
	TPlusPlus:   UpdateOperatorIncrement,
	TMinusMinus: UpdateOperatorDecrement,
}

type UpdateExpressionType int

const (
	UpdateExpressionTypePrefix UpdateExpressionType = iota
	UpdateExpressionTypePostfix
)

func (u UpdateExpressionType) String() string {
	switch u {
	case UpdateExpressionTypePrefix:
		return "prefix"
	case UpdateExpressionTypePostfix:
		return "postfix"
	}
	return ""
}

type ExpressionUpdate struct {
	Expression
	Type     UpdateExpressionType
	Operator UpdateOperator
	Operand  Expression
}

func (e *ExpressionUpdate) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
}
func (e *ExpressionUpdate) Analyze(a AnalyzeQuery) bool {
	return false
}
func (e *ExpressionUpdate) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Operand.Bytecode(ex, c)
	ex.AddInstruction(InsPushReference)
	if e.Operand.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	ex.AddInstruction(InsToNumber)
	if e.Type == UpdateExpressionTypePrefix {
		if e.Operator == UpdateOperatorIncrement {
			ex.AddInstruction(&IIncrement{})
		} else {
			ex.AddInstruction(&IDecrement{})
		}

		ex.AddInstruction(InsPutValue)
		ex.AddInstruction(InsPopReference)
	} else {
		ex.AddInstruction(InsLoad)
		if e.Operator == UpdateOperatorIncrement {
			ex.AddInstruction(&IIncrement{})
		} else {
			ex.AddInstruction(&IDecrement{})
		}
		ex.AddInstruction(InsPutValue)
		ex.AddInstruction(InsPopReference)
		ex.AddInstruction(InsStore)
	}
}
func (e *ExpressionUpdate) String() string {
	if e.Type == UpdateExpressionTypePrefix {
		return e.Operator.String() + e.Operand.String()
	}
	return e.Operand.String() + e.Operator.String()
}

// MARK: - AssignmentExpression

type AssignmentOperator int

func (a AssignmentOperator) String() string {
	switch a {
	case AssignmentOperatorAssign:
		return "="
	case AssignmentOperatorAddition:
		return "+="
	case AssignmentOperatorSubtraction:
		return "-="
	case AssignmentOperatorMultiplication:
		return "*="
	case AssignmentOperatorDivision:
		return "/="
	case AssignmentOperatorRemainder:
		return "%="
	case AssignmentOperatorLeftShift:
		return "<<="
	case AssignmentOperatorRightShift:
		return ">>="
	case AssignmentOperatorUnsignedRightShift:
		return ">>>="
	case AssignmentOperatorBitwiseAnd:
		return "&="
	case AssignmentOperatorBitwiseXor:
		return "^="
	case AssignmentOperatorBitwiseOr:
		return "|="
	case AssignmentOperatorExponentiation:
		return "**="
	case AssignmentOperatorAnd:
		return "&&="
	case AssignmentOperatorOr:
		return "||="
	case AssignmentOperatorNullishCoalescing:
		return "??="
	}
	return ""
}

const (
	AssignmentOperatorAssign AssignmentOperator = iota
	AssignmentOperatorAddition
	AssignmentOperatorSubtraction
	AssignmentOperatorMultiplication
	AssignmentOperatorDivision
	AssignmentOperatorRemainder
	AssignmentOperatorLeftShift
	AssignmentOperatorRightShift
	AssignmentOperatorUnsignedRightShift
	AssignmentOperatorBitwiseAnd
	AssignmentOperatorBitwiseXor
	AssignmentOperatorBitwiseOr
	AssignmentOperatorExponentiation
	AssignmentOperatorAnd
	AssignmentOperatorOr
	AssignmentOperatorNullishCoalescing
)

var operatorAssignmentMap = map[TokenType]AssignmentOperator{
	TEquals:                   AssignmentOperatorAssign,
	TPlusEquals:               AssignmentOperatorAddition,
	TMinusEquals:              AssignmentOperatorSubtraction,
	TStarEquals:               AssignmentOperatorMultiplication,
	TDivideEquals:             AssignmentOperatorDivision,
	TPercentEquals:            AssignmentOperatorRemainder,
	TLeftShiftEquals:          AssignmentOperatorLeftShift,
	TRightShiftEquals:         AssignmentOperatorRightShift,
	TUnsignedRightShiftEquals: AssignmentOperatorUnsignedRightShift,
	TBitwiseAndEquals:         AssignmentOperatorBitwiseAnd,
	TBitwiseOrEquals:          AssignmentOperatorBitwiseOr,
	TBitwiseXorEquals:         AssignmentOperatorBitwiseXor,
	TStarStarEquals:           AssignmentOperatorExponentiation,
	TAmpersandAmpersandEquals: AssignmentOperatorAnd,
	TPipePipeEquals:           AssignmentOperatorOr,
	TQuestionQuestionEquals:   AssignmentOperatorNullishCoalescing,
}

type ExpressionAssignmentExpression struct {
	Expression
	Left     Expression
	Operator AssignmentOperator
	Right    Expression
}

func (e *ExpressionAssignmentExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (e *ExpressionAssignmentExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (e *ExpressionAssignmentExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	if e.Operator == AssignmentOperatorAssign {
		e.Left.Bytecode(ex, c)
		ex.AddInstruction(&IPushReference{})

		e.Right.Bytecode(ex, c)
		if e.Right.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		ex.AddInstruction(&IPutValue{})
		ex.AddInstruction(&IPopReference{})
	} else if e.Operator != AssignmentOperatorAnd &&
		e.Operator != AssignmentOperatorOr &&
		e.Operator != AssignmentOperatorNullishCoalescing {
		e.Left.Bytecode(ex, c)
		ex.AddInstruction(&IPushReference{})

		if e.Left.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}
		ex.AddInstruction(InsLoad)

		e.Right.Bytecode(ex, c)

		if e.Right.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}
		ex.AddInstruction(InsLoad)

		var operatorMap = map[AssignmentOperator]BinaryOperator{
			AssignmentOperatorAddition:           BinaryOperatorAddition,
			AssignmentOperatorSubtraction:        BinaryOperatorSubtraction,
			AssignmentOperatorMultiplication:     BinaryOperatorMultiplication,
			AssignmentOperatorDivision:           BinaryOperatorDivision,
			AssignmentOperatorRemainder:          BinaryOperatorRemainder,
			AssignmentOperatorLeftShift:          BinaryOperatorLeftShift,
			AssignmentOperatorRightShift:         BinaryOperatorRightShift,
			AssignmentOperatorUnsignedRightShift: BinaryOperatorUnsignedRightShift,
			AssignmentOperatorBitwiseAnd:         BinaryOperatorBitwiseAnd,
			AssignmentOperatorBitwiseXor:         BinaryOperatorBitwiseXor,
			AssignmentOperatorBitwiseOr:          BinaryOperatorBitwiseOr,
			AssignmentOperatorExponentiation:     BinaryOperatorExponentiation,
		}
		op, _ := operatorMap[e.Operator]
		ex.AddInstruction(&IApplyStringOrNumericBinaryOperator{
			Operator: op,
		})

		ex.AddInstruction(InsPutValue)
		ex.AddInstruction(InsPopReference)
	} else if e.Operator == AssignmentOperatorAnd {
		e.Left.Bytecode(ex, c)
		ex.AddInstruction(InsPushReference)

		if e.Left.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}
		ex.AddInstruction(InsLoad)

		jumpIfTrue := &IJumpIfTrue{}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.Target = len(ex.Instructions) - 1

		e.Right.Bytecode(ex, c)
		if e.Right.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		ex.AddInstruction(InsPutValue)

		endJump := &IJump{}
		ex.AddInstruction(endJump)

		jumpIfTrue.TargetElse = len(ex.Instructions) - 1
		ex.AddInstruction(InsStore)

		endJump.Target = len(ex.Instructions) - 1
		ex.AddInstruction(InsPopReference)
	} else if e.Operator == AssignmentOperatorOr {
		e.Left.Bytecode(ex, c)
		ex.AddInstruction(InsPushReference)

		if e.Left.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		jumpIfTrue := &IJumpIfTrue{}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.TargetElse = len(ex.Instructions) - 1

		e.Right.Bytecode(ex, c)
		if e.Right.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		ex.AddInstruction(InsPutValue)

		jumpIfTrue.Target = len(ex.Instructions) - 1

		ex.AddInstruction(InsPopReference)
	} else if e.Operator == AssignmentOperatorNullishCoalescing {
		e.Left.Bytecode(ex, c)
		ex.AddInstruction(InsPushReference)

		if e.Left.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}
		ex.AddInstruction(InsLoad)

		ex.AddInstruction(InsLoad)
		ex.AddInstruction(&ILoadConstant{
			Value: UndefinedValue,
		})
		ex.AddInstruction(InsLooselyEqual)

		jumpIfTrue := &IJumpIfTrue{}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.Target = len(ex.Instructions) - 1

		e.Right.Bytecode(ex, c)
		if e.Right.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		ex.AddInstruction(InsPutValue)
		endJump := &IJump{}
		ex.AddInstruction(endJump)

		jumpIfTrue.TargetElse = len(ex.Instructions) - 1
		ex.AddInstruction(InsStore)

		endJump.Target = len(ex.Instructions) - 1
		ex.AddInstruction(InsPopReference)
	}
}

func (e *ExpressionAssignmentExpression) String() string {
	return e.Left.String() + " " + e.Operator.String() + " " + e.Right.String()
}

// MARK: - NewExpression

type ExpressionNewExpression struct {
	Expression
	Callee    Expression
	Arguments Arguments
}

func (e *ExpressionNewExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (e *ExpressionNewExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (e *ExpressionNewExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Callee.Bytecode(ex, c)
	if e.Callee.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	ex.AddInstruction(InsLoad)

	for _, arg := range e.Arguments {
		arg.Bytecode(ex, c)
		if arg.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}
		ex.AddInstruction(InsLoad)
	}
	ex.AddInstruction(&INew{ArgumentCount: len(e.Arguments)})
}

func (e *ExpressionNewExpression) String() string {
	sb := "new " + e.Callee.String() + "("
	for i, arg := range e.Arguments {
		if i != 0 {
			sb += ", "
		}
		sb += arg.String()
	}
	sb += ")"
	return sb
}

// MARK: - BinaryExpression

type BinaryOperator int

func (b BinaryOperator) String() string {
	switch b {
	case BinaryOperatorMultiplication:
		return "*"
	case BinaryOperatorExponentiation:
		return "**"
	case BinaryOperatorDivision:
		return "/"
	case BinaryOperatorRemainder:
		return "%"
	case BinaryOperatorAddition:
		return "+"
	case BinaryOperatorSubtraction:
		return "-"
	case BinaryOperatorLeftShift:
		return "<<"
	case BinaryOperatorRightShift:
		return ">>"
	case BinaryOperatorUnsignedRightShift:
		return ">>>"
	case BinaryOperatorBitwiseAnd:
		return "&"
	case BinaryOperatorBitwiseXor:
		return "^"
	case BinaryOperatorBitwiseOr:
		return "|"
	}
	return ""
}

const (
	BinaryOperatorMultiplication BinaryOperator = iota
	BinaryOperatorExponentiation
	BinaryOperatorDivision
	BinaryOperatorRemainder
	BinaryOperatorAddition
	BinaryOperatorSubtraction
	BinaryOperatorLeftShift
	BinaryOperatorRightShift
	BinaryOperatorUnsignedRightShift
	BinaryOperatorBitwiseAnd
	BinaryOperatorBitwiseXor
	BinaryOperatorBitwiseOr
)

var operatorBinaryMap = map[TokenType]BinaryOperator{
	TStar:               BinaryOperatorMultiplication,
	TStarStar:           BinaryOperatorExponentiation,
	TSlash:              BinaryOperatorDivision,
	TPercent:            BinaryOperatorRemainder,
	TPlus:               BinaryOperatorAddition,
	TMinus:              BinaryOperatorSubtraction,
	TLeftShift:          BinaryOperatorLeftShift,
	TRightShift:         BinaryOperatorRightShift,
	TUnsignedRightShift: BinaryOperatorUnsignedRightShift,
	TAmpersand:          BinaryOperatorBitwiseAnd,
	TCaret:              BinaryOperatorBitwiseXor,
	TPipe:               BinaryOperatorBitwiseOr,
}

type ExpressionBinaryExpression struct {
	Expression
	Left     Expression
	Operator BinaryOperator
	Right    Expression
}

func (b *ExpressionBinaryExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (b *ExpressionBinaryExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (b *ExpressionBinaryExpression) Bytecode(e *Executable, c *BytecodeContext) {
	b.Left.Bytecode(e, c)
	if b.Left.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)

	b.Right.Bytecode(e, c)
	if b.Right.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)

	e.AddInstruction(&IApplyStringOrNumericBinaryOperator{
		Operator: b.Operator,
	})
}

func (b *ExpressionBinaryExpression) String() string {
	return b.Left.String() + " " + b.Operator.String() + " " + b.Right.String()
}

// MARK: - SequenceExpression

type ExpressionSequenceExpression struct {
	Expression
	Expressions []Expression
}

func (e *ExpressionSequenceExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (e *ExpressionSequenceExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	for _, expr := range e.Expressions {
		expr.Bytecode(ex, c)
		if expr.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}
	}
}

func (e *ExpressionSequenceExpression) String() string {
	var sb string
	for i, expr := range e.Expressions {
		if i != 0 {
			sb += ", "
		}
		sb += expr.String()
	}
	return sb
}

// MARK: - ConditionalExpression

type ExpressionConditionalExpression struct {
	Expression
	Test       Expression
	Consequent Expression
	Alternate  Expression
}

func (e *ExpressionConditionalExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (e *ExpressionConditionalExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Test.Bytecode(ex, c)
	if e.Test.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
	ex.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(ex.Instructions) - 1
	e.Consequent.Bytecode(ex, c)

	if e.Consequent.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}

	jump := &IJump{Target: 0}
	ex.AddInstruction(jump)

	jumpIfTrue.TargetElse = len(ex.Instructions)
	e.Alternate.Bytecode(ex, c)

	if e.Alternate.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}

	jump.Target = len(ex.Instructions)
}

func (e *ExpressionConditionalExpression) String() string {
	return e.Test.String() + " ? " + e.Consequent.String() + " : " + e.Alternate.String()
}

// MARK: - LogicalExpression

type LogicalOperator int

const (
	LogicalOperatorAnd LogicalOperator = iota
	LogicalOperatorOr
	LogicalOperatorNullishCoalescing
)

func (l LogicalOperator) String() string {
	switch l {
	case LogicalOperatorAnd:
		return "&&"
	case LogicalOperatorOr:
		return "||"
	case LogicalOperatorNullishCoalescing:
		return "??"
	}
	return ""
}

var operatorLogicalMap = map[TokenType]LogicalOperator{
	TAmpersandAmpersand: LogicalOperatorAnd,
	TPipePipe:           LogicalOperatorOr,
	TQuestionQuestion:   LogicalOperatorNullishCoalescing,
}

type ExpressionLogicalExpression struct {
	Expression
	Left     Expression
	Operator LogicalOperator
	Right    Expression
}

func (e *ExpressionLogicalExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (e *ExpressionLogicalExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Left.Bytecode(ex, c)
	if e.Left.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}

	switch e.Operator {
	case LogicalOperatorAnd:
		jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.Target = len(ex.Instructions)
		e.Right.Bytecode(ex, c)

		if e.Right.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		jumpIfTrue.TargetElse = len(ex.Instructions)
	case LogicalOperatorOr:
		jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.TargetElse = len(ex.Instructions)
		e.Right.Bytecode(ex, c)
		if e.Right.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		jumpIfTrue.Target = len(ex.Instructions)
	case LogicalOperatorNullishCoalescing:
		ex.AddInstruction(InsLoad)

		ex.AddInstruction(InsLoad)
		ex.AddInstruction(&ILoadConstant{
			Value: UndefinedValue,
		})
		ex.AddInstruction(InsLooselyEqual)

		jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.Target = len(ex.Instructions)
		ex.AddInstruction(InsStore)

		e.Right.Bytecode(ex, c)
		if e.Right.Analyze(AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		jump := &IJump{Target: 0}
		ex.AddInstruction(jump)

		jumpIfTrue.TargetElse = len(ex.Instructions)
		ex.AddInstruction(InsStore)

		jump.Target = len(ex.Instructions)
	}
}

func (e *ExpressionLogicalExpression) String() string {
	return e.Left.String() + " " + e.Operator.String() + " " + e.Right.String()
}

// MARK: - EqualityExpression

type EqualityOperator int

const (
	EqualityOperatorEqual EqualityOperator = iota
	EqualityOperatorNotEqual
	EqualityOperatorStrictEqual
	EqualityOperatorStrictNotEqual
)

var operatorEqualityMap = map[TokenType]EqualityOperator{
	TEquals:          EqualityOperatorEqual,
	TNotEquals:       EqualityOperatorNotEqual,
	TStrictEquals:    EqualityOperatorStrictEqual,
	TStrictNotEquals: EqualityOperatorStrictNotEqual,
}

func (e EqualityOperator) String() string {
	switch e {
	case EqualityOperatorEqual:
		return "=="
	case EqualityOperatorNotEqual:
		return "!="
	case EqualityOperatorStrictEqual:
		return "==="
	case EqualityOperatorStrictNotEqual:
		return "!=="
	}
	return ""
}

type ExpressionEqualityExpression struct {
	Expression
	Left     Expression
	Operator EqualityOperator
	Right    Expression
}

func (e *ExpressionEqualityExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (e *ExpressionEqualityExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Left.Bytecode(ex, c)
	if e.Left.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	ex.AddInstruction(InsLoad)

	e.Right.Bytecode(ex, c)
	if e.Right.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	ex.AddInstruction(InsLoad)

	switch e.Operator {
	case EqualityOperatorEqual:
		ex.AddInstruction(InsLooselyEqual)
	case EqualityOperatorNotEqual:
		ex.AddInstruction(InsLooselyEqual)
		ex.AddInstruction(InsLogicalNot)
	case EqualityOperatorStrictEqual:
		ex.AddInstruction(InsStrictlyEqual)
	case EqualityOperatorStrictNotEqual:
		ex.AddInstruction(InsStrictlyEqual)
		ex.AddInstruction(InsLogicalNot)
	}
}

func (e *ExpressionEqualityExpression) String() string {
	return e.Left.String() + " " + e.Operator.String() + " " + e.Right.String()
}

// MARK: - RelationalExpression

type RelationalOperator int

func (r RelationalOperator) String() string {
	switch r {
	case RelationalOperatorLessThan:
		return "<"
	case RelationalOperatorGreaterThan:
		return ">"
	case RelationalOperatorLessThanOrEqual:
		return "<="
	case RelationalOperatorGreaterThanOrEqual:
		return ">="
	case RelationalOperatorInstanceof:
		return "instanceof"
	case RelationalOperatorIn:
		return "in"
	}
	return ""
}

const (
	RelationalOperatorLessThan RelationalOperator = iota
	RelationalOperatorGreaterThan
	RelationalOperatorLessThanOrEqual
	RelationalOperatorGreaterThanOrEqual
	RelationalOperatorInstanceof
	RelationalOperatorIn
)

var operatorRelationMap = map[TokenType]RelationalOperator{
	TLessThan:          RelationalOperatorLessThan,
	TGreaterThan:       RelationalOperatorGreaterThan,
	TLessThanEquals:    RelationalOperatorLessThanOrEqual,
	TGreaterThanEquals: RelationalOperatorGreaterThanOrEqual,
	TInstanceof:        RelationalOperatorInstanceof,
	TIn:                RelationalOperatorIn,
}

type ExpressionRelationalExpression struct {
	Expression
	Left     Expression
	Operator RelationalOperator
	Right    Expression
}

func (e *ExpressionRelationalExpression) Analyze(a AnalyzeQuery) bool {
	return false
}

func (e *ExpressionRelationalExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Left.Bytecode(ex, c)
	if e.Left.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	ex.AddInstruction(InsLoad)

	e.Right.Bytecode(ex, c)
	if e.Right.Analyze(AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	ex.AddInstruction(InsLoad)

	switch e.Operator {
	case RelationalOperatorLessThan:
		ex.AddInstruction(InsLessThan)
	case RelationalOperatorGreaterThan:
		ex.AddInstruction(InsGreaterThan)
	case RelationalOperatorLessThanOrEqual:
		ex.AddInstruction(InsLessThanEquals)
	case RelationalOperatorGreaterThanOrEqual:
		ex.AddInstruction(InsGreaterThanEquals)
	case RelationalOperatorInstanceof:
		ex.AddInstruction(InsInstanceOf)
	case RelationalOperatorIn:
		ex.AddInstruction(InsHasProperty)
	}
}

func (e *ExpressionRelationalExpression) String() string {
	return e.Left.String() + " " + e.Operator.String() + " " + e.Right.String()
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
		u.Expression.Bytecode(e, c)
		if !u.Expression.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(&IStoreConstant{
				Value: NewBooleanValue(true),
			})
		} else {
			e.AddInstruction(&IDelete{})
		}
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

	e.AddInstruction(&IPushReference{})
	isReference := c.Callee.Analyze(AnalyzeQueryIsReference)
	if isReference {
		e.AddInstruction(InsGetValue)
	}

	e.AddInstruction(InsLoad)
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

	e.AddInstruction(&IPopReference{})
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
	VarScopedDeclarations() []*VariableDeclaration
}

// MARK: - VariableStatement

type StatementVariable struct {
	Statement
	DeclarationList *VariableDeclarationList
}

var _ Statement = (*StatementVariable)(nil)

func (s *StatementVariable) VarScopedDeclarations() (l []*VariableDeclaration) {
	return s.DeclarationList.VarScopedDeclarations()
}

func (s *StatementVariable) Analyze(a AnalyzeQuery) bool {
	return false
}

func (s *StatementVariable) Bytecode(e *Executable, c *BytecodeContext) {
	s.DeclarationList.Bytecode(e, c)
}

func (s *StatementVariable) String() string {
	return "var " + s.DeclarationList.String()
}

type VariableDeclarationList struct {
	ASTNode
	Items []*VariableDeclaration
}

func (v *VariableDeclarationList) VarScopedDeclarations() (l []*VariableDeclaration) {
	return v.Items
}

func (v *VariableDeclarationList) Bytecode(e *Executable, c *BytecodeContext) {
	for _, item := range v.Items {
		item.Bytecode(e, c)
	}
}

func (v *VariableDeclarationList) String() string {
	var sb string
	for i, item := range v.Items {
		if i != 0 {
			sb += ", "
		}
		sb += item.String()
	}
	return sb
}

type VariableDeclaration struct {
	ASTNode
	Identifier  IdentifierName
	Initializer Expression
}

func (v *VariableDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	if v.Initializer == nil {
		return
	}

	e.AddInstruction(InsLoad)
	e.AddInstruction(&IResolveBinding{
		Name: v.Identifier,
	})
	_ = c.containedInStrictCode
	e.AddInstruction(InsPushReference)

	v.Initializer.Bytecode(e, c)

	if v.Initializer.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsPutValue)
	e.AddInstruction(InsPopReference)
}

func (v *VariableDeclaration) String() string {
	if v.Initializer != nil {
		return string(v.Identifier) + " = " + v.Initializer.String()
	}
	return string(v.Identifier)
}

// MARK: - BlockStatement

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

// MARK: - EmptyStatement

type StatementEmpty struct {
	Statement
}

var _ Statement = (*StatementEmpty)(nil)

func (s *StatementEmpty) Analyze(a AnalyzeQuery) bool {
	return false
}

func (s *StatementEmpty) VarScopedDeclarations() (l []*VariableDeclaration) {
	return l
}

func (s *StatementEmpty) Bytecode(e *Executable, c *BytecodeContext) {
	// empty
}
func (s *StatementEmpty) String() string {
	return ""
}

// MARK: - TryStatement

type CatchParameter IdentifierName

func (c CatchParameter) ToIdentifier() IdentifierName {
	return IdentifierName(c)
}

type StatementTry struct {
	Statement
	CatchParameter CatchParameter
	TryBlock       *Block
	CatchBlock     *Block
	FinallyBlock   *Block
}

var _ Statement = (*StatementTry)(nil)

func (t *StatementTry) Analyze(a AnalyzeQuery) bool {
	return false
}

func (t *StatementTry) VarScopedDeclarations() (l []*VariableDeclaration) {
	l = append(l, t.TryBlock.StatementList.VarScopedDeclarations()...)
	if t.CatchBlock != nil {
		l = append(l, t.CatchBlock.StatementList.VarScopedDeclarations()...)
	}
	if t.FinallyBlock != nil {
		l = append(l, t.FinallyBlock.StatementList.VarScopedDeclarations()...)
	}
	return
}

func (t *StatementTry) Bytecode(e *Executable, c *BytecodeContext) {
	if t.FinallyBlock == nil {
		exceptionJumpToCatch := &IPushExceptionJumpTarget{}
		e.AddInstruction(exceptionJumpToCatch)

		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
		t.TryBlock.Bytecode(e, c)
		e.AddInstruction(&IPopExceptionJumpTarget{})
		exceptionJumpToEnd := &IJump{}
		e.AddInstruction(exceptionJumpToEnd)

		exceptionJumpToCatch.Target = len(e.Instructions) - 1
		e.AddInstruction(&IPopExceptionJumpTarget{})
		if t.CatchParameter != "" {
			e.AddInstruction(&ICreateCatchBinding{
				IdentifierName: t.CatchParameter.ToIdentifier(),
			})
		}
		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
		t.CatchBlock.Bytecode(e, c)

		exceptionJumpToEnd.Target = len(e.Instructions) - 1
	} else if t.CatchBlock == nil {
		exceptionJumpToFinally := &IPushExceptionJumpTarget{}
		e.AddInstruction(exceptionJumpToFinally)

		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
		t.TryBlock.Bytecode(e, c)

		exceptionJumpToFinally.Target = len(e.Instructions) - 1
		e.AddInstruction(&IPopExceptionJumpTarget{})
		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
		t.FinallyBlock.Bytecode(e, c)
		e.AddInstruction(&IRethrowExceptionIfAny{})

	} else {
		exceptionJumpToCatch := &IPushExceptionJumpTarget{}
		e.AddInstruction(exceptionJumpToCatch)

		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
		t.TryBlock.Bytecode(e, c)
		jumpToFinally := &IJump{}
		e.AddInstruction(jumpToFinally)

		exceptionJumpToCatch.Target = len(e.Instructions) - 1
		e.AddInstruction(&IPopExceptionJumpTarget{})
		exceptionJumpToFinally := &IPushExceptionJumpTarget{}
		e.AddInstruction(exceptionJumpToFinally)
		if t.CatchParameter != "" {
			e.AddInstruction(&ICreateCatchBinding{
				IdentifierName: t.CatchParameter.ToIdentifier(),
			})
		}
		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
		t.CatchBlock.Bytecode(e, c)

		jumpToFinally.Target = len(e.Instructions) - 1
		exceptionJumpToFinally.Target = len(e.Instructions) - 1
		e.AddInstruction(&IPopExceptionJumpTarget{})
		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
		t.FinallyBlock.Bytecode(e, c)
		e.AddInstruction(&IRethrowExceptionIfAny{})
	}
}

func (t *StatementTry) String() string {
	return "try"
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

func (b *BreakableStatement) Analyze(a AnalyzeQuery) bool {
	return false
}

func (b *BreakableStatement) VarScopedDeclarations() (l []*VariableDeclaration) {
	return b.IterationStatement.VarScopedDeclarations()
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
	return "CompletionTypeThrow " + s.Expression.String()
}

// MARK: - Function

type FunctionType int

const (
	FunctionTypeNormal FunctionType = iota
	FunctionTypeGenerator
	FunctionTypeAsync
	FunctionTypeAsyncGenerator
)

type FunctionBody struct {
	ASTNode
	StatementList StatementList
	Strict        bool
	Type          FunctionType
}

func (f *FunctionBody) VarScopedDeclarations() (l []*VariableDeclaration) {
	return f.StatementList.VarScopedDeclarations()
}

func (f *FunctionBody) Bytecode(e *Executable, c *BytecodeContext) {
	strictBefore := c.containedInStrictCode
	c.containedInStrictCode = c.containedInStrictCode || f.FunctionBodyContainsUseStrict()
	defer func() {
		c.containedInStrictCode = strictBefore
	}()
	f.StatementList.Bytecode(e, c)
}

func (f *FunctionBody) String() string {
	return f.StatementList.String()
}

func (f *FunctionBody) FunctionBodyContainsUseStrict() bool {
	return f.StatementList.ContainsDirective("use strict")
}

// MARK: - FormalParameters

type FormalParameters struct {
	Items []FormalParametersItem
}

func (f *FormalParameters) IsSimpleParameterList() bool {
	for _, item := range f.Items {
		if _, ok := item.(*FormalParameter); !ok {
			return false
		}
	}
	return true
}
func (f *FormalParameters) ContainsExpression() bool {
	for _, item := range f.Items {
		if _, ok := item.(Expression); ok {
			return true
		}
	}
	return false
}

func (f *FormalParameters) BoundNames() (l []IdentifierName) {
	for _, item := range f.Items {
		l = append(l, item.(*FormalParameter).BindingElement.Identifier)
	}
	return
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

var _ Statement = (*StatementIf)(nil)

func (s *StatementIf) VarScopedDeclarations() (l []*VariableDeclaration) {
	l = append(l, s.Consequent.VarScopedDeclarations()...)
	if s.Alternate != nil {
		l = append(l, s.Alternate.VarScopedDeclarations()...)
		return
	}
	return
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
	VarScopedDeclarations() []*VariableDeclaration
	ASTNode
}

// MARK: - WhileStatement

type StatementWhile struct {
	IterationStatement
	Condition Expression
	Body      Statement
}

var _ IterationStatement = (*StatementWhile)(nil)

func (s *StatementWhile) Analyze(a AnalyzeQuery) bool {
	return false
}

func (s *StatementWhile) VarScopedDeclarations() []*VariableDeclaration {
	return s.Body.VarScopedDeclarations()
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
	e.AddInstruction(InsStore)
	s.Body.Bytecode(e, c)
	e.AddInstruction(&ILoad{})

	e.AddInstruction(&IJump{Target: condition})

	jumpIfTrue.TargetElse = len(e.Instructions)
	e.AddInstruction(InsStore)
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

func (s *StatementDoWhile) Analyze(a AnalyzeQuery) bool {
	return false
}

func (s *StatementDoWhile) VarScopedDeclarations() []*VariableDeclaration {
	return s.Body.VarScopedDeclarations()
}

func (s *StatementDoWhile) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&ILoadConstant{Value: UndefinedValue})

	bodyIndex := len(e.Instructions)

	e.AddInstruction(InsStore)
	s.Body.Bytecode(e, c)
	e.AddInstruction(&ILoad{})

	s.Condition.Bytecode(e, c)
	if s.Condition.Analyze(AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}

	e.AddInstruction(&ILoad{})
	jumpIfTrue := &IJumpIfTrue{Target: bodyIndex, TargetElse: 0}
	jumpIfTrue.TargetElse = len(e.Instructions)

	e.AddInstruction(InsStore)
}

func (s *StatementDoWhile) String() string {
	sb := "DoWhile"
	sb += " " + s.Condition.String() + " \n"
	sb += s.Body.String()
	return sb
}

// MARK: - ForStatement

type ForStatementInitializer interface {
	ASTNode
}
type ForStatementInitializerExpression struct {
	ForStatementInitializer
	Expression Expression
}

func (f *ForStatementInitializerExpression) String() string {
	return f.Expression.String()
}

type ForStatementInitializerVariable struct {
	ForStatementInitializer
	VariableStatement *StatementVariable
}

func (f *ForStatementInitializerVariable) String() string {
	return f.VariableStatement.String()
}

type ForStatementInitializerLexicalDeclaration struct {
	ForStatementInitializer
	LexicalDeclaration *DeclarationLexical
}

func (f *ForStatementInitializerLexicalDeclaration) String() string {
	return f.LexicalDeclaration.String()
}

type StatementFor struct {
	IterationStatement
	Initializer ForStatementInitializer
	Condition   Expression
	Increment   Expression
	Body        Statement
}

func (s *StatementFor) Analyze(a AnalyzeQuery) bool {
	return false
}
func (s *StatementFor) VarScopedDeclarations() (l []*VariableDeclaration) {
	if s.Initializer != nil {
		if varStatement, ok := s.Initializer.(*ForStatementInitializerVariable); ok {
			l = append(l, varStatement.VariableStatement.DeclarationList.VarScopedDeclarations()...)
		}
	}
	l = append(l, s.Body.VarScopedDeclarations()...)
	return
}
func (s *StatementFor) Bytecode(e *Executable, c *BytecodeContext) {
	if s.Initializer != nil {
		switch initializer := s.Initializer.(type) {
		case *ForStatementInitializerExpression:
			initializer.Expression.Bytecode(e, c)
			if initializer.Expression.Analyze(AnalyzeQueryIsReference) {
				e.AddInstruction(InsGetValue)
			}
		case *ForStatementInitializerVariable:
			initializer.VariableStatement.Bytecode(e, c)
		case *ForStatementInitializerLexicalDeclaration:
			initializer.LexicalDeclaration.Bytecode(e, c)
		}
	}

	e.AddInstruction(&ILoadConstant{Value: UndefinedValue})

	conditionIndex := len(e.Instructions) - 1
	var endJump *IJumpIfTrue
	if s.Condition != nil {
		s.Condition.Bytecode(e, c)
		if s.Condition.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}

		jumpIfTrue := &IJumpIfTrue{}
		e.AddInstruction(jumpIfTrue)
		jumpIfTrue.Target = len(e.Instructions) - 1
		jumpIfTrue.TargetElse = len(e.Instructions) - 1
		endJump = jumpIfTrue
	}

	e.AddInstruction(InsStore)
	s.Body.Bytecode(e, c)
	e.AddInstruction(&ILoad{})

	if s.Increment != nil {
		s.Increment.Bytecode(e, c)
		if s.Increment.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
	}

	e.AddInstruction(&IJump{Target: conditionIndex})
	if endJump != nil {
		endJump.TargetElse = len(e.Instructions) - 1
	}
	e.AddInstruction(InsStore)
}

func (s *StatementFor) String() string {
	sb := "For"
	if s.Initializer != nil {
		sb += " " + s.Initializer.String()
	}
	if s.Condition != nil {
		sb += " " + s.Condition.String()
	}
	if s.Increment != nil {
		sb += " " + s.Increment.String()
	}
	sb += " \n"
	sb += s.Body.String()
	return sb
}

// MARK: - ReturnStatement

type StatementReturn struct {
	Statement
	Expression Expression
}

func (s *StatementReturn) Analyze(a AnalyzeQuery) bool {
	return false
}

func (s *StatementReturn) Bytecode(e *Executable, c *BytecodeContext) {
	if s.Expression != nil {
		s.Expression.Bytecode(e, c)
		if s.Expression.Analyze(AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
	} else {
		e.AddInstruction(&IStoreConstant{Value: UndefinedValue})
	}
	e.AddInstruction(InsReturn)
}

func (s *StatementReturn) String() string {
	if s.Expression != nil {
		return "CompletionTypeReturn " + s.Expression.String()
	}
	return "CompletionTypeReturn"
}

// MARK: - Declaration

type Declaration interface {
	ASTNode
	Analyze(a AnalyzeQuery) bool
}

// MARK: - HoistableDeclaration

type DeclarationHoistable interface {
	Declaration
}

// MARK: - FunctionDeclaration

type DeclarationHoistableFunction struct {
	DeclarationHoistable
	FunctionDeclaration *FunctionDeclaration
}

func (d *DeclarationHoistableFunction) Analyze(a AnalyzeQuery) bool {
	return false
}

func (d *DeclarationHoistableFunction) Bytecode(e *Executable, c *BytecodeContext) {
	d.FunctionDeclaration.Bytecode(e, c)
}

func (d *DeclarationHoistableFunction) String() string {
	return d.FunctionDeclaration.String()
}

// MARK: - AsyncFunctionDeclaration

type DeclarationHoistableAsyncFunction struct {
	DeclarationHoistable
	AsyncFunctionDeclaration *AsyncFunctionDeclaration
}

func (d *DeclarationHoistableAsyncFunction) Analyze(a AnalyzeQuery) bool {
	return false
}
func (d *DeclarationHoistableAsyncFunction) Bytecode(e *Executable, c *BytecodeContext) {
	d.AsyncFunctionDeclaration.Bytecode(e, c)
}
func (d *DeclarationHoistableAsyncFunction) String() string {
	return d.AsyncFunctionDeclaration.String()
}

type AsyncFunctionDeclaration struct {
	ASTNode
	Identifier       IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (d *AsyncFunctionDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	realm := c.agent.CurrentRealm()
	env := realm.GlobalEnv
	function := d.instantiateAsyncFunctionObject(c.agent, env, nil)
	realm.GlobalEnv.ObjectRecord.BindingObject.Set(NewStringPropertyKey(string(d.Identifier)), NewValueFromObject(function), setThrowTypeIgnore)
}
func (d *AsyncFunctionDeclaration) instantiateAsyncFunctionObject(agent *Agent, env EnvironmentRecord, privateEnv *PrivateEnvironment) ObjectType {
	realm := agent.CurrentRealm()
	name := d.Identifier
	sourceText := d.SourceText
	function := OrdinaryFunctionCreate(
		agent,
		realm.Intrinsics.AsyncFunctionPrototype,
		sourceText,
		d.FormalParameters,
		d.Body,
		functionCreateThisModeNonLexical,
		env,
		privateEnv,
	)
	SetFunctionName(function.Object, NewStringPropertyKey(string(name)), "")
	return function
}
func (d *AsyncFunctionDeclaration) String() string {
	return "AsyncFunctionDeclaration " + string(d.Identifier)
}

// MARK: - AsyncGeneratorDeclaration

type DeclarationHoistableAsyncGenerator struct {
	DeclarationHoistable
	AsyncGeneratorDeclaration *AsyncGeneratorDeclaration
}

func (d *DeclarationHoistableAsyncGenerator) Analyze(a AnalyzeQuery) bool {
	return false
}
func (d *DeclarationHoistableAsyncGenerator) Bytecode(e *Executable, c *BytecodeContext) {
	d.AsyncGeneratorDeclaration.Bytecode(e, c)
}
func (d *DeclarationHoistableAsyncGenerator) String() string {
	return d.AsyncGeneratorDeclaration.String()
}

type AsyncGeneratorDeclaration struct {
	ASTNode
	Identifier       IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (d *AsyncGeneratorDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	realm := c.agent.CurrentRealm()
	env := realm.GlobalEnv
	function := d.instantiateAsyncGeneratorFunctionObject(c.agent, env, nil)
	realm.GlobalEnv.ObjectRecord.BindingObject.Set(NewStringPropertyKey(string(d.Identifier)), NewValueFromObject(function), setThrowTypeIgnore)
}

// 15.6.3
func (d *AsyncGeneratorDeclaration) instantiateAsyncGeneratorFunctionObject(agent *Agent, env EnvironmentRecord, privateEnv *PrivateEnvironment) ObjectType {
	realm := agent.CurrentRealm()
	name := d.Identifier
	sourceText := d.SourceText
	function := OrdinaryFunctionCreate(
		agent,
		realm.Intrinsics.FunctionPrototype,
		sourceText,
		d.FormalParameters,
		d.Body,
		functionCreateThisModeNonLexical,
		env,
		privateEnv,
	)
	SetFunctionName(function.Object, NewStringPropertyKey(string(name)), "")
	prototype := OrdinaryObjectCreate(agent, realm.Intrinsics.AsyncGeneratorFunctionPrototypePrototype, nil)
	function.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
		Value:        NewValueFromObject(prototype),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	return function
}
func (d *AsyncGeneratorDeclaration) String() string {
	return "AsyncGeneratorDeclaration " + string(d.Identifier)
}

// MARK: - GeneratorDeclaration

type DeclarationHoistableGenerator struct {
	DeclarationHoistable
	GeneratorDeclaration *GeneratorDeclaration
}

func (d *DeclarationHoistableGenerator) Analyze(a AnalyzeQuery) bool {
	return false
}

func (d *DeclarationHoistableGenerator) Bytecode(e *Executable, c *BytecodeContext) {
	d.GeneratorDeclaration.Bytecode(e, c)
}
func (d *DeclarationHoistableGenerator) String() string {
	return d.GeneratorDeclaration.String()
}

type GeneratorDeclaration struct {
	ASTNode
	Identifier       IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (d *GeneratorDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	realm := c.agent.CurrentRealm()
	env := realm.GlobalEnv
	function := d.instantiateOrdinaryFunctionObject(c.agent, env, nil)
	realm.GlobalEnv.ObjectRecord.BindingObject.Set(NewStringPropertyKey(string(d.Identifier)), NewValueFromObject(function), setThrowTypeIgnore)
}

// 15.5.3
func (d *GeneratorDeclaration) instantiateOrdinaryFunctionObject(agent *Agent, env EnvironmentRecord, privateEnv *PrivateEnvironment) ObjectType {
	realm := agent.CurrentRealm()
	name := d.Identifier
	sourceText := d.SourceText
	function := OrdinaryFunctionCreate(
		agent,
		realm.Intrinsics.FunctionPrototype,
		sourceText,
		d.FormalParameters,
		d.Body,
		functionCreateThisModeNonLexical,
		env,
		privateEnv,
	)

	SetFunctionName(function.Object, NewStringPropertyKey(string(name)), "")
	prototype := OrdinaryObjectCreate(agent, realm.Intrinsics.GeneratorFunctionPrototype, nil)
	function.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
		Value:        NewValueFromObject(prototype),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	return function
}

func (d *GeneratorDeclaration) String() string {
	return "GeneratorDeclaration " + string(d.Identifier)
}

// MARK: - LexicalDeclaration

type LexicalDeclarationType int

const (
	LexicalDeclarationTypeLet LexicalDeclarationType = iota
	LexicalDeclarationTypeConst
)

type DeclarationLexical struct {
	Declaration
	Type        LexicalDeclarationType
	BindingList *BindingList
}

func (d *DeclarationLexical) Analyze(a AnalyzeQuery) bool {
	return false
}
func (d *DeclarationLexical) Bytecode(e *Executable, c *BytecodeContext) {
	d.BindingList.Bytecode(e, c)
}
func (d *DeclarationLexical) String() string {
	return "LexicalDeclaration " + d.BindingList.String()
}

type BindingList struct {
	ASTNode
	Items []*LexicalBinding
}

func (b *BindingList) Bytecode(e *Executable, c *BytecodeContext) {
	for _, item := range b.Items {
		item.Bytecode(e, c)
	}
}
func (b *BindingList) String() string {
	var sb string
	for i, item := range b.Items {
		if i != 0 {
			sb += ", "
		}
		sb += item.String()
	}
	return sb
}

type LexicalBinding struct {
	ASTNode
	Identifier  IdentifierName
	Initializer Expression
}

func (l *LexicalBinding) Bytecode(e *Executable, c *BytecodeContext) {
	variableDecl := &VariableDeclaration{
		Identifier:  l.Identifier,
		Initializer: l.Initializer,
	}
	variableDecl.Bytecode(e, c)
}
func (l *LexicalBinding) String() string {
	if l.Initializer != nil {
		return string(l.Identifier) + " = " + l.Initializer.String()
	}
	return string(l.Identifier)
}

// MARK: - FunctionDeclaration

type FunctionDeclaration struct {
	ASTNode
	Identifier       IdentifierName
	Body             *FunctionBody
	FormalParameters *FormalParameters
	SourceText       string
}

// 15.2.2
func (f *FunctionDeclaration) functionBodyContainsUseStrict() bool {
	return f.Body.StatementList.ContainsDirective("use strict")
}

// 15.2.4
func (f *FunctionDeclaration) instantiateOrdinaryFunctionObject(agent *Agent, env EnvironmentRecord, privateEnv *PrivateEnvironment) ObjectType {
	realm := agent.CurrentRealm()
	name := f.Identifier
	sourceText := f.SourceText
	function := OrdinaryFunctionCreate(
		agent,
		realm.Intrinsics.FunctionPrototype,
		sourceText,
		f.FormalParameters,
		f.Body,
		functionCreateThisModeNonLexical,
		env,
		privateEnv,
	)

	SetFunctionName(function.Object, NewStringPropertyKey(string(name)), "")
	MakeConstructor(function, false, nil)
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

func (b *BlockStatementBlock) Analyze(a AnalyzeQuery) bool {
	return false
}

func (b *BlockStatementBlock) VarScopedDeclarations() []*VariableDeclaration {
	return b.Block.StatementList.VarScopedDeclarations()
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

func (s StatementList) VarScopedDeclarations() []*VariableDeclaration {
	var vars []*VariableDeclaration
	for _, item := range s {
		switch stmt := item.(type) {
		case *StatementListItemStatement:
			if v, ok := stmt.Statement.(*StatementVariable); ok {
				for _, varDeclaration := range v.DeclarationList.Items {
					vars = append(vars, varDeclaration)
				}
			}
		case *StatementListItemDeclaration:
			switch decl := item.(type) {
			case *DeclarationLexical:
				for _, bindingItem := range decl.BindingList.Items {
					vars = append(vars, &VariableDeclaration{
						Identifier:  bindingItem.Identifier,
						Initializer: bindingItem.Initializer,
					})
				}
			}
		}
	}
	return vars
}

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

func (s *StatementListItemStatement) VarScopedDeclarations() []*VariableDeclaration {
	return s.Statement.VarScopedDeclarations()
}

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

func (s *StatementListItemDeclaration) VarScopedDeclarations() []*VariableDeclaration {
	return nil
}

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
