package coldmoon

import "fmt"

func printObject(object ObjectType) string {
	switch o := object.(type) {
	case *Object:
	case *NumberObject:
	case *StringObject:
		return o.Data
	}
	return "object"
}

func printValue(v Value) string {
	switch vv := v.(type) {
	case *undefinedValue, *nullValue:
		return vv.String()
	case *ObjectValue:
		return printObject(vv.Object)
	case *StringValue:
		return vv.Data
	case *NumberValue:
		return fmt.Sprintf("%f", vv.Data)
	}
	return "value"
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
	node.Bytecode(exe, c)
	return c
}

func GenerateAndRunBytecode(agent *Agent, node ASTNode) CompletionValue {
	bytecode := GenerateBytecode(agent, node)

	if Debug.PrintAST {
		fmt.Println("AST: ", node.String())
	}
	if Debug.PrintBytecode {
		fmt.Println("Executable: ", bytecode.exe.String())
	}
	result := bytecode.Run()
	if result.Data() != nil {
		fmt.Println("Result: ", result.Data().String())
	} else if result.IsError() {
		fmt.Println("Error Result: ", result.Error().String())
	}

	return result
}
