package vm

import (
	"fmt"
	"github.com/Seeingu/coldmoon/code"
	"github.com/Seeingu/coldmoon/compiler"
	"github.com/Seeingu/coldmoon/object"
)

var JSTrue = &object.BooleanObject{Value: true}
var JSFalse = &object.BooleanObject{Value: false}
var JSNull = &object.NullObject{}
var JSUndefined = &object.UndefinedObject{}

const MaxFrames = 1024
const GlobalsSize = 65536
const StackSize = 2048

type VM struct {
	constants []object.Object

	stack []object.Object
	// stack pointer: point at next free slot
	sp int

	globals []object.Object

	frames     []*Frame
	frameIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:  bytecode.Constants,
		stack:      make([]object.Object, StackSize),
		sp:         0,
		globals:    make([]object.Object, GlobalsSize),
		frames:     frames,
		frameIndex: 1,
	}
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(JSTrue)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(JSFalse)
			if err != nil {
				return err
			}
		case code.OpGreaterThan, code.OpEqual, code.OpNotEqual, code.OpLessThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpNot:
			err := vm.executeNot()
			if err != nil {
				return err
			}
		case code.OpNegate:
			err := vm.executeNegate()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpFalse:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(JSNull)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()
		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements
			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpObject:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			o, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements

			err = vm.push(o)
			if err != nil {
				return err
			}
		case code.OpIndex:
			i := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, i)
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.callFunction(int(numArgs))
			if err != nil {
				return err
			}

		case code.OpReturnValue:
			returnValue := vm.pop()
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			err := vm.push(JSUndefined)
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()
		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

// MARK: Private

// MARK: Frame

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

// MARK: Execute

func (vm *VM) executeNot() error {
	operand := vm.pop()
	switch operand {
	case JSTrue:
		return vm.push(JSFalse)
	case JSFalse:
		return vm.push(JSTrue)
	case JSNull:
		return vm.push(JSTrue)
	default:
		return vm.push(JSFalse)

	}
}

func (vm *VM) executeNegate() error {
	operand := vm.pop()
	if operand.Type() != object.TypeInt {
		return fmt.Errorf("negate: operand must be an integer")
	}
	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.TypeInt && right.Type() == object.TypeInt {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator %d", op)
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	case code.OpLessThan:
		return vm.push(nativeBoolToBooleanObject(leftValue < rightValue))
	default:
		return fmt.Errorf("integer comparison: unknown operator %d", op)
	}

}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()
	if leftType == object.TypeInt && rightType == object.TypeInt {
		return vm.executeIntegerOperation(op, left, right)
	}
	if leftType == object.TypeString && rightType == object.TypeString {
		return vm.executeBinaryStringOperation(op, left, right)
	}
	return fmt.Errorf("unknown operator %d", op)

}

func (vm *VM) callFunction(numArgs int) error {
	fn, ok := vm.stack[vm.sp-1-numArgs].(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("expected a compiled function")
	}
	if numArgs != fn.NumParameters {
		return fmt.Errorf("expected %d parameters, got %d", numArgs, fn.NumParameters)
	}
	frame := NewFrame(fn, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + fn.NumLocals
	return nil
}

// MARK: Utils

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) executeIntegerOperation(op code.Opcode, left object.Object, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left object.Object, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("string: unknown operator %d", op)
	}
	leftValue := left.(*object.StringObject).Value
	rightValue := right.(*object.StringObject).Value
	return vm.push(&object.StringObject{Value: leftValue + rightValue})
}

func (vm *VM) buildArray(startIndex int, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i] = vm.stack[i]
	}
	return &object.ArrayObject{Elements: elements}
}

func (vm *VM) buildHash(startIndex int, endIndex int) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("invalid hash key type: %T", key)
		}
		pairs[hashKey.HashKey()] = pair
	}
	return &object.ObjectObject{Pairs: pairs}, nil
}

func (vm *VM) executeIndexExpression(left object.Object, i object.Object) error {
	switch {
	case left.Type() == object.TypeArray && i.Type() == object.TypeInt:
		return vm.executeArrayIndex(left, i)
	case left.Type() == object.TypeObject:
		return vm.executeObjectIndex(left, i)
	default:
		return fmt.Errorf("index expression: unknown type %s", left.Type().String())
	}
}

func (vm *VM) executeArrayIndex(left object.Object, i object.Object) error {
	a := left.(*object.ArrayObject)
	index := i.(*object.Integer).Value

	maxIndex := int64(len(a.Elements) - 1)
	if index < 0 || index > maxIndex {
		return vm.push(JSUndefined)
	}

	return vm.push(a.Elements[index])
}

func (vm *VM) executeObjectIndex(left object.Object, i object.Object) error {
	o := left.(*object.ObjectObject)
	key, ok := i.(object.Hashable)
	if !ok {
		return fmt.Errorf("invalid object key type: %T", key)
	}

	pair, ok := o.Pairs[key.HashKey()]
	if !ok {
		return vm.push(JSUndefined)
	}

	return vm.push(pair.Value)
}

func nativeBoolToBooleanObject(input bool) object.Object {
	if input {
		return JSTrue
	} else {
		return JSFalse
	}
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.BooleanObject:
		return obj.Value
	case *object.NullObject:
		return false
	default:
		return true
	}
}
