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
