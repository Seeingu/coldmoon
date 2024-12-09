package coldmoon

import (
	"fmt"
)

type Executable struct {
	Instructions []Instruction
	debugInfo    map[int]string
}

func NewExecutable() *Executable {
	return &Executable{
		debugInfo: make(map[int]string),
	}
}

func (e *Executable) AddInstruction(i Instruction) {
	e.Instructions = append(e.Instructions, i)
}

func (e *Executable) AddDebug(m string) {
	e.debugInfo[len(e.Instructions)] = m
}

func (e *Executable) AddInstructionDebug(i Instruction, m string) {
	e.debugInfo[len(e.Instructions)] = m
	e.AddInstruction(i)
}

func (e *Executable) String() string {
	sb := ""
	for i, ins := range e.Instructions {
		if m, ok := e.debugInfo[i]; ok {
			sb += fmt.Sprintf("INFO: %s\n", m)
		}
		sb += fmt.Sprintf("\t%d: %s\n", i, ins.String())
	}
	return sb
}
