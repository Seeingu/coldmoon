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
	object := o.Object
	return object.InternalMethods().Call(object, this, argumentsList)
}

func (o *ObjectValue) CallNoArgs(this Value) Value {
	return o.Call(this, nil)
}

func (o *ObjectValue) ToCompletion() CompletionValue {
	return NewCompletionValue(o)
}

// String is for internal use only.
// TODO: JS standard toString() should use ToString() instead.
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
	default:
		panic("unimplemented")

	}
}
