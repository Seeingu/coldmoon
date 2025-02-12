package coldmoon

type VM2 struct {
	agent                 *Agent
	containedInStrictCode bool
	// IsJSONParse handle is parsed from JSON.parse
	// 25.5.1: Step 7
	IsJSONParse bool
}

func NewVM2(agent *Agent) *VM2 {
	return &VM2{
		agent: agent,
	}
}

// 13.3.4
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
	if r, ok := ref.(*ReferenceRecordValue); ok {
		rr := r.ReferenceRecord
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

func CreatePerIterationEnvironment(perIterationBindings []string) {
	// TODO
}

// 14.7.1.1
func LoopContinues(result Value, labelSet LabelSet) bool {
	// TODO
	return true
}

// 6.2.4.3
func UpdateEmpty(result, V Value) Value {
	// TODO
	return UndefinedValue
}
