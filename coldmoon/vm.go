package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type VM struct {
	agent  *Agent
	stack  pkg.Stack[Value]
	result Value
}

func NewVM(agent *Agent) *VM {
	return &VM{
		agent: agent,
	}
}

func (vm *VM) Run(executable *Executable) Value {
	ip := 0
	for ip < len(executable.Instructions) {
		i := executable.Instructions[ip]
		switch ins := i.(type) {
		case *ILoad:
			vm.stack.Push(vm.result)
		case *ILoadConstant:
			vm.stack.Push(ins.Value)
		case *IStore:
			vm.result = vm.stack.Pop()
		case *IStoreConstant:
			vm.result = ins.Value
		case *IResolveThisBinding:
			vm.result = vm.agent.ResolveThisBinding()
		}
		ip += 1
	}
	return vm.result
}
