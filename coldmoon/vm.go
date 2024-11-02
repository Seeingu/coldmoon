package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type evaluateCallContext struct {
	reference *ReferenceRecord
}
type VM struct {
	agent               *Agent
	stack               pkg.Stack[Value]
	result              Value
	ip                  int
	evaluateCallContext evaluateCallContext
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
			vm.evaluateCallContext.reference = reference
		case *ICall:
			argumentCount := ins.ArgumentCount
			arguments := make([]Value, argumentCount)
			for i := argumentCount - 1; i >= 0; i-- {
				arguments[i] = vm.stack.Pop()
			}
			this := vm.stack.Pop()
			function := vm.stack.Pop()
			vm.result = evaluateCall(
				vm.agent,
				function,
				this,
				arguments,
			)
		case *IPrepareCall:
			isReference := ins.IsReference
			var reference *ReferenceRecord
			if isReference {
				reference = vm.evaluateCallContext.reference
			} else {
				reference = nil
			}
			this := evaluateCallGetThisValue(reference)
			vm.stack.Push(this)

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
		case *IThrow:
			value := vm.stack.Pop()
			vm.agent.exception = value
			panic("Throw")
		}
		vm.ip += 1
	}
	return vm.result
}

// 13.3.6.2
func evaluateCall(agent *Agent, function Value, this Value, arguments []Value) Value {
	if _, ok := function.(*ObjectValue); !ok {
		panic("TypeError: function is not an object")
	}
	if !IsCallable(function) {
		panic("TypeError: function is not callable")
	}
	return CallAssumeCallable(function, this, arguments)
}

func evaluateCallGetThisValue(reference *ReferenceRecord) Value {
	if reference == nil {
		return nil
	}
	if reference.IsPropertyReference() {
		return reference.GetThisValue()
	}
	refEnv := reference.Base.(*ReferenceRecordBaseEnvironment).Environment
	if o := refEnv.WithBaseObject(); o != nil {
		return NewValueFromObject(o)
	}
	return nil
}
