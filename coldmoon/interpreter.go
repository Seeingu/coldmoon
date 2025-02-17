package coldmoon

type VM2 struct {
	agent                 *Agent
	containedInStrictCode bool
	// IsJSONParse handle is parsed from JSON.parse
	// 25.5.1: Step 7
	IsJSONParse bool
	// isReturn indicates the current execution context is a return statement
	isReturn bool
}

func NewVM2(agent *Agent) *VM2 {
	return &VM2{
		agent: agent,
	}
}

func (v *VM2) RunningLexicalEnvironment() EnvironmentRecord {
	return v.agent.RunningExecutionContext().ECMAScriptCode.LexicalEnvironment
}

func (v *VM2) RunningPrivateEnvironment() *PrivateEnvironment {
	return v.agent.RunningExecutionContext().ECMAScriptCode.PrivateEnvironment
}

// InitializeBoundName
// spec: 8.6.2.1
func (v *VM2) InitializeBoundName(name string, value Value, env EnvironmentRecord) {
	if env == nil {
		lhs := v.agent.ResolveBinding(name, nil, true)
		lhs.PutValue(v.agent, value)
	} else {
		env.InitializeBinding(name, value)
	}
}

// EvaluatePropertyAccessWithExpressionKey
// spec: 13.3.3
func (v *VM2) EvaluatePropertyAccessWithExpressionKey(
	baseValue Value, expression Expression, strict bool,
) *ReferenceRecord {
	propertyNameReference := expression.Evaluation(v)
	propertyNameValue := propertyNameReference.GetValue(v.agent)
	propertyKey := ToPropertyKey(v.agent, propertyNameValue)
	return NewReferenceRecord(NewReferenceRecordBaseValue(baseValue), propertyKey.ToReference(), strict, UndefinedValue)
}

// EvaluatePropertyAccessWithIdentifierKey
// spec: 13.3.4
func (v *VM2) EvaluatePropertyAccessWithIdentifierKey(baseValue Value, identifierName IdentifierName, strict bool) *ReferenceRecord {
	// TODO: use 13.1.2 Static Semantics: StringValue
	propertyNameString := identifierName
	return NewReferenceRecord(
		&ReferenceRecordBase{value: baseValue},
		&ReferencedName{String: propertyNameString},
		strict,
		// TODO: EMPTY
		UndefinedValue,
	)
}

// 13.15.3
func (v *VM2) ApplyStringOrNumericBinaryOperator(left, right Value, op BinaryOperator) Value {
	return ApplyStringOrNumericBinaryOperator(
		v.agent, left, right, op,
	)
}

// EvaluateCall ( func, ref, arguments, tailPosition )
// 13.3.6.2
func (v *VM2) EvaluateCall(fun, ref Value, arguments []Value, tailPosition bool) Value {
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
	return evaluateCall(agent, fun, thisValue, arguments)
}

type LabelSet = []string

// 14.7.4.3
func (v *VM2) ForBodyEvaluation(test, increment Expression, stmt Statement, perIterationBindings []string, labelSet LabelSet) Value {
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
func (v *VM2) ForInOfHeadEvaluation(
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
func (v *VM2) ForInOfBodyEvaluation(
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

// 14.7.1.1
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
