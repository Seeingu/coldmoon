package code

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
		Make(OpSetLocal, 1),
		Make(OpGetLocal, 255),
		Make(OpClosure, 2, 1),
	}

	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
0007 OpSetLocal 1
0009 OpGetLocal 255
0011 OpClosure 2 1
`

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	assert.Equal(t, concatted.String(), expected)
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}

	for _, tt := range tests {
		ins := Make(tt.op, tt.operands...)
		def, err := Lookup(byte(tt.op))
		assert.NoError(t, err, "definition not found")

		operandsRead, n := ReadOperands(def, ins[1:])
		assert.Equal(t, tt.bytesRead, n)
		for i, want := range tt.operands {
			assert.Equal(t, want, operandsRead[i])
		}
	}
}
