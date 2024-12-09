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

func GenerateAndRunBytecode(agent *Agent, node ASTNode) CompletionValue {
	vm := NewVM(agent)
	exe := NewExecutable()

	c := &BytecodeContext{
		agent:                 agent,
		containedInStrictCode: false,
	}
	node.Bytecode(exe, c)

	if Debug.PrintAST {
		fmt.Println("AST: ", node.String())
	}
	if Debug.PrintBytecode {
		fmt.Println("Executable: ", exe.String())
	}
	result := vm.Run(exe)
	if Debug.PrintBytecode {
		if result.Data() != nil {
			fmt.Println("Result: ", result.Data().String())
		} else {
			fmt.Println("Result: nil")
		}
	}

	return result
}
