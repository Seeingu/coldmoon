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
	_primaryExpression()
}

// MARK: - ClassExpression

type PrimaryExpressionClassExpression struct {
	PrimaryExpression
	IdentifierName IdentifierName
	ClassTail      *ClassTail
	SourceText     string
}

func (p *PrimaryExpressionClassExpression) _primaryExpression() {}
func (p *PrimaryExpressionClassExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}
func (p *PrimaryExpressionClassExpression) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IClassDefinitionEvaluation{ClassExpression: p})
}
func (p *PrimaryExpressionClassExpression) String() string {
	return "ClassExpression"
}

// MARK: - RegularExpressionLiteral

type PrimaryExpressionRegularExpressionLiteral struct {
	PrimaryExpression
	Pattern string
	Flags   string
}

func (p *PrimaryExpressionRegularExpressionLiteral) _primaryExpression() {}
func (p *PrimaryExpressionRegularExpressionLiteral) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}
func (p *PrimaryExpressionRegularExpressionLiteral) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&ILoadConstant{Value: NewStringValue(p.Pattern)})
	e.AddInstruction(&ILoadConstant{Value: NewStringValue(p.Flags)})
	e.AddInstruction(&IRegExpCreate{})
}
func (p *PrimaryExpressionRegularExpressionLiteral) String() string {
	return "/" + p.Pattern + "/" + p.Flags
}
func (p *PrimaryExpressionRegularExpressionLiteral) IsValidRegularExpressionLiteral() bool {
	// TODO
	return true
}

// MARK: - IdentifierReference

type IdentifierName string
type PrivateIdentifierName string
type PrimaryExpressionIdentifierReference struct {
	PrimaryExpression
	Identifier IdentifierName
}

func (p *PrimaryExpressionIdentifierReference) _primaryExpression() {}
func (p *PrimaryExpressionIdentifierReference) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
}

func (p *PrimaryExpressionIdentifierReference) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IResolveBinding{Name: p.Identifier, Strict: c.containedInStrictCode})
}
func (p *PrimaryExpressionIdentifierReference) String() string {
	return string(p.Identifier)
}

// MARK: - Literal

type PrimaryExpressionLiteral struct {
	PrimaryExpression
	Literal Literal
}

func (p *PrimaryExpressionLiteral) _primaryExpression() {}
func (p *PrimaryExpressionLiteral) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (p *PrimaryExpressionLiteral) Bytecode(e *Executable, c *BytecodeContext) {
	p.Literal.Bytecode(e, c)
}
func (p *PrimaryExpressionLiteral) String() string {
	return p.Literal.String()
}

// MARK: - AsyncFunctionExpression

type PrimaryExpressionAsyncFunctionExpression struct {
	PrimaryExpression
	Identifier       IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (p *PrimaryExpressionAsyncFunctionExpression) _primaryExpression() {}
func (p *PrimaryExpressionAsyncFunctionExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}
func (p *PrimaryExpressionAsyncFunctionExpression) Bytecode(e *Executable, c *BytecodeContext) {
	strict := c.containedInStrictCode || p.Body.FunctionBodyContainsUseStrict()
	p.Body.Strict = strict
	e.AddInstruction(&IInstantiateAsyncFunctionExpression{FunctionExpression: p})
}
func (p *PrimaryExpressionAsyncFunctionExpression) String() string {
	return "AsyncFunctionExpression"
}

// MARK: - GeneratorExpression

type PrimaryExpressionGeneratorExpression struct {
	PrimaryExpression
	IdentifierName   IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (p *PrimaryExpressionGeneratorExpression) _primaryExpression() {}
func (p *PrimaryExpressionGeneratorExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
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

func (p *PrimaryExpressionAsyncGeneratorExpression) _primaryExpression() {}
func (p *PrimaryExpressionAsyncGeneratorExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
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

func (p *PrimaryExpressionThis) _primaryExpression() {}
func (p *PrimaryExpressionThis) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (p *PrimaryExpressionThis) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IResolveThisBinding{})
}
func (p *PrimaryExpressionThis) String() string {
	return "this"
}

// MARK: - ParenthesizedExpression

type PrimaryExpressionParenthesizedExpression struct {
	PrimaryExpression
	Expression Expression
}

func (p *PrimaryExpressionParenthesizedExpression) _primaryExpression() {}
func ParenthesizedExpressionAnalyze(p *PrimaryExpressionParenthesizedExpression, query AnalyzeQuery) bool {
	switch query {
	case AnalyzeQueryIsReference:
		return ExpressionAnalyze(p.Expression, query)
	default:
		return false
	}
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
type ArrayElementSpread struct {
	ArrayElement
	Spread Expression
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
			if ExpressionAnalyze(element.Expression, AnalyzeQueryIsReference) {
				e.AddInstruction(InsGetValue)
			}
			e.AddInstruction(InsLoad)
			e.AddInstruction(&IArraySetValue{Index: i})
			e.AddInstruction(InsLoad)
		case *ArrayElementElision:
			e.AddInstruction(InsStore)
			e.AddInstruction(&IArrayPushValue{})
			e.AddInstruction(InsLoad)
		case *ArrayElementSpread:
			element.Spread.Bytecode(e, c)
			if ExpressionAnalyze(element.Spread, AnalyzeQueryIsReference) {
				e.AddInstruction(InsGetValue)
			}
			e.AddInstruction(InsLoad)
			e.AddInstruction(&IArraySpread{})
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

// MARK: - ObjectLiteral

type PrimaryExpressionObjectLiteral struct {
	PrimaryExpression
	PropertyList *PropertyDefinitionList
}

func (p *PrimaryExpressionObjectLiteral) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
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

	if ExpressionAnalyze(p.Expression, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
	e.AddInstruction(&IObjectSetProperty{})
	e.AddInstruction(InsLoad)
}

func (p *PropertyDefinitionNameAndExpression) String() string {
	return p.Name.String() + ": " + p.Expression.String()
}

type PropertyDefinitionSpread struct {
	PropertyDefinition
	Spread Expression
}

func (p *PropertyDefinitionSpread) Bytecode(e *Executable, c *BytecodeContext) {
	p.Spread.Bytecode(e, c)
	if ExpressionAnalyze(p.Spread, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
	e.AddInstruction(&IObjectSpreadValue{})
}
func (p *PropertyDefinitionSpread) String() string {
	return "..." + p.Spread.String()
}

// MARK: - Method Definition

type MethodDefinitionType int

const (
	MethodDefinitionTypeNil MethodDefinitionType = iota
	MethodDefinitionTypeMethod
	MethodDefinitionTypeGet
	MethodDefinitionTypeSet
	MethodDefinitionTypeGenerator
	MethodDefinitionTypeAsync
	MethodDefinitionTypeAsyncGenerator
)

type MethodDefinition struct {
	PropertyDefinition
	Type                     MethodDefinitionType
	PropertyName             PropertyName
	FunctionExpression       *PrimaryExpressionFunctionExpression
	GeneratorExpression      *PrimaryExpressionGeneratorExpression
	AsyncFunctionExpression  *PrimaryExpressionAsyncFunctionExpression
	AsyncGeneratorExpression *PrimaryExpressionAsyncGeneratorExpression
}

func (p *MethodDefinition) Bytecode(e *Executable, c *BytecodeContext) {
	strict := c.containedInStrictCode
	if p.FunctionExpression != nil {
		strict = strict || p.FunctionExpression.Body.FunctionBodyContainsUseStrict()
		p.FunctionExpression.Body.Strict = strict
	}
	p.PropertyName.Bytecode(e, c)
	e.AddInstruction(InsLoad)
	e.AddInstruction(&IObjectDefineMethod{
		FunctionExpression:       p.FunctionExpression,
		MethodType:               p.Type,
		GeneratorExpression:      p.GeneratorExpression,
		AsyncFunctionExpression:  p.AsyncFunctionExpression,
		AsyncGeneratorExpression: p.AsyncGeneratorExpression,
	})
}

// MARK: - FieldDefinition

type FieldDefinition struct {
	PropertyName PropertyName
	Initializer  Expression
}

// MARK: - PropertyName

type PropertyName interface {
	ASTNode
}
type PropertyNameLiteral interface {
	PropertyName
	LiteralString() string
}
type PropertyNameLiteralIdentifier struct {
	PropertyNameLiteral
	Identifier IdentifierName
}

func (p *PropertyNameLiteralIdentifier) LiteralString() string {
	return string(p.Identifier)
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

func (p *PropertyNameLiteralString) LiteralString() string {
	return p.StringLiteral.Value
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
func (p *PropertyNameLiteralNumeric) LiteralString() string {
	return p.NumericLiteral.Value
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
	if ExpressionAnalyze(p.Expression, AnalyzeQueryIsReference) {
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

func (p *PrimaryExpressionFunctionExpression) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IInstantiateOrdinaryFunctionExpression{FunctionExpression: p})
}

func (p *PrimaryExpressionFunctionExpression) String() string {
	return "FunctionExpression " + string(p.Identifier)
}

// MARK: - AsyncArrowFunction

type PrimaryExpressionAsyncArrowFunction struct {
	PrimaryExpression
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

func (p *PrimaryExpressionAsyncArrowFunction) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}
func (p *PrimaryExpressionAsyncArrowFunction) Bytecode(e *Executable, c *BytecodeContext) {
	strict := c.containedInStrictCode || p.Body.FunctionBodyContainsUseStrict()
	p.Body.Strict = strict
	e.AddInstruction(&IInstantiateAsyncArrowFunctionExpression{FunctionExpression: p})
}
func (p *PrimaryExpressionAsyncArrowFunction) String() string {
	return "AsyncArrowFunction"
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

func (m *MemberExpression) Bytecode(e *Executable, c *BytecodeContext) {
	m.Member.Bytecode(e, c)
	if ExpressionAnalyze(m.Member, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
	strict := c.containedInStrictCode

	switch prop := m.Property.(type) {
	case *ASTPropertyExpression:
		prop.Expression.Bytecode(e, c)
		if ExpressionAnalyze(prop.Expression, AnalyzeQueryIsReference) {
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

func LiteralAnalyze(l Literal, a AnalyzeQuery) bool {
	switch a {
	case AnalyzeQueryIsReference:
		return false
	case AnalyzeQueryIsStringLiteral:
		_, ok := l.(*LiteralString)
		return ok
	}
	panic("unreachable")
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

type LiteralUndefined struct {
	Literal
}

var _ Literal = (*LiteralUndefined)(nil)

func (l *LiteralUndefined) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IStoreConstant{Value: UndefinedValue})
}
func (l *LiteralUndefined) String() string {
	return "undefined"
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

// MARK: - Expression

type AssignmentTargetType int

const (
	AssignmentTargetTypeSimple AssignmentTargetType = iota
	AssignmentTargetTypeInvalid
)

type Expression interface {
	ASTNode
	AssignmentTargetType() AssignmentTargetType
}

func ExpressionAnalyze(e Expression, a AnalyzeQuery) bool {
	switch a {
	case AnalyzeQueryIsReference:
		switch ee := e.(type) {
		case PrimaryExpression:
			return PrimaryExpressionAnalyze(ee, a)
		case *MemberExpression, SuperProperty:
			return true
		default:
			return false
		}
	case AnalyzeQueryIsStringLiteral:
		switch ee := e.(type) {
		case *ExpressionPrimary:
			return PrimaryExpressionAnalyze(ee.PrimaryExpression, a)
		default:
			return false
		}
	}
	panic("unreachable")
}

// MARK: - MetaProperty

type MetaProperty interface {
	Expression
}

type MetaPropertyNewTarget struct {
	MetaProperty
}

func (m *MetaPropertyNewTarget) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}
func (m *MetaPropertyNewTarget) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(InsGetNewTarget)
}
func (m *MetaPropertyNewTarget) String() string {
	return "new.target"
}

// MARK: - SuperProperty

type SuperProperty interface {
	Expression
	_superProperty()
}

type SuperPropertyExpression struct {
	SuperProperty
	Expression Expression
}

func (s *SuperPropertyExpression) _superProperty() {}

func (s *SuperPropertyExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
}
func (s *SuperPropertyExpression) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(InsLoadThisValueSuper)
	s.Expression.Bytecode(e, c)
	if ExpressionAnalyze(s.Expression, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)

	strict := c.containedInStrictCode
	e.AddInstruction(&IMakeSuperPropertyReference{
		Strict: strict,
	})
}
func (s *SuperPropertyExpression) String() string {
	return "super." + s.Expression.String()
}

type SuperPropertyIdentifier struct {
	SuperProperty
	IdentifierName IdentifierName
}

func (s *SuperPropertyIdentifier) _superProperty() {}
func (s *SuperPropertyIdentifier) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
}
func (s *SuperPropertyIdentifier) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(InsLoadThisValueSuper)
	e.AddInstruction(&ILoadConstant{
		Value: NewStringValue(string(s.IdentifierName)),
	})
	strict := c.containedInStrictCode
	e.AddInstruction(&IMakeSuperPropertyReference{
		Strict: strict,
	})
}
func (s *SuperPropertyIdentifier) String() string {
	return "super." + string(s.IdentifierName)
}

// MARK: - SuperCall

type ExpressionSuperCall struct {
	Expression
	Arguments Arguments
}

func (e *ExpressionSuperCall) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}
func (e *ExpressionSuperCall) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Arguments.Bytecode(ex, c)
	ex.AddInstruction(&IEvaluateSuperCall{
		ArgumentCount: len(e.Arguments),
	})
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

func PrimaryExpressionAnalyze(e PrimaryExpression, a AnalyzeQuery) bool {
	switch a {
	case AnalyzeQueryIsReference:
		switch pe := e.(type) {
		case *PrimaryExpressionIdentifierReference:
			return true
		case *PrimaryExpressionParenthesizedExpression:
			return ParenthesizedExpressionAnalyze(pe, a)
		default:
			return false
		}
	case AnalyzeQueryIsStringLiteral:
		switch pe := e.(type) {
		case *PrimaryExpressionLiteral:
			return LiteralAnalyze(pe.Literal, a)
		default:
			return false
		}
	default:
		panic("unreachable")
	}
}

// MARK: - TemplateLiteral

type PrimaryExpressionTemplateLiteral struct {
	PrimaryExpression
	TemplateLiteral *TemplateLiteral
	SourceText      string
}

func (t *PrimaryExpressionTemplateLiteral) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IStoreConstant{
		Value: NewStringValue(t.SourceText),
	})
}

func (t *PrimaryExpressionTemplateLiteral) String() string {
	return t.SourceText
}

type TemplateSpan struct {
	Text       string
	Expression Expression
}
type TemplateLiteral struct {
	TemplateHead string
	Spans        []*TemplateSpan
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
func (e *ExpressionUpdate) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Operand.Bytecode(ex, c)
	ex.AddInstruction(InsPushReference)
	if ExpressionAnalyze(e.Operand, AnalyzeQueryIsReference) {
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

func (e *ExpressionAssignmentExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	if e.Operator == AssignmentOperatorAssign {
		e.Left.Bytecode(ex, c)
		ex.AddInstruction(&IPushReference{})

		e.Right.Bytecode(ex, c)
		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		ex.AddInstruction(&IPutValue{})
		ex.AddInstruction(&IPopReference{})
	} else if e.Operator != AssignmentOperatorAnd &&
		e.Operator != AssignmentOperatorOr &&
		e.Operator != AssignmentOperatorNullishCoalescing {
		e.Left.Bytecode(ex, c)
		ex.AddInstruction(&IPushReference{})

		if ExpressionAnalyze(e.Left, AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}
		ex.AddInstruction(InsLoad)

		e.Right.Bytecode(ex, c)

		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
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

		if ExpressionAnalyze(e.Left, AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}
		ex.AddInstruction(InsLoad)

		jumpIfTrue := &IJumpIfTrue{}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.Target = len(ex.Instructions) - 1

		e.Right.Bytecode(ex, c)
		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
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

		if ExpressionAnalyze(e.Left, AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		jumpIfTrue := &IJumpIfTrue{}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.TargetElse = len(ex.Instructions) - 1

		e.Right.Bytecode(ex, c)
		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		ex.AddInstruction(InsPutValue)

		jumpIfTrue.Target = len(ex.Instructions) - 1

		ex.AddInstruction(InsPopReference)
	} else if e.Operator == AssignmentOperatorNullishCoalescing {
		e.Left.Bytecode(ex, c)
		ex.AddInstruction(InsPushReference)

		if ExpressionAnalyze(e.Left, AnalyzeQueryIsReference) {
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
		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
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

func (e *ExpressionNewExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Callee.Bytecode(ex, c)
	if ExpressionAnalyze(e.Callee, AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	ex.AddInstruction(InsLoad)

	for _, arg := range e.Arguments {
		arg.Bytecode(ex, c)
		if ExpressionAnalyze(arg, AnalyzeQueryIsReference) {
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

func (b *ExpressionBinaryExpression) Bytecode(e *Executable, c *BytecodeContext) {
	b.Left.Bytecode(e, c)
	if ExpressionAnalyze(b.Left, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)

	b.Right.Bytecode(e, c)
	if ExpressionAnalyze(b.Right, AnalyzeQueryIsReference) {
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

func (e *ExpressionSequenceExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	for _, expr := range e.Expressions {
		expr.Bytecode(ex, c)
		if ExpressionAnalyze(expr, AnalyzeQueryIsReference) {
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

func (e *ExpressionConditionalExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Test.Bytecode(ex, c)
	if ExpressionAnalyze(e.Test, AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
	ex.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(ex.Instructions) - 1
	e.Consequent.Bytecode(ex, c)

	if ExpressionAnalyze(e.Consequent, AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}

	jump := &IJump{Target: 0}
	ex.AddInstruction(jump)

	jumpIfTrue.TargetElse = len(ex.Instructions)
	e.Alternate.Bytecode(ex, c)

	if ExpressionAnalyze(e.Alternate, AnalyzeQueryIsReference) {
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

func (e *ExpressionLogicalExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Left.Bytecode(ex, c)
	if ExpressionAnalyze(e.Left, AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}

	switch e.Operator {
	case LogicalOperatorAnd:
		jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.Target = len(ex.Instructions)
		e.Right.Bytecode(ex, c)

		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		jumpIfTrue.TargetElse = len(ex.Instructions)
	case LogicalOperatorOr:
		jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.TargetElse = len(ex.Instructions)
		e.Right.Bytecode(ex, c)
		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
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
		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
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

func (e *ExpressionEqualityExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Left.Bytecode(ex, c)
	if ExpressionAnalyze(e.Left, AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	ex.AddInstruction(InsLoad)

	e.Right.Bytecode(ex, c)
	if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
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

func (e *ExpressionRelationalExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Left.Bytecode(ex, c)
	if ExpressionAnalyze(e.Left, AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}
	ex.AddInstruction(InsLoad)

	e.Right.Bytecode(ex, c)
	if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
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

func (u UnaryOperator) String() string {
	switch u {
	case UnaryOperatorDelete:
		return "delete"
	case UnaryOperatorVoid:
		return "void"
	case UnaryOperatorTypeof:
		return "typeof"
	case UnaryOperatorAddition:
		return "+"
	case UnaryOperatorSubtraction:
		return "-"
	case UnaryOperatorLogicalNot:
		return "!"
	case UnaryOperatorBitwiseNot:
		return "~"
	}
	return ""
}

type UnaryExpression struct {
	Expression
	Operator UnaryOperator
	Operand  Expression
}

func (u *UnaryExpression) Bytecode(e *Executable, c *BytecodeContext) {
	u.Operand.Bytecode(e, c)
	switch u.Operator {
	case UnaryOperatorDelete:
		u.Expression.Bytecode(e, c)
		if !ExpressionAnalyze(u.Expression, AnalyzeQueryIsReference) {
			e.AddInstruction(&IStoreConstant{
				Value: NewBooleanValue(true),
			})
		} else {
			e.AddInstruction(&IDelete{})
		}
	case UnaryOperatorVoid:
		u.Operand.Bytecode(e, c)
		if ExpressionAnalyze(u.Operand, AnalyzeQueryIsReference) {
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

		if ExpressionAnalyze(u.Operand, AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(InsToNumber)
	case UnaryOperatorSubtraction:
		u.Operand.Bytecode(e, c)

		if ExpressionAnalyze(u.Operand, AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(InsToNumeric)
		e.AddInstruction(InsUnaryMinus)
	case UnaryOperatorLogicalNot:
		u.Operand.Bytecode(e, c)

		if ExpressionAnalyze(u.Operand, AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(&ILogicalNot{})
	case UnaryOperatorBitwiseNot:
		u.Operand.Bytecode(e, c)

		if ExpressionAnalyze(u.Operand, AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(&IBitwiseNot{})
	}
}

func (u *UnaryExpression) String() string {
	return "UnaryExpression " + u.Operator.String() + u.Operand.String()
}

// MARK: - CallExpression

type Arguments []Expression

func (a Arguments) Bytecode(e *Executable, c *BytecodeContext) {
	for _, arg := range a {
		arg.Bytecode(e, c)
		if ExpressionAnalyze(arg, AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(InsLoad)
	}
}

type CallExpression struct {
	Expression
	Callee    Expression
	Arguments Arguments
}

var _ Expression = (*CallExpression)(nil)

func (c *CallExpression) Bytecode(e *Executable, bc *BytecodeContext) {
	c.Callee.Bytecode(e, bc)

	e.AddInstruction(&IPushReference{})
	isReference := ExpressionAnalyze(c.Callee, AnalyzeQueryIsReference)
	if isReference {
		e.AddInstruction(InsGetValue)
	}

	e.AddInstruction(InsLoad)
	e.AddInstruction(InsLoadThisValue)
	for _, arg := range c.Arguments {
		arg.Bytecode(e, bc)
		if ExpressionAnalyze(arg, AnalyzeQueryIsReference) {
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
	VarScopedDeclarations() []*VariableDeclaration
	_statement()
}

func StatementAnalyze(s Statement, a AnalyzeQuery) bool {
	exprStmt, isExpr := s.(*StatementExpression)
	switch a {
	case AnalyzeQueryIsReference:
		if isExpr {
			return ExpressionAnalyze(exprStmt.Expression, a)
		}
		return false
	case AnalyzeQueryIsStringLiteral:
		if isExpr {
			return ExpressionAnalyze(exprStmt.Expression, a)
		}
		return false
	}
	panic("unreachable")
}

// MARK: - VariableStatement

type StatementVariable struct {
	Statement
	DeclarationList *VariableDeclarationList
}

var _ Statement = (*StatementVariable)(nil)

func (s *StatementVariable) _statement() {}
func (s *StatementVariable) VarScopedDeclarations() (l []*VariableDeclaration) {
	return s.DeclarationList.VarScopedDeclarations()
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

	if ExpressionAnalyze(v.Initializer, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsPutValue)
	e.AddInstruction(InsPopReference)

	e.AddInstruction(InsStore)
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

func (s *StatementBlock) _statement() {}
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

func (s *StatementEmpty) _statement() {}

var _ Statement = (*StatementEmpty)(nil)

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

func (t *StatementTry) _statement() {}

var _ Statement = (*StatementTry)(nil)

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

func (s *StatementExpression) _statement() {}
func (s *StatementExpression) Bytecode(e *Executable, c *BytecodeContext) {
	s.Expression.Bytecode(e, c)
	if ExpressionAnalyze(s.Expression, AnalyzeQueryIsReference) {
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
	if ExpressionAnalyze(s.Expression, AnalyzeQueryIsReference) {
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
		var name IdentifierName
		switch p := item.(type) {
		case *FormalParameter:
			name = p.BindingElement.Identifier
		case *FormalParameterFunctionRestParameter:
			name = p.BindingRestElement.(*BindingRestElementIdentifier).Identifier
		}
		l = append(l, name)
	}
	return
}

// 15.1.5
func (f *FormalParameters) ExpectedArgumentCount() int {
	l := len(f.Items)
	if l > 0 {
		if _, ok := f.Items[l-1].(*FormalParameterFunctionRestParameter); ok {
			return l - 1
		}
	}
	return l
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

// MARK:-
type FormalParameterFunctionRestParameter struct {
	FormalParametersItem
	BindingRestElement BindingRestElement
}

// MARK: - BindingRestElement

type BindingRestElement interface {
}
type BindingRestElementIdentifier struct {
	BindingRestElement
	Identifier IdentifierName
}

func (b *BindingRestElementIdentifier) String() string {
	return "..." + string(b.Identifier)
}

// MARK: - FormalParameter

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

// 14.6.2
func (s *StatementIf) Bytecode(e *Executable, c *BytecodeContext) {
	s.Condition.Bytecode(e, c)

	if ExpressionAnalyze(s.Condition, AnalyzeQueryIsReference) {
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

func (s *StatementWhile) VarScopedDeclarations() []*VariableDeclaration {
	return s.Body.VarScopedDeclarations()
}

func (s *StatementWhile) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&ILoadConstant{Value: UndefinedValue})

	condition := len(e.Instructions)
	s.Condition.Bytecode(e, c)
	if ExpressionAnalyze(s.Condition, AnalyzeQueryIsReference) {
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
	if ExpressionAnalyze(s.Condition, AnalyzeQueryIsReference) {
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
			if ExpressionAnalyze(initializer.Expression, AnalyzeQueryIsReference) {
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
		if ExpressionAnalyze(s.Condition, AnalyzeQueryIsReference) {
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
		if ExpressionAnalyze(s.Increment, AnalyzeQueryIsReference) {
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

func (s *StatementReturn) Bytecode(e *Executable, c *BytecodeContext) {
	if s.Expression != nil {
		s.Expression.Bytecode(e, c)
		if ExpressionAnalyze(s.Expression, AnalyzeQueryIsReference) {
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
	_declaration()
}

func DeclarationAnalyze(d Declaration, a AnalyzeQuery) bool {
	return false
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

func (d *DeclarationHoistableFunction) _declaration() {}
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

func (d *DeclarationHoistableAsyncFunction) _declaration() {}
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

// MARK: - ClassDeclaration

type DeclarationClass struct {
	Declaration
	IdentifierName IdentifierName
	ClassTail      *ClassTail
	SourceText     string
}

func (d *DeclarationClass) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(InsLoad)
	e.AddInstruction(&IBindingClassDeclarationEvaluation{
		ClassDeclaration: d,
	})
	e.AddInstruction(InsStore)
}
func (d *DeclarationClass) String() string {
	return "ClassDeclaration " + string(d.IdentifierName)
}

type ClassTail struct {
	ClassHeritage Expression
	ClassBody     *ClassBody
}

type ClassBody struct {
	ASTNode
	ClassElementList *ClassElementList
}

func (c *ClassBody) ConstructorMethod() *MethodDefinition {
	for _, item := range c.ClassElementList.Items {
		if item.ClassElementKind() == ClassElementKindConstructorMethod {
			return item.(*ClassElementMethodDefinition).MethodDefinition
		}
	}
	return nil
}

func (c *ClassBody) PrivateBoundIdentifiers() (l []PrivateIdentifierName) {
	// TODO
	return
}

func (c *ClassBody) NonConstructorElements() (l []ClassElement) {
	for _, item := range c.ClassElementList.Items {
		if item.ClassElementKind() == ClassElementKindNonConstructorMethod {
			l = append(l, item)
		}
	}
	return l
}

func (c *ClassBody) Bytecode(e *Executable, cx *BytecodeContext) {
	containedInStrictCode := cx.containedInStrictCode
	cx.containedInStrictCode = true
	defer func() {
		cx.containedInStrictCode = containedInStrictCode
	}()
	c.ClassElementList.Bytecode(e, cx)
}
func (c *ClassBody) String() string {
	return c.ClassElementList.String()
}

// MARK: - ClassElementList

type ClassElementList struct {
	ASTNode
	Items []ClassElement
}

func (c *ClassElementList) Bytecode(e *Executable, cx *BytecodeContext) {
	for _, item := range c.Items {
		item.Bytecode(e, cx)
	}
}
func (c *ClassElementList) String() string {
	var sb string
	for i, item := range c.Items {
		if i != 0 {
			sb += ", "
		}
		sb += item.String()
	}
	return sb
}

// MARK: - Class: ClassElement

type ClassElementKind int

const (
	ClassElementKindConstructorMethod ClassElementKind = iota
	ClassElementKindNonConstructorMethod
	ClassElementKindEmpty
)

type ClassElement interface {
	ASTNode
	ClassElementKind() ClassElementKind
}

// 15.7.4
func ClassElementIsStatic(c ClassElement) bool {
	switch c.(type) {
	case *ClassElementStaticMethodDefinition:
		return true
	default:
		return false
	}
}

// MARK: - ClassElement: ClassStaticBlock

type ClassElementStaticBlock struct {
	ClassElement
	StatementList StatementList
}

func (c *ClassElementStaticBlock) ClassElementKind() ClassElementKind {
	return ClassElementKindNonConstructorMethod
}

// MARK: - ClassElement: FieldDefinition, StaticFieldDefinition

type ClassElementFieldDefinition struct {
	ClassElement
	FieldDefinition *FieldDefinition
}

func (c *ClassElementFieldDefinition) ClassElementKind() ClassElementKind {
	return ClassElementKindNonConstructorMethod
}

type ClassElementStaticFieldDefinition struct {
	ClassElement
	FieldDefinition *FieldDefinition
}

func (c *ClassElementStaticFieldDefinition) ClassElementKind() ClassElementKind {
	return ClassElementKindNonConstructorMethod
}

// MARK: - ClassElement: EmptyStatement

type ClassElementEmpty struct {
	ClassElement
}

func (c *ClassElementEmpty) ClassElementKind() ClassElementKind {
	return ClassElementKindEmpty
}

// MARK: - ClassElement: StaticMethodDefinition

type ClassElementStaticMethodDefinition struct {
	ClassElement
	MethodDefinition *MethodDefinition
}

func (c *ClassElementStaticMethodDefinition) ClassElementKind() ClassElementKind {
	return ClassElementKindNonConstructorMethod
}

// MARK: - ClassElement: MethodDefinition

type ClassElementMethodDefinition struct {
	ClassElement
	MethodDefinition *MethodDefinition
}

func (c *ClassElementMethodDefinition) ClassElementKind() ClassElementKind {
	switch pn := c.MethodDefinition.PropertyName.(type) {
	case PropertyNameLiteral:
		if l, ok := pn.(*PropertyNameLiteralIdentifier); ok {
			if l.Identifier == "constructor" {
				return ClassElementKindConstructorMethod
			}
		}
	}
	return ClassElementKindNonConstructorMethod
}

func (c *ClassElementMethodDefinition) Bytecode(e *Executable, cx *BytecodeContext) {
	c.MethodDefinition.Bytecode(e, cx)
}
func (c *ClassElementMethodDefinition) String() string {
	return c.MethodDefinition.String()
}

// MARK: - Class: MethodDefinition

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
		if !StatementListItemAnalyze(item, AnalyzeQueryIsStringLiteral) {
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
}

func StatementListItemAnalyze(s StatementListItem, a AnalyzeQuery) bool {
	switch t := s.(type) {
	case *StatementListItemStatement:
		return StatementAnalyze(t.Statement, a)
	case *StatementListItemDeclaration:
		return DeclarationAnalyze(t.Declaration, a)
	}
	return false
}

type StatementListItemStatement struct {
	StatementListItem
	Statement Statement
}

var _ ASTNode = (*StatementListItemStatement)(nil)

func (s *StatementListItemStatement) VarScopedDeclarations() []*VariableDeclaration {
	return s.Statement.VarScopedDeclarations()
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

func (s *StatementListItemDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	s.Declaration.Bytecode(e, c)
}

func (s *StatementListItemDeclaration) String() string {
	return s.Declaration.String()
}

type ExpressionStatement struct {
	Expression Expression
}

// 14.5.1
func (e *ExpressionStatement) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Expression.Bytecode(ex, c)
}

func (e *ExpressionStatement) String() string {
	return e.Expression.String()
}

// MARK: - Script

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

// MARK: - Module

type Module struct {
	ASTNode
	ModuleItemList ModuleItemList
}

func (m *Module) Bytecode(e *Executable, c *BytecodeContext) {
	m.ModuleItemList.Bytecode(e, c)
}

type ModuleItemList []ModuleItem

func (m ModuleItemList) Bytecode(e *Executable, c *BytecodeContext) {
	for _, item := range m {
		item.Bytecode(e, c)
	}
}

// MARK: - ModuleItem

type ModuleItem interface {
	ASTNode
	_moduleItem()
}

// MARK: - ModuleItem: StatementListItem

type ModuleItemStatementListItem struct {
	ModuleItem
	StatementListItem StatementListItem
}

func (m *ModuleItemStatementListItem) Bytecode(e *Executable, c *BytecodeContext) {
	m.StatementListItem.Bytecode(e, c)
}

func (m *ModuleItemStatementListItem) _moduleItem() {}

// MARK: - ModuleItem: ImportDeclaration

type ModuleItemImportDeclaration struct {
	ModuleItem
}

func (m *ModuleItemImportDeclaration) _moduleItem() {}

// MARK: - ModuleItem: ExportDeclaration

type ModuleItemExportDeclaration struct {
	ModuleItem
}

func (m *ModuleItemExportDeclaration) _moduleItem() {}
