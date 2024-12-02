package coldmoon

type ObjectValue struct {
	Value
	Object ObjectType
}

var _ Value = (*ObjectValue)(nil)

func (o *ObjectValue) Call(this Value, argumentsList ArgumentsList) Value {
	if !IsCallable(o) {
		panic("TypeError")
	}
	return o.CallAssumeCallable(this, argumentsList)
}

func (o *ObjectValue) ToCompletion() CompletionValue {
	return NewCompletionValue(o)
}

func (o *ObjectValue) CallAssumeCallable(value Value, argumentsList ArgumentsList) Value {
	object := o.Object
	return object.InternalMethods().Call(object, value, argumentsList)
}

func (o *ObjectValue) String() string {
	primValue := ToPrimitive(o.Object.Agent(), o, PreferredTypeString)
	if _, isObject := primValue.(*ObjectValue); isObject {
		panic("")
	}
	return primValue.String()
}
