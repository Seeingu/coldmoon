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

// MARK: - Resolve

type IResolveThisBinding struct {
	Instruction
}

func (i *IResolveThisBinding) String() string {
	return "IResolveThisBinding"
}

type IResolveBinding struct {
	Instruction
	Name IdentifierName
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

// MARK: - Instruction Constant

var InsLoad = &ILoad{}
var InsThrow = &IThrow{}
var InsLoadThisValue = &ILoadThisValue{}
var InsGetValue = &IGetValue{}
var InsTypeof = &ITypeof{}
