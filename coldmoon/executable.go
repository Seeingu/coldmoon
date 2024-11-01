package coldmoon

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
	sb := ""
	for _, ins := range e.Instructions {
		sb += ins.String() + "\n"
	}
	return sb
}
