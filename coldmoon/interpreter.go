package coldmoon

import (
	"reflect"

	"github.com/Seeingu/coldmoon/pkg"
)

type VM struct {
	agent                 *Agent
	containedInStrictCode bool
	// IsJSONParse handle is parsed from JSON.parse
	// 25.5.1: Step 7
	IsJSONParse bool
	// isReturn indicates the current execution context is a return statement
	isReturn bool
	// isYield indicates the current execution context is a yield statement
	isYield bool
	// visitedNodesMap is used to store visited nodes
	suspendedGeneratorBody GeneratorBody
	loopNodeStack          pkg.Stack[IterationStatement]
	isInLoop               bool
}

func NewVM2(agent *Agent) *VM {
	return &VM{
		agent: agent,
	}
}

func (v *VM) RunningLexicalEnvironment() EnvironmentRecord {
	return v.agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment
}

func (v *VM) RunningPrivateEnvironment() *PrivateEnvironment {
	return v.agent.RunningExecutionContext().ECMAScriptCode.PrivateEnvironment
}

// TODO(BM): error handling
func (v *VM) panic(err Value) {
	panic(err)
}

// InitializeBoundName
// spec: 8.6.2.1
func (v *VM) InitializeBoundName(name string, value Value, env EnvironmentRecord) {
	if env == nil {
		lhs := v.agent.ResolveBinding(name, nil, true)
		lhs.PutValue(v.agent, value)
	} else {
		env.InitializeBinding(name, value)
	}
}

// EvaluatePropertyAccessWithExpressionKey
// spec: 13.3.3
func (v *VM) EvaluatePropertyAccessWithExpressionKey(
	baseValue Value, expression Expression, strict bool,
) *ReferenceRecord {
	propertyNameReference := expression.Evaluation(v)
	propertyNameValue := propertyNameReference.GetValue(v.agent)
	propertyKey := ToPropertyKey(v.agent, propertyNameValue)
	return NewReferenceRecord(NewReferenceRecordBaseValue(baseValue), propertyKey.ToReference(), strict, UndefinedValue)
}

// EvaluatePropertyAccessWithIdentifierKey
// spec: 13.3.4
func (v *VM) EvaluatePropertyAccessWithIdentifierKey(baseValue Value, identifierName IdentifierName, strict bool) *ReferenceRecord {
	propertyNameString := identifierName
	return NewReferenceRecord(
		NewReferenceRecordBaseValue(baseValue),
		&ReferencedName{String: propertyNameString},
		strict,
		// EMPTY
		UndefinedValue,
	)
}

// InstanceOfOperator
// spec: 13.10.2
func (v *VM) InstanceOfOperator(value Value, target Value) bool {
	agent := v.agent
	if _, ok := target.(*ObjectValue); !ok {
		agent.ThrowTypeError("target is not an object")
		return false
	}
	symbol := WellKnownSymbols[WellKnownSymbolsHasInstance]
	instOfHandler := GetMethod(
		agent,
		target,
		NewSymbolPropertyKey(symbol))
	if instOfHandler != nil {
		return instOfHandler.ToValue().Call(target, []Value{value}).ToBoolean()
	}

	if !IsCallable(target) {
		agent.ThrowTypeError("target is not callable")
		return false
	}
	return v.OrdinaryHasInstance(target, value).Data()
}

// 7.3.21
func (v *VM) OrdinaryHasInstance(c Value, value Value) Completion[bool] {
	agent := v.agent
	if !IsCallable(c) {
		return NewNormalCompletion(false)
	}
	o := MustGetObject(c)
	if b, ok := o.(*BoundFunctionObject); ok {
		bc := b.BoundTargetFunction
		return NewNormalCompletion(v.InstanceOfOperator(value, bc.ToValue()))
	}

	objectValue, ok := value.(*ObjectValue)
	if !ok {
		return NewNormalCompletion(false)
	}

	proto := o.Get(NewStringPropertyKey("prototype"))
	protoObject, ok := proto.(*ObjectValue)
	if !ok {
		return NewThrowCompletion[bool](agent.ThrowException(TypeError, "prototype is not an object"))
	}

	object := objectValue.Object
	for {
		object = object.InternalMethods().GetPrototypeOf(object)
		if object == nil {
			return NewNormalCompletion(false)
		}
		if protoObject.Object == object {
			return NewNormalCompletion(true)
		}
	}
}

// spec: 13.15.3
func (v *VM) ApplyStringOrNumericBinaryOperator(left, right Value, op BinaryOperator) Value {
	agent := v.agent
	lhs := left
	rhs := right
	finalLval := lhs
	finalRval := rhs
	if op == BinaryOperatorAddition {
		lprim := lhs.ToPrimitive(agent, PreferredTypeDefault)
		rprim := rhs.ToPrimitive(agent, PreferredTypeDefault)
		_, lprimIsString := lprim.(*StringValue)
		_, rprimIsString := rprim.(*StringValue)
		if lprimIsString || rprimIsString {
			lstr := lprim.String()
			rstr := rprim.String()
			return NewStringValue(lstr + rstr)
		}

		finalLval = lprim
		finalRval = rprim
	}

	lnum := ToNumeric(agent, finalLval)
	rnum := ToNumeric(agent, finalRval)
	if reflect.TypeOf(lnum) != reflect.TypeOf(rnum) {
		panic("TypeError: lnum and rnum are not the same type")
	}

	lNumber, isNumber := lnum.(*NumberValue)
	lBigInt, _ := lnum.(*BigIntValue)
	rNumber, _ := rnum.(*NumberValue)
	rBigInt, _ := rnum.(*BigIntValue)

	switch op {
	case BinaryOperatorExponentiation:
		if isNumber {
			return lNumber.Exponentiate(rNumber)
		} else {
			return lBigInt.Exponentiate(rBigInt)
		}
	case BinaryOperatorMultiplication:
		if isNumber {
			return lNumber.Multiply(rNumber)
		} else {
			return lBigInt.Multiply(rBigInt)
		}
	case BinaryOperatorAddition:
		if isNumber {
			return lNumber.Add(rNumber)
		} else {
			return lBigInt.Add(rBigInt)
		}
	case BinaryOperatorSubtraction:
		if isNumber {
			return lNumber.Subtract(rNumber)
		} else {
			return lBigInt.Subtract(rBigInt)
		}
	case BinaryOperatorDivision:
		if isNumber {
			return lNumber.Divide(rNumber)
		} else {
			return lBigInt.Divide(rBigInt)
		}
	case BinaryOperatorRemainder:
		if isNumber {
			return lNumber.Remainder(rNumber)
		} else {
			return lBigInt.Remainder(rBigInt)
		}
	case BinaryOperatorLeftShift:
		if isNumber {
			return lNumber.LeftShift(rNumber)
		} else {
			return lBigInt.LeftShift(rBigInt)
		}
	case BinaryOperatorRightShift:
		if isNumber {
			return lNumber.SignedRightShift(rNumber)
		} else {
			return lBigInt.SignedRightShift(rBigInt)
		}
	case BinaryOperatorUnsignedRightShift:
		if isNumber {
			return lNumber.UnsignedRightShift(rNumber)
		} else {
			return lBigInt.UnsignedRightShift(rBigInt)
		}
	case BinaryOperatorBitwiseAnd:
		if isNumber {
			return lNumber.BitwiseAnd(rNumber)
		} else {
			return lBigInt.BitwiseAnd(rBigInt)
		}
	case BinaryOperatorBitwiseOr:
		if isNumber {
			return lNumber.BitwiseOr(rNumber)
		} else {
			return lBigInt.BitwiseOr(rBigInt)
		}
	case BinaryOperatorBitwiseXor:
		if isNumber {
			return lNumber.BitwiseXor(rNumber)
		} else {
			return lBigInt.BitwiseXor(rBigInt)
		}
	}
	panic("unreachable")
}

// EvaluateCall ( func, ref, arguments, tailPosition )
// spec: 13.3.6.2
func (v *VM) EvaluateCall(fun, ref Value, arguments []Value, tailPosition bool) Value {
	agent := v.agent
	var thisValue Value
	if rr, ok := ref.ReferenceRecord(); ok {
		if rr.IsPropertyReference() {
			thisValue = rr.GetThisValue()
		} else {
			refEnv, ok := rr.Base.Env()
			Assert(ok)
			// TODO(SM): we should provide undefined object
			if o := refEnv.WithBaseObject(); o != nil {
				thisValue = o.ToValue()
			} else {
				thisValue = UndefinedValue
			}
		}
	} else {
		thisValue = UndefinedValue
	}
	if !ValueIsObject(fun) {
		return agent.ThrowTypeError("function is not an object")
	}
	if !IsCallable(fun) {
		return agent.ThrowTypeError("function is not callable")
	}
	// TODO: WIP: tailPosition
	// TODO(SM): argumentsList
	return fun.Call(thisValue, arguments)
}

type LabelSet = []string

// 14.7.4.3
func (v *VM) ForBodyEvaluation(test, increment Expression, stmt Statement, perIterationBindings []string, labelSet LabelSet) Value {
	var V Value = UndefinedValue
	CreatePerIterationEnvironment(perIterationBindings)
	for {
		if test != nil {
			testRef := test.Evaluation(v)
			testValue := testRef.GetValue(v.agent)
			if !testValue.ToBoolean() {
				return V
			}
			result := stmt.Evaluation(v)
			if !LoopContinues(result, labelSet) {
				return UpdateEmpty(result, V)
			}
			if !IsUndefinedOrNil(result) {
				V = result
				CreatePerIterationEnvironment(perIterationBindings)
				if increment != nil {
					incRef := increment.Evaluation(v)
					incRef.GetValue(v.agent)
				}
			}
		}
	}
}

// ForInOfHeadEvaluation
// spec: 14.7.5.6
func (v *VM) ForInOfHeadEvaluation(
	uninitializedBoundNames []string,
	expr Expression,
	iterationKind ForInOfIterationKind,
) (iterator *IteratorRecord, err Value) {
	agent := v.agent
	oldEnv := v.RunningLexicalEnvironment()
	if len(uninitializedBoundNames) > 0 {
		// TODO: Assert: uninitializedBoundNames has no duplicate entries.
		newEnv := NewDeclarativeEnvironment(oldEnv)
		for _, name := range uninitializedBoundNames {
			newEnv.CreateMutableBinding(name, false)
		}
		agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = newEnv
	}
	exprRef := expr.Evaluation(v)
	agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = oldEnv
	exprValue := exprRef.GetValue(agent)
	if iterationKind == ForInOfIterationKindEnumerate {
		if IsUndefinedOrNil(exprValue) || exprValue == NullValue {
			// TODO: return { [[Type]]: BREAK, [[Value]]: EMPTY, [[Target]]: EMPTY }
			return &IteratorRecord{}, nil
		}
		obj := exprValue.ToObject(agent)
		iterator := obj.EnumerateObjectProperties()
		nextMethod := GetV(agent, iterator.ToValue(), NewStringPropertyKey("next"))
		return &IteratorRecord{
			Iterator:   iterator,
			NextMethod: nextMethod,
		}, nil
	} else {
		var iteratorKind IteratorKind
		if iterationKind == ForInOfIterationKindAsyncIterate {
			iteratorKind = IteratorKindAsync
		} else {
			iteratorKind = IteratorKindSync
		}
		completion := GetIterator(agent, exprValue, iteratorKind)
		if completion.IsError() {
			return nil, completion.Error()
		}
		return completion.Data(), nil
	}
}

// ForInOfBodyEvaluation
// spec: 14.7.5.7
// iteratorKind is optional
func (v *VM) ForInOfBodyEvaluation(
	lhs Expression,
	stmt Statement,
	iteratorRecord *IteratorRecord,
	iterationKind ForInOfIterationKind,
	lhsKind ForInOfLhsKind,
	labelSet LabelSet,
	iteratorKind IteratorKind,
) (value Value, err Value) {
	agent := v.agent
	oldEnv := v.RunningLexicalEnvironment()
	var V Value = UndefinedValue
	// TODO:
	destructuring := false
	if destructuring {
	}
	for {
		nextResultValue := iteratorRecord.NextMethod.Call(iteratorRecord.Iterator.ToValue(), nil)
		if iteratorKind == IteratorKindAsync {
			// TODO: Await
		}
		nextResult, ok := ValueGetObject(nextResultValue)
		if !ok {
			return nil, v.agent.ThrowTypeError("Iterator result is not an object")
		}
		done := IteratorComplete(nextResult)
		if done {
			return V, nil
		}
		nextValue := IteratorValue(nextResult)
		if lhsKind == ForInOfLhsKindAssignment || lhsKind == ForInOfLhsKindVarBinding {
			if destructuring {
				if lhsKind == ForInOfLhsKindAssignment {
					// TODO
				} else {
					// TODO
				}
			} else {
				lhs.Evaluation(v)
				// TODO: check is abrupt
			}
		} else {
			f := lhs.(*ForDeclaration)
			iterationEnv := NewDeclarativeEnvironment(oldEnv)
			f.ForDeclarationBindingInstantiation(v, iterationEnv)
			agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = iterationEnv
			if destructuring {
				// TODO
			} else {
				lhsName := lhs.(StaticSemanticsBoundNames).BoundNames()[0]
				lhsRef := agent.ResolveBinding(lhsName, nil, true)
				// TODO: should return status
				lhsRef.InitializeReferencedBinding(nextValue)
			}
		}
		// TODO: handle status
		result := stmt.Evaluation(v)
		agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = oldEnv
		if !LoopContinues(result, labelSet) {
			if iterationKind == ForInOfIterationKindEnumerate {
				return UpdateEmpty(result, V), nil
			} else {
				Assert(iterationKind == ForInOfIterationKindIterate)
				UpdateEmpty(result, V)
				if iteratorKind == IteratorKindAsync {
					// TODO: AsyncIteratorClose
				}
				// TODO: handle return value
				iteratorRecord.IteratorClose()
				return UndefinedValue, nil
			}
		}
		if result != nil {
			V = result
		}
	}
}

func CreatePerIterationEnvironment(perIterationBindings []string) {
	// TODO
}

// LoopContinues
// spec: 14.7.1.1
func LoopContinues(result Value, labelSet LabelSet) bool {
	// TODO
	return true
}

// TODO: return completion
// UpdateEmpty
// spec: 6.2.4.3
func UpdateEmpty(result, V Value) Value {
	if result != nil {
		return result
	}
	return V
}

func (v *VM) GetValueOrPanic(result Value, err Value) Value {
	if result == nil {
		v.panic(err)
	}
	return result
}

func (v *VM) GetCompletionValueOrPanic(c CompletionValue) Value {
	if c.IsError() {
		v.panic(c.Error())
	}
	return c.Data()
}

func RunNode(agent *Agent, node ASTNode) CompletionValue {
	vm2 := NewVM2(agent)
	value := node.Evaluation(vm2)
	// TODO: use completion return
	return NewCompletionReturnValue(value)
}
