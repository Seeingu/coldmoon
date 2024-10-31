package coldmoon

type Instruction interface {
}

type ILoad struct {
	Instruction
}

type ILoadConstant struct {
	Instruction
	Value Value
}

type IStore struct {
	Instruction
}

type IStoreConstant struct {
	Instruction
	Value Value
}

type Executable struct {
	Instructions []Instruction
	Constants    []Value
}

func NewExecutable() *Executable {
	return &Executable{}
}

func (e *Executable) AddInstruction(i Instruction) {
	e.Instructions = append(e.Instructions, i)
}

func (e *Executable) AddConstant(i Instruction, v Value) {
	e.AddInstruction(i)
	e.Constants = append(e.Constants, v)
	e.AddInstruction(len(e.Constants) - 1)
}
