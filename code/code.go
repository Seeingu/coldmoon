package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// MARK: Instruction

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			_, _ = fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		_, _ = fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += 1 + read
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand count mismatch: %d != %d", len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s: %d", def.Name, operandCount)
}

// MARK: Op

type Opcode byte

const (
	OpConstant Opcode = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpPop
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterEqual
	OpLessThan
	OpLessEqual
	OpNegate
	OpNot
	OpJump
	OpJumpFalse
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal
	OpArray
	OpObject
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
	OpGetBuiltin
	OpClosure
	OpGetFree
	OpCurrentClosure
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant:       {"OpConstant", []int{2}},
	OpAdd:            {"OpAdd", []int{}},
	OpSub:            {"OpSub", []int{}},
	OpMul:            {"OpMul", []int{}},
	OpDiv:            {"OpDiv", []int{}},
	OpPop:            {"OpPop", []int{}},
	OpTrue:           {"OpTrue", []int{}},
	OpFalse:          {"OpFalse", []int{}},
	OpEqual:          {"OpEqual", []int{}},
	OpNotEqual:       {"OpNotEqual", []int{}},
	OpGreaterEqual:   {"OpGreaterEqual", []int{}},
	OpGreaterThan:    {"OpGreaterThan", []int{}},
	OpLessThan:       {"OpLessThan", []int{}},
	OpLessEqual:      {"OpLessEqual", []int{}},
	OpNegate:         {"OpNegate", []int{}},
	OpNot:            {"OpNot", []int{}},
	OpJump:           {"OpJump", []int{2}},
	OpJumpFalse:      {"OpJumpFalse", []int{2}},
	OpNull:           {"OpNull", []int{}},
	OpGetGlobal:      {"OpGetGlobal", []int{2}},
	OpSetGlobal:      {"OpSetGlobal", []int{2}},
	OpGetLocal:       {"OpGetLocal", []int{1}},
	OpSetLocal:       {"OpSetLocal", []int{1}},
	OpArray:          {"OpArray", []int{2}},
	OpObject:         {"OpObject", []int{2}},
	OpIndex:          {"OpIndex", []int{}},
	OpCall:           {"OpCall", []int{1}},
	OpReturnValue:    {"OpReturnValue", []int{}},
	OpReturn:         {"OpReturn", []int{}},
	OpGetBuiltin:     {"OpGetBuiltin", []int{1}},
	OpClosure:        {"OpClosure", []int{2, 1}},
	OpGetFree:        {"OpGetFree", []int{1}},
	OpCurrentClosure: {"OpCurrentClosure", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode: %d undefined", op)
	}
	return def, nil
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}
	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}
