package coldmoon

import "fmt"

type Instruction interface {
	String() string
}

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
