package coldmoon

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

func (e *Executable) String() string {
	i := 0
	sb := ""
	for i < len(e.Instructions) {
		ins := e.Instructions[i]
		sb += ins.String() + "\n"
		i += 1
	}
	return sb
}
