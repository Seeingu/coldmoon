package coldmoon

import (
	"reflect"
)

type VM struct {
	agent                 *Agent
	containedInStrictCode bool
	// IsJSONParse handle is parsed from JSON.parse
	// 25.5.1: Step 7
	IsJSONParse bool
}

func NewVM2(agent *Agent) *VM {
	return &VM{
		agent: agent,
	}
}

func (v *VM) RunningLexicalEnvironment() EnvironmentRecord {
	return v.agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment
}

func (v *VM) SetRunningLexicalEnvironment(env EnvironmentRecord) {
	v.agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = env
}

func (v *VM) RunningPrivateEnvironment() *PrivateEnvironment {
	return v.agent.RunningExecutionContext().ECMAScriptCode.PrivateEnvironment
}

// TODO(BM): error handling
func (v *VM) panic(err Value) {
	panic(err)
}

// TODO(BM): rename completion
// Deprecated
// Completion
// spec: 5.2.3.1
func CompletionHandle[T any](completion Completion[T]) Completion[T] {
	// do nothing
	// - received `completion` is already a completion record
	return completion
}

// CompletionHandleV2
// spec: 5.2.3.1
func CompletionHandleV2[T any](c CompletionConvertable[T]) Completion[T] {
	if c == nil {
		panic("Completion: c is nil")
	}
	return c.ToCompletion()
}

// ReturnIfAbrupt
// spec: 5.2.3.3
// caller should return `rt` if `isAbrupt` is true
// TODO: can we simplify caller code?
func ReturnIfAbrupt[T any, RT any](completion Completion[T], returnCompletion Completion[RT]) (value T, isAbrupt bool, rt Completion[RT]) {
	value = completion.value
	if completion.IsAbrupt() {
		isAbrupt = true
		rt = CompletionFrom(returnCompletion, completion)
		return
	}
	return
}

// ReturnAssertNormal
//
// not standard: 5.2.3.4 ReturnIfAbrupt Shorthands
func ReturnAssertNormal[T any](c Completion[T]) T {
	Assert(!c.IsAbrupt())
	return c.value
}

// InitializeBoundName
// spec: 8.6.2.1
// return UNUSED or abrupt
func (v *VM) InitializeBoundName(name string, value Value, env EnvironmentRecord) (co CompletionValue) {
	if env != nil {
		env.InitializeBinding(name, value)
		return
	} else {
		lhs := v.agent.ResolveBinding(name, nil, true)
		return lhs.PutValue(v.agent, value)
	}
}

// EvaluatePropertyAccessWithExpressionKey
// spec: 13.3.3
func (v *VM) EvaluatePropertyAccessWithExpressionKey(
	baseValue Value, expression Expression, strict bool,
) (co Completion[*ReferenceRecord]) {
	propertyNameReference, isAbrupt, rt := ReturnIfAbrupt(expression.Evaluation(v), co)
	if isAbrupt {
		return rt
	}
	propertyNameValue, isAbrupt, rt := ReturnIfAbrupt(propertyNameReference.GetValue(v.agent), co)
	if isAbrupt {
		return rt
	}
	propertyKey, isAbrupt, rt := ReturnIfAbrupt(ToPropertyKey(v.agent, propertyNameValue), co)
	if isAbrupt {
		return rt
	}
	co.value = NewReferenceRecord(NewReferenceRecordBaseValue(baseValue), propertyKey.ToReference(), strict, UndefinedValue)
	return
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
		return instOfHandler.Call(target, []Value{value}).value.ToBoolean()
	}

	if !IsCallable(target) {
		agent.ThrowTypeError("target is not callable")
		return false
	}
	completion := v.OrdinaryHasInstance(target, value)
	return completion.Data()
}

// 7.3.21
func (v *VM) OrdinaryHasInstance(c Value, value Value) (co Completion[bool]) {
	agent := v.agent
	if !IsCallable(c) {
		co.value = false
		return
	}
	o := MustGetObject(c)
	if b, ok := o.(*BoundFunctionObject); ok {
		bc := b.BoundTargetFunction
		co.value = v.InstanceOfOperator(value, bc.ToValue())
		return
	}

	objectValue, ok := value.(*ObjectValue)
	if !ok {
		co.value = false
		return
	}

	proto := o.Get(NewStringPropertyKey("prototype"))
	protoObject, ok := proto.(*ObjectValue)
	if !ok {
		co.err = agent.ThrowException(TypeError, "prototype is not an object")
		return
	}

	object := objectValue.Object
	for {
		object = object.InternalMethods().GetPrototypeOf(object)
		if object == nil {
			co.value = false
			return
		}
		if protoObject.Object == object {
			co.value = true
			return
		}
	}
}

// ApplyStringOrNumericBinaryOperator
// spec: 13.15.3
func (v *VM) ApplyStringOrNumericBinaryOperator(left, right Value, op BinaryOperator) (co CompletionValue) {
	agent := v.agent
	lhs := left
	rhs := right
	finalLval := lhs
	finalRval := rhs
	if op == BinaryOperatorAddition {
		lprim, isAbrupt, rt := ReturnIfAbrupt(lhs.ToPrimitive(agent, PreferredTypeDefault), co)
		if isAbrupt {
			return rt
		}
		rprim, isAbrupt, rt := ReturnIfAbrupt(rhs.ToPrimitive(agent, PreferredTypeDefault), co)
		if isAbrupt {
			return rt
		}
		_, lprimIsString := lprim.(*StringValue)
		_, rprimIsString := rprim.(*StringValue)
		if lprimIsString || rprimIsString {
			lstr, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, lprim), co)
			if isAbrupt {
				return rt
			}
			rstr, isAbrupt, rt := ReturnIfAbrupt(ToStringCompletion(agent, rprim), co)
			if isAbrupt {
				return rt
			}
			return NewStringValue(lstr.Data + rstr.Data).ToCompletion()
		}

		finalLval = lprim
		finalRval = rprim
	}

	lnum, isAbrupt, rt := ReturnIfAbrupt(ToNumeric(agent, finalLval), co)
	if isAbrupt {
		return rt
	}
	rnum, isAbrupt, rt := ReturnIfAbrupt(ToNumeric(agent, finalRval), co)
	if isAbrupt {
		return rt
	}
	if reflect.TypeOf(lnum) != reflect.TypeOf(rnum) {
		return co.ThrowTypeError(agent, "TypeError: lnum and rnum are not the same type")
	}

	lNumber, isNumber := lnum.(*NumberValue)
	lBigInt, _ := lnum.(*BigIntValue)
	rNumber, _ := rnum.(*NumberValue)
	rBigInt, _ := rnum.(*BigIntValue)

	switch op {
	case BinaryOperatorExponentiation:
		if isNumber {
			return lNumber.Exponentiate(rNumber).ToCompletion()
		} else {
			return lBigInt.Exponentiate(rBigInt).ToCompletion()
		}
	case BinaryOperatorMultiplication:
		if isNumber {
			return lNumber.Multiply(rNumber).ToCompletion()
		} else {
			return lBigInt.Multiply(rBigInt).ToCompletion()
		}
	case BinaryOperatorAddition:
		if isNumber {
			return lNumber.Add(rNumber).ToCompletion()
		} else {
			return lBigInt.Add(rBigInt).ToCompletion()
		}
	case BinaryOperatorSubtraction:
		if isNumber {
			return lNumber.Subtract(rNumber).ToCompletion()
		} else {
			return lBigInt.Subtract(rBigInt).ToCompletion()
		}
	case BinaryOperatorDivision:
		if isNumber {
			return lNumber.Divide(rNumber).ToCompletion()
		} else {
			return lBigInt.Divide(rBigInt).ToCompletion()
		}
	case BinaryOperatorRemainder:
		if isNumber {
			return lNumber.Remainder(rNumber).ToCompletion()
		} else {
			return lBigInt.Remainder(rBigInt).ToCompletion()
		}
	case BinaryOperatorLeftShift:
		if isNumber {
			return lNumber.LeftShift(rNumber).ToCompletion()
		} else {
			return lBigInt.LeftShift(rBigInt).ToCompletion()
		}
	case BinaryOperatorRightShift:
		if isNumber {
			return lNumber.SignedRightShift(rNumber).ToCompletion()
		} else {
			return lBigInt.SignedRightShift(rBigInt).ToCompletion()
		}
	case BinaryOperatorUnsignedRightShift:
		if isNumber {
			return lNumber.UnsignedRightShift(rNumber).ToCompletion()
		} else {
			return lBigInt.UnsignedRightShift(rBigInt).ToCompletion()
		}
	case BinaryOperatorBitwiseAnd:
		if isNumber {
			return lNumber.BitwiseAnd(rNumber).ToCompletion()
		} else {
			return lBigInt.BitwiseAnd(rBigInt).ToCompletion()
		}
	case BinaryOperatorBitwiseOr:
		if isNumber {
			return lNumber.BitwiseOr(rNumber).ToCompletion()
		} else {
			return lBigInt.BitwiseOr(rBigInt).ToCompletion()
		}
	case BinaryOperatorBitwiseXor:
		if isNumber {
			return lNumber.BitwiseXor(rNumber).ToCompletion()
		} else {
			return lBigInt.BitwiseXor(rBigInt).ToCompletion()
		}
	}
	panic("unreachable")
}

// EvaluateCall ( func, ref, arguments, tailPosition )
// spec: 13.3.6.2
func (v *VM) EvaluateCall(fun, ref Value, arguments []Value, tailPosition bool) (co CompletionValue) {
	agent := v.agent
	var thisValue Value
	if rr, ok := ref.ReferenceRecord(); ok {
		if rr.IsPropertyReference() {
			thisValue = rr.GetThisValue()
		} else {
			refEnv, ok := rr.Base.Env()
			Assert(ok)
			if o := refEnv.WithBaseObject(); o != nil {
				thisValue = o.ToValue()
			} else {
				thisValue = UndefinedValue
			}
		}
	} else {
		thisValue = UndefinedValue
	}
	if !fun.IsObject() {
		co.err = agent.ThrowTypeError("function is not an object")
		return
	}
	if !IsCallable(fun) {
		co.err = agent.ThrowTypeError("function is not callable")
		return
	}
	// TODO: WIP: tailPosition
	value, isAbrupt, rt := ReturnIfAbrupt(fun.Call(agent, thisValue, arguments), co)
	if isAbrupt {
		return rt
	}
	co.value = value
	return
}

type LabelSet = []string

// ForBodyEvaluation
// spec: 14.7.4.3
func (v *VM) ForBodyEvaluation(
	test, increment Expression,
	stmt Statement,
	perIterationBindings []string,
	labelSet LabelSet,
) (co CompletionValue) {
	var V Value = UndefinedValue
	CreatePerIterationEnvironment(perIterationBindings)
	for {
		if test != nil {
			testValue, _, isAbrupt, rt := v.EvalAndGetValue(test, co)
			if isAbrupt {
				return rt
			}
			if !testValue.ToBoolean() {
				return V.ToCompletion()
			}
		}
		result := stmt.Evaluation(v)
		if !LoopContinues(result, labelSet) {
			return UpdateEmpty(result, V)
		}
		if !IsUndefinedOrNil(result.value) {
			V = result.value
		}
		CreatePerIterationEnvironment(perIterationBindings)
		if increment != nil {
			incRef := increment.Evaluation(v)
			incRef.value.GetValue(v.agent)
		}
	}
}

// ForInOfHeadEvaluation
// spec: 14.7.5.6
func (v *VM) ForInOfHeadEvaluation(
	uninitializedBoundNames []string,
	expr Expression,
	iterationKind ForInOfIterationKind,
) (co Completion[*IteratorRecord]) {
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
	exprRef, isAbrupt, rt := ReturnIfAbrupt(expr.Evaluation(v), co)
	if isAbrupt {
		return rt
	}
	agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment = oldEnv
	exprValue, isAbrupt, rt := ReturnIfAbrupt(exprRef.GetValue(agent), co)
	if isAbrupt {
		return rt
	}
	if iterationKind == ForInOfIterationKindEnumerate {
		if IsUndefinedOrNil(exprValue) || exprValue == NullValue {
			// TODO: return { [[Type]]: BREAK, [[Value]]: EMPTY, [[Target]]: EMPTY }
			co.t = CompletionTypeBreak
			// return EMPTY
			return
		}
		obj := exprValue.ToObject(agent).value
		iterator := obj.EnumerateObjectProperties()
		nextMethod := ReturnAssertNormal(GetV(agent, iterator.ToValue(), NewStringPropertyKey("next")))
		co.value = &IteratorRecord{
			Iterator:   iterator,
			NextMethod: nextMethod,
		}
		return
	} else {
		var iteratorKind IteratorKind
		if iterationKind == ForInOfIterationKindAsyncIterate {
			iteratorKind = IteratorKindAsync
		} else {
			iteratorKind = IteratorKindSync
		}
		value, isAbrupt, rt := ReturnIfAbrupt(GetIterator(agent, exprValue, iteratorKind), co)
		if isAbrupt {
			return rt
		}
		co.value = value
		return
	}
}

// ForInOfBodyEvaluation
// spec: 14.7.5.7
// iteratorKind is optional
func (v *VM) ForInOfBodyEvaluation(
	lhs ASTNode,
	stmt Statement,
	iteratorRecord *IteratorRecord,
	iterationKind ForInOfIterationKind,
	lhsKind ForInOfLhsKind,
	labelSet LabelSet,
	iteratorKind IteratorKind,
) (co CompletionValue) {
	agent := v.agent
	oldEnv := v.RunningLexicalEnvironment()
	var V Value = UndefinedValue
	// TODO:
	destructuring := false
	if destructuring {
	}
	for {
		nextResultValue := ReturnAssertNormal(iteratorRecord.NextMethod.Call(agent, iteratorRecord.Iterator.ToValue(), nil))
		if iteratorKind == IteratorKindAsync {
			panic("unimplemented")
		}
		nextResult, ok := nextResultValue.GetObject()
		if !ok {
			co.err = v.agent.ThrowTypeError("Iterator result is not an object")
			return
		}
		done, isAbrupt, rt := ReturnIfAbrupt(IteratorComplete(nextResult), co)
		if isAbrupt {
			return rt
		}
		if done {
			co.value = V
			return
		}
		nextValue, isAbrupt, rt := ReturnIfAbrupt(IteratorValue(nextResult), co)
		if isAbrupt {
			return rt
		}
		if lhsKind == ForInOfLhsKindAssignment || lhsKind == ForInOfLhsKindVarBinding {
			if destructuring {
				if lhsKind == ForInOfLhsKindAssignment {
					// TODO
				} else {
					// TODO
				}
			} else {
				if lhsKind == ForInOfLhsKindVarBinding {
					lhsName := lhs.(*ForBinding).BindingIdentifier
					lhsRef := agent.ResolveBinding(lhsName, nil, true)
					_, isAbrupt, rt := ReturnIfAbrupt(lhsRef.PutValue(agent, nextValue), co)
					if isAbrupt {
						return rt
					}
				} else {
					lhsValue, isAbrupt, rt := ReturnIfAbrupt(lhs.Evaluation(v), co)
					if isAbrupt {
						return rt
					}
					lhsRef, ok := lhsValue.ReferenceRecord()
					if !ok {
						return co.ThrowTypeError(agent, "for-in target is not assignable")
					}
					_, isAbrupt, rt = ReturnIfAbrupt(lhsRef.PutValue(agent, nextValue), co)
					if isAbrupt {
						return rt
					}
				}
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
				return UpdateEmpty(result, V)
			} else {
				Assert(iterationKind == ForInOfIterationKindIterate)
				UpdateEmpty(result, V)
				if iteratorKind == IteratorKindAsync {
					// TODO: AsyncIteratorClose
					panic("unimplemented")
				}
				return iteratorRecord.IteratorClose(result)
			}
		}
		if result.value != nil {
			V = result.value
		}
	}
}

// BlockDeclarationInstantiation
// spec: 14.2.3, B.3.2.6
func (v *VM) BlockDeclarationInstantiation(code StaticSemanticsLexicallyScopedDeclarations, env EnvironmentRecord) {
	declarations := code.LexicallyScopedDeclarations()
	privateEnv := v.RunningPrivateEnvironment()
	for _, decl := range declarations {
		boundNames := decl.(StaticSemanticsBoundNames).BoundNames()
		for _, dn := range boundNames {
			if IsConstantDeclaration(decl) {
				env.CreateImmutableBinding(dn, true)
			} else {
				if !env.HasBinding(dn) {
					env.CreateMutableBinding(dn, false)
				}
			}
		}
		if d, ok := decl.(RuntimeSemanticsInstantiateFunctionObject); ok {
			fn := boundNames[0]
			fo := d.InstantiateFunctionObject(env, privateEnv)
			if !env.HasBinding(fn) {
				env.InitializeBinding(fn, fo.ToValue())
			} else {
				_, ok := decl.(*FunctionDeclaration)
				Assert(ok)
				env.SetMutableBinding(fn, fo.ToValue(), false)
			}
		} else {
			panic("unreachable")
		}
	}
}

// EvalAndGetValue is a helper function to evaluate a node and get the value
func (v *VM) EvalAndGetValue(node ASTNode, co CompletionValue) (Value, Value, bool, CompletionValue) {
	ref, isAbrupt, rt := ReturnIfAbrupt(node.Evaluation(v), co)
	if isAbrupt {
		return nil, nil, true, rt
	}
	Assert(ref != nil)
	value, isAbrupt, rt := ReturnIfAbrupt(ref.GetValue(v.agent), co)
	if isAbrupt {
		return nil, nil, true, rt
	}
	return value, ref, false, rt
}

func CreatePerIterationEnvironment(perIterationBindings []string) {
	// TODO
}

// LoopContinues
// spec: 14.7.1.1
func LoopContinues(result CompletionValue, labelSet LabelSet) bool {
	if result.t == CompletionTypeNormal && result.err == nil {
		return true
	}
	// TODO: check target
	return false
}

// UpdateEmpty
// spec: 6.2.4.3
func UpdateEmpty(result CompletionValue, V Value) CompletionValue {
	if result.value != nil || result.err != nil {
		return result
	}
	return V.ToCompletion()
}

func RunNode(agent *Agent, node ASTNode) (co Completion[Value]) {
	vm2 := NewVM2(agent)
	if functionBody, ok := node.(*FunctionBody); ok {
		vm2.containedInStrictCode = functionBody.Strict
	}
	return node.Evaluation(vm2)
}
