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

// MARK: - Return

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

type ISetEvaluationContextReference struct {
	Instruction
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

// MARK: - Throw

type IThrow struct {
	Instruction
}

func (i *IThrow) String() string {
	return "IThrow"
}

type IGetValue struct {
	Instruction
}

type ILoadThisValue struct {
	Instruction
}

func (i *ILoadThisValue) String() string {
	return "ILoadThisValue"
}

// MARK: - Typeof

type ITypeof struct {
	Instruction
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

// MARK: - Instruction Constant

var InsLoad = &ILoad{}
var InsThrow = &IThrow{}
var InsLoadThisValue = &ILoadThisValue{}
var InsGetValue = &IGetValue{}
var InsTypeof = &ITypeof{}
var InsToNumber = &IToNumber{}
var InsToNumeric = &IToNumeric{}
var InsUnaryMinus = &IUnaryMinus{}
var InsReturn = &IReturn{}
var InsStore = &IStore{}
var InsLessThan = &ILessThan{}
var InsLessThanEquals = &ILessThanEquals{}
var InsGreaterThan = &IGreaterThan{}
var InsGreaterThanEquals = &IGreaterThanEquals{}
var InsInstanceOf = &IInstanceOf{}
var InsHasProperty = &IHasProperty{}
var InsStrictlyEqual = &IStrictlyEqual{}
var InsLooselyEqual = &ILooselyEqual{}
var InsLogicalNot = &ILogicalNot{}
