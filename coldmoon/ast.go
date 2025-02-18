package coldmoon

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/Seeingu/coldmoon/pkg"
	"github.com/samber/lo"
)

// Deprecated
type boundName interface {
	BoundNames() (l []IdentifierName)
}

type ASTNode interface {
	String() string
	Bytecode(e *Executable, c *BytecodeContext)
	RuntimeSemanticsEvaluation
}

// MARK: - AnalyzeQuery

type AnalyzeQuery int

const (
	AnalyzeQueryIsReference AnalyzeQuery = iota
	AnalyzeQueryIsIdentifierReference
	AnalyzeQueryIsStringLiteral
)

// MARK: - Scope Declaration

// VarScopedDeclaration Enum
type VarScopedDeclaration struct {
	VariableDeclaration  *VariableDeclaration
	HoistableDeclaration DeclarationHoistable
}

// LexicallyScopedDeclaration Enum
type LexicallyScopedDeclaration struct {
	HoistableDeclaration DeclarationHoistable
	// TODO
}

// MARK: - PrimaryExpression

type PrimaryExpression interface {
	Expression
	AssignmentTargetType() AssignmentTargetType
	_primaryExpression()
}

// MARK: - ClassExpression

// ClassExpression [Yield, Await] :
// - class BindingIdentifier[?Yield, ?Await] opt ClassTail[?Yield, ?Await]
type ClassExpression struct {
	PrimaryExpression
	IdentifierName IdentifierName
	ClassTail      *ClassTail
	SourceText     string
}

func (p *ClassExpression) _primaryExpression() {}
func (p *ClassExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (p *ClassExpression) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IClassDefinitionEvaluation{ClassExpression: p})
}

func (p *ClassExpression) astHasIdentifier() bool {
	return p.IdentifierName != ""
}

// Evaluation
// spec: 15.7.16
func (p *ClassExpression) Evaluation(vm *VM2) Value {
	if p.astHasIdentifier() {
		className := p.IdentifierName
		value, err := p.ClassTail.ClassDefinitionEvaluation(vm, className, NewStringPropertyKey(className))
		if err != nil {
			vm.panic(err)
		}
		// TODO: set sourceText
		return value.ToValue()
	} else {
		value, err := p.ClassTail.ClassDefinitionEvaluation(vm, "", NewStringPropertyKey(""))
		if err != nil {
			vm.panic(err)
		}
		// TODO: set sourceText
		return value.ToValue()
	}
}

func (p *ClassExpression) String() string {
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

type (
	IdentifierName        = string
	PrivateIdentifierName string
)

// TODO(SM): should be rename to IdentifierReference
// IdentifierReference [Yield, Await] :
// Identifier
// [~Yield] yield
// [~Await] await
type PrimaryExpressionIdentifierReference struct {
	PrimaryExpression
	Identifier IdentifierName
}

func (p *PrimaryExpressionIdentifierReference) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
}

// 13.1.3
func (p *PrimaryExpressionIdentifierReference) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IResolveBinding{Name: p.Identifier, Strict: c.containedInStrictCode})
}

// Evaluation 13.1.3
func (p *PrimaryExpressionIdentifierReference) Evaluation(vm *VM2) Value {
	return NewReferenceRecordValue(
		vm.agent.ResolveBinding(
			p.Identifier,
			nil, vm.containedInStrictCode,
		),
	)
}

func (p *PrimaryExpressionIdentifierReference) String() string {
	return string(p.Identifier)
}

// MARK: - Literal

// TODO(SM): should be rename to Literal
// Literal :
// NullLiteral
// BooleanLiteral
// NumericLiteral
// StringLiteral
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

// Evaluation Literal 13.2.3.1
func (p *PrimaryExpressionLiteral) Evaluation(vm *VM2) Value {
	return p.Literal.Evaluation(vm)
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

// GeneratorExpression :
//   - function * BindingIdentifier[+Yield, ~Await] opt ( FormalParameters[+Yield, ~Await]
//     ) { GeneratorBody }
type GeneratorExpression struct {
	PrimaryExpression
	IdentifierName   IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

var _ RuntimeSemanticsInstantiateGeneratorFunctionExpression = (*GeneratorExpression)(nil)

func (p *GeneratorExpression) _primaryExpression() {}
func (p *GeneratorExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (p *GeneratorExpression) Bytecode(e *Executable, c *BytecodeContext) {
	strict := c.containedInStrictCode || p.Body.FunctionBodyContainsUseStrict()
	p.Body.Strict = strict
	e.AddInstruction(&IInstantiateGeneratorFunctionExpression{FunctionExpression: p})
}

func (p *GeneratorExpression) astHasIdentifierName() bool {
	return p.IdentifierName != ""
}

// InstantiateGeneratorFunctionExpression
// spec: 15.5.4
func (p *GeneratorExpression) InstantiateGeneratorFunctionExpression(
	vm *VM2,
	name PropertyKeyOrPrivateName,
) (fun ObjectType) {
	agent := vm.agent
	realm := agent.CurrentRealm()
	if p.astHasIdentifierName() {
		env := vm.RunningLexicalEnvironment()
		privateEnv := vm.RunningPrivateEnvironment()
		sourceText := p.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.GeneratorFunctionPrototype,
			sourceText,
			p.FormalParameters,
			p.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		SetFunctionName(closure, name, "")
		prototype := OrdinaryObjectCreate(agent, realm.Intrinsics.GeneratorFunctionPrototypePrototype, nil)
		closure.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
			Value:        prototype.ToValue(),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})
		return closure
	} else {
		outerEnv := vm.RunningLexicalEnvironment()
		funcEnv := NewDeclarativeEnvironment(outerEnv)
		funcEnv.CreateImmutableBinding(p.IdentifierName, false)
		privateEnv := vm.RunningPrivateEnvironment()
		sourceText := p.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.GeneratorFunctionPrototype,
			sourceText,
			p.FormalParameters,
			p.Body,
			functionCreateThisModeNonLexical,
			funcEnv,
			privateEnv,
		)
		SetFunctionName(closure, NewStringPropertyKey(p.IdentifierName), "")
		prototype := OrdinaryObjectCreate(agent, realm.Intrinsics.GeneratorFunctionPrototypePrototype, nil)
		closure.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
			Value:        prototype.ToValue(),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})
		return closure
	}
}

func (p *GeneratorExpression) Evaluation(vm *VM2) Value {
	// TODO: name
	name := NewStringPropertyKey("")
	return p.InstantiateGeneratorFunctionExpression(vm, name).ToValue()
}

func (p *GeneratorExpression) String() string {
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

// This keyword
// spec: 13.2.1
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

// Evaluation
// spec: 13.2.1.1
func (p *PrimaryExpressionThis) Evaluation(vm *VM2) Value {
	return vm.agent.ResolveThisBinding()
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

type (
	ArrayElement        any
	ArrayElementElision struct {
		ArrayElement
	}
)

type ArrayElementExpression struct {
	ArrayElement
	Expression Expression
}
type ArrayElementSpread struct {
	ArrayElement
	Spread Expression
}

// TODO: handle elision
// ElementList [Yield, Await] :
// Elisionopt AssignmentExpression[+In, ?Yield, ?Await]
// Elisionopt SpreadElement[?Yield, ?Await]
// ElementList[?Yield, ?Await] , Elisionopt AssignmentExpression[+In, ?Yield, ?Await]
// ElementList[?Yield, ?Await] , Elisionopt SpreadElement[?Yield, ?Await]
type ElementList []ArrayElement

var _ RuntimeSemanticsArrayAccumulation = ElementList{}

// ElementList : Elisionopt AssignmentExpression
func (e ElementList) astIsOnlyHasAssigmentExpression() (expr *ArrayElementExpression, ok bool) {
	if len(e) != 1 {
		return
	}
	expr, ok = e[0].(*ArrayElementExpression)
	return
}

func (e ElementList) astHasElementListAndAssignmentExpression() (list ElementList, expr *ArrayElementExpression, ok bool) {
	if len(e) < 2 {
		return
	}
	list = e[:len(e)-1]
	expr, ok = e[len(e)-1].(*ArrayElementExpression)
	return
}

// ArrayAccumulation
// spec: 13.2.4.1
func (e ElementList) ArrayAccumulation(vm *VM2, array *ArrayObject, nextIndex JSInt) (index JSInt, err Value) {
	if list, expr, ok := e.astHasElementListAndAssignmentExpression(); ok {
		nextIndex, err = list.ArrayAccumulation(vm, array, nextIndex)
		if err != nil {
			return
		}
		// TODO: check elision
		initResult := expr.Expression.Evaluation(vm)
		initValue := initResult.GetValue(vm.agent)
		array.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(nextIndex), initValue)
		return nextIndex + 1, nil
	} else if expr, ok = e.astIsOnlyHasAssigmentExpression(); ok {
		// TODO: check elision
		initResult := expr.Expression.Evaluation(vm)
		initValue := initResult.GetValue(vm.agent)
		array.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(nextIndex), initValue)
		return nextIndex + 1, nil
	} else {
		panic("unimplemented")
	}
}

// ArrayLiteral [Yield, Await] :
// - [ Elisionopt ]
// - [ ElementList[?Yield, ?Await] ]
// - [ ElementList[?Yield, ?Await] , Elisionopt ]
type ArrayLiteral struct {
	PrimaryExpression
	ElementList ElementList
}

func (p *ArrayLiteral) Bytecode(e *Executable, c *BytecodeContext) {
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

// TODO: Elision
func (p *ArrayLiteral) astHasElision() bool {
	return false
}

func (p *ArrayLiteral) astElementListEmpty() bool {
	return len(p.ElementList) == 0
}

// Evaluation
// spec: 13.2.4.2
func (p *ArrayLiteral) Evaluation(vm *VM2) Value {
	array := ArrayCreate(vm.agent, 0, nil)
	if p.astElementListEmpty() {
		return array.ToValue()
	}
	if p.astHasElision() {
	} else {
		p.ElementList.ArrayAccumulation(vm, array, 0)
	}
	return array.ToValue()
}

func (p *ArrayLiteral) String() string {
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

// ObjectLiteral [Yield, Await] :
// - { }
// - { PropertyDefinitionList[?Yield, ?Await] }
// - { PropertyDefinitionList[?Yield, ?Await] , }
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

func (p *PrimaryExpressionObjectLiteral) astIsEmpty() bool {
	return len(p.PropertyList.Items) == 0
}

// 13.2.5.4
func (p *PrimaryExpressionObjectLiteral) Evaluation(vm *VM2) Value {
	if p.astIsEmpty() {
		return OrdinaryObjectCreate(vm.agent, vm.agent.CurrentRealm().Intrinsics.ObjectPrototype, nil).ToValue()
	}

	obj := OrdinaryObjectCreate(vm.agent, vm.agent.CurrentRealm().Intrinsics.ObjectPrototype, nil)
	p.PropertyList.PropertyDefinitionEvaluation(vm, obj)
	return obj.ToValue()
}

func (p *PrimaryExpressionObjectLiteral) String() string {
	sb := "{"
	sb += p.PropertyList.String()
	sb += "}"
	return sb
}

// PropertyDefinitionList [Yield, Await] :
// PropertyDefinition[?Yield, ?Await]
// PropertyDefinitionList[?Yield, ?Await] , PropertyDefinition[?Yield, ?Await]
type PropertyDefinitionList struct {
	ASTNode
	Items []PropertyDefinition
}

var _ RuntimeSemanticsPropertyDefinitionEvaluation = (*PropertyDefinitionList)(nil)

// 13.2.5.4
func (p *PropertyDefinitionList) PropertyDefinitionEvaluation(vm *VM2, obj ObjectType) {
	for _, item := range p.Items {
		item.PropertyDefinitionEvaluation(vm, obj)
	}
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

// TODO(SM): remove
// PropertyDefinition [Yield, Await] :
// - IdentifierReference[?Yield, ?Await]
// - CoverInitializedName[?Yield, ?Await]
// - PropertyName[?Yield, ?Await] : AssignmentExpression[+In, ?Yield, ?Await]
// - MethodDefinition[?Yield, ?Await]
// - ... AssignmentExpression[+In, ?Yield, ?Await]
type PropertyDefinition interface {
	ASTNode
	RuntimeSemanticsPropertyDefinitionEvaluation
}

type PropertyDefinitionMethodDefinition struct {
	*MethodDefinition
}

var _ RuntimeSemanticsPropertyDefinitionEvaluation = (*PropertyDefinitionMethodDefinition)(nil)

func (p *PropertyDefinitionMethodDefinition) PropertyDefinitionEvaluation(vm *VM2, obj ObjectType) {
	p.MethodDefinition.MethodDefinitionEvaluation(vm, obj, true)
}

type PropertyDefinitionIdentifierReference struct {
	PropertyDefinition
	IdentifierReference *PrimaryExpressionIdentifierReference
}

func (p *PropertyDefinitionIdentifierReference) Bytecode(e *Executable, c *BytecodeContext) {
	propName := p.IdentifierReference.Identifier
	e.AddInstruction(&ILoadConstant{
		Value: NewStringValue(propName),
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
	e.AddInstruction(InsLoad)

	p.Expression.Bytecode(e, c)

	if ExpressionAnalyze(p.Expression, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
	e.AddInstruction(&IObjectSetProperty{})
}

func (p *PropertyDefinitionNameAndExpression) PropertyDefinitionEvaluation(vm *VM2, object ObjectType) {
	propKey := p.Name.Evaluation(vm)
	var isProtoSetter bool
	if vm.IsJSONParse {
		isProtoSetter = false
	} else if propKey.String() == "__proto__" && !IsComputedPropertyKeyDefault(p.Name) {
		isProtoSetter = true
	} else {
		isProtoSetter = false
	}

	var propValue Value
	if IsAnonymousFunctionDefinition(p.Expression) {
		panic("unimplemented")
	} else {
		exprValueRef := p.Expression.Evaluation(vm)
		propValue = exprValueRef.GetValue(vm.agent)
	}
	if isProtoSetter {
		if ValueIsObject(propValue) || propValue == NullValue {
			object.SetPrototype(propValue.ToObject(vm.agent))
		}
		return
	}
	// TODO: assert no non-configurable properties
	Assert(object.IsOrdinary() && object.IsExtensible())
	object.CreateDataPropertyOrThrow(
		ToPropertyKey(vm.agent, propKey), propValue)
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

// MethodDefinition [Yield, Await] :
// - ClassElementName[?Yield, ?Await] ( UniqueFormalParameters[~Yield, ~Await] ) {
// - FunctionBody[~Yield, ~Await] }
// - GeneratorMethod[?Yield, ?Await]
// - AsyncMethod[?Yield, ?Await]
// - AsyncGeneratorMethod[?Yield, ?Await]
// - get ClassElementName[?Yield, ?Await] ( ) { FunctionBody[~Yield, ~Await] }
// - set ClassElementName[?Yield, ?Await] ( PropertySetParameterList ) {
// - FunctionBody[~Yield, ~Await] }
type MethodDefinition struct {
	PropertyDefinition
	Type                     MethodDefinitionType
	PropertyName             PropertyName
	FunctionExpression       *FunctionExpression
	GeneratorExpression      *GeneratorExpression
	AsyncFunctionExpression  *PrimaryExpressionAsyncFunctionExpression
	AsyncGeneratorExpression *PrimaryExpressionAsyncGeneratorExpression
}

var (
	_ RuntimeSemanticsMethodDefinitionEvaluation = (*MethodDefinition)(nil)
	_ RuntimeSemanticsDefineMethod               = (*MethodDefinition)(nil)
)

func (p *MethodDefinition) sourceText() string {
	if p.FunctionExpression != nil {
		return p.FunctionExpression.SourceText
	} else if p.GeneratorExpression != nil {
		return p.GeneratorExpression.SourceText
	} else if p.AsyncFunctionExpression != nil {
		return p.AsyncFunctionExpression.SourceText
	} else if p.AsyncGeneratorExpression != nil {
		return p.AsyncGeneratorExpression.SourceText
	}
	return ""
}

func (p *MethodDefinition) formalParameters() *FormalParameters {
	if p.FunctionExpression != nil {
		return p.FunctionExpression.FormalParameters
	} else if p.GeneratorExpression != nil {
		return p.GeneratorExpression.FormalParameters
	} else if p.AsyncFunctionExpression != nil {
		return p.AsyncFunctionExpression.FormalParameters
	} else if p.AsyncGeneratorExpression != nil {
		return p.AsyncGeneratorExpression.FormalParameters
	}
	return nil
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

func (p *MethodDefinition) astIsClassElementName() bool {
	return p.PropertyName != nil
}

func (p *MethodDefinition) astIsGet() bool {
	return p.Type == MethodDefinitionTypeGet
}

// DefineMethod
// spec: 15.4.4
func (p *MethodDefinition) DefineMethod(vm *VM2, obj ObjectType, proto ObjectType) (record DefineMethodRecord, err Value) {
	agent := vm.agent
	realm := agent.CurrentRealm()
	if p.astIsClassElementName() {
		propKey := p.PropertyName.Evaluation(vm)
		env := vm.RunningLexicalEnvironment()
		privateEnv := vm.RunningPrivateEnvironment()
		var prototype ObjectType
		if proto != nil {
			prototype = proto
		} else {
			prototype = realm.Intrinsics.FunctionPrototype
		}
		sourceText := p.sourceText()
		closure := OrdinaryFunctionCreate(
			agent,
			prototype,
			sourceText,
			p.FunctionExpression.FormalParameters,
			p.FunctionExpression.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		MakeMethod(closure, obj)
		record.Key = ToPropertyKey(agent, propKey)
		record.Closure = closure
		return
	} else {
		panic("unreachable")
	}
}

func (p *MethodDefinition) MethodDefinitionEvaluation(vm *VM2, obj ObjectType, enumerable bool) (pe *PrivateElement, err Value) {
	agent := vm.agent
	if p.astIsGet() {
		propKey := p.PropertyName.Evaluation(vm)
		env := vm.RunningLexicalEnvironment()
		privateEnv := vm.RunningPrivateEnvironment()
		sourceText := p.sourceText()
		formalParameterList := p.formalParameters()
		closure := OrdinaryFunctionCreate(
			agent,
			agent.CurrentRealm().Intrinsics.FunctionPrototype,
			sourceText,
			formalParameterList,
			p.FunctionExpression.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		MakeMethod(closure, obj)
		// TODO: check propKey is private name
		SetFunctionName(closure, propKey.ToPropertyKey(), "get")
		// TODO: check propKey is private name
		isPrivateName := false
		if isPrivateName {
			panic("unimplemented")
		} else {
			desc := &PropertyDescriptor{
				Get:          closure,
				Enumerable:   enumerable,
				Configurable: true,
			}
			obj.DefinePropertyOrThrow(propKey.ToPropertyKey(), desc)
			// return UNUSED
			return
		}
	} else if p.astIsClassElementName() {
		methodDef, err := p.DefineMethod(vm, obj, nil)
		if err != nil {
			return nil, err
		}
		SetFunctionName(methodDef.Closure, methodDef.Key, "")
		return DefineMethodProperty(obj, methodDef.Key, methodDef.Closure, enumerable), nil
	} else {
		panic("unimplemented")
	}
}

func (p *MethodDefinition) String() string {
	return "MethodDefinition " + p.PropertyName.String()
}

// MARK: - FieldDefinition

type FieldDefinition struct {
	PropertyName PropertyName
	Initializer  Expression
}

// MARK: - PropertyName

// PropertyName [Yield, Await] :
// - LiteralPropertyName
// - ComputedPropertyName[?Yield, ?Await]
type PropertyName interface {
	ASTNode
}
type LiteralPropertyName interface {
	PropertyName
	LiteralString() string
}
type PropertyNameLiteralIdentifier struct {
	LiteralPropertyName
	Identifier IdentifierName
}

func (p *PropertyNameLiteralIdentifier) LiteralString() string {
	return string(p.Identifier)
}

func (p *PropertyNameLiteralIdentifier) String() string {
	return string(p.Identifier)
}

func (p *PropertyNameLiteralIdentifier) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IStoreConstant{Value: NewStringValue(string(p.Identifier))})
}

func (p *PropertyNameLiteralIdentifier) Evaluation(vm *VM2) Value {
	return NewStringValue(p.Identifier)
}

type PropertyNameLiteralString struct {
	LiteralPropertyName
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
	LiteralPropertyName
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

// ComputedPropertyName [Yield, Await] :
// - [ AssignmentExpression[+In, ?Yield, ?Await] ]
type ComputedPropertyName struct {
	PropertyName
	Expression Expression
}

func (p *ComputedPropertyName) String() string {
	return p.Expression.String()
}

func (p *ComputedPropertyName) Bytecode(e *Executable, c *BytecodeContext) {
	p.Expression.Bytecode(e, c)
	if ExpressionAnalyze(p.Expression, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
}

// Evaluation
// spec: 13.2.5.4
func (p *ComputedPropertyName) Evaluation(vm *VM2) Value {
	exprValue := p.Expression.Evaluation(vm)
	propName := exprValue.GetValue(vm.agent)
	return ToPropertyKey(vm.agent, propName).ToValue()
}

// MARK: - FunctionExpression

// FunctionExpression :
//   - function BindingIdentifier[~Yield, ~Await] opt ( FormalParameters[~Yield, ~Await] )
//     { FunctionBody[~Yield, ~Await] }
type FunctionExpression struct {
	PrimaryExpression
	Identifier       IdentifierName
	FormalParameters *FormalParameters
	Body             *FunctionBody
	SourceText       string
}

var _ RuntimeSemanticsInstantiateOrdinaryFunctionExpression = (*FunctionExpression)(nil)

func (p *FunctionExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (p *FunctionExpression) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IInstantiateOrdinaryFunctionExpression{FunctionExpression: p})
}

func (p *FunctionExpression) astHasIdentifier() bool {
	return p.Identifier != ""
}

// InstantiateOrdinaryFunctionExpression
// spec: 15.2.5
func (p *FunctionExpression) InstantiateOrdinaryFunctionExpression(vm *VM2, propertyKeyOrPrivateName PropertyKeyOrPrivateName) (fun ObjectType) {
	agent := vm.agent
	realm := agent.CurrentRealm()
	if p.astHasIdentifier() {
		Assert(propertyKeyOrPrivateName == nil)
		name := p.Identifier
		outerEnv := vm.RunningLexicalEnvironment()
		funcEnv := NewDeclarativeEnvironment(outerEnv)
		funcEnv.CreateImmutableBinding(name, false)
		privateEnv := vm.RunningPrivateEnvironment()
		sourceText := p.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.FunctionPrototype,
			sourceText,
			p.FormalParameters,
			p.Body,
			functionCreateThisModeNonLexical,
			funcEnv,
			privateEnv,
		)
		SetFunctionName(closure, NewStringPropertyKey(name), "")
		MakeConstructor(closure, false, nil)

		funcEnv.InitializeBinding(name, closure.ToValue())
		return closure
	} else {
		var name PropertyKeyOrPrivateName
		if propertyKeyOrPrivateName == nil {
			name = NewStringPropertyKey("")
		} else {
			name = propertyKeyOrPrivateName
		}
		env := vm.RunningLexicalEnvironment()
		privateEnv := vm.RunningPrivateEnvironment()
		sourceText := p.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.FunctionPrototype,
			sourceText,
			p.FormalParameters,
			p.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		SetFunctionName(closure, name, "")
		MakeConstructor(closure, false, nil)
		return closure
	}
}

// NamedEvaluation
// spec: 8.4.5
func (p *FunctionExpression) NamedEvaluation(vm *VM2, name PropertyKeyOrPrivateName) Value {
	return p.InstantiateOrdinaryFunctionExpression(vm, name).ToValue()
}

// TODO: non standard
func (p *FunctionExpression) Evaluation(vm *VM2) Value {
	return p.NamedEvaluation(vm, NewStringPropertyKey(""))
}

func (p *FunctionExpression) String() string {
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

// ArrowFunction [In, Yield, Await] :
// - ArrowParameters[?Yield, ?Await] [no LineTerminator here] => ConciseBody[?In]
type ArrowFunction struct {
	PrimaryExpression
	// TODO: change to ArrowParameters
	// ArrowParameters [Yield, Await] :
	// - BindingIdentifier[?Yield, ?Await]
	// - CoverParenthesizedExpressionAndArrowParameterList[?Yield, ?Await]
	FormalParameters *FormalParameters
	// ConciseBody[In] :
	// - [lookahead ≠ {] ExpressionBody[?In, ~Await]
	// - { FunctionBody[~Yield, ~Await] }
	// ExpressionBody[In, Await] :
	// - AssignmentExpression[?In, ~Yield, ?Await]
	Body       *FunctionBody
	SourceText string
}

var _ RuntimeSemanticsInstantiateArrowFunctionExpression = (*ArrowFunction)(nil)

func (p *ArrowFunction) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (p *ArrowFunction) Bytecode(e *Executable, c *BytecodeContext) {
	strict := c.containedInStrictCode || p.Body.FunctionBodyContainsUseStrict()
	p.Body.Strict = strict
	e.AddInstruction(&IInstantiateArrowFunctionExpression{FunctionExpression: p})
}

func (p *ArrowFunction) Evaluation(vm *VM2) Value {
	return p.InstantiateArrowFunctionExpression(vm, "")
}

func (p *ArrowFunction) InstantiateArrowFunctionExpression(vm *VM2, name string) Value {
	agent := vm.agent
	realm := agent.CurrentRealm()
	env := vm.RunningLexicalEnvironment()
	privateEnv := vm.RunningPrivateEnvironment()
	sourceText := p.SourceText
	closure := OrdinaryFunctionCreate(
		agent,
		realm.Intrinsics.FunctionPrototype,
		sourceText,
		p.FormalParameters,
		p.Body,
		functionCreateThisModeLexical,
		env,
		privateEnv,
	)
	SetFunctionName(closure, CMString(name).ToPropertyKey(), "")
	return closure.ToValue()
}

func (p *ArrowFunction) String() string {
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

// TODO(BM): refactor: match spec
// MemberExpression [Yield, Await] :
// - PrimaryExpression[?Yield, ?Await]
// - MemberExpression[?Yield, ?Await] [ Expression[+In, ?Yield, ?Await] ]
// - MemberExpression[?Yield, ?Await]. IdentifierName
// - MemberExpression[?Yield, ?Await] TemplateLiteral[?Yield, ?Await, +Tagged]
// - SuperProperty[?Yield, ?Await]
// - MetaProperty
// - new MemberExpression[?Yield, ?Await] Arguments[?Yield, ?Await]
// - MemberExpression[?Yield, ?Await]. PrivateIdentifier
type MemberExpression struct {
	Expression
	Member   Expression
	Property ASTProperty
}

func (m *MemberExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
}

func (m *MemberExpression) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddDebug("MemberExpression: " + m.String())
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

// astIsMemberExpression check if Member is MemberExpression
func (m *MemberExpression) astIsPropertyExpression() bool {
	_, ok := m.Property.(*ASTPropertyExpression)
	return ok
}

// Evaluation
// spec: 13.3.2.1
func (m *MemberExpression) Evaluation(vm *VM2) Value {
	baseReference := m.Member.Evaluation(vm)
	baseValue := baseReference.GetValue(vm.agent)
	strict := vm.containedInStrictCode

	switch prop := m.Property.(type) {
	case *ASTPropertyExpression:
		// TODO: check source text is strict
		return NewReferenceRecordValue(
			vm.EvaluatePropertyAccessWithExpressionKey(baseValue, prop.Expression, strict),
		)
	case *ASTPropertyIdentifier:
		return NewReferenceRecordValue(
			vm.EvaluatePropertyAccessWithIdentifierKey(baseValue, prop.Identifier, strict),
		)
	}
	panic("unimplemented")
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

func (l *LiteralNull) Evaluation(vm *VM2) Value {
	return NullValue
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

// 13.2.3.1
func (l *LiteralUndefined) Evaluation(vm *VM2) Value {
	return UndefinedValue
}

func (l *LiteralUndefined) String() string {
	return "LiteralUndefined"
}

type LiteralBoolean struct {
	Literal
	Bool bool
}

var _ Literal = (*LiteralBoolean)(nil)

func (l *LiteralBoolean) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IStoreConstant{Value: NewBooleanValue(l.Bool)})
}

func (l *LiteralBoolean) Evaluation(vm *VM2) Value {
	return NewBooleanValue(l.Bool)
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
	Type  NumericType
}

var _ Literal = (*LiteralNumeric)(nil)

// 12.9.3.3
func (l *LiteralNumeric) NumericValue() (Value, error) {
	if l.Type == NumericTypeBigInt {
		bi := big.NewInt(0)
		bi.SetString(l.Value, 10)
		return NewBigIntValue(bi), nil
	}
	num, err := strconv.ParseFloat(l.Value, 64)
	if err != nil {
		return nil, err
	}
	return NewNumberValue(JSNumber(num)), nil
}

func (l *LiteralNumeric) Bytecode(e *Executable, c *BytecodeContext) {
	v, err := l.NumericValue()
	if err != nil {
		panic(err)
	}
	e.AddInstruction(&IStoreConstant{Value: v})
}

func (l *LiteralNumeric) Evaluation(vm *VM2) Value {
	v, err := l.NumericValue()
	if err != nil {
		panic(err)
	}
	return v
}

func (l *LiteralNumeric) String() string {
	return l.Value
}

// MARK: - LiteralString

// TODO(SM): rename to StringLiteral
type LiteralString struct {
	Literal
	Value string
}

var _ Literal = (*LiteralString)(nil)

// 12.9.4.2: SV
func (l *LiteralString) StringValue() Value {
	return NewStringValue(l.Value)
}

func (l *LiteralString) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IStoreConstant{Value: l.StringValue()})
}

// 13.2.3.1
func (l *LiteralString) Evaluation(vm *VM2) Value {
	return l.StringValue()
}

func (l *LiteralString) String() string {
	return "StringLiteral: " + l.Value
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

type expressionDefaultImpl struct {
	Expression
}

func (e *expressionDefaultImpl) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func ExpressionAnalyze(e Expression, a AnalyzeQuery) bool {
	switch ee := e.(type) {
	case PrimaryExpression:
		return PrimaryExpressionAnalyze(ee, a)
	case *ExpressionPrimary:
		return PrimaryExpressionAnalyze(ee.PrimaryExpression, a)
	}
	switch a {
	case AnalyzeQueryIsReference:
		switch e.(type) {
		case *MemberExpression, SuperProperty:
			return true
		default:
			return false
		}
	case AnalyzeQueryIsStringLiteral:
		switch e.(type) {
		default:
			return false
		}
	case AnalyzeQueryIsIdentifierReference:
		return false
	}
	panic("unreachable")
}

// MARK: - ImportCall

type ExpressionImportCall struct {
	*expressionDefaultImpl
	Expression Expression
}

func (i *ExpressionImportCall) Bytecode(e *Executable, c *BytecodeContext) {
	i.Expression.Bytecode(e, c)
	if ExpressionAnalyze(i.Expression, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)
	e.AddInstruction(&IImportCall{})
}

func (i *ExpressionImportCall) String() string {
	return "import(" + i.Expression.String() + ")"
}

// MARK: - OptionalExpression

// OptionalExpression [Yield, Await] :
// - MemberExpression[?Yield, ?Await] OptionalChain[?Yield, ?Await]
// - CallExpression[?Yield, ?Await] OptionalChain[?Yield, ?Await]
// - OptionalExpression[?Yield, ?Await] OptionalChain[?Yield, ?Await]
//
// OptionalChain[Yield, Await] :
// - ?. Arguments[?Yield, ?Await]
// - ?. [ Expression[+In, ?Yield, ?Await] ]
// - ?. IdentifierName
// - ?. TemplateLiteral[?Yield, ?Await, +Tagged]
// - ?. PrivateIdentifier
// - OptionalChain[?Yield, ?Await] Arguments[?Yield, ?Await]
// - OptionalChain[?Yield, ?Await] [ Expression[+In, ?Yield, ?Await] ]
// - OptionalChain[?Yield, ?Await]. IdentifierName
// - OptionalChain[?Yield, ?Await] TemplateLiteral[?Yield, ?Await, +Tagged]
// - OptionalChain[?Yield, ?Await]. PrivateIdentifier
type OptionalExpression struct {
	Expression
	// Property is OptionalChain
	Property *OptionalExpressionProperty
	// Expr is MemberExpression, CallExpression, or OptionalExpression
	Expr Expression
}

func (o *OptionalExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (o *OptionalExpression) Bytecode(e *Executable, c *BytecodeContext) {
	o.Expr.Bytecode(e, c)
	if o.Property.Arguments != nil {
		e.AddInstruction(InsPushReference)
	}
	if ExpressionAnalyze(o.Expr, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	e.AddInstruction(InsLoad)

	e.AddInstruction(InsLoad)
	e.AddInstruction(&ILoadConstant{
		Value: UndefinedValue,
	})
	e.AddInstruction(InsLooselyEqual)

	jumpIfTrue := &IJumpIfTrue{}
	e.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(e.Instructions) - 1
	e.AddInstruction(InsStore)
	e.AddInstruction(&IStoreConstant{
		Value: UndefinedValue,
	})
	endJump := &IJump{}

	jumpIfTrue.TargetElse = len(e.Instructions) - 1
	strict := c.containedInStrictCode

	if o.Property.Arguments != nil {
		e.AddInstruction(InsLoadThisValue)
		for _, arg := range o.Property.Arguments {
			arg.Bytecode(e, c)
			if ExpressionAnalyze(arg, AnalyzeQueryIsReference) {
				e.AddInstruction(InsGetValue)
			}
			e.AddInstruction(InsLoad)
		}

		e.AddInstructionDebug(&ICall{
			ArgumentCount: len(o.Property.Arguments),
			Strict:        strict,
		}, "OptionalExpression "+o.String())
	} else if o.Property.Expression != nil {
		expr := o.Property.Expression
		expr.Bytecode(e, c)
		if ExpressionAnalyze(expr, AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(InsLoad)
		e.AddInstruction(&IEvaluatePropertyAccessWithExpressionKey{
			Strict: strict,
		})
	} else if o.Property.Identifier != "" {
		e.AddInstruction(&IEvaluatePropertyAccessWithIdentifierKey{
			Strict: strict,
			Name:   o.Property.Identifier,
		})
	}

	if o.Property.Arguments == nil {
		e.AddInstruction(InsGetValue)
	}

	endJump.Target = len(e.Instructions) - 1
}

func (o *OptionalExpression) Evaluation(vm *VM2) Value {
	baseReference := o.Expr.Evaluation(vm)
	baseValue := baseReference.GetValue(vm.agent)
	if IsUndefinedOrNull(baseValue) {
		return UndefinedValue
	}
	result, err := o.Property.ChainEvaluation(vm, baseValue, baseReference)
	if err != nil {
		vm.panic(err)
	}
	return result
}

func (o *OptionalExpression) String() string {
	return o.Expr.String() + "?." + o.Property.String()
}

// OptionalExpressionProperty Enum
// TODO(BM): rename to OptionalChain
// TODO: handle chaining
// TODO: handle private identifier
type OptionalExpressionProperty struct {
	Arguments  Arguments
	Expression Expression
	Identifier IdentifierName
}

var _ RuntimeSemanticsChainEvaluation = (*OptionalExpressionProperty)(nil)

func (o *OptionalExpressionProperty) astIsArguments() bool {
	return o.Arguments != nil
}

func (o *OptionalExpressionProperty) astIsExpression() bool {
	return o.Expression != nil
}

func (o *OptionalExpressionProperty) astIsIdentifier() bool {
	return o.Identifier != ""
}

func (o *OptionalExpressionProperty) isStrict() bool {
	// TODO: check strict
	return true
}

func (o *OptionalExpressionProperty) ChainEvaluation(vm *VM2, baseValue Value, baseReference Value) (value Value, err Value) {
	if o.astIsArguments() {
		thisChain := o
		// TODO
		tailCall := false
		arguments := thisChain.Arguments.Evaluation(vm)
		return vm.EvaluateCall(baseValue, baseReference, arguments.(*ListValue).Values, tailCall), nil
	} else if o.astIsExpression() {
		strict := o.isStrict()
		return NewReferenceRecordValue(
			vm.EvaluatePropertyAccessWithExpressionKey(baseValue, o.Expression, strict),
		), nil
	} else if o.astIsIdentifier() {
		strict := o.isStrict()
		return NewReferenceRecordValue(
			vm.EvaluatePropertyAccessWithIdentifierKey(baseValue, o.Identifier, strict),
		), nil
	} else {
		panic("unimplemented")
	}
}

func (o *OptionalExpressionProperty) String() string {
	if o.Arguments != nil {
		return "(" + o.Arguments.String() + ")"
	}
	if o.Expression != nil {
		return "[" + o.Expression.String() + "]"
	}
	return string(o.Identifier)
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

type MetaPropertyImportMeta struct {
	MetaProperty
}

func (m *MetaPropertyImportMeta) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (m *MetaPropertyImportMeta) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(&IGetOrCreateImportMeta{})
}

func (m *MetaPropertyImportMeta) String() string {
	return "import.meta"
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

// SuperCall [Yield, Await] :
// - super Arguments[?Yield, ?Await]
type ExpressionSuperCall struct {
	Expression
	Arguments Arguments
}

func (e *ExpressionSuperCall) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

// Evaluation
// spec: 13.3.7.1
func (e *ExpressionSuperCall) Evaluation(vm *VM2) Value {
	agent := vm.agent
	newTarget := agent.GetNewTarget()
	// alias of func
	f := agent.GetSuperConstructor()
	argList := e.Arguments.ArgumentListEvaluation(vm)
	if !IsConstructor(f) {
		return agent.ThrowTypeError("SuperCall: not a constructor")
	}
	result := MustGetObject(f).Construct(argList, newTarget)
	thisER := agent.GetThisEnvironment().(*FunctionEnvironment)
	thisER.BindThisValue(result.ToValue())
	F := thisER.FunctionObject
	result.InitializeInstanceElements(F)
	return result.ToValue()
}

func (e *ExpressionSuperCall) Bytecode(ex *Executable, c *BytecodeContext) {
	e.Arguments.Bytecode(ex, c)
	ex.AddInstruction(&IEvaluateSuperCall{
		ArgumentCount: len(e.Arguments),
	})
}

func (e *ExpressionSuperCall) String() string {
	return "super(" + e.Arguments.String() + ")"
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
	case AnalyzeQueryIsReference, AnalyzeQueryIsIdentifierReference:
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

// TODO(BM): rename
type PrimaryExpressionTemplateLiteral struct {
	PrimaryExpression
	TemplateLiteral *TemplateLiteral
	SourceText      string
}

// 13.2.8.6
func (t *PrimaryExpressionTemplateLiteral) Bytecode(e *Executable, c *BytecodeContext) {
	hasHead := t.TemplateLiteral.TemplateHead != nil
	if hasHead {
		e.AddInstruction(&ILoadConstant{
			Value: t.TemplateLiteral.TemplateHead.TV().ToValue(),
		})
	} else {
		// NoSubstitutionTemplate
		span := t.TemplateLiteral.Spans[0]
		e.AddInstruction(&IStoreConstant{
			Value: span.TV().ToValue(),
		})
		return
	}
	for i, span := range t.TemplateLiteral.Spans {
		// --- Expression
		if span.Expression != nil {
			span.Expression.Bytecode(e, c)
			if ExpressionAnalyze(span.Expression, AnalyzeQueryIsReference) {
				e.AddInstruction(InsGetValue)
			}
			// TODO: to string
			e.AddInstruction(InsLoad)
			e.AddInstruction(&IApplyStringOrNumericBinaryOperator{Operator: BinaryOperatorAddition})
			e.AddInstruction(InsLoad)
		}

		// --- Text
		e.AddInstruction(&ILoadConstant{
			Value: span.TV().ToValue(),
		})
		e.AddInstruction(&IApplyStringOrNumericBinaryOperator{Operator: BinaryOperatorAddition})
		if i < len(t.TemplateLiteral.Spans)-1 {
			e.AddInstruction(InsLoad)
		}
	}
}

func (t *PrimaryExpressionTemplateLiteral) astNoSubstitution() bool {
	return t.TemplateLiteral.TemplateHead == nil
}

// Evaluation
// spec: 13.2.8.6
func (t *PrimaryExpressionTemplateLiteral) Evaluation(vm *VM2) Value {
	if t.astNoSubstitution() {
		// NoSubstitutionTemplate
		span := t.TemplateLiteral.Spans[0]
		return span.TV().ToValue()
	} else {
		// TODO(BM): after remove old vm
	}
	panic("unimplemented")
}

func (t *PrimaryExpressionTemplateLiteral) String() string {
	return t.SourceText
}

type TemplateSpan struct {
	Text       string
	Expression Expression
}

// 12.9.6.1
func (t *TemplateSpan) TV() CMString {
	start := 0
	end := len(t.Text)
	if t.Text[0] == '`' || t.Text[0] == '}' {
		start = 1
	}
	if strings.HasSuffix(t.Text, "${") {
		end -= 2
	}
	if strings.HasSuffix(t.Text, "}") || strings.HasSuffix(t.Text, "`") {
		end--
	}
	if start > end {
		return CMString("")
	}
	return CMString(t.Text[start:end])
}

type TemplateLiteral struct {
	// TemplateHead is string only
	TemplateHead *TemplateSpan
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

// UpdateExpression [Yield, Await] :
// LeftHandSideExpression[?Yield, ?Await]
// LeftHandSideExpression[?Yield, ?Await] [no LineTerminator here] ++
// LeftHandSideExpression[?Yield, ?Await] [no LineTerminator here]--
// ++ UnaryExpression[?Yield, ?Await]
// -- UnaryExpression[?Yield, ?Await]
type UpdateExpression struct {
	Expression
	Type     UpdateExpressionType
	Operator UpdateOperator
	Operand  Expression
}

func (e *UpdateExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeSimple
}

func (e *UpdateExpression) Bytecode(ex *Executable, c *BytecodeContext) {
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

func (e *UpdateExpression) isPrefix() bool {
	return e.Type == UpdateExpressionTypePrefix
}

// Evaluation
// postfix: 13.4.2.1/13.4.3.1
// prefix: 13.4.4.1/13.4.5.1
func (e *UpdateExpression) Evaluation(vm *VM2) Value {
	expr := e.Operand.Evaluation(vm)
	oldValue := ToNumeric(vm.agent, expr.GetValue(vm.agent))
	var newValue Value
	if n, bi, ok := oldValue.NumberOrBigInt(); ok {
		if n != nil {
			one := NewNumberValue(1)
			if e.Operator == UpdateOperatorIncrement {
				newValue = n.Add(one)
			} else {
				newValue = n.Subtract(one)
			}
		} else {
			one := NewBigIntValue(big.NewInt(1))
			if e.Operator == UpdateOperatorIncrement {
				newValue = bi.Add(one)
			} else {
				newValue = bi.Subtract(one)
			}
		}
	} else {
		Assert(false)
	}
	if r, ok := expr.ReferenceRecord(); ok {
		r.PutValue(vm.agent, newValue)
	} else {
		panic("unreachable")
	}
	if e.isPrefix() {
		return newValue
	}
	return oldValue
}

func (e *UpdateExpression) String() string {
	if e.Type == UpdateExpressionTypePrefix {
		return e.Operator.String() + e.Operand.String()
	}
	return e.Operand.String() + e.Operator.String()
}

// MARK: - AssignmentExpression

// TODO: standardlize
// AssignmentOperator : one of
// *= /= %= += -= <<= >>= >>>= &= ^= |= **=
type AssignmentOperator int

func (a AssignmentOperator) ToBinaryOperator() BinaryOperator {
	switch a {
	case AssignmentOperatorAddition:
		return BinaryOperatorAddition
	case AssignmentOperatorSubtraction:
		return BinaryOperatorSubtraction
	case AssignmentOperatorMultiplication:
		return BinaryOperatorMultiplication
	case AssignmentOperatorDivision:
		return BinaryOperatorDivision
	case AssignmentOperatorRemainder:
		return BinaryOperatorRemainder
	case AssignmentOperatorLeftShift:
		return BinaryOperatorLeftShift
	case AssignmentOperatorRightShift:
		return BinaryOperatorRightShift
	case AssignmentOperatorUnsignedRightShift:
		return BinaryOperatorUnsignedRightShift
	case AssignmentOperatorBitwiseAnd:
		return BinaryOperatorBitwiseAnd
	case AssignmentOperatorBitwiseXor:
		return BinaryOperatorBitwiseXor
	case AssignmentOperatorBitwiseOr:
		return BinaryOperatorBitwiseOr
	case AssignmentOperatorExponentiation:
		return BinaryOperatorExponentiation
	}
	panic("unreachable")
}

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

// AssignmentExpression [In, Yield, Await] :
// ConditionalExpression[?In, ?Yield, ?Await]
// [+Yield] YieldExpression[?In, ?Await]
// ArrowFunction[?In, ?Yield, ?Await]
// AsyncArrowFunction[?In, ?Yield, ?Await]
// LeftHandSideExpression[?Yield, ?Await] = AssignmentExpression[?In, ?Yield, ?Await]
// LeftHandSideExpression[?Yield, ?Await] AssignmentOperator
// AssignmentExpression[?In, ?Yield, ?Await]
// LeftHandSideExpression[?Yield, ?Await] &&=
// AssignmentExpression[?In, ?Yield, ?Await]
// LeftHandSideExpression[?Yield, ?Await] ||=
// AssignmentExpression[?In, ?Yield, ?Await]
// LeftHandSideExpression[?Yield, ?Await] ??=
// AssignmentExpression[?In, ?Yield, ?Await]
type AssignmentExpression struct {
	Expression
	Left     Expression
	Operator AssignmentOperator
	Right    Expression
}

func (e *AssignmentExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (e *AssignmentExpression) Bytecode(ex *Executable, c *BytecodeContext) {
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

		operatorMap := map[AssignmentOperator]BinaryOperator{
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
		op := operatorMap[e.Operator]
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

func (e *AssignmentExpression) astIsAssign() bool {
	return e.Operator == AssignmentOperatorAssign
}

// Evaluation
// spec: 13.15.2
func (e *AssignmentExpression) Evaluation(vm *VM2) Value {
	if e.astIsAssign() {
		_, isObjectLiteral := e.Left.(*PrimaryExpressionObjectLiteral)
		_, isArrayLiteral := e.Left.(*ArrayLiteral)
		if !isObjectLiteral && !isArrayLiteral {
			lref := e.Left.Evaluation(vm)
			var rval Value
			if IsAnonymousFunctionDefinition(e.Right) {
				// FIXME: handle named evaluation
				panic("")
			} else {
				rref := e.Right.Evaluation(vm)
				rval = rref.GetValue(vm.agent)
			}
			if ref, ok := lref.ReferenceRecord(); ok {
				ref.PutValue(vm.agent, rval)
			} else {
				panic("unreachable")
			}
			return rval
		} else {
			// TODO(BM): DestructuringAssignmentEvaluation
			// assignmentPattern := e.Left
			// rref := e.Right.Evaluation(vm)
			// rval := rref.GetValue(vm.agent)
			panic("unimplemented")
		}
	} else {
		// TODO: handle &&= ||=, ??=
		lref := e.Left.Evaluation(vm)
		lval := lref.GetValue(vm.agent)
		rref := e.Right.Evaluation(vm)
		rval := rref.GetValue(vm.agent)
		r := ApplyStringOrNumericBinaryOperator(vm.agent, lval, rval, e.Operator.ToBinaryOperator())

		if ref, ok := lref.ReferenceRecord(); ok {
			ref.PutValue(vm.agent, r)
		} else {
			panic("unreachable")
		}
		return r
	}
}

func (e *AssignmentExpression) String() string {
	return e.Left.String() + " " + e.Operator.String() + " " + e.Right.String()
}

// MARK: - NewExpression

// NewExpression [Yield, Await] :
// MemberExpression[?Yield, ?Await]
// new NewExpression[?Yield, ?Await]
type NewExpression struct {
	Expression
	Callee    Expression
	Arguments Arguments
}

func (e *NewExpression) AssignmentTargetType() AssignmentTargetType {
	return AssignmentTargetTypeInvalid
}

func (e *NewExpression) Bytecode(ex *Executable, c *BytecodeContext) {
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

// 13.3.5.1
func (e *NewExpression) Evaluation(vm *VM2) Value {
	return e.EvaluateNew(vm)
}

// 13.3.5.1.1
func (e *NewExpression) EvaluateNew(vm *VM2) Value {
	ref := e.Callee.Evaluation(vm)
	constructor := ref.GetValue(vm.agent)
	var argList []Value
	if len(e.Arguments) == 0 {
		argList = []Value{}
	} else {
		argList = e.Arguments.ArgumentListEvaluation(vm)
	}
	if !IsConstructor(constructor) {
		return vm.agent.ThrowTypeError("constructor is not a constructor")
	}
	o := MustGetObject(constructor)
	return o.Construct(argList, nil).ToValue()
}

func (e *NewExpression) String() string {
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

// 13.15.4
func (b *ExpressionBinaryExpression) EvaluateStringOrNumericBinaryExpression(vm *VM2) Value {
	lref := b.Left.Evaluation(vm)
	lval := lref.GetValue(vm.agent)

	rref := b.Right.Evaluation(vm)
	rval := rref.GetValue(vm.agent)
	return vm.ApplyStringOrNumericBinaryOperator(lval, rval, b.Operator)
}

func (b *ExpressionBinaryExpression) Evaluation(vm *VM2) Value {
	return b.EvaluateStringOrNumericBinaryExpression(vm)
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

// ConditionalExpression [In, Yield, Await] :
//   - ShortCircuitExpression[?In, ?Yield, ?Await]
//   - ShortCircuitExpression[?In, ?Yield, ?Await] ?
//     AssignmentExpression[+In, ?Yield, ?Await] :
//     AssignmentExpression[?In, ?Yield, ?Await]
type ConditionalExpression struct {
	Expression
	Test       Expression
	Consequent Expression
	Alternate  Expression
}

func (e *ConditionalExpression) Bytecode(ex *Executable, c *BytecodeContext) {
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

	jumpIfTrue.TargetElse = len(ex.Instructions) - 1
	e.Alternate.Bytecode(ex, c)

	if ExpressionAnalyze(e.Alternate, AnalyzeQueryIsReference) {
		ex.AddInstruction(InsGetValue)
	}

	jump.Target = len(ex.Instructions) - 1
}

// Evaluation
// spec: 13.14.1
func (e *ConditionalExpression) Evaluation(vm *VM2) Value {
	agent := vm.agent
	lref := e.Test.Evaluation(vm)
	lval := lref.GetValue(agent)
	if lval.ToBoolean() {
		trueRef := e.Consequent.Evaluation(vm)
		return trueRef.GetValue(agent)
	} else {
		falseRef := e.Alternate.Evaluation(vm)
		return falseRef.GetValue(agent)
	}
}

func (e *ConditionalExpression) String() string {
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

// LogicalANDExpression[In, Yield, Await] :
//   - BitwiseORExpression[?In, ?Yield, ?Await]
//   - LogicalANDExpression[?In, ?Yield, ?Await] &&
//     BitwiseORExpression[?In, ?Yield, ?Await]
//
// LogicalORExpression[In, Yield, Await] :
//   - LogicalANDExpression[?In, ?Yield, ?Await]
//   - LogicalORExpression[?In, ?Yield, ?Await] ||
//     LogicalANDExpression[?In, ?Yield, ?Await]
//
// CoalesceExpression[In, Yield, Await] :
//   - CoalesceExpressionHead[?In, ?Yield, ?Await] ??
//     BitwiseORExpression[?In, ?Yield, ?Await]
//
// CoalesceExpressionHead[In, Yield, Await] :
// - CoalesceExpression[?In, ?Yield, ?Await]
// - BitwiseORExpression[?In, ?Yield, ?Await]
// ShortCircuitExpression[In, Yield, Await] :
// - LogicalORExpression[?In, ?Yield, ?Await]
// - CoalesceExpression[?In, ?Yield, ?Await]
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

		jumpIfTrue.Target = len(ex.Instructions) - 1
		e.Right.Bytecode(ex, c)

		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		jumpIfTrue.TargetElse = len(ex.Instructions) - 1
	case LogicalOperatorOr:
		jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.TargetElse = len(ex.Instructions) - 1
		e.Right.Bytecode(ex, c)
		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		jumpIfTrue.Target = len(ex.Instructions) - 1
	case LogicalOperatorNullishCoalescing:
		ex.AddInstruction(InsLoad)

		ex.AddInstruction(InsLoad)
		ex.AddInstruction(&ILoadConstant{
			Value: UndefinedValue,
		})
		ex.AddInstruction(InsLooselyEqual)

		jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
		ex.AddInstruction(jumpIfTrue)

		jumpIfTrue.Target = len(ex.Instructions) - 1
		ex.AddInstruction(InsStore)

		e.Right.Bytecode(ex, c)
		if ExpressionAnalyze(e.Right, AnalyzeQueryIsReference) {
			ex.AddInstruction(InsGetValue)
		}

		jump := &IJump{Target: 0}
		ex.AddInstruction(jump)

		jumpIfTrue.TargetElse = len(ex.Instructions) - 1
		ex.AddInstruction(InsStore)

		jump.Target = len(ex.Instructions) - 1
	}
}

func (e *ExpressionLogicalExpression) astIsOr() bool {
	return e.Operator == LogicalOperatorOr
}

func (e *ExpressionLogicalExpression) astNullishCoalescing() bool {
	return e.Operator == LogicalOperatorNullishCoalescing
}

// Evaluation
// spec: 13.13.1
func (e *ExpressionLogicalExpression) Evaluation(vm *VM2) Value {
	if e.astIsOr() {
		lref := e.Left.Evaluation(vm)
		lval := lref.GetValue(vm.agent)
		lbool := lval.ToBoolean()
		if lbool {
			return lval
		}
		rref := e.Right.Evaluation(vm)
		return rref.GetValue(vm.agent)
	} else if e.astNullishCoalescing() {
		lref := e.Left.Evaluation(vm)
		lval := lref.GetValue(vm.agent)
		if IsUndefinedOrNil(lval) || lval == NullValue {
			rref := e.Right.Evaluation(vm)
			return rref.GetValue(vm.agent)
		} else {
			return lval
		}
	} else {
		panic("unimplemented")
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

// EqualityExpression [In, Yield, Await] :
// RelationalExpression[?In, ?Yield, ?Await]
// EqualityExpression[?In, ?Yield, ?Await] == RelationalExpression[?In, ?Yield, ?Await]
// EqualityExpression[?In, ?Yield, ?Await] != RelationalExpression[?In, ?Yield, ?Await]
// EqualityExpression[?In, ?Yield, ?Await] ===
// RelationalExpression[?In, ?Yield, ?Await]
// EqualityExpression[?In, ?Yield, ?Await] !==
// RelationalExpression[?In, ?Yield, ?Await]
type EqualityExpression struct {
	Expression
	// EqualityExpression
	Left     Expression
	Operator EqualityOperator
	// RelationalExpression
	Right Expression
}

func (e *EqualityExpression) Bytecode(ex *Executable, c *BytecodeContext) {
	ex.AddDebug("Equality: " + e.String())
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

// 13.11.1
func (e *EqualityExpression) Evaluation(vm *VM2) Value {
	lref := e.Left.Evaluation(vm)
	lval := lref.GetValue(vm.agent)

	rref := e.Right.Evaluation(vm)
	rval := rref.GetValue(vm.agent)

	switch e.Operator {
	case EqualityOperatorEqual:
		return NewBooleanValue(IsLooselyEqual(vm.agent, lval, rval))
	case EqualityOperatorNotEqual:
		return NewBooleanValue(!IsLooselyEqual(vm.agent, lval, rval))
	case EqualityOperatorStrictEqual:
		return NewBooleanValue(IsStrictlyEqual(lval, rval))
	case EqualityOperatorStrictNotEqual:
		return NewBooleanValue(!IsStrictlyEqual(lval, rval))
	}
	panic("unreachable")
}

func (e *EqualityExpression) String() string {
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

// RelationalExpression [In, Yield, Await] :
// ShiftExpression[?Yield, ?Await]
// RelationalExpression[?In, ?Yield, ?Await] < ShiftExpression[?Yield, ?Await]
// RelationalExpression[?In, ?Yield, ?Await] > ShiftExpression[?Yield, ?Await]
// RelationalExpression[?In, ?Yield, ?Await] <= ShiftExpression[?Yield, ?Await]
// RelationalExpression[?In, ?Yield, ?Await] >= ShiftExpression[?Yield, ?Await]
// RelationalExpression[?In, ?Yield, ?Await] instanceof
// ShiftExpression[?Yield, ?Await]
// [+In] RelationalExpression[+In, ?Yield, ?Await] in ShiftExpression[?Yield, ?Await]
// [+In] PrivateIdentifier in ShiftExpression[?Yield, ?Await]
type RelationalExpression struct {
	Expression
	Left     Expression
	Operator RelationalOperator
	Right    Expression
}

func (e *RelationalExpression) Bytecode(ex *Executable, c *BytecodeContext) {
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

// Evaluation
// spec: 13.10.1
func (e *RelationalExpression) Evaluation(vm *VM2) Value {
	lref := e.Left.Evaluation(vm)
	lval := lref.GetValue(vm.agent)

	rref := e.Right.Evaluation(vm)
	rval := rref.GetValue(vm.agent)

	switch e.Operator {
	case RelationalOperatorLessThan, RelationalOperatorGreaterThan:
		var order isLessThanOrder
		if e.Operator == RelationalOperatorLessThan {
			order = IsLessThanOrderLeftFirst
		} else {
			order = IsLessThanOrderRightFirst
		}
		return NewBooleanValue(IsLessThan(vm.agent, lval, rval, order))
	case RelationalOperatorLessThanOrEqual, RelationalOperatorGreaterThanOrEqual:
		var order isLessThanOrder
		if e.Operator == RelationalOperatorLessThanOrEqual {
			order = IsLessThanOrderRightFirst
		} else {
			order = IsLessThanOrderLeftFirst
		}
		return NewBooleanValue(!IsLessThan(vm.agent, rval, lval, order))
	case RelationalOperatorInstanceof:
		return NewBooleanValue(InstanceOfOperator(vm.agent, lval, rval))
	case RelationalOperatorIn:
		panic("unimplemented")
	}
	panic("unreachable")
}

func (e *RelationalExpression) String() string {
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

// UnaryExpression [Yield, Await] :
// - UpdateExpression[?Yield, ?Await]
// - delete UnaryExpression[?Yield, ?Await]
// - void UnaryExpression[?Yield, ?Await]
// - typeof UnaryExpression[?Yield, ?Await]
// - + UnaryExpression[?Yield, ?Await]
// - - UnaryExpression[?Yield, ?Await]
// - ~ UnaryExpression[?Yield, ?Await]
// - ! UnaryExpression[?Yield, ?Await]
// - [+Await] AwaitExpression[?Yield]
type UnaryExpression struct {
	Expression
	Operator UnaryOperator
	Operand  Expression
}

func (u *UnaryExpression) Bytecode(e *Executable, c *BytecodeContext) {
	switch u.Operator {
	case UnaryOperatorDelete:
		u.Operand.Bytecode(e, c)
		if !ExpressionAnalyze(u.Operand, AnalyzeQueryIsReference) {
			e.AddInstruction(&IStoreConstant{
				Value: TrueValue,
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
		if ExpressionAnalyze(u.Operand, AnalyzeQueryIsIdentifierReference) {
			identifier := u.Operand.(*PrimaryExpressionIdentifierReference).Identifier
			e.AddInstruction(&ITypeOfIdentifier{
				IdentifierName: string(identifier),
			})
		} else {
			u.Operand.Bytecode(e, c)
			if ExpressionAnalyze(u.Operand, AnalyzeQueryIsReference) {
				e.AddInstruction(InsGetValue)
			}
			e.AddInstruction(InsTypeof)
		}
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

func (u *UnaryExpression) astIsAdd() bool {
	return u.Operator == UnaryOperatorAddition
}

func (u *UnaryExpression) astIsSubtract() bool {
	return u.Operator == UnaryOperatorSubtraction
}

func (u *UnaryExpression) astIsLogicalNot() bool {
	return u.Operator == UnaryOperatorLogicalNot
}

func (u *UnaryExpression) astIsBitwiseNot() bool {
	return u.Operator == UnaryOperatorBitwiseNot
}

func (u *UnaryExpression) astIsTypeof() bool {
	return u.Operator == UnaryOperatorTypeof
}

func (u *UnaryExpression) astIsDelete() bool {
	return u.Operator == UnaryOperatorDelete
}

func (u *UnaryExpression) astIsVoid() bool {
	return u.Operator == UnaryOperatorVoid
}

// TODO: use Number::[op], BigInt::[op]
// spec: 13.5.3.1, 13.5.4.1, 13.5.5.1, 13.5.6.1, 13.5.7.1
func (u *UnaryExpression) Evaluation(vm *VM2) Value {
	agent := vm.agent
	switch {
	case u.astIsTypeof():
		// 13.5.3.1
		val := u.Operand.Evaluation(vm)
		if r, ok := val.ReferenceRecord(); ok {
			if r.IsUnresolvableReference() {
				return NewStringValue("undefined")
			}
		}
		val = val.GetValue(agent)
		// TODO: B.3.6.3 isHTMLDDA
		return NewStringValue(val.TypeString())
	case u.astIsAdd():
		// 13.5.4.1
		expr := u.Operand.Evaluation(vm)
		return ToNumber(agent, expr.GetValue(agent))
	case u.astIsSubtract():
		// 13.5.5.1
		expr := u.Operand.Evaluation(vm)
		oldValue := ToNumeric(agent, expr.GetValue(agent))
		if n, b, ok := oldValue.NumberOrBigInt(); ok {
			if n != nil {
				return NewNumberValue(-n.Data)
			} else {
				return NewBigIntValue(b.Data.Neg(nil))
			}
		} else {
			Assert(false)
		}
	case u.astIsBitwiseNot():
		// 13.5.6.1
		expr := u.Operand.Evaluation(vm)
		oldValue := ToNumeric(agent, expr.GetValue(agent))
		if n, b, ok := oldValue.NumberOrBigInt(); ok {
			if n != nil {
				return NewNumberValue(JSNumber(^int64(n.Data)))
			} else {
				return NewBigIntValue(b.Data.Not(nil))
			}
		} else {
			Assert(false)
		}
	case u.astIsLogicalNot():
		// 13.5.7.1
		expr := u.Operand.Evaluation(vm)
		oldValue := expr.GetValue(agent).ToBoolean()
		if oldValue {
			return FalseValue
		} else {
			return TrueValue
		}
	case u.astIsTypeof():
		// 13.5.3.1
		val := u.Operand.Evaluation(vm)
		if ref, ok := val.ReferenceRecord(); ok {
			if ref.IsUnresolvableReference() {
				return NewStringValue("undefined")
			}
		}
		val = val.GetValue(agent)
		return NewStringValue(val.TypeString())
	case u.astIsVoid():
		// 13.5.2.1
		expr := u.Operand.Evaluation(vm)
		expr.GetValue(agent)
		return UndefinedValue
	case u.astIsDelete():
		// 13.5.1.2
		refValue := u.Operand.Evaluation(vm)
		ref, ok := refValue.ReferenceRecord()
		if !ok {
			return TrueValue
		}
		if ref.IsUnresolvableReference() {
			Assert(!ref.Strict)
			return TrueValue
		}
		if ref.IsPropertyReference() {
			Assert(!ref.IsPrivateReference())
			if ref.IsSuperReference() {
				vm.panic(agent.ThrowException(ReferenceError, "cannot delete super property"))
			}
			v, _ := ref.Base.Value()
			baseObj := v.ToObject(agent)
			var referencedName PropertyKey
			if ref.ReferencedName.PrivateName != nil {
				panic("unreachable")
			} else if ref.ReferencedName.Symbol != nil {
				referencedName = NewSymbolPropertyKey(ref.ReferencedName.Symbol)
			} else {
				referencedName = NewStringPropertyKey(ref.ReferencedName.String)
			}
			deleteStatus := baseObj.InternalMethods().Delete(baseObj, referencedName)
			if !deleteStatus && ref.Strict {
				vm.panic(agent.ThrowTypeError("cannot delete property"))
			}
			return NewBooleanValue(deleteStatus)
		} else {
			base, ok := ref.Base.Env()
			Assert(ok)
			return NewBooleanValue(base.DeleteBinding(ref.ReferencedName.String))
		}
	}
	panic("unimplemented")
}

func (u *UnaryExpression) String() string {
	return "UnaryExpression " + u.Operator.String() + " " + u.Operand.String()
}

// MARK: - CallExpression

// TODO: handle spread expression
// ArgumentList[Yield, Await] :
// AssignmentExpression[+In, ?Yield, ?Await]
// ... AssignmentExpression[+In, ?Yield, ?Await]
// ArgumentList[?Yield, ?Await] , AssignmentExpression[+In, ?Yield, ?Await]
// ArgumentList[?Yield, ?Await] , ... AssignmentExpression[+In, ?Yield, ?Await]
type Arguments []Expression

var _ RuntimeSemanticsArgumentListEvaluation = Arguments{}

func (a Arguments) astIsAssignmentExpression() bool {
	return len(a) == 1
}

// TODO
func (a Arguments) astIsSpreadElement() bool {
	return false
}

func (a Arguments) isEmpty() bool {
	return len(a) == 0
}

func (a Arguments) ArgumentListEvaluation(vm *VM2) []Value {
	if a.isEmpty() {
		return []Value{}
	}
	if a.astIsAssignmentExpression() {
		// TODO: handle spread
		if a.astIsSpreadElement() {
			panic("unimplemented")
		}
		ref := a[0].Evaluation(vm)
		arg := ref.GetValue(vm.agent)
		return []Value{arg}
	} else {
		// TODO: handle spread
		last := a[len(a)-1].Evaluation(vm)
		return append(a[:len(a)-1].ArgumentListEvaluation(vm), last.GetValue(vm.agent))
	}
}

func (a Arguments) String() string {
	var sb string
	for i, arg := range a {
		if i != 0 {
			sb += ", "
		}
		sb += arg.String()
	}
	return sb
}

// 13.3.8.1 ArgumentListEvaluation
func (a Arguments) Bytecode(e *Executable, c *BytecodeContext) {
	for _, arg := range a {
		arg.Bytecode(e, c)
		if ExpressionAnalyze(arg, AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(InsLoad)
		// TODO: Spread
	}
}

// TODO(WIP): Spread, Template
// Evaluation 13.3.8.1
// ArgumentListEvaluation
// TODO(XXX): can we split it out based on different types of arguments?
func (a Arguments) Evaluation(vm *VM2) Value {
	// Assigment
	if len(a) == 0 {
		return NewListValue([]Value{})
	} else if len(a) == 1 {
		ref := a[0].Evaluation(vm)
		arg := ref.GetValue(vm.agent)
		return NewListValue([]Value{arg})
	} else if len(a) > 1 {
		precedingArgs := a[:len(a)-1].Evaluation(vm)
		ref := a[len(a)-1].Evaluation(vm)
		arg := ref.GetValue(vm.agent)
		return NewListValue(append(precedingArgs.(*ListValue).Values, arg))
	} else {
		panic("unimplemented")
	}
}

// CallExpression [Yield, Await] :
// - CoverCallExpressionAndAsyncArrowHead[?Yield, ?Await]
// - SuperCall[?Yield, ?Await]
// - ImportCall[?Yield, ?Await]
// - CallExpression[?Yield, ?Await] Arguments[?Yield, ?Await]
// - CallExpression[?Yield, ?Await] [ Expression[+In, ?Yield, ?Await] ]
// - CallExpression[?Yield, ?Await]. IdentifierName
// - CallExpression[?Yield, ?Await] TemplateLiteral[?Yield, ?Await, +Tagged]
// - CallExpression[?Yield, ?Await]. PrivateIdentifier
// CoverCallExpressionAndAsyncArrowHead[Yield, Await] :
// - MemberExpression[?Yield, ?Await] Arguments[?Yield, ?Await]
type CallExpression struct {
	Expression
	Callee    Expression
	Arguments Arguments
}

var _ Expression = (*CallExpression)(nil)

// astIsCover matches CoverCallExpressionAndAsyncArrowHead
func (c *CallExpression) astIsCover() bool {
	_, ok := c.Callee.(*MemberExpression)
	return ok
}

// astIsFunctionCall matches CallExpression[?Yield, ?Await] Arguments[?Yield, ?Await]
func (c *CallExpression) astIsFunctionCall() bool {
	return true
}

// 13.3.6.1
func (c *CallExpression) Bytecode(e *Executable, bc *BytecodeContext) {
	c.Callee.Bytecode(e, bc)

	e.AddInstruction(&IPushReference{})
	if ExpressionAnalyze(c.Callee, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}

	e.AddInstruction(InsLoad)
	e.AddInstruction(InsLoadThisValue)
	c.Arguments.Bytecode(e, bc)

	strict := bc.containedInStrictCode

	e.AddInstructionDebug(
		&ICall{
			ArgumentCount: len(c.Arguments),
			Strict:        strict,
		},
		"CallExpression: "+c.String(),
	)

	e.AddInstruction(&IPopReference{})
}

// TODO(SM): WIP
// Evaluation 13.3.6.1
func (c *CallExpression) Evaluation(vm *VM2) Value {
	if c.astIsCover() {
		return c.coverCallExpressionAndAsyncArrowHead(vm)
	}
	if !c.astIsFunctionCall() {
		panic("unimplemented")
	}
	ref := c.Callee.Evaluation(vm)
	// rename from func
	f := ref.GetValue(vm.agent)

	// TODO: this ref
	// i.This(c)
	// i.Let("thisCall")

	// TODO: IsInTailPosition
	// i.IsInTailPosition()
	// i.Let("tailPosition")

	arguments := c.Arguments.Evaluation(vm)
	return vm.EvaluateCall(f, ref, arguments.(*ListValue).Values, false)
}

// coverCallExpressionAndAsyncArrowHead matches CallExpression : CoverCallExpressionAndAsyncArrowHead
// spec: 13.3.6.1
func (c *CallExpression) coverCallExpressionAndAsyncArrowHead(vm *VM2) Value {
	callee := c.Callee.(*MemberExpression)
	expr := callee
	memberExpr := expr
	// TODO(XXX): is pass Arguments to evaluate directly?
	arguments := c.Arguments.Evaluation(vm)
	ref := memberExpr.Evaluation(vm)
	f := ref.GetValue(vm.agent)
	if r, ok := ref.ReferenceRecord(); ok {
		if !r.IsPropertyReference() && r.ReferencedName.String == "eval" {
			panic("unimplemented")
		}
	}

	// TODO: check is tailPosition
	tailCall := false

	return vm.EvaluateCall(f, ref, arguments.(*ListValue).Values, tailCall)
}

func (c *CallExpression) String() string {
	sb := "CallExpression "
	sb += c.Callee.String() + "("
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
	VarDeclaredNames() []IdentifierName
	_statement()
}

type StatementDefaultImpl struct {
	Statement
}

func (s *StatementDefaultImpl) _statement() {}
func (s *StatementDefaultImpl) VarScopedDeclarations() (l []*VariableDeclaration) {
	return nil
}

func (s *StatementDefaultImpl) VarDeclaredNames() (l []IdentifierName) {
	return
}

func (s *StatementDefaultImpl) Bytecode(e *Executable, c *BytecodeContext) {
	panic("should implement")
}

func (s *StatementDefaultImpl) String() string {
	panic("should implement")
}

func StatementAnalyze(s Statement, a AnalyzeQuery) bool {
	exprStmt, isExpr := s.(*StatementExpression)
	if isExpr {
		return ExpressionAnalyze(exprStmt.Expression, a)
	}
	switch a {
	case AnalyzeQueryIsIdentifierReference:
		return false
	case AnalyzeQueryIsReference:
		return false
	case AnalyzeQueryIsStringLiteral:
		return false
	}
	panic("unreachable")
}

// MARK: - VariableStatement

// TODO(SM): remove
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

func (s *StatementVariable) Evaluation(vm *VM2) Value {
	return s.DeclarationList.Evaluation(vm)
}

func (s *StatementVariable) String() string {
	return "var " + s.DeclarationList.String()
}

// VariableDeclarationList [In, Yield, Await] :
// VariableDeclaration[?In, ?Yield, ?Await]
// VariableDeclarationList[?In, ?Yield, ?Await] , VariableDeclaration[?In, ?Yield, ?Await]
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

func (v *VariableDeclarationList) Evaluation(vm *VM2) Value {
	var lastValue Value
	for _, item := range v.Items {
		lastValue = item.Evaluation(vm)
	}
	if lastValue == nil {
		return UndefinedValue
	}
	return lastValue
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

// VariableDeclaration [In] :
// - BindingIdentifier Initializer[?In]
type VariableDeclaration struct {
	ASTNode
	BindingIdentifier IdentifierName
	Initializer       Expression
}

func (v *VariableDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	if v.Initializer == nil {
		return
	}

	e.AddInstruction(InsLoad)
	e.AddInstruction(&IResolveBinding{
		Name: v.BindingIdentifier,
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

// VariableDeclaration : BindingIdentifier Initializer
func (v *VariableDeclaration) astHasInitializer() bool {
	return v.Initializer != nil
}

func (v *VariableDeclaration) Evaluation(vm *VM2) Value {
	if !v.astHasInitializer() {
		return UndefinedValue
	} else {
		bindingId := v.BindingIdentifier
		lhs := vm.agent.ResolveBinding(bindingId, nil, false)
		var value Value
		if IsAnonymousFunctionDefinition(v.Initializer) {
			// FIXME: handle named evaluation
			panic("")
		} else {
			rhs := v.Initializer.Evaluation(vm)
			value = rhs.GetValue(vm.agent)
		}
		lhs.PutValue(vm.agent, value)
	}

	// return EMPTY
	return UndefinedValue
}

func (v *VariableDeclaration) String() string {
	if v.Initializer != nil {
		return string(v.BindingIdentifier) + " = " + v.Initializer.String()
	}
	return string(v.BindingIdentifier)
}

// MARK: - BlockStatement

type StatementBlock struct {
	*StatementDefaultImpl
	BlockStatement BlockStatement
}

var _ Statement = (*StatementBlock)(nil)

func (s *StatementBlock) VarScopedDeclarations() (l []*VariableDeclaration) {
	return s.BlockStatement.VarScopedDeclarations()
}

func (s *StatementBlock) VarDeclaredNames() (l []IdentifierName) {
	return s.BlockStatement.VarDeclaredNames()
}
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
		e.AddDebug("Try Catch Start: " + t.String())
		exceptionJumpToCatch := &IPushExceptionJumpTarget{}
		e.AddInstruction(exceptionJumpToCatch)

		e.AddInstruction(&IStoreConstant{
			Value: UndefinedValue,
		})
		t.TryBlock.Bytecode(e, c)
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
		e.AddDebug("Catch block start")
		t.CatchBlock.Bytecode(e, c)

		exceptionJumpToEnd.Target = len(e.Instructions) - 1
		e.AddDebug("Try Catch end")
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
	sb := "try " + t.TryBlock.String()
	if t.CatchBlock != nil {
		sb += " catch (" + string(t.CatchParameter) + ") " + t.CatchBlock.String()
	}
	if t.FinallyBlock != nil {
		sb += " finally " + t.FinallyBlock.String()
	}
	return sb
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

func (s *StatementExpression) VarScopedDeclarations() (l []*VariableDeclaration) {
	return
}

func (s *StatementExpression) Bytecode(e *Executable, c *BytecodeContext) {
	s.Expression.Bytecode(e, c)
	if ExpressionAnalyze(s.Expression, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
}

func (s *StatementExpression) Evaluation(vm *VM2) Value {
	return s.Expression.Evaluation(vm)
}

func (s *StatementExpression) String() string {
	e := s.Expression.String()
	return "ExpressionStatement " + e
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

func (b *BreakableStatement) Evaluation(vm *VM2) Value {
	return b.IterationStatement.Evaluation(vm)
}

func (b *BreakableStatement) String() string {
	return b.IterationStatement.String()
}

// MARK: - ThrowStatement

type StatementThrow struct {
	*StatementDefaultImpl
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

// FunctionBody [Yield, Await] :
// - FunctionStatementList[?Yield, ?Await]
//
// FunctionStatementList[Yield, Await] :
// - StatementList[?Yield, ?Await, +Return] opt
type FunctionBody struct {
	ASTNode
	StatementList StatementList
	Strict        bool
	Type          FunctionType
}

func (f *FunctionBody) VarScopedDeclarations() (l []*VariableDeclaration) {
	return f.StatementList.VarScopedDeclarations()
}

func (f *FunctionBody) VarDeclaredNames() (l []IdentifierName) {
	return f.StatementList.VarDeclaredNames()
}

func (f *FunctionBody) LexicallyDeclaredNames() (l []IdentifierName) {
	return f.StatementList.TopLevelLexicallyDeclaredNames()
}

func (f *FunctionBody) Bytecode(e *Executable, c *BytecodeContext) {
	strictBefore := c.containedInStrictCode
	c.containedInStrictCode = c.containedInStrictCode || f.FunctionBodyContainsUseStrict()
	defer func() {
		c.containedInStrictCode = strictBefore
	}()
	f.StatementList.Bytecode(e, c)
}

// TODO(spec)
func (f *FunctionBody) Evaluation(vm *VM2) Value {
	return f.StatementList.Evaluation(vm)
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
		switch f := item.(type) {
		case *FormalParameter:
			if !f.BindingElement.IsSimpleParameterList() {
				return false
			}
		case *FormalParameterFunctionRestParameter:
			return false
		}
	}
	return true
}

func (f *FormalParameters) ContainsExpression() bool {
	for _, item := range f.Items {
		switch p := item.(type) {
		case *FormalParameter:
			if p.BindingElement.ContainsExpression() {
				return true
			}
		case *FormalParameterFunctionRestParameter:
			if p.BindingRestElement.ContainsExpression() {
				return true
			}
		}
	}
	return false
}

func (f *FormalParameters) BoundNames() (l []IdentifierName) {
	for _, item := range f.Items {
		var name IdentifierName
		switch p := item.(type) {
		case *FormalParameter:
			l = append(l, p.BindingElement.BoundNames()...)
			continue
		case *FormalParameterFunctionRestParameter:
			name = p.BindingRestElement.(*BindingRestElementIdentifier).Identifier
		}
		l = append(l, name)
	}
	return
}

// 15.1.5
func (f *FormalParameters) ExpectedArgumentCount() JSInt {
	l := JSInt(len(f.Items))
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
	_formalParametersItem()
	ContainsExpression() bool
}

// MARK:-
type FormalParameterFunctionRestParameter struct {
	FormalParametersItem
	BindingRestElement BindingRestElement
}

func (f *FormalParameterFunctionRestParameter) ContainsExpression() bool {
	return f.BindingRestElement.ContainsExpression()
}

// MARK: - BindingRestElement

type BindingRestElement interface {
	ASTNode
	_bindingRestElement()
	ContainsExpression() bool
}
type BindingRestElementIdentifier struct {
	BindingRestElement
	Identifier IdentifierName
}

func (b *BindingRestElementIdentifier) ContainsExpression() bool {
	return false
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

// MARK: - BindingElement

// SingleNameBinding [Yield, Await] :
// - BindingIdentifier[?Yield, ?Await] Initializer[+In, ?Yield, ?Await] opt
type SingleNameBinding struct {
	ASTNode
	boundName
	BindingIdentifier IdentifierName
	Initializer       Expression
}

func (s *SingleNameBinding) BoundNames() (l []IdentifierName) {
	l = append(l, s.BindingIdentifier)
	return
}

func (s *SingleNameBinding) String() string {
	if s.Initializer != nil {
		return string(s.BindingIdentifier) + " = " + s.Initializer.String()
	}
	return string(s.BindingIdentifier)
}

// BindingProperty [Yield, Await] :
// - SingleNameBinding[?Yield, ?Await]
// - PropertyName[?Yield, ?Await] : BindingElement[?Yield, ?Await]
type BindingProperty struct {
	ASTNode
	SingleNameBinding             *SingleNameBinding
	PropertyNameAndBindingElement *struct {
		PropertyName   PropertyName
		BindingElement *BindingElement
	}
}

func (b *BindingProperty) String() string {
	if b.SingleNameBinding != nil {
		return b.SingleNameBinding.String()
	}
	return b.PropertyNameAndBindingElement.PropertyName.String() + " : " + b.PropertyNameAndBindingElement.BindingElement.String()
}

// BindingRestProperty [Yield, Await] :
// - ... BindingIdentifier[?Yield, ?Await]
type BindingRestProperty struct {
	ASTNode
	BindingIdentifier IdentifierName
}

func (b *BindingRestProperty) String() string {
	return "..." + string(b.BindingIdentifier)
}

// ObjectBindingPattern : { }
// ObjectBindingPattern : { BindingPropertyList, BindingRestProperty }
type ObjectBindingPattern struct {
	Properties []*struct {
		BindingPropertyList []BindingProperty
		BindingRestProperty *BindingRestProperty
	}
}

func (o *ObjectBindingPattern) ContainsExpression() bool {
	// TODO
	return false
}

func (o *ObjectBindingPattern) String() string {
	var sb string
	for i, p := range o.Properties {
		if i != 0 {
			sb += ", "
		}
		for j, bp := range p.BindingPropertyList {
			if j != 0 {
				sb += ", "
			}
			sb += bp.String()
		}
		if p.BindingRestProperty != nil {
			sb += ", " + p.BindingRestProperty.String()
		}
	}
	return sb
}

// ArrayBindingPattern [Yield, Await] :
// - [ Elision(opt) BindingRestElement[?Yield, ?Await] opt ]
// - [ BindingElementList[?Yield, ?Await] ]
// - [ BindingElementList[?Yield, ?Await] , Elision(opt)
// - BindingRestElement[?Yield, ?Await] opt ]
type ArrayBindingPattern struct {
	Elements []*struct {
		Elision            bool
		BindingRestElement BindingRestElement
		BindingElement     *BindingElement
	}
}

func (a *ArrayBindingPattern) ContainsExpression() bool {
	// TODO
	return false
}

func (a *ArrayBindingPattern) String() string {
	var sb string
	for i, e := range a.Elements {
		if i != 0 {
			sb += ", "
		}
		if e.BindingElement != nil {
			sb += e.BindingElement.String()
		} else {
			sb += e.BindingRestElement.String()
		}
	}
	return sb
}

// BindingPattern : ObjectBindingPattern | ArrayBindingPattern
type BindingPattern struct {
	boundName
	ObjectBindingPattern *ObjectBindingPattern
	ArrayBindingPattern  *ArrayBindingPattern
}

func (b *BindingPattern) BoundNames() (l []IdentifierName) {
	if b.ObjectBindingPattern != nil {
		for _, p := range b.ObjectBindingPattern.Properties {
			for _, bp := range p.BindingPropertyList {
				if bp.SingleNameBinding != nil {
					l = append(l, bp.SingleNameBinding.BindingIdentifier)
				}
			}
		}
	}
	if b.ArrayBindingPattern != nil {
		for _, e := range b.ArrayBindingPattern.Elements {
			if e.BindingElement != nil {
				l = append(l, e.BindingElement.BoundNames()...)
			}
		}
	}
	return
}

func (b *BindingPattern) ContainsExpression() bool {
	if b.ObjectBindingPattern != nil {
		return b.ObjectBindingPattern.ContainsExpression()
	}
	return b.ArrayBindingPattern.ContainsExpression()
}

func (b *BindingPattern) String() string {
	if b.ObjectBindingPattern != nil {
		return b.ObjectBindingPattern.String()
	}
	return b.ArrayBindingPattern.String()
}

// BindingElement [Yield, Await] :
// - SingleNameBinding[?Yield, ?Await]
// - BindingPattern[?Yield, ?Await] Initializer[+In, ?Yield, ?Await] opt
type BindingElement struct {
	BindingPattern    *BindingPattern
	SingleNameBinding *SingleNameBinding
}

func (b *BindingElement) IsSimpleParameterList() bool {
	return b.SingleNameBinding != nil && b.SingleNameBinding.Initializer == nil
}

func (b *BindingElement) ContainsExpression() bool {
	if b.SingleNameBinding != nil {
		return b.SingleNameBinding.Initializer != nil
	}
	if b.BindingPattern.ObjectBindingPattern != nil {
		return b.BindingPattern.ContainsExpression()
	}
	panic("unreachable")
}

func (b *BindingElement) BoundNames() (l []IdentifierName) {
	if b.SingleNameBinding != nil {
		return b.SingleNameBinding.BoundNames()
	}
	return b.BindingPattern.BoundNames()
}

func (b *BindingElement) String() string {
	if b.SingleNameBinding != nil {
		return b.SingleNameBinding.String()
	}
	return b.BindingPattern.String()
}

// MARK: - IfStatement

// IfStatement [Yield, Await, Return] :
//   - if ( Expression[+In, ?Yield, ?Await] ) Statement[?Yield, ?Await, ?Return] else
//     Statement[?Yield, ?Await, ?Return]
//   - if ( Expression[+In, ?Yield, ?Await] ) Statement[?Yield, ?Await, ?Return]
//     [lookahead ≠ else]
type IfStatement struct {
	Statement
	Condition  Expression
	Consequent Statement
	Alternate  Statement
}

var _ Statement = (*IfStatement)(nil)

func (s *IfStatement) VarScopedDeclarations() (l []*VariableDeclaration) {
	l = append(l, s.Consequent.VarScopedDeclarations()...)
	if s.Alternate != nil {
		l = append(l, s.Alternate.VarScopedDeclarations()...)
		return
	}
	return
}

// 14.6.2
func (s *IfStatement) Bytecode(e *Executable, c *BytecodeContext) {
	s.Condition.Bytecode(e, c)

	if ExpressionAnalyze(s.Condition, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
	e.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(e.Instructions) - 1
	e.AddInstruction(&IStoreConstant{Value: UndefinedValue})
	s.Consequent.Bytecode(e, c)
	endJump := &IJump{Target: 0}
	e.AddInstruction(endJump)

	// else
	jumpIfTrue.TargetElse = len(e.Instructions) - 1
	e.AddInstruction(&IStoreConstant{Value: UndefinedValue})

	if s.Alternate != nil {
		s.Alternate.Bytecode(e, c)
	}
	endJump.Target = len(e.Instructions) - 1
}

func (s *IfStatement) astHasElse() bool {
	return s.Alternate != nil
}

// Evaluation
// spec: 14.6.2
func (s *IfStatement) Evaluation(vm *VM2) Value {
	if s.astHasElse() {
		exprRef := s.Condition.Evaluation(vm)
		exprValue := exprRef.GetValue(vm.agent)
		var stmtCompletion Value
		if exprValue.ToBoolean() {
			stmtCompletion = s.Consequent.Evaluation(vm)
		} else {
			stmtCompletion = s.Alternate.Evaluation(vm)
		}
		return UpdateEmpty(stmtCompletion, UndefinedValue)
	} else {
		exprRef := s.Condition.Evaluation(vm)
		exprValue := exprRef.GetValue(vm.agent)
		if !exprValue.ToBoolean() {
			return UndefinedValue
		} else {
			stmtCompletion := s.Consequent.Evaluation(vm)
			return UpdateEmpty(stmtCompletion, UndefinedValue)
		}
	}
}

func (s *IfStatement) String() string {
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

	conditionIndex := len(e.Instructions) - 1
	s.Condition.Bytecode(e, c)
	if ExpressionAnalyze(s.Condition, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}

	jumpIfTrue := &IJumpIfTrue{Target: 0, TargetElse: 0}
	e.AddInstruction(jumpIfTrue)

	jumpIfTrue.Target = len(e.Instructions) - 1
	e.AddInstruction(InsStore)
	s.Body.Bytecode(e, c)
	bodyEndIndex := len(e.Instructions) - 1
	e.AddInstruction(InsLoad)

	e.AddInstruction(&IJump{Target: conditionIndex})

	jumpIfTrue.TargetElse = len(e.Instructions) - 1
	e.AddInstruction(InsStore)

	for _, index := range c.continueJumpIndices.Data() {
		index.Target = bodyEndIndex
	}
	c.continueJumpIndices.Clear()
	for _, index := range c.breakJumpIndices.Data() {
		index.Target = len(e.Instructions) - 1
	}
	c.breakJumpIndices.Clear()
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

	bodyStartIndex := len(e.Instructions) - 1

	e.AddInstruction(InsStore)
	s.Body.Bytecode(e, c)
	bodyEndIndex := len(e.Instructions) - 1
	e.AddInstruction(InsLoad)

	s.Condition.Bytecode(e, c)
	if ExpressionAnalyze(s.Condition, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}

	e.AddInstruction(InsLoad)
	jumpIfTrue := &IJumpIfTrue{Target: bodyStartIndex, TargetElse: 0}
	jumpIfTrue.TargetElse = len(e.Instructions) - 1

	e.AddInstruction(InsStore)

	for _, index := range c.continueJumpIndices.Data() {
		index.Target = bodyEndIndex
	}
	c.continueJumpIndices.Clear()
	if c.Label != "" {
		// TODO: Label Jump
	}
	for _, index := range c.breakJumpIndices.Data() {
		index.Target = len(e.Instructions) - 1
	}
	c.breakJumpIndices.Clear()
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

func (f *ForStatementInitializerVariable) Evaluation(vm *VM2) Value {
	return f.VariableStatement.Evaluation(vm)
}

type ForStatementInitializerLexicalDeclaration struct {
	ForStatementInitializer
	LexicalDeclaration *LexicalDeclaration
}

func (f *ForStatementInitializerLexicalDeclaration) String() string {
	return f.LexicalDeclaration.String()
}

// ForStatement :
// for ( LexicalDeclaration ; ) Statement
// for ( LexicalDeclaration ; Expression ) Statement
// for ( LexicalDeclaration Expression ; ) Statement
// for ( LexicalDeclaration Expression ; Expression ) Statement
type ForStatement struct {
	IterationStatement
	Initializer ForStatementInitializer
	Condition   Expression
	Increment   Expression
	Body        Statement
}

func (s *ForStatement) VarScopedDeclarations() (l []*VariableDeclaration) {
	if s.Initializer != nil {
		if varStatement, ok := s.Initializer.(*ForStatementInitializerVariable); ok {
			l = append(l, varStatement.VariableStatement.DeclarationList.VarScopedDeclarations()...)
		}
	}
	l = append(l, s.Body.VarScopedDeclarations()...)
	return
}

func (s *ForStatement) Bytecode(e *Executable, c *BytecodeContext) {
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
	bodyEndIndex := len(e.Instructions) - 1
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

	for _, index := range c.continueJumpIndices.Data() {
		index.Target = bodyEndIndex
	}
	c.continueJumpIndices.Clear()
	for _, index := range c.breakJumpIndices.Data() {
		index.Target = len(e.Instructions) - 1
	}
	c.breakJumpIndices.Clear()
}

// isVariableDeclarationList identifies
// ForStatement : for ( var VariableDeclarationList ; Expression opt ; Expression opt ) Statement
func (s *ForStatement) isVariableDeclarationList() bool {
	_, ok := s.Initializer.(*ForStatementInitializerVariable)
	return ok
}

// ForLoopEvaluation 14.7.4.2
func (s *ForStatement) ForLoopEvaluation(vm *VM2) Value {
	// TODO: label set
	if s.isVariableDeclarationList() {
		s.Initializer.Evaluation(vm)
		var test Expression
		if s.Condition != nil {
			test = s.Condition
		}
		var increment Expression
		if s.Increment != nil {
			increment = s.Increment
		}
		var perIterationBindings []string
		return vm.ForBodyEvaluation(test, increment, s.Body, perIterationBindings, nil)
	}
	panic("unimplemented")
}

func (s *ForStatement) Evaluation(vm *VM2) Value {
	return s.ForLoopEvaluation(vm)
}

func (s *ForStatement) String() string {
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

// MARK: - ForInOfStatement

type ForInOfStatementType int

const (
	ForInOfStatementTypeIn ForInOfStatementType = iota
	ForInOfStatementTypeOf
)

// ForInOfStatement [Yield, Await, GetLastValue] :
// - for ( [lookahead ≠ let [] LeftHandSideExpression[?Yield, ?Await] in
//   - Expression[+In, ?Yield, ?Await] ) Statement[?Yield, ?Await, ?GetLastValue]
//
// - for ( var ForBinding[?Yield, ?Await] in Expression[+In, ?Yield, ?Await] )
//   - Statement[?Yield, ?Await, ?GetLastValue]
//
// - for ( ForDeclaration[?Yield, ?Await] in Expression[+In, ?Yield, ?Await] )
//   - Statement[?Yield, ?Await, ?GetLastValue]
//
// - for ( [lookahead ∉ { let, async of}] LeftHandSideExpression[?Yield, ?Await] of
//   - AssignmentExpression[+In, ?Yield, ?Await] ) Statement[?Yield, ?Await, ?GetLastValue]
//
// - for ( var ForBinding[?Yield, ?Await] of AssignmentExpression[+In, ?Yield, ?Await]
//   - ) Statement[?Yield, ?Await, ?GetLastValue]
//
// - for ( ForDeclaration[?Yield, ?Await] of AssignmentExpression[+In, ?Yield, ?Await]
//   - ) Statement[?Yield, ?Await, ?GetLastValue]
//
// - [+Await] for await ( [lookahead ≠ let] LeftHandSideExpression[?Yield, ?Await] of
//   - AssignmentExpression[+In, ?Yield, ?Await] ) Statement[?Yield, ?Await, ?GetLastValue]
//
// - [+Await] for await ( var ForBinding[?Yield, ?Await] of
//   - AssignmentExpression[+In, ?Yield, ?Await] ) Statement[?Yield, ?Await, ?GetLastValue]
//
// - [+Await] for await ( ForDeclaration[?Yield, ?Await] of
//   - AssignmentExpression[+In, ?Yield, ?Await] ) Statement[?Yield, ?Await, ?GetLastValue]
type ForInOfStatement struct {
	IterationStatement
	Type        ForInOfStatementType
	IsAwait     bool
	Body        Statement
	isVar       bool
	Initializer *ForInOfStatementInitializer
	// Expression is Expression, LeftHandSideExpression or AssignmentExpression
	Expression Expression
}

var _ RuntimeSemanticsForInOfLoopEvaluation = (*ForInOfStatement)(nil)

func (f *ForInOfStatement) VarScopedDeclarations() (l []*VariableDeclaration) {
	l = append(l, f.Body.VarScopedDeclarations()...)
	return
}

func (f *ForInOfStatement) BoundNames() (l []IdentifierName) {
	if f.Initializer.ForBinding != nil {
		l = append(l, f.Initializer.ForBinding.BoundNames()...)
		l = append(l, f.Body.VarDeclaredNames()...)
	}
	return
}

func (f *ForInOfStatement) Evaluation(vm *VM2) Value {
	// TODO
	var labelSet []string
	result, err := f.ForInOfLoopEvaluation(vm, labelSet)
	if err != nil {
		panic(err)
	}
	return result
}

func (f *ForInOfStatement) astIsLeftHandSideExpression() bool {
	return f.Initializer.LeftHandSideExpression != nil
}

func (f *ForInOfStatement) astIsIn() bool {
	return f.Type == ForInOfStatementTypeIn
}

func (f *ForInOfStatement) astIsOf() bool {
	return f.Type == ForInOfStatementTypeOf
}

func (f *ForInOfStatement) astIsForDeclaration() bool {
	return f.Initializer.ForDeclaration != nil
}

// ForInOfLoopEvaluation
// spec: 14.7.5.5
func (f *ForInOfStatement) ForInOfLoopEvaluation(vm *VM2, labelSet []string) (value Value, err Value) {
	if f.astIsForDeclaration() {
		forDeclaration := f.Initializer.ForDeclaration
		if f.astIsOf() {
			keyResult, err := vm.ForInOfHeadEvaluation(forDeclaration.BoundNames(), f.Expression, ForInOfIterationKindIterate)
			if err != nil {
				return nil, err
			}
			return vm.ForInOfBodyEvaluation(
				forDeclaration,
				f.Body,
				keyResult,
				ForInOfIterationKindIterate,
				ForInOfLhsKindLexicalBinding,
				labelSet,
				IteratorKindSync,
			)
		} else {
			// for in
			keyResult, err := vm.ForInOfHeadEvaluation(forDeclaration.BoundNames(), f.Expression, ForInOfIterationKindEnumerate)
			if err != nil {
				return nil, err
			}
			return vm.ForInOfBodyEvaluation(
				forDeclaration,
				f.Body,
				keyResult,
				ForInOfIterationKindEnumerate,
				ForInOfLhsKindLexicalBinding,
				labelSet,
				IteratorKindSync,
			)
		}
	}
	panic("unimplemented")
}

type ForInOfIterationKind int

const (
	ForInOfIterationKindEnumerate ForInOfIterationKind = iota
	ForInOfIterationKindIterate
	ForInOfIterationKindAsyncIterate
)

type ForInOfLhsKind int

const (
	ForInOfLhsKindAssignment ForInOfLhsKind = iota
	ForInOfLhsKindVarBinding
	ForInOfLhsKindLexicalBinding
)

func (f *ForInOfStatement) Bytecode(e *Executable, c *BytecodeContext) {
	var iterationKind ForInOfIterationKind
	if f.Type == ForInOfStatementTypeIn {
		iterationKind = ForInOfIterationKindEnumerate
	} else {
		if f.IsAwait {
			iterationKind = ForInOfIterationKindAsyncIterate
		} else {
			iterationKind = ForInOfIterationKindIterate
		}
	}
	var lhsKind ForInOfLhsKind
	if f.Initializer.ForBinding != nil {
		lhsKind = ForInOfLhsKindVarBinding
	} else if f.Initializer.ForDeclaration != nil {
		lhsKind = ForInOfLhsKindLexicalBinding
	} else {
		lhsKind = ForInOfLhsKindAssignment
	}
	var iteratorKind IteratorKind
	if iterationKind == ForInOfIterationKindAsyncIterate {
		iteratorKind = IteratorKindAsync
	} else {
		iteratorKind = IteratorKindSync
	}

	index := f.forInOfHeadEvaluation(e, c, f.Expression, iterationKind, iteratorKind)
	f.forInOfBodyEvaluation(e, c, index, lhsKind, iteratorKind)
}

// ForIn/OfHeadEvaluation
func (f *ForInOfStatement) forInOfHeadEvaluation(e *Executable, c *BytecodeContext, expr Expression, iterationKind ForInOfIterationKind, iteratorKind IteratorKind) *IJumpIfTrue {
	// TODO: BoundNames of expr

	expr.Bytecode(e, c)
	if ExpressionAnalyze(expr, AnalyzeQueryIsReference) {
		e.AddInstruction(InsGetValue)
	}
	if iterationKind == ForInOfIterationKindEnumerate {
		e.AddInstruction(InsLoad)
		// a. If exprValue is either undefined or null, then
		//     i. GetLastValue Completion Record { [[Type]]: break, [[Value]]: empty, [[Target]]: empty }.
		e.AddInstruction(InsLoad)
		e.AddInstruction(&ILoadConstant{Value: UndefinedValue})
		e.AddInstruction(InsLooselyEqual)
		jumpIfTrue := &IJumpIfTrue{}
		e.AddInstruction(jumpIfTrue)
		jumpIfTrue.TargetElse = len(e.Instructions) - 1

		e.AddInstruction(InsStore)
		e.AddInstruction(&IForInIterator{})
		return jumpIfTrue
	} else {
		e.AddInstruction(&IGetIterator{
			IteratorKind: iteratorKind,
		})
		return nil
	}
}

func (f *ForInOfStatement) IsDestructuring() bool {
	// TODO:
	return f.Initializer.ForBinding != nil && f.Initializer.ForBinding.BindingPattern != nil
}

// ForIn/OfBodyEvaluation
func (f *ForInOfStatement) forInOfBodyEvaluation(
	e *Executable,
	c *BytecodeContext,
	jumpIndex *IJumpIfTrue,
	lhsKind ForInOfLhsKind,
	iteratorKind IteratorKind,
) {
	body := f.Body
	e.AddInstruction(InsPushLexicalEnvironment)
	e.AddInstruction(&ILoadConstant{Value: UndefinedValue})

	destructuring := f.IsDestructuring()
	lhs := f.Initializer

	if destructuring && lhsKind == ForInOfLhsKindAssignment {
		// TODO
	}

	startIndex := len(e.Instructions) - 1
	e.AddInstruction(&ILoadIterator{})
	e.AddInstruction(&ICall{})

	if iteratorKind == IteratorKindAsync {
		// e.AddInstruction(InsAwait)
	}

	e.AddInstruction(InsLoad)

	e.AddInstruction(InsLoad)
	e.AddInstruction(&IEvaluatePropertyAccessWithIdentifierKey{
		Name:   "done",
		Strict: false,
	})
	e.AddInstruction(InsGetValue)

	jumpIfTrue := &IJumpIfTrue{}
	e.AddInstruction(jumpIfTrue)

	jumpIfTrue.TargetElse = len(e.Instructions) - 1

	e.AddInstruction(&IEvaluatePropertyAccessWithIdentifierKey{
		Name:   "value",
		Strict: false,
	})
	e.AddInstruction(InsGetValue)

	if lhsKind == ForInOfLhsKindAssignment || lhsKind == ForInOfLhsKindVarBinding {
		if destructuring {
			if lhsKind == ForInOfLhsKindAssignment {
				// TODO
			} else {
				// TODO
			}
		} else {
			if lhs.LeftHandSideExpression != nil {
				lhs.LeftHandSideExpression.Bytecode(e, c)
			} else if lhs.ForBinding != nil {
				e.AddInstruction(&IResolveBinding{
					Name: lhs.ForBinding.BindingIdentifier,
				})
			} else {
				panic("unreachable")
			}
			e.AddInstruction(InsPushReference)
			e.AddInstruction(InsPutValue)
			e.AddInstruction(InsPopReference)
		}
	} else {
		lexicalDeclaration := &LexicalDeclaration{
			BindingList: &BindingList{
				Items: make([]*LexicalBinding, 0),
			},
		}
		lexicalBinding := &LexicalBinding{}
		if lhs.ForDeclaration.ForBinding.BindingIdentifier != "" {
			lexicalBinding.Identifier = lhs.ForDeclaration.ForBinding.BindingIdentifier
		} else {
			lexicalBinding.BindingPattern = lhs.ForDeclaration.ForBinding.BindingPattern
		}
		lexicalDeclaration.BindingList.Items = append(lexicalDeclaration.BindingList.Items, lexicalBinding)
		e.AddInstruction(&IForDeclarationBindingInstantiation{
			LexicalDeclaration: lexicalDeclaration,
		})

		if destructuring {
			// TODO
			panic("unimplemented")
		} else {
			lhsName := lhs.ForDeclaration.ForBinding.BindingIdentifier
			e.AddInstruction(&IResolveBinding{
				Name:   lhsName,
				Strict: c.containedInStrictCode,
			})
			e.AddInstruction(&IInitializeReferencedBinding{})
		}
	}

	e.AddInstruction(InsStore)
	body.Bytecode(e, c)
	continueIndex := len(e.Instructions) - 1
	e.AddInstruction(InsLoad)
	e.AddInstruction(InsRestoreLexicalEnvironment)

	e.AddInstruction(&IJump{Target: startIndex})
	jumpIfTrue.Target = len(e.Instructions) - 1
	e.AddInstruction(InsStore)
	e.AddInstruction(InsStore)
	for _, index := range c.continueJumpIndices.Data() {
		index.Target = continueIndex
	}
	c.continueJumpIndices.Clear()
	for _, index := range c.breakJumpIndices.Data() {
		index.Target = len(e.Instructions) - 1
	}
	c.breakJumpIndices.Clear()

	if jumpIndex != nil {
		skipJump := &IJump{}
		e.AddInstruction(skipJump)
		jumpIndex.Target = len(e.Instructions) - 1

		e.AddInstruction(InsStore)
		e.AddInstruction(&IStoreConstant{Value: UndefinedValue})
		skipJump.Target = len(e.Instructions) - 1
	}
}

func (f *ForInOfStatement) String() string {
	s := "ForInOfStatement "
	if f.IsAwait {
		s += "await "
	}
	s += f.Initializer.String() + " "
	if f.Type == ForInOfStatementTypeIn {
		s += "In "
	} else {
		s += "Of "
	}
	s += f.Expression.String() + " " + f.Body.String()
	return s
}

// ForInOfStatementInitializer Enum
type ForInOfStatementInitializer struct {
	// LeftHandSideExpression [Yield, Await] :
	// - NewExpression[?Yield, ?Await]
	// - CallExpression[?Yield, ?Await]
	// - OptionalExpression[?Yield, ?Await]
	LeftHandSideExpression Expression
	ForBinding             *ForBinding
	ForDeclaration         *ForDeclaration
}

func (f *ForInOfStatementInitializer) String() string {
	if f.ForBinding != nil {
		return f.ForBinding.String()
	} else if f.ForDeclaration != nil {
		return f.ForDeclaration.String()
	} else {
		return f.LeftHandSideExpression.String()
	}
}

// ForBinding [Yield, Await] :
// - BindingIdentifier[?Yield, ?Await]
// - BindingPattern[?Yield, ?Await]
type ForBinding struct {
	BindingIdentifier IdentifierName
	BindingPattern    *BindingPattern
}

var _ StaticSemanticsBoundNames = (*ForBinding)(nil)

func (f *ForBinding) String() string {
	if f.BindingPattern != nil {
		return f.BindingPattern.String()
	}
	return string(f.BindingIdentifier)
}

func (f *ForBinding) BoundNames() (l []IdentifierName) {
	if f.BindingPattern != nil {
		return f.BindingPattern.BoundNames()
	}
	return []IdentifierName{f.BindingIdentifier}
}

// ForDeclaration [Yield, Await] :
// - LetOrConst ForBinding[?Yield, ?Await]
type ForDeclaration struct {
	// TODO(BM): unused, remove
	Expression
	LetOrConst LetOrConst
	ForBinding *ForBinding
}

var (
	_ StaticSemanticsBoundNames                          = (*ForDeclaration)(nil)
	_ RuntimeSemanticsForDeclarationBindingInstantiation = (*ForDeclaration)(nil)
)

func (f *ForDeclaration) ForDeclarationBindingInstantiation(vm *VM2, env EnvironmentRecord) {
	for _, name := range f.ForBinding.BoundNames() {
		if f.LetOrConst.IsConstantDeclaration() {
			env.CreateImmutableBinding(name, true)
		} else {
			env.CreateMutableBinding(name, false)
		}
	}
	// return UNUSED
}

func (f *ForDeclaration) BoundNames() (l []IdentifierName) {
	return f.ForBinding.BoundNames()
}

func (f *ForDeclaration) String() string {
	var letOrConst string
	if f.LetOrConst == LetOrConstLet {
		letOrConst = "Let"
	} else {
		letOrConst = "Const"
	}
	return letOrConst + " " + f.ForBinding.String()
}

// MARK: - BreakStatement

type StatementBreak struct {
	*StatementDefaultImpl
	Label IdentifierName
}

func (s *StatementBreak) Bytecode(e *Executable, c *BytecodeContext) {
	jump := &IJump{}
	if s.Label != "" {
		e.AddInstruction(jump)
		if list, ok := c.labelBreakJumpIndexMap[string(s.Label)]; ok {
			list.Push(jump)
		}
	} else {
		e.AddInstruction(jump)
		c.breakJumpIndices.Push(jump)
	}
}

func (s *StatementBreak) String() string {
	if s.Label != "" {
		return "Break " + string(s.Label)
	}
	return "Break"
}

// MARK: - ContinueStatement

type StatementContinue struct {
	Statement
	Label IdentifierName
}

func (s *StatementContinue) Bytecode(e *Executable, c *BytecodeContext) {
	jump := &IJump{}
	if s.Label != "" {
		e.AddInstruction(jump)
		if list, ok := c.labelContinueJumpIndexMap[string(s.Label)]; ok {
			list.Push(jump)
		}
	} else {
		e.AddInstruction(jump)
		c.continueJumpIndices.Push(jump)
	}
}

func (s *StatementContinue) String() string {
	if s.Label != "" {
		return "Continue " + string(s.Label)
	}
	return "Continue"
}

// MARK: - ReturnStatement

// ReturnStatement [Yield, Await] :
// - return ;
// - return [no LineTerminator here] Expression[+In, ?Yield, ?Await] ;
type StatementReturn struct {
	*StatementDefaultImpl
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

func (s *StatementReturn) Evaluation(vm *VM2) Value {
	defer func() {
		vm.isReturn = true
	}()
	if s.Expression == nil {
		return UndefinedValue
	} else {
		exprRef := s.Expression.Evaluation(vm)
		exprValue := exprRef.GetValue(vm.agent)
		// TODO: GetGeneratorKind
		return exprValue
	}
}

func (s *StatementReturn) String() string {
	if s.Expression != nil {
		return "CompletionTypeReturn " + s.Expression.String()
	}
	return "CompletionTypeReturn"
}

// MARK: - Declaration

// Declaration [Yield, Await] :
// HoistableDeclaration[?Yield, ?Await, ~Default]
// ClassDeclaration[?Yield, ?Await, ~Default]
// LexicalDeclaration[+In, ?Yield, ?Await]
type Declaration interface {
	ASTNode
	BoundNames() []IdentifierName
}
type declarationDefaultImpl struct {
	Declaration
}

func DeclarationBoundNames(d Declaration) (l []IdentifierName) {
	switch decl := d.(type) {
	case *DeclarationHoistableFunction, *DeclarationHoistableAsyncFunction:
		return
	case *ClassDeclaration:
		return decl.BoundNames()
	case *LexicalDeclaration:
		return decl.BoundNames()
	}
	panic("unreachable")
}

func DeclarationAnalyze(d Declaration, a AnalyzeQuery) bool {
	return false
}

// MARK: - HoistableDeclaration

// TODO: use HoistableDeclaration
// HoistableDeclaration [Yield, Await, Default] :
// - FunctionDeclaration[?Yield, ?Await, ?Default]
// - GeneratorDeclaration[?Yield, ?Await, ?Default]
// - AsyncFunctionDeclaration[?Yield, ?Await, ?Default]
// - AsyncGeneratorDeclaration[?Yield, ?Await, ?Default]
type DeclarationHoistable interface {
	Declaration
}

// MARK: - FunctionDeclaration

// TODO: use FunctionDeclaration directly
type DeclarationHoistableFunction struct {
	DeclarationHoistable
	*declarationDefaultImpl
	FunctionDeclaration *FunctionDeclaration
}

func (d *DeclarationHoistableFunction) Bytecode(e *Executable, c *BytecodeContext) {
	d.FunctionDeclaration.Bytecode(e, c)
}

func (d *DeclarationHoistableFunction) Evaluation(vm *VM2) Value {
	return d.FunctionDeclaration.Evaluation(vm)
}

func (d *DeclarationHoistableFunction) BoundNames() []IdentifierName {
	return d.FunctionDeclaration.BoundNames()
}

func (d *DeclarationHoistableFunction) String() string {
	return d.FunctionDeclaration.String()
}

// MARK: - AsyncFunctionDeclaration

type DeclarationHoistableAsyncFunction struct {
	DeclarationHoistable
	*declarationDefaultImpl
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
	realm.GlobalEnv.ObjectRecord.BindingObject.Set(NewStringPropertyKey(string(d.Identifier)), (function).ToValue(), setThrowTypeIgnore)
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
	realm.GlobalEnv.ObjectRecord.BindingObject.Set(NewStringPropertyKey(string(d.Identifier)), (function).ToValue(), setThrowTypeIgnore)
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
	SetFunctionName(function, NewStringPropertyKey(string(name)), "")
	prototype := OrdinaryObjectCreate(agent, realm.Intrinsics.AsyncGeneratorFunctionPrototypePrototype, nil)
	function.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
		Value:        (prototype).ToValue(),
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
	function := d.instantiateGeneratorFunctionObject(c.agent, env, nil)
	realm.GlobalEnv.ObjectRecord.BindingObject.Set(CMString(d.Identifier).ToPropertyKey(), function.ToValue(), setThrowTypeIgnore)
}

// 15.5.3
func (d *GeneratorDeclaration) instantiateGeneratorFunctionObject(agent *Agent, env EnvironmentRecord, privateEnv *PrivateEnvironment) ObjectType {
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
		Value:        (prototype).ToValue(),
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

// ClassDeclaration [Yield, Await, Default] :
// - class BindingIdentifier[?Yield, ?Await] ClassTail[?Yield, ?Await]
// - [+Default] class ClassTail[?Yield, ?Await]
type ClassDeclaration struct {
	Declaration
	IdentifierName IdentifierName
	ClassTail      *ClassTail
	SourceText     string
}

var (
	_ RuntimeSemanticsEvaluation                        = (*ClassDeclaration)(nil)
	_ RuntimeSemanticsBindingClassDeclarationEvaluation = (*ClassDeclaration)(nil)
)

func (d *ClassDeclaration) LexicallyDeclaredNames() (l []IdentifierName) {
	return d.BoundNames()
}

func (d *ClassDeclaration) BoundNames() (l []IdentifierName) {
	if d.IdentifierName != "" {
		l = append(l, d.IdentifierName)
	} else {
		l = append(l, "default")
	}
	return
}

func (d *ClassDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	e.AddInstruction(InsLoad)
	e.AddInstruction(&IBindingClassDeclarationEvaluation{
		ClassDeclaration: d,
	})
	e.AddInstruction(InsStore)
}

func (d *ClassDeclaration) astHasIdentifier() bool {
	return d.IdentifierName != ""
}

// 15.7.16
func (d *ClassDeclaration) Evaluation(vm *VM2) Value {
	d.BindingClassDeclarationEvaluation(vm)
	// return EMPTY
	return UndefinedValue
}

// BindingClassDeclarationEvaluation
// spec: 15.7.15
func (d *ClassDeclaration) BindingClassDeclarationEvaluation(vm *VM2) (obj ObjectType, err Value) {
	if d.astHasIdentifier() {
		className := d.IdentifierName
		value, err := d.ClassTail.ClassDefinitionEvaluation(vm, className, NewStringPropertyKey(className))
		if err != nil {
			return nil, err
		}
		// TODO: set [[SourceText]]
		env := vm.RunningLexicalEnvironment()
		vm.InitializeBoundName(className, value.ToValue(), env)
		return value, nil
	} else {
		value, err := d.ClassTail.ClassDefinitionEvaluation(vm, "", NewStringPropertyKey("default"))
		if err != nil {
			return nil, err
		}
		// TODO: set [[SourceText]]
		return value, nil
	}
}

func (d *ClassDeclaration) String() string {
	return "ClassDeclaration " + string(d.IdentifierName)
}

type ClassTail struct {
	ClassHeritage Expression
	ClassBody     *ClassBody
}

var _ RuntimeSemanticsClassDefinitionEvaluation = (*ClassTail)(nil)

// ClassDefinitionEvaluation
// spec: 15.7.14
func (c *ClassTail) ClassDefinitionEvaluation(vm *VM2, classBinding string, className PropertyKeyOrPrivateName) (obj ObjectType, err Value) {
	agent := vm.agent
	realm := agent.CurrentRealm()
	// outer env of class
	env := agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment
	// the class env
	classEnv := NewDeclarativeEnvironment(env)
	if classBinding != "" {
		classEnv.CreateImmutableBinding(classBinding, true)
	}

	outerPrivateEnvironment := agent.RunningExecutionContext().ECMAScriptCode.PrivateEnvironment
	classPrivateEnvironment := NewPrivateEnvironment(outerPrivateEnvironment)

	if len(c.ClassBody.ClassElementList.Items) > 0 {
		privateBoundIdentifiers := c.ClassBody.PrivateBoundIdentifiers()
		for _, privateBoundIdentifier := range privateBoundIdentifiers {
			names := lo.Map(classPrivateEnvironment.Names, func(item PrivateName, index int) string {
				return item.Symbol.Description
			})
			if lo.Contains(names, string(privateBoundIdentifier)) {
			} else {
				classPrivateEnvironment.Names = append(classPrivateEnvironment.Names, PrivateName{
					Symbol: agent.CreateSymbol(string(privateBoundIdentifier)),
				})
			}
		}
	}

	var protoParent ObjectType
	var constructorParent ObjectType
	if c.ClassHeritage == nil {
		protoParent = realm.Intrinsics.ObjectPrototype
		constructorParent = realm.Intrinsics.FunctionPrototype
	} else {
		agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = classEnv
		superclassRef := GenerateAndRunBytecode(agent,
			&StatementExpression{
				Expression: c.ClassHeritage,
			},
		)
		agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = env
		superclass := superclassRef.Data().GetValue(agent)
		if superclass == nil {
			protoParent = nil
			constructorParent = realm.Intrinsics.FunctionPrototype
		} else if !IsConstructor(superclass) {
			panic("TypeError: superclass is not a constructor")
		} else {
			protoParentValue := MustGetObject(superclass).Get(NewStringPropertyKey("prototype"))
			if !ValueIsObject(protoParentValue) {
				panic("TypeError: prototype is not an object")
			}
			protoParent = MustGetObject(protoParentValue)
			constructorParent = MustGetObject(superclass)
		}
	}

	proto := OrdinaryObjectCreate(agent, protoParent, nil)
	constructor := c.ClassBody.ConstructorMethod()
	agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = classEnv
	agent.RunningExecutionContext().ECMAScriptCode.PrivateEnvironment = classPrivateEnvironment

	var function ObjectType
	if constructor == nil {
		var defaultConstructor BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
			args := arguments
			if newTarget == nil {
				return agent.ThrowTypeError("class must be invoked with 'new'")
			}

			F := agent.ActiveFunctionObject()
			classConstructorFields := ObjectAs[*BuiltinFunction](F).AdditionalFields.ClassConstructorFields
			var result ObjectType
			if classConstructorFields.ConstructorKind == ConstructorKindDerived {
				fun := function.InternalMethods().GetPrototypeOf(function)
				if !IsConstructor((fun).ToValue()) {
					panic("TypeError: prototype is not a constructor")
				}
				result = fun.Construct(args, newTarget)
			} else {
				result = OrdinaryCreateFromConstructor(agent, newTarget, "%Object.prototype", nil)
			}
			return result.ToValue()
		}

		function = CreateBuiltinFunction(
			agent,
			defaultConstructor,
			0,
			CMString("constructor"),
			builtinFunctionArgs{
				prototype:     constructorParent,
				realm:         realm,
				isConstructor: true,
				additionalFields: &AdditionalFields{
					ClassConstructorFields: &ClassConstructorFields{},
				},
			})
	} else {
		constructorInfo := DefineMethod(agent, constructor.FunctionExpression, NewStringValue("constructor"), proto, constructorParent)
		F := constructorInfo.Closure
		MakeClassConstructor(F.(*ECMAScriptFunction))
		SetFunctionName(F, className.(PropertyKey), "")
		function = F
	}

	MakeConstructor(function, false, proto)
	if c.ClassHeritage != nil {
		if f, ok := function.(*ECMAScriptFunction); ok {
			f.ConstructorKind = ConstructorKindDerived
		} else if b, ok := function.(*BuiltinFunction); ok {
			b.AdditionalFields.ClassConstructorFields.ConstructorKind = ConstructorKindDerived
		} else {
			panic("unreachable")
		}
	}

	DefineMethodProperty(proto, NewStringPropertyKey("constructor"), function, false)

	elements := c.ClassBody.NonConstructorElements()

	instancePrivateMethods := &pkg.Stack[*PrivateElement]{}
	staticPrivateMethods := &pkg.Stack[*PrivateElement]{}

	instanceFields := make([]*ClassFieldDefinition, 0)

	staticClassFields := make([]*ClassFieldDefinition, 0)
	staticStaticBlocks := make([]*ClassStaticBlockDefinition, 0)

	for _, classElement := range elements {
		var err Value
		var result classEvaluationResult
		if !ClassElementIsStatic(classElement) {
			result, err = classElement.ClassElementEvaluation(vm, proto)
		} else {
			result, err = classElement.ClassElementEvaluation(vm, function)
		}
		if err != nil {
			agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = env
			agent.RunningExecutionContext().ECMAScriptCode.PrivateEnvironment = outerPrivateEnvironment
			vm.panic(err)
		}
		if result.classFieldDefinition != nil {
			if !ClassElementIsStatic(classElement) {
				instanceFields = append(instanceFields, result.classFieldDefinition)
			} else {
				staticClassFields = append(staticClassFields, result.classFieldDefinition)
			}
		} else if result.staticBlockDefinition != nil {
			staticStaticBlocks = append(staticStaticBlocks, result.staticBlockDefinition)
		} else if result.privateElement != nil {
			pme := result.privateElement
			Assert(pme.Kind == PrivateElementKindMethod || pme.Kind == PrivateElementKindAccessor)
			var container *pkg.Stack[*PrivateElement]
			if !ClassElementIsStatic(classElement) {
				container = instancePrivateMethods
			} else {
				container = staticPrivateMethods
			}

			element := pme
			var found bool
			for i, pe := range container.Data() {
				if pe.Key.Equal(element.Key) {
					found = true
					Assert(
						pe.Kind == PrivateElementKindAccessor &&
							pe.Kind == element.Kind)
					var combined *PrivateElement
					if element.Get == nil {
						combined = &PrivateElement{
							Key:  element.Key,
							Kind: PrivateElementKindAccessor,
							Set:  element.Set,
							Get:  pe.Get,
						}
					} else {
						combined = &PrivateElement{
							Key:  element.Key,
							Kind: PrivateElementKindAccessor,
							Set:  pe.Set,
							Get:  element.Get,
						}
					}
					container.Data()[i] = combined
				}
			}
			if !found {
				container.Push(pme)
			}
		}
	}

	agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = env
	if classBinding != "" {
		classEnv.InitializeBinding(classBinding, function.ToValue())
	}

	if ObjectIs[*ECMAScriptFunction](function) {
		e := ObjectAs[*ECMAScriptFunction](function)
		e.privateMethods = instancePrivateMethods.Data()
		function.(InternalSlotFields).SetFields(instanceFields)
	}

	for _, method := range staticPrivateMethods.Data() {
		function.PrivateMethodOrAccessorAdd(method)
	}
	for _, element := range staticClassFields {
		function.DefineField(element)
	}
	for _, block := range staticStaticBlocks {
		block.BodyFunction.ToValue().Call(function.ToValue(), nil)
	}

	agent.RunningExecutionContext().ECMAScriptCode.PrivateEnvironment = outerPrivateEnvironment

	return function, nil
}

type ClassBody struct {
	ASTNode
	ClassElementList *ClassElementList
}

// 15.7.3
func (c *ClassBody) ConstructorMethod() *MethodDefinition {
	for _, item := range c.ClassElementList.Items {
		if item.ClassElementKind() == ClassElementKindConstructorMethod {
			return item.(*ClassElementMethodDefinition).MethodDefinition
		}
	}
	return nil
}

func (c *ClassBody) PrivateBoundIdentifiers() (l []PrivateIdentifierName) {
	var propertyName PropertyName
	for _, item := range c.ClassElementList.Items {
		switch i := item.(type) {
		case *ClassElementStaticBlock, *ClassElementEmpty:
			// ignore
		case *ClassElementMethodDefinition:
			propertyName = i.MethodDefinition.PropertyName
		case *ClassElementStaticMethodDefinition:
			propertyName = i.MethodDefinition.PropertyName
		case *ClassElementFieldDefinition:
			propertyName = i.FieldDefinition.PropertyName
		case *ClassElementStaticFieldDefinition:
			propertyName = i.FieldDefinition.PropertyName
		}
	}
	switch p := propertyName.(type) {
	case *PropertyNameLiteralIdentifier:
		l = append(l, PrivateIdentifierName(p.Identifier))
	}

	return
}

// 15.7.5
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

// TODO(BM): use struct
// ClassElement [Yield, Await] :
// - MethodDefinition[?Yield, ?Await]
// - static MethodDefinition[?Yield, ?Await]
// - FieldDefinition[?Yield, ?Await] ;
// - static FieldDefinition[?Yield, ?Await] ;
// - ClassStaticBlock
// - ;
type ClassElement interface {
	ASTNode
	RuntimeSemanticsClassElementEvaluation
	ClassElementKind() ClassElementKind
}

// ClassElementIsStatic
// spec: 15.7.4
func ClassElementIsStatic(c ClassElement) bool {
	switch ce := c.(type) {
	case *ClassElementStaticMethodDefinition, *ClassElementStaticFieldDefinition:
		return true
	case *ClassElementFieldDefinition:
		return ce.IsStatic
	case *ClassElementMethodDefinition:
		return ce.IsStatic
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
	IsStatic        bool
}

var (
	_ RuntimeSemanticsClassElementEvaluation         = (*ClassElementFieldDefinition)(nil)
	_ RuntimeSemanticsClassFieldDefinitionEvaluation = (*ClassElementFieldDefinition)(nil)
)

func (c *ClassElementFieldDefinition) ClassElementEvaluation(vm *VM2, obj ObjectType) (result classEvaluationResult, err Value) {
	field, err := c.ClassFieldDefinitionEvaluation(vm, obj)
	result.classFieldDefinition = field
	return
}

func (c *ClassElementFieldDefinition) ClassFieldDefinitionEvaluation(vm *VM2, homeObject ObjectType) (field *ClassFieldDefinition, err Value) {
	agent := vm.agent
	realm := agent.CurrentRealm()
	var name PropertyKeyOrPrivateName
	value := GenerateAndRunBytecode(agent, c.FieldDefinition.PropertyName)
	if value.Data() != nil {
		name = ToPropertyKey(agent, value.Data())
	}
	var initializer ObjectType
	if c.FieldDefinition.Initializer != nil {
		formalParameterList := &FormalParameters{}
		env := agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := agent.RunningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := ""
		functionBody := &FunctionBody{
			StatementList: StatementList{
				&StatementListItemStatement{
					Statement: &StatementReturn{
						Expression: c.FieldDefinition.Initializer,
					},
				},
			},
			Strict: true,
			Type:   FunctionTypeNormal,
		}
		initializer = OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.FunctionPrototype,
			sourceText,
			formalParameterList,
			functionBody,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		MakeMethod(initializer.(*ECMAScriptFunction), homeObject)
		initializer.(InternalSlotClassFieldInitializerName).SetClassFieldInitializerName(name)
	} else {
		return &ClassFieldDefinition{
			Name: name,
		}, nil
	}

	return &ClassFieldDefinition{
		Name:        name,
		Initializer: initializer.(*ECMAScriptFunction),
	}, nil
}

func (c *ClassElementFieldDefinition) ClassElementKind() ClassElementKind {
	return ClassElementKindNonConstructorMethod
}

// Deprecated
// TODO(BM): use ClassElementFieldDefinition
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

func (c *ClassElementEmpty) ClassElementEvaluation(vm *VM2, function ObjectType) (result classEvaluationResult, err Value) {
	// return UNUSED
	return
}

// MARK: - ClassElement: StaticMethodDefinition

// Deprecated
// TODO(BM): remove
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
	IsStatic         bool
}

func (c *ClassElementMethodDefinition) ClassElementEvaluation(vm *VM2, obj ObjectType) (result classEvaluationResult, err Value) {
	privateElement, err := c.MethodDefinition.MethodDefinitionEvaluation(vm, obj, false)
	if err != nil {
		return result, err
	}
	result.privateElement = privateElement
	return
}

func (c *ClassElementMethodDefinition) ClassElementKind() ClassElementKind {
	switch pn := c.MethodDefinition.PropertyName.(type) {
	case LiteralPropertyName:
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

type LetOrConst int

const (
	LetOrConstLet LetOrConst = iota
	LetOrConstConst
)

var _ StaticSemanticsIsConstantDeclaration = LetOrConst(0)

func (l LetOrConst) IsConstantDeclaration() bool {
	return l == LetOrConstConst
}

// LexicalDeclaration [In, Yield, Await] :
// LetOrConst BindingList[?In, ?Yield, ?Await] ;
type LexicalDeclaration struct {
	Declaration
	Type        LetOrConst
	BindingList *BindingList
}

func (d *LexicalDeclaration) IsConstantDeclaration() bool {
	return d.Type == LetOrConstConst
}

func (d *LexicalDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	d.BindingList.Bytecode(e, c)
}

// Evaluation 14.3.1.2
func (d *LexicalDeclaration) Evaluation(vm *VM2) Value {
	d.BindingList.Evaluation(vm)
	// return EMPTY
	return UndefinedValue
}

func (d *LexicalDeclaration) String() string {
	return "LexicalDeclaration " + d.BindingList.String()
}

func (d *LexicalDeclaration) BoundNames() (l []IdentifierName) {
	return d.BindingList.BoundNames()
}

// BindingList [In, Yield, Await] :
// - LexicalBinding[?In, ?Yield, ?Await]
// - BindingList[?In, ?Yield, ?Await] , LexicalBinding[?In, ?Yield, ?Await]
type BindingList struct {
	ASTNode
	Items []*LexicalBinding
}

func (b *BindingList) BoundNames() (l []IdentifierName) {
	for _, item := range b.Items {
		l = append(l, item.BoundNames()...)
	}
	return
}

func (b *BindingList) Bytecode(e *Executable, c *BytecodeContext) {
	for _, item := range b.Items {
		item.Bytecode(e, c)
	}
}

func (b *BindingList) Evaluation(vm *VM2) Value {
	var list []Value
	for _, item := range b.Items {
		list = append(list, item.Evaluation(vm))
	}
	return NewListValue(list)
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

// LexicalBinding [In, Yield, Await] :
// - BindingIdentifier[?Yield, ?Await] Initializer[?In, ?Yield, ?Await] opt
// - BindingPattern[?Yield, ?Await] Initializer[?In, ?Yield, ?Await]
type LexicalBinding struct {
	ASTNode
	// BindingIdentifier
	Identifier     IdentifierName
	BindingPattern *BindingPattern
	Initializer    Expression
}

func (l *LexicalBinding) BoundNames() (list []IdentifierName) {
	if l.Identifier != "" {
		list = append(list, l.Identifier)
	} else {
		list = append(list, l.BindingPattern.BoundNames()...)
	}
	return
}

func (l *LexicalBinding) Bytecode(e *Executable, c *BytecodeContext) {
	if l.Initializer == nil {
		panic("unimplemented: bindingidentifier")
	} else {
		e.AddInstruction(InsLoad)
		e.AddInstruction(&IResolveBinding{
			Name:   l.Identifier,
			Strict: c.containedInStrictCode,
		})

		l.Initializer.Bytecode(e, c)
		if ExpressionAnalyze(l.Initializer, AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
		e.AddInstruction(&IInitializeReferencedBinding{})
		e.AddInstruction(InsStore)
	}
}

// 14.3.1.2
func (l *LexicalBinding) Evaluation(vm *VM2) Value {
	switch {
	case l.Identifier != "" && l.Initializer == nil:
		lhs := vm.agent.ResolveBinding(l.Identifier, nil, false)
		lhs.InitializeReferencedBinding(UndefinedValue)
	case l.Identifier != "":
		// LexicalBinding : BindingIdentifier Initializer
		lhs := vm.agent.ResolveBinding(l.Identifier, nil, false)
		if IsAnonymousFunctionDefinition(l.Initializer) {
			// FIXME: handle named evaluation
			panic("")
		} else {
			rhs := l.Initializer.Evaluation(vm)
			value := rhs.GetValue(vm.agent)
			lhs.InitializeReferencedBinding(value)
		}
		// return EMPTY
		return UndefinedValue
	}
	panic("unimplemented")
}

func (l *LexicalBinding) String() string {
	if l.Initializer != nil {
		return string(l.Identifier) + " = " + l.Initializer.String()
	}
	return string(l.Identifier)
}

// MARK: - FunctionDeclaration

// FunctionDeclaration [Yield, Await, Default] :
// - function BindingIdentifier[?Yield, ?Await] ( FormalParameters[~Yield, ~Await] ) {
// - FunctionBody[~Yield, ~Await] }
// - [+Default] function ( FormalParameters[~Yield, ~Await] ) {
// - FunctionBody[~Yield, ~Await] }
type FunctionDeclaration struct {
	ASTNode
	Identifier       IdentifierName
	Body             *FunctionBody
	FormalParameters *FormalParameters
	SourceText       string
}

var _ StaticSemanticsBoundNames = (*FunctionDeclaration)(nil)

func (f *FunctionDeclaration) BoundNames() (l []IdentifierName) {
	if f.Identifier != "" {
		l = append(l, f.Identifier)
	}
	return
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
	realm.GlobalEnv.ObjectRecord.BindingObject.Set(NewStringPropertyKey(string(f.Identifier)), function.ToValue(), setThrowTypeIgnore)
}

// Evaluation 15.2.6
func (f *FunctionDeclaration) Evaluation(vm *VM2) Value {
	// TODO(SM): check spec
	realm := vm.agent.CurrentRealm()
	env := realm.GlobalEnv
	function := f.instantiateOrdinaryFunctionObject(vm.agent, env, nil)
	realm.GlobalEnv.ObjectRecord.BindingObject.Set(NewStringPropertyKey(string(f.Identifier)), function.ToValue(), setThrowTypeIgnore)
	// return EMPTY
	return UndefinedValue
}

func (f *FunctionDeclaration) String() string {
	return "FunctionDeclaration " + string(f.Identifier)
}

// MARK: - BlockStatement

type BlockStatement interface {
	Statement
}

// TODO(SM): remove
type BlockStatementBlock struct {
	Block *Block
}

func (b *BlockStatementBlock) _statement() {}
func (b *BlockStatementBlock) VarScopedDeclarations() []*VariableDeclaration {
	return b.Block.StatementList.VarScopedDeclarations()
}

func (b *BlockStatementBlock) VarDeclaredNames() []IdentifierName {
	return b.Block.StatementList.VarDeclaredNames()
}

func (b *BlockStatementBlock) Bytecode(e *Executable, c *BytecodeContext) {
	b.Block.Bytecode(e, c)
}

func (b *BlockStatementBlock) Evaluation(vm *VM2) Value {
	return b.Block.Evaluation(vm)
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

// 14.2.2
func (b *Block) Evaluation(vm *VM2) Value {
	return b.StatementList.Evaluation(vm)
}

func (b *Block) String() string {
	return b.StatementList.String()
}

type StatementList []StatementListItem

func (s StatementList) TopLevelLexicallyDeclaredNames() (l []IdentifierName) {
	for _, item := range s {
		l = append(l, StatementListItemTopLevelLexicallyDeclaredNames(item)...)
	}
	return
}

func StatementListItemTopLevelLexicallyDeclaredNames(item StatementListItem) (l []IdentifierName) {
	switch t := item.(type) {
	case *StatementListItemDeclaration:
		return DeclarationBoundNames(t.Declaration)
	case *StatementListItemStatement:
		return
	}
	return
}

func (s StatementList) LexicallyDeclaredNames() (l []IdentifierName) {
	for _, item := range s {
		l = append(l, item.LexicallyDeclaredNames()...)
	}
	return
}

func (s StatementList) VarScopedDeclarations() (l []*VariableDeclaration) {
	for _, item := range s {
		l = append(l, item.VarScopedDeclarations()...)
	}
	return
}

func (s StatementList) VarDeclaredNames() (l []IdentifierName) {
	for _, item := range s {
		l = append(l, item.VarDeclaredNames()...)
	}
	return
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

func (s StatementList) Evaluation(vm *VM2) Value {
	var lastValue Value
	for _, item := range s {
		lastValue = item.Evaluation(vm)
		if vm.isReturn {
			return lastValue
		}
	}
	if lastValue == nil {
		return UndefinedValue
	}
	return lastValue
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
		str += "\n"
	}
	return str
}

// StatementListItem [Yield, Await, GetLastValue] :
// Statement[?Yield, ?Await, ?GetLastValue]
// Declaration[?Yield, ?Await]
type StatementListItem interface {
	ASTNode
	VarScopedDeclarations() []*VariableDeclaration
	VarDeclaredNames() []IdentifierName
	LexicallyDeclaredNames() []IdentifierName
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

func (s *StatementListItemStatement) TopLevelLexicallyDeclaredNames() (l []IdentifierName) {
	return
}

func (s *StatementListItemStatement) VarDeclaredNames() (l []IdentifierName) {
	return s.Statement.VarDeclaredNames()
}

func (s *StatementListItemStatement) VarScopedDeclarations() (l []*VariableDeclaration) {
	vars := s.Statement.VarScopedDeclarations()
	if vars == nil {
		return
	}
	return vars
}

func (s *StatementListItemStatement) Bytecode(e *Executable, c *BytecodeContext) {
	s.Statement.Bytecode(e, c)
}

func (s *StatementListItemStatement) Evaluation(vm *VM2) Value {
	return s.Statement.Evaluation(vm)
}

func (s *StatementListItemStatement) String() string {
	return s.Statement.String()
}

type StatementListItemDeclaration struct {
	StatementListItem
	Declaration Declaration
}

var _ ASTNode = (*StatementListItemDeclaration)(nil)

func (s *StatementListItemDeclaration) TopLevelLexicallyDeclaredNames() (l []IdentifierName) {
	return s.Declaration.BoundNames()
}

func (s *StatementListItemDeclaration) VarDeclaredNames() (l []IdentifierName) {
	return
}

func (s *StatementListItemDeclaration) VarScopedDeclarations() (l []*VariableDeclaration) {
	switch d := s.Declaration.(type) {
	case *LexicalDeclaration:
		for _, bindingItem := range d.BindingList.Items {
			l = append(l, &VariableDeclaration{
				BindingIdentifier: bindingItem.Identifier,
				Initializer:       bindingItem.Initializer,
			})
		}
	default:
	}
	return
}

func (s *StatementListItemDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	s.Declaration.Bytecode(e, c)
}

func (s *StatementListItemDeclaration) Evaluation(vm *VM2) Value {
	return s.Declaration.Evaluation(vm)
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

func (e *ExpressionStatement) Evaluation(vm *VM2) Value {
	return e.Expression.Evaluation(vm)
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

func (s *Script) Evaluation(vm *VM2) Value {
	return s.StatementList.Evaluation(vm)
}

func (s *Script) String() string {
	return s.StatementList.String()
}

func (s *Script) IsStrict() bool {
	return s.StatementList.ContainsDirective("use strict")
}

// MARK: - Module

// Module :
// - ModuleBody[opt]
type Module struct {
	ModuleItemList ModuleItemList
}

var (
	_ ASTNode                       = (*Module)(nil)
	_ StaticSemanticsModuleRequests = (*Module)(nil)
	_ StaticSemanticsImportEntries  = (*Module)(nil)
	_ StaticSemanticsExportEntries  = (*Module)(nil)
)

func (m *Module) Bytecode(e *Executable, c *BytecodeContext) {
	for _, moduleItem := range m.ModuleItemList {
		if stmt, ok := moduleItem.(*ModuleItemStatementListItem); ok {
			stmt.Bytecode(e, c)
		}
		if stmt, ok := moduleItem.(*ModuleItemExportDeclaration); ok {
			stmt.Bytecode(e, c)
		}

	}
}

// TODO: spec reference
func (m *Module) Evaluation(vm *VM2) Value {
	var list []Value
	for _, moduleItem := range m.ModuleItemList {
		switch stmt := moduleItem.(type) {
		case *ModuleItemImportDeclaration:
			panic("unimplemented")
		case *ModuleItemStatementListItem:
			list = append(list, stmt.Evaluation(vm))
		default:
			panic("unimplemented")
		}
	}
	return NewListValue(list)
}

func (m *Module) String() string {
	return m.ModuleItemList.String()
}

func (m *Module) moduleRequests() []string {
	return m.ModuleItemList.moduleRequests()
}

func (m *Module) importEntries() (l []ImportEntryRecord) {
	for _, item := range m.ModuleItemList {
		switch stmt := item.(type) {
		case *ModuleItemImportDeclaration:
			moduleRequest := stmt.ImportDeclaration.ModuleSpecifier.StringValue().String()
			if stmt.ImportDeclaration.ImportClause != nil {
				i := stmt.ImportDeclaration.ImportClause
				if i.ImportedDefaultBinding != "" {
					l = append(l, ImportEntryRecord{
						ImportName:    "default",
						ModuleRequest: moduleRequest,
						LocalName:     string(i.ImportedDefaultBinding),
					})
				} else if i.NamespaceImport != "" {
					l = append(l, ImportEntryRecord{
						ImportName:    ImportNameNamespaceObject,
						ModuleRequest: moduleRequest,
						LocalName:     string(i.NamespaceImport),
					})
				} else if i.NamedImports != nil {
					for _, specifier := range i.NamedImports.Items {
						if specifier.ModuleExportName != nil {
							var importName ImportName
							if specifier.ModuleExportName.IdentifierName != "" {
								importName = ImportName(specifier.ModuleExportName.IdentifierName)
							} else {
								importName = ImportName(specifier.ModuleExportName.StringLiteral.StringValue().String())
							}
							localName := string(specifier.ImportedBinding)
							l = append(l, ImportEntryRecord{
								ImportName:    importName,
								ModuleRequest: moduleRequest,
								LocalName:     localName,
							})
						} else {
							l = append(l, ImportEntryRecord{
								ImportName:    ImportName(specifier.ImportedBinding),
								ModuleRequest: moduleRequest,
								LocalName:     string(specifier.ImportedBinding),
							})
						}
					}
				} else {
					panic("unimplemented")
				}
			}
		default:
			continue
		}
	}
	return
}

func (m *Module) exportEntries() (l []ExportEntry) {
	for _, item := range m.ModuleItemList {
		switch stmt := item.(type) {
		case *ModuleItemStatementListItem, *ModuleItemImportDeclaration:
			continue
		case *ModuleItemExportDeclaration:
			l = append(l, stmt.exportEntries()...)
		default:
			panic("unimplemented")
		}
	}
	return
}

type ModuleItemList []ModuleItem

var _ StaticSemanticsModuleRequests = (ModuleItemList)(nil)

func (m ModuleItemList) String() string {
	var sb string
	for _, item := range m {
		sb += item.String()
		sb += "\n"
	}
	return sb
}

func (m ModuleItemList) moduleRequests() (l []string) {
	for _, item := range m {
		l = append(l, item.moduleRequests()...)
	}
	return
}

func (m ModuleItemList) VarScopedDeclarations() (l []*VariableDeclaration) {
	for _, item := range m {
		switch stmt := item.(type) {
		case *ModuleItemStatementListItem:
			l = append(l, stmt.StatementListItem.VarScopedDeclarations()...)
		case *ModuleItemExportDeclaration:
			if stmt.VariableStatement != nil {
				l = append(l, stmt.VariableStatement.VarScopedDeclarations()...)
			}
		case *ModuleItemImportDeclaration:
			continue
		default:
			panic("unimplemented")
		}
	}
	return
}

// MARK: - ModuleItem

// TODO: convert to struct
// ModuleItem :
// - ImportDeclaration
// - ExportDeclaration
// - StatementListItem[~Yield, +Await, ~GetLastValue]
type ModuleItem interface {
	String() string
	StaticSemanticsModuleRequests
}

// MARK: - ModuleItem: StatementListItem

type ModuleItemStatementListItem struct {
	StatementListItem StatementListItem
}

var _ ModuleItem = (*ModuleItemStatementListItem)(nil)

func (m *ModuleItemStatementListItem) moduleRequests() (l []string) {
	return
}

func (m *ModuleItemStatementListItem) Bytecode(e *Executable, c *BytecodeContext) {
	m.StatementListItem.Bytecode(e, c)
}

func (m *ModuleItemStatementListItem) Evaluation(vm *VM2) Value {
	return m.StatementListItem.Evaluation(vm)
}

func (m *ModuleItemStatementListItem) String() string {
	return m.StatementListItem.String()
}

// MARK: - ModuleItem: ImportDeclaration

type ModuleItemImportDeclaration struct {
	ImportDeclaration *ImportDeclaration
}

var _ ModuleItem = (*ModuleItemImportDeclaration)(nil)

func (m *ModuleItemImportDeclaration) moduleRequests() (l []string) {
	l = append(l, m.ImportDeclaration.ModuleSpecifier.StringValue().String())
	return
}

func (m *ModuleItemImportDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	panic("unimplemented")
}

func (m *ModuleItemImportDeclaration) String() string {
	return m.ImportDeclaration.String()
}

// MARK: - ModuleItem: ExportDeclaration

// TODO: rename
// ExportDeclaration :
//   - export ExportFromClause FromClause ;
//   - export NamedExports ;
//   - export VariableStatement[~Yield, +Await]
//   - export Declaration[~Yield, +Await]
//   - export default HoistableDeclaration[~Yield, +Await, +Default]
//   - export default ClassDeclaration[~Yield, +Await, +Default]
//   - export default [lookahead ∉ { function, async [no LineTerminator here] function,
//     class}] AssignmentExpression[+In, ~Yield, +Await] ;
type ModuleItemExportDeclaration struct {
	ExportFrom                  *ExportFrom
	NamedExports                *NamedExports
	Declaration                 Declaration
	VariableStatement           *StatementVariable
	DefaultHoistableDeclaration DeclarationHoistable
	DefaultClassDeclaration     *ClassDeclaration
	DefaultExpression           Expression
}

var (
	_ ModuleItem                   = (*ModuleItemExportDeclaration)(nil)
	_ StaticSemanticsExportEntries = (*ModuleItemExportDeclaration)(nil)
	_ ASTNode                      = (*ModuleItemExportDeclaration)(nil)
)

func (m *ModuleItemExportDeclaration) Bytecode(e *Executable, c *BytecodeContext) {
	switch {
	case m.ExportFrom != nil || m.NamedExports != nil:
		return
	case m.Declaration != nil:
		m.Declaration.Bytecode(e, c)
	case m.VariableStatement != nil:
		m.VariableStatement.Bytecode(e, c)
	case m.DefaultHoistableDeclaration != nil:
		m.DefaultHoistableDeclaration.Bytecode(e, c)
	case m.DefaultClassDeclaration != nil:
		m.DefaultClassDeclaration.Bytecode(e, c)
	case m.DefaultExpression != nil:
		m.DefaultExpression.Bytecode(e, c)
	default:
		panic("unreachable")
	}
}

func (m *ModuleItemExportDeclaration) Evaluation(vm *VM2) Value {
	panic("unimplemented")
}

func (m *ModuleItemExportDeclaration) String() string {
	switch {
	case m.ExportFrom != nil:
		panic("unimplemented")
	case m.NamedExports != nil:
		panic("unimplemented")
	case m.Declaration != nil:
		return m.Declaration.String()
	case m.VariableStatement != nil:
		return m.VariableStatement.String()
	case m.DefaultHoistableDeclaration != nil:
		return m.DefaultHoistableDeclaration.String()
	case m.DefaultClassDeclaration != nil:
		return m.DefaultClassDeclaration.String()
	case m.DefaultExpression != nil:
		return m.DefaultExpression.String()
	}
	return "ModuleItemExportDeclaration"
}

func (m *ModuleItemExportDeclaration) moduleRequests() (l []string) {
	if m.ExportFrom != nil {
		l = append(l, m.ExportFrom.ModuleSpecifier.StringValue().String())
	}
	return
}

func (m *ModuleItemExportDeclaration) exportEntries() (l []ExportEntry) {
	if m.ExportFrom != nil {
		panic("unimplemented")
	} else if m.NamedExports != nil {
		panic("unimplemented")
	} else if m.Declaration != nil {
		boundNames := m.Declaration.BoundNames()
		for _, name := range boundNames {
			l = append(l, ExportEntry{
				ExportName: string(name),
				LocalName:  string(name),
			})
		}
	} else if m.VariableStatement != nil {
		panic("unimplemented")
	} else if m.DefaultHoistableDeclaration != nil {
		panic("unimplemented")
	} else if m.DefaultClassDeclaration != nil {
		panic("unimplemented")
	} else if m.DefaultExpression != nil {
		panic("unimplemented")
	}
	return
}

type ExportFrom struct {
	ExportFromClause *ExportFromClause
	ModuleSpecifier  *LiteralString
}

// Enum
type ExportFromClause struct {
	Star         bool
	StarAs       *ModuleExportName
	NamedExports *NamedExports
}

type NamedExports struct {
	ExportsList *ExportsList
}
type ExportsList struct {
	Items []*ExportSpecifier
}
type ExportSpecifier struct {
	Name  *ModuleExportName
	Alias *ModuleExportName
}

// ModuleExportName :
// - IdentifierName
// - StringLiteral
type ModuleExportName struct {
	IdentifierName IdentifierName
	StringLiteral  *LiteralString
}

func (m *ModuleExportName) String() string {
	if m.IdentifierName != "" {
		return string(m.IdentifierName)
	}
	return m.StringLiteral.String()
}

func (m *ModuleItemExportDeclaration) _moduleItem() {}

// MARK: - Import

// ImportDeclaration :
// - import ImportClause FromClause ;
// - import ModuleSpecifier ;
type ImportDeclaration struct {
	// FromClause :
	// - from ModuleSpecifier
	// optional
	ImportClause *ImportClause
	// Not null
	ModuleSpecifier *LiteralString
}

func (i *ImportDeclaration) BoundNames() (l []IdentifierName) {
	if i.ImportClause != nil {
		return i.ImportClause.BoundNames()
	}
	return
}

// 16.2.2.2

func (i *ImportDeclaration) String() string {
	if i.ImportClause != nil {
		return "ImportDeclaration " + i.ImportClause.String()
	}
	return "ImportDeclaration"
}

// ImportClause :
// - ImportedDefaultBinding
// - NameSpaceImport
// - NamedImports
// - ImportedDefaultBinding, NameSpaceImport
// - ImportedDefaultBinding, NamedImports
type ImportClause struct {
	ImportedDefaultBinding IdentifierName
	NamespaceImport        IdentifierName
	// NamedImports :
	// - { }
	// - { ImportsList }
	// - { ImportsList, }
	NamedImports *ImportsList
}

func (i *ImportClause) BoundNames() (l []IdentifierName) {
	if i.ImportedDefaultBinding != "" {
		l = append(l, i.ImportedDefaultBinding)
	} else if i.NamedImports != nil {
		for _, s := range i.NamedImports.Items {
			if s.ImportedBinding != "" {
				l = append(l, s.ImportedBinding)
			}
		}
	} else {
		panic("unimplemented")
	}
	return
}

func (i *ImportClause) String() string {
	var sb string
	if i.ImportedDefaultBinding != "" {
		sb += string(i.ImportedDefaultBinding)
	}
	if i.NamespaceImport != "" {
		sb += " " + string(i.NamespaceImport)
	}
	if i.NamedImports != nil {
		sb += " " + i.NamedImports.String()
	}
	return sb
}

// ImportsList :
// - ImportSpecifier
// - ImportsList, ImportSpecifier
type ImportsList struct {
	Items []*ImportSpecifier
}

func (i *ImportsList) String() string {
	var sb string
	for i, item := range i.Items {
		if i != 0 {
			sb += ", "
		}
		sb += item.String()
	}
	return sb
}

// ImportSpecifier :
// - ImportedBinding
// - ModuleExportName as ImportedBinding
type ImportSpecifier struct {
	// ImportedBinding :
	// - BindingIdentifier[~Yield, +Await]
	// TODO: change to BindingIdentifier
	ImportedBinding  IdentifierName
	ModuleExportName *ModuleExportName
}

func (i *ImportSpecifier) String() string {
	if i.ModuleExportName != nil {
		return i.ModuleExportName.String() + " as " + string(i.ImportedBinding)
	}
	return string(i.ImportedBinding)
}

// MARK: - YieldExpression

// YieldExpression [In, Await] :
// yield
// yield [no LineTerminator here] AssignmentExpression[?In, +Yield, ?Await]
// yield [no LineTerminator here] * AssignmentExpression[?In, +Yield, ?Await]
type YieldExpression struct {
	Expression
	// TODO: Check AssignmentExpression type
	AssignmentExpression Expression
}

var _ Expression = (*YieldExpression)(nil)

func (y *YieldExpression) String() string {
	if y.AssignmentExpression != nil {
		return "YieldExpression " + y.AssignmentExpression.String()
	}
	return "YieldExpression"
}

func (y *YieldExpression) Bytecode(e *Executable, c *BytecodeContext) {
	if y.AssignmentExpression != nil {
		y.AssignmentExpression.Bytecode(e, c)
		if ExpressionAnalyze(y.AssignmentExpression, AnalyzeQueryIsReference) {
			e.AddInstruction(InsGetValue)
		}
	} else {
		e.AddInstruction(&IStoreConstant{Value: UndefinedValue})
	}
	e.AddInstruction(InsYield)
}
