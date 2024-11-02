package coldmoon

import (
	"fmt"
)

type Executable struct {
	Instructions []Instruction
}

func NewExecutable() *Executable {
	return &Executable{}
}

func (e *Executable) AddInstruction(i Instruction) {
	e.Instructions = append(e.Instructions, i)
}

func (e *Executable) String() string {
	sb := ""
	for i, ins := range e.Instructions {
		sb += fmt.Sprintf("\t%d: %s\n", i, ins.String())
	}
	return sb
}
