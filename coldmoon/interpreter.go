package coldmoon

type VM2 struct {
	agent                 *Agent
	containedInStrictCode bool
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
