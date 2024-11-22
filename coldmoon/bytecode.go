package coldmoon

import "fmt"

func GenerateAndRunBytecode(agent *Agent, node ASTNode) *CompletionValue {
	vm := NewVM(agent)
	exe := NewExecutable()

	c := &BytecodeContext{
		agent:                 agent,
		containedInStrictCode: false,
	}
	node.Bytecode(exe, c)

	fmt.Println("Executable: ", exe.String())
	result := vm.Run(exe)

	if result.Value != nil {
		fmt.Println("Result: ", result.Value.String())
	} else {
		fmt.Println("Result: nil")
	}
	return result
}
