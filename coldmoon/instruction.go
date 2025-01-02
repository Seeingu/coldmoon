package coldmoon

import "fmt"

type Instruction interface {
	String() string
}

// MARK: - Load, Store

type ILoad struct {
	Instruction
}

func (i *ILoad) String() string {
	return "ILoad"
}

type ILoadConstant struct {
	Instruction
	Value Value
}

func (i *ILoadConstant) String() string {
	return "ILoadConstant " + i.Value.String()
}

type IStore struct {
	Instruction
}

func (i *IStore) String() string {
	return "IStore"
}

type IStoreConstant struct {
	Instruction
	Value Value
}

func (i *IStoreConstant) String() string {
	return "IStoreConstant " + i.Value.String()
}

// MARK: - CompletionTypeReturn

type IReturn struct {
	Instruction
}

func (i *IReturn) String() string {
	return "IReturn"
}

// MARK: - Resolve

type IResolveThisBinding struct {
	Instruction
}

func (i *IResolveThisBinding) String() string {
	return "IResolveThisBinding"
}

type IResolveBinding struct {
	Instruction
	Name   IdentifierName
	Strict bool
}

func (i *IResolveBinding) String() string {
	return "IResolveBinding " + string(i.Name)
}

type IPushReference struct {
	Instruction
}

func (i *IPushReference) String() string {
	return "IPushReference"
}

type IPopReference struct {
	Instruction
}

func (i *IPopReference) String() string {
	return "IPopReference"
}

type IInitializeReferencedBinding struct {
	Instruction
}

func (i *IInitializeReferencedBinding) String() string {
	return "IInitializeReferencedBinding"
}

// MARK: - Exception

type IPopExceptionJumpTarget struct {
	Instruction
}

func (i *IPopExceptionJumpTarget) String() string {
	return "IPopExceptionJumpTarget"
}

type IPushExceptionJumpTarget struct {
	Instruction
	Target int
}

func (i *IPushExceptionJumpTarget) String() string {
	return "IPushExceptionJumpTarget " + fmt.Sprintf("%d", i.Target)
}

type IRethrowExceptionIfAny struct {
	Instruction
}

func (i *IRethrowExceptionIfAny) String() string {
	return "IRethrowExceptionIfAny"
}

// MARK: - Jump, JumpIfTrue

type IJump struct {
	Instruction
	Target int
}

func (i *IJump) String() string {
	return fmt.Sprintf("IJump %d", i.Target)
}

type IJumpIfTrue struct {
	Instruction
	Target     int
	TargetElse int
}

func (i *IJumpIfTrue) String() string {
	return fmt.Sprintf("IJumpIfTrue %d %d", i.Target, i.TargetElse)
}

// MARK: - CompletionTypeThrow

type IThrow struct {
	Instruction
}

func (i *IThrow) String() string {
	return "IThrow"
}

type IGetValue struct {
	Instruction
}

func (i *IGetValue) String() string {
	return "IGetValue"
}

type ILoadThisValue struct {
	Instruction
}

func (i *ILoadThisValue) String() string {
	return "ILoadThisValue"
}

type ILoadThisValueSuper struct {
	Instruction
}

func (i *ILoadThisValueSuper) String() string {
	return "ILoadThisValueSuper"
}

// MARK: - Typeof

type ITypeof struct {
	Instruction
}

func (i *ITypeof) String() string {
	return "ITypeof"
}

// MARK: - Call

type ICall struct {
	Instruction
	ArgumentCount int
	Strict        bool
}

func (i *ICall) String() string {
	return "ICall " + fmt.Sprintf("%d", i.ArgumentCount)
}

// MARK: - Property Access

type IEvaluatePropertyAccessWithExpressionKey struct {
	Instruction
	Strict bool
}

func (i *IEvaluatePropertyAccessWithExpressionKey) String() string {
	return "IEvaluatePropertyAccessWithExpressionKey " + fmt.Sprintf("%t", i.Strict)
}

type IEvaluatePropertyAccessWithIdentifierKey struct {
	Instruction
	Strict bool
	Name   IdentifierName
}

func (i *IEvaluatePropertyAccessWithIdentifierKey) String() string {
	return "IEvaluatePropertyAccessWithIdentifierKey " + fmt.Sprintf("%t %s", i.Strict, i.Name)
}

// MARK: - ToNumber

type IToNumber struct {
	Instruction
}

func (i *IToNumber) String() string {
	return "IToNumber"
}

type IToNumeric struct {
	Instruction
}

func (i *IToNumeric) String() string {
	return "IToNumeric"
}

type IUnaryMinus struct {
	Instruction
}

func (i *IUnaryMinus) String() string {
	return "IUnaryMinus"
}

// MARK: - Not

type ILogicalNot struct {
	Instruction
}

func (i *ILogicalNot) String() string {
	return "ILogicalNot"
}

type IBitwiseNot struct {
	Instruction
}

func (i *IBitwiseNot) String() string {
	return "IBitwiseNot"
}

// MARK: - Function

type IInstantiateOrdinaryFunctionExpression struct {
	Instruction
	FunctionExpression *PrimaryExpressionFunctionExpression
}

func (i *IInstantiateOrdinaryFunctionExpression) String() string {
	return "IInstantiateOrdinaryFunctionExpression " + i.FunctionExpression.String()
}

type IInstantiateArrowFunctionExpression struct {
	Instruction
	FunctionExpression *PrimaryExpressionArrowFunction
}

func (i *IInstantiateArrowFunctionExpression) String() string {
	return "IInstantiateArrowFunctionExpression"
}

// MARK: - Array

type IArrayCreate struct {
	Instruction
}

func (i *IArrayCreate) String() string {
	return "IArrayCreate"
}

type IArraySetValue struct {
	Instruction
	Index int
}

func (i *IArraySetValue) String() string {
	return "IArraySetValue"
}

type IArraySetLength struct {
	Instruction
	Length int
}

func (i *IArraySetLength) String() string {
	return "IArraySetLength " + fmt.Sprintf("%d", i.Length)
}

type IArrayPushValue struct {
	Instruction
}

func (i *IArrayPushValue) String() string {
	return "IArrayPushValue"
}

type IArraySpread struct {
	Instruction
}

func (i *IArraySpread) String() string {
	return "IArraySpread"
}

// MARK: - Object

type IObjectCreate struct {
	Instruction
}

func (i *IObjectCreate) String() string {
	return "IObjectCreate"
}

type IObjectSetProperty struct {
	Instruction
}

func (i *IObjectSetProperty) String() string {
	return "IObjectSetProperty"
}

type IObjectDefineMethod struct {
	Instruction
	MethodType               MethodDefinitionType
	FunctionExpression       *PrimaryExpressionFunctionExpression
	GeneratorExpression      *PrimaryExpressionGeneratorExpression
	AsyncFunctionExpression  *PrimaryExpressionAsyncFunctionExpression
	AsyncGeneratorExpression *PrimaryExpressionAsyncGeneratorExpression
}

func (i *IObjectDefineMethod) String() string {
	return "IObjectDefineMethod"
}

type IObjectSpreadValue struct {
	Instruction
}

func (i *IObjectSpreadValue) String() string {
	return "IObjectSpreadValue"
}

// MARK: - Relation

type IGreaterThan struct {
	Instruction
}

func (i *IGreaterThan) String() string {
	return "IGreaterThan"
}

type IGreaterThanEquals struct {
	Instruction
}

func (i *IGreaterThanEquals) String() string {
	return "IGreaterThanEquals"
}

type IHasProperty struct {
	Instruction
}

func (i *IHasProperty) String() string {
	return "IHasProperty"
}

type IInstanceOf struct {
	Instruction
}

func (i *IInstanceOf) String() string {
	return "IInstanceOf"
}

type ILessThan struct {
	Instruction
}

func (i *ILessThan) String() string {
	return "ILessThan"
}

type ILessThanEquals struct {
	Instruction
}

func (i *ILessThanEquals) String() string {
	return "ILessThanEquals"
}

// MARK: - Equality

type ILooselyEqual struct {
	Instruction
}

func (i *ILooselyEqual) String() string {
	return "ILooselyEqual"
}

type IStrictlyEqual struct {
	Instruction
}

func (i *IStrictlyEqual) String() string {
	return "IStrictlyEqual"
}

// MARK: - Logical

type IIncrement struct {
	Instruction
}

func (i *IIncrement) String() string {
	return "IIncrement"
}

type IDecrement struct {
	Instruction
}

func (i *IDecrement) String() string {
	return "IDecrement"
}

// MARK: - Numeric Binary

type IApplyStringOrNumericBinaryOperator struct {
	Instruction
	Operator BinaryOperator
}

func (i *IApplyStringOrNumericBinaryOperator) String() string {
	return "IApplyStringOrNumericBinaryOperator " + i.Operator.String()
}

// MARK: - New

type INew struct {
	Instruction
	ArgumentCount int
}

func (i *INew) String() string {
	return "INew " + fmt.Sprintf("%d", i.ArgumentCount)
}

// MARK: - Put Value

type IPutValue struct {
	Instruction
}

func (i *IPutValue) String() string {
	return "IPutValue"
}

// MARK: - Catch

type ICreateCatchBinding struct {
	Instruction
	IdentifierName IdentifierName
}

func (i *ICreateCatchBinding) String() string {
	return "ICreateCatchBinding " + string(i.IdentifierName)
}

// MARK: - Delete

type IDelete struct {
	Instruction
}

func (i *IDelete) String() string {
	return "IDelete"
}

// MARK: - Generator

type IInstantiateGeneratorFunctionExpression struct {
	Instruction
	FunctionExpression *PrimaryExpressionGeneratorExpression
}

func (i *IInstantiateGeneratorFunctionExpression) String() string {
	return "IInstantiateGeneratorFunctionExpression " + i.FunctionExpression.String()
}

type IInstantiateAsyncGeneratorFunctionExpression struct {
	Instruction
	FunctionExpression *PrimaryExpressionAsyncGeneratorExpression
}

func (i *IInstantiateAsyncGeneratorFunctionExpression) String() string {
	return "IInstantiateAsyncGeneratorFunctionExpression " + i.FunctionExpression.String()
}

type IInstantiateAsyncArrowFunctionExpression struct {
	Instruction
	FunctionExpression *PrimaryExpressionAsyncArrowFunction
}

func (i *IInstantiateAsyncArrowFunctionExpression) String() string {
	return "IInstantiateAsyncArrowFunctionExpression"
}

type IInstantiateAsyncFunctionExpression struct {
	Instruction
	FunctionExpression *PrimaryExpressionAsyncFunctionExpression
}

func (i *IInstantiateAsyncFunctionExpression) String() string {
	return "IInstantiateAsyncFunctionExpression " + i.FunctionExpression.String()
}

type IYield struct {
	Instruction
}

func (i *IYield) String() string {
	return "IYield"
}

// MARK: - GetNewTarget

type IGetNewTarget struct {
	Instruction
}

func (i *IGetNewTarget) String() string {
	return "IGetNewTarget"
}

// MARK: - RegExp

type IRegExpCreate struct {
	Instruction
}

func (i *IRegExpCreate) String() string {
	return "IRegExpCreate"
}

// MARK: - Class

type IBindingClassDeclarationEvaluation struct {
	Instruction
	ClassDeclaration *DeclarationClass
}

func (i *IBindingClassDeclarationEvaluation) String() string {
	return "IBindingClassDeclarationEvaluation"
}

type IClassDefinitionEvaluation struct {
	Instruction
	ClassExpression *PrimaryExpressionClassExpression
}

func (i *IClassDefinitionEvaluation) String() string {
	return "IClassDefinitionEvaluation"
}

// MARK: - Super

type IMakeSuperPropertyReference struct {
	Instruction
	Strict bool
}

func (i *IMakeSuperPropertyReference) String() string {
	return "IMakeSuperPropertyReference"
}

type IEvaluateSuperCall struct {
	Instruction
	ArgumentCount int
}

func (i *IEvaluateSuperCall) String() string {
	return "IEvaluateSuperCall"
}

// MARK: - Import

type IGetOrCreateImportMeta struct {
	Instruction
}

func (i *IGetOrCreateImportMeta) String() string {
	return "GetOrCreateImportMeta"
}

type IImportCall struct {
	Instruction
}

func (i *IImportCall) String() string {
	return "IEvaluateImportCall"
}

// MARK: - Iteration

type IForInIterator struct {
	Instruction
}

func (i *IForInIterator) String() string {
	return "IForInIterator"
}

type IGetIterator struct {
	Instruction
	IteratorKind IteratorKind
}

func (i *IGetIterator) String() string {
	return "IGetIterator"
}

type ILoadIterator struct {
	Instruction
}

func (i *ILoadIterator) String() string {
	return "ILoadIterator"
}

// 14.7.5.4
type IForDeclarationBindingInstantiation struct {
	Instruction
	LexicalDeclaration *DeclarationLexical
}

func (i *IForDeclarationBindingInstantiation) String() string {
	return "IForDeclarationBindingInstantiation"
}

// MARK: - Env

type IRestoreLexicalEnvironment struct {
	Instruction
}

func (i *IRestoreLexicalEnvironment) String() string {
	return "IRestoreLexicalEnvironment"
}

type IPushLexicalEnvironment struct {
	Instruction
}

func (i *IPushLexicalEnvironment) String() string {
	return "IPushLexicalEnvironment"
}

type IPopLexicalEnvironment struct {
	Instruction
}

func (i *IPopLexicalEnvironment) String() string {
	return "IPopLexicalEnvironment"
}

// MARK: - Instruction Constant

var (
	InsLoad                      = &ILoad{}
	InsThrow                     = &IThrow{}
	InsLoadThisValue             = &ILoadThisValue{}
	InsLoadThisValueSuper        = &ILoadThisValueSuper{}
	InsGetValue                  = &IGetValue{}
	InsTypeof                    = &ITypeof{}
	InsToNumber                  = &IToNumber{}
	InsToNumeric                 = &IToNumeric{}
	InsUnaryMinus                = &IUnaryMinus{}
	InsReturn                    = &IReturn{}
	InsStore                     = &IStore{}
	InsLessThan                  = &ILessThan{}
	InsLessThanEquals            = &ILessThanEquals{}
	InsGreaterThan               = &IGreaterThan{}
	InsGreaterThanEquals         = &IGreaterThanEquals{}
	InsInstanceOf                = &IInstanceOf{}
	InsHasProperty               = &IHasProperty{}
	InsStrictlyEqual             = &IStrictlyEqual{}
	InsLooselyEqual              = &ILooselyEqual{}
	InsLogicalNot                = &ILogicalNot{}
	InsPushReference             = &IPushReference{}
	InsPopReference              = &IPopReference{}
	InsPutValue                  = &IPutValue{}
	InsGetNewTarget              = &IGetNewTarget{}
	InsRegExpCreate              = &IRegExpCreate{}
	InsRestoreLexicalEnvironment = &IRestoreLexicalEnvironment{}
	InsPushLexicalEnvironment    = &IPushLexicalEnvironment{}
	InsYield                     = &IYield{}
)
