package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
	"strconv"
)

type evaluateContext struct {
	reference *ReferenceRecord
}
type VM struct {
	agent                *Agent
	stack                pkg.Stack[Value]
	result               Value
	ip                   int
	reference            *ReferenceRecord
	evaluateContextStack pkg.Stack[*evaluateContext]
}

func NewVM(agent *Agent) *VM {
	return &VM{
		agent: agent,
	}
}

func (vm *VM) Run(executable *Executable) *CompletionRecord {
	for vm.ip < len(executable.Instructions) {
		i := executable.Instructions[vm.ip]
		switch ins := i.(type) {
		case *ILoad:
			vm.stack.Push(vm.result)
		case *ILoadConstant:
			vm.stack.Push(ins.Value)
		case *ISetEvaluationContextReference:
			vm.evaluateContextStack.Push(&evaluateContext{
				reference: vm.reference,
			})
		case *IStore:
			vm.result = vm.stack.Pop()
		case *IStoreConstant:
			vm.result = ins.Value
		case *IResolveBinding:
			// TODO: maybe ins.Name can pass to ResolveBinding directly
			vm.reference = vm.agent.ResolveBinding(string(ins.Name), nil, ins.Strict)
		case *ICall:
			argumentCount := ins.ArgumentCount
			arguments := make([]Value, argumentCount)
			strict := ins.Strict
			for i := argumentCount - 1; i >= 0; i-- {
				arguments[i] = vm.stack.Pop()
			}
			this := vm.stack.Pop()
			function := vm.stack.Pop()

			realm := vm.agent.CurrentRealm()
			eval := realm.Intrinsics.Eval

			evaluateContext := vm.evaluateContextStack.Pop()
			if evaluateContext.reference != nil {
				ref := evaluateContext.reference
				refName, ok :=
					ref.ReferencedName.(*ReferencedNameString)
				if ref.IsPropertyReference() &&
					ok &&
					refName.String == "eval" &&
					pkg.FuncEqual(function.(*ObjectValue).Object, eval) {
					vm.result = directEval(vm.agent, arguments, strict)
					continue
				}
			}

			vm.result = evaluateCall(
				vm.agent,
				function,
				this,
				arguments,
			)
		case *ILoadThisValue:
			this := evaluateCallGetThisValue(vm.evaluateContextStack.Peek())
			vm.stack.Push(this)
		case *IResolveThisBinding:
			vm.result = vm.agent.ResolveThisBinding()
		case *IReturn:
			return NewNormalCompletion(vm.result)
		case *IJump:
			vm.ip = ins.Target
		case *IJumpIfTrue:
			value := vm.result
			if value.ToBoolean() {
				vm.ip = ins.Target
			} else {
				vm.ip = ins.TargetElse
			}
		case *IThrow:
			value := vm.result
			vm.agent.exception = value
			panic("Throw")
		case *ITypeof:
			if vm.reference != nil {
				if vm.reference.IsUnresolvableReference() {
					vm.result = NewStringValue("undefined")
					continue
				}
			}
			var value = vm.result
			if vm.reference != nil {
				value = vm.reference.GetValue()
			}

			switch v := value.(type) {
			case *undefinedValue:
				vm.result = NewStringValue("undefined")
			case *nullValue:
				vm.result = NewStringValue("object")
			case *BooleanValue:
				vm.result = NewStringValue("boolean")
			case *NumberValue:
				vm.result = NewStringValue("number")
			case *StringValue:
				vm.result = NewStringValue("string")
			case *SymbolValue:
				vm.result = NewStringValue("symbol")
			case *BigIntValue:
				vm.result = NewStringValue("bigint")
			case *ObjectValue:
				if v.Object.InternalMethods().Call != nil {
					vm.result = NewStringValue("function")
				} else {
					vm.result = NewStringValue("object")
				}
			default:
				panic("unreachable")
			}

		case *IToNumber:
			value := vm.result
			vm.result = ToNumber(value, vm.agent)
		case *IToNumeric:
			value := vm.result
			vm.result = ToNumeric(value, vm.agent)
		case *IUnaryMinus:
			value := vm.result
			switch v := value.(type) {
			case *BigIntValue:
				vm.result = v.UnaryMinus()
			case *NumberValue:
				vm.result = v.UnaryMinus()
			default:
				panic("unreachable")
			}
		case *ILogicalNot:
			value := vm.result
			vm.result = NewBooleanValue(!value.ToBoolean())
		case *IBitwiseNot:
			value := vm.result
			switch v := value.(type) {
			case *BigIntValue:
				vm.result = v.BitwiseNot()
			case *NumberValue:
				vm.result = v.BitwiseNot()
			default:
				panic("unreachable")
			}
		case *IEvaluatePropertyAccessWithExpressionKey:
			// 13.3.3
			propertyNameValue := vm.stack.Pop()
			strict := ins.Strict
			baseValue := vm.stack.Pop()
			propertyKey := ToPropertyKey(propertyNameValue, vm.agent)

			var referencedName ReferencedName
			switch p := propertyKey.(type) {
			case StringPropertyKey:
				referencedName = &ReferencedNameString{
					String: p.Value,
				}
			case SymbolPropertyKey:
				referencedName = &ReferencedNameSymbol{
					Symbol: p.Value,
				}
			case IntegerIndexPropertyKey:
				referencedName = &ReferencedNameString{
					String: strconv.Itoa(p.Value),
				}
			}
			vm.reference = &ReferenceRecord{
				Base: &ReferenceRecordBaseValue{
					Value: baseValue,
				},
				ReferencedName: referencedName,
				Strict:         strict,
				ThisValue:      nil,
			}
		case *IEvaluatePropertyAccessWithIdentifierKey:
			// 13.3.4
			propertyNameString := ins.Name
			strict := ins.Strict
			baseValue := vm.stack.Pop()

			referencedName := &ReferencedNameString{
				String: string(propertyNameString),
			}
			vm.reference = &ReferenceRecord{
				Base: &ReferenceRecordBaseValue{
					Value: baseValue,
				},
				ReferencedName: referencedName,
				Strict:         strict,
				ThisValue:      nil,
			}
		case *IGetValue:
			if vm.reference != nil {
				vm.result = vm.reference.GetValue()
			}
			vm.reference = nil
		}
		vm.ip += 1
	}
	return NewNormalCompletion(vm.result)
}

// 13.3.6.2
func evaluateCall(agent *Agent, function Value, this Value, arguments []Value) Value {
	if _, ok := function.(*ObjectValue); !ok {
		panic("TypeError: function is not an object")
	}
	if !IsCallable(function) {
		panic("TypeError: function is not callable")
	}
	return function.CallAssumeCallable(this, arguments)
}

func evaluateCallGetThisValue(ctx *evaluateContext) Value {
	reference := ctx.reference
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

func directEval(agent *Agent, arguments []Value, strict bool) Value {
	if len(arguments) == 0 {
		return nil
	}
	evalArg := arguments[0]
	strictCaller := strict
	return PerformEval(agent, evalArg, strictCaller, true)
}
