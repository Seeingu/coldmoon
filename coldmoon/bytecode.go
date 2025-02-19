package coldmoon

import (
	"fmt"

	"github.com/Seeingu/coldmoon/pkg"
)

type BytecodeContext struct {
	vm                        *VM
	vm2                       *VM2
	exe                       *Executable
	agent                     *Agent
	containedInStrictCode     bool
	labelContinueJumpIndexMap map[string]pkg.Stack[*IJump]
	labelBreakJumpIndexMap    map[string]pkg.Stack[*IJump]
	continueJumpIndices       pkg.Stack[*IJump]
	breakJumpIndices          pkg.Stack[*IJump]
	Label                     string
}

// Run Instructions from last `ip` position from `VM`
func (b *BytecodeContext) Run() CompletionValue {
	if Debug.PrintBytecode {
		fmt.Println("Executable: ", b.exe.String())
	}
	result := b.vm.Run(b.exe)
	if result.Data() != nil {
		fmt.Println("Result: ", result.Data().String())
	} else if result.IsError() {
		fmt.Println("Error Result: ", result.Error().String())
	}
	return result
}

func (b *BytecodeContext) IsFinished() bool {
	return b.vm.ip >= len(b.exe.Instructions)
}

func GenerateBytecode(agent *Agent, node ASTNode) *BytecodeContext {
	vm := NewVM(agent)
	exe := NewExecutable()
	c := &BytecodeContext{
		vm:                    vm,
		exe:                   exe,
		agent:                 agent,
		containedInStrictCode: false,
	}
	if Debug.PrintAST {
		fmt.Println("AST: ", node.String())
	}
	node.Bytecode(exe, c)

	return c
}

func GenerateAndRunBytecode(agent *Agent, node ASTNode) CompletionValue {
	vm2 := NewVM2(agent)
	value := node.Evaluation(vm2)
	// TODO: use completion return
	return NewCompletionReturnValue(value)
}
