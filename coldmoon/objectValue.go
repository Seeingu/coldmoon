package coldmoon

type ObjectValue struct {
	Value
	Object ObjectType
}

var _ Value = (*ObjectValue)(nil)

func (o *ObjectValue) Call(agent *Agent, this Value, argumentsList ArgumentsList) (co CompletionValue) {
	result := o.Object.Call(this, argumentsList)
	// only returns normal or throw completion
	if result.t != CompletionTypeThrow {
		result.t = CompletionTypeNormal
	}
	return result
}

func (o *ObjectValue) CallNoArgs(this Value) CompletionValue {
	return o.Call(o.Object.Agent(), this, nil)
}

func (o *ObjectValue) ToCompletion() (co Completion[Value]) {
	co.value = o
	return
}

func (o *ObjectValue) ToString() CMString {
	pk := CMString("toString").ToPropertyKey()
	return o.Object.Get(pk).CallNoArgs(o).value.ToString()
}

// String is for internal use only.
func (o *ObjectValue) String() string {
	return "ObjectValue: " + o.Object.String()
}

func (o *ObjectValue) ToBoolean() bool {
	switch oo := o.Object.(type) {
	case *BooleanObject:
		return oo.getData()
	case *StringObject:
		return oo.Data != ""
	case *Object:
		return true
	case *NumberObject:
		return oo.Data != 0
	default:
		panic("unimplemented")

	}
}
