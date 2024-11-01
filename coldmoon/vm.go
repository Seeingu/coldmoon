package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type VM struct {
	agent  *Agent
	stack  pkg.Stack[Value]
	result Value
	ip     int
}

func NewVM(agent *Agent) *VM {
	return &VM{
		agent: agent,
	}
}

func (vm *VM) Run(executable *Executable) Value {
	for vm.ip < len(executable.Instructions) {
		i := executable.Instructions[vm.ip]
		switch ins := i.(type) {
		case *ILoad:
			vm.stack.Push(vm.result)
		case *ILoadConstant:
			vm.stack.Push(ins.Value)
		case *IStore:
			vm.result = vm.stack.Pop()
		case *IStoreConstant:
			vm.result = ins.Value
		case *IResolveBinding:
			// TODO: maybe ins.Name can pass to ResolveBinding directly
			reference := vm.agent.ResolveBinding(string(ins.Name), nil)
			vm.result = reference.GetValue()
		case *IResolveThisBinding:
			vm.result = vm.agent.ResolveThisBinding()
		case *IJump:
			vm.ip = ins.Target
		case *IJumpIfTrue:
			value := vm.stack.Pop()
			if value.ToBoolean() {
				vm.ip = ins.Target
			} else {
				vm.ip = ins.TargetElse
			}
		}
		vm.ip += 1
	}
	return vm.result
}
