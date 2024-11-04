package coldmoon

import "fmt"

func GenerateAndRunBytecode(agent *Agent, node ASTNode) *CompletionRecord {
	vm := NewVM(agent)
	exe := NewExecutable()

	c := &BytecodeContext{
		agent:                 agent,
		containedInStrictCode: false,
	}
	node.Bytecode(exe, c)

	fmt.Println("Executable: ", exe.String())
	result := vm.Run(exe)

	fmt.Println("Result: ", result.Value.String())
	return result
}
