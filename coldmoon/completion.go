package coldmoon

type CompletionType int

const (
	CompletionTypeNormal CompletionType = iota
	CompletionTypeBreak
	CompletionTypeContinue
	CompletionTypeReturn
	CompletionTypeThrow
)

// 6.2.4
type Completion interface {
	IsError() bool
	IsUndefined() bool
	IsNull() bool
	IsAbrupt() bool
}

// MARK: - PropertyDescriptor

type CompletionPropertyDescriptor struct {
	Completion
	PropertyDescriptor *PropertyDescriptor
	isUndefined        bool
}

func NewCompletionPropertyDescriptorUndefined() *CompletionPropertyDescriptor {
	return &CompletionPropertyDescriptor{
		isUndefined: true,
	}
}
func NewCompletionPropertyDescriptor(desc *PropertyDescriptor) *CompletionPropertyDescriptor {
	return &CompletionPropertyDescriptor{
		PropertyDescriptor: desc,
	}
}
func (c *CompletionPropertyDescriptor) IsUndefined() bool {
	return c.isUndefined
}
func (c *CompletionPropertyDescriptor) IsError() bool {
	return false
}

// MARK: - Object

type CompletionObject struct {
	Completion
	Type   CompletionType
	Object ObjectType
	Error  Value
	isNull bool
}

func NewCompletionObject(obj ObjectType) *CompletionObject {
	return &CompletionObject{
		Type:   CompletionTypeNormal,
		Object: obj,
	}
}
func NewCompletionObjectError(err Value) *CompletionObject {
	return &CompletionObject{
		Type:  CompletionTypeThrow,
		Error: err,
	}
}
func NewCompletionObjectNull() *CompletionObject {
	return &CompletionObject{
		Type:   CompletionTypeNormal,
		isNull: true,
	}
}
func (c *CompletionObject) IsError() bool {
	return c.Error != nil
}
func (c *CompletionObject) IsUndefined() bool {
	return false
}
func (c *CompletionObject) IsNull() bool {
	return c.isNull
}

// MARK: - Value

type CompletionValue struct {
	Completion
	Type        CompletionType
	Value       Value
	isUndefined bool
	Error       Value
}

func NewCompletionValue(value Value) *CompletionValue {
	return &CompletionValue{
		Type:  CompletionTypeNormal,
		Value: value,
	}
}
func NewCompletionReturnValue(value Value) *CompletionValue {
	return &CompletionValue{
		Type:  CompletionTypeReturn,
		Value: value,
	}
}

func NewCompletionValueError(value Value) *CompletionValue {
	return &CompletionValue{
		Type:  CompletionTypeThrow,
		Error: value,
	}
}
func NewCompletionValueUndefined() *CompletionValue {
	return &CompletionValue{
		Type:        CompletionTypeNormal,
		isUndefined: true,
	}
}
func (c *CompletionValue) IsUndefined() bool {
	return c.isUndefined
}
func (c *CompletionValue) IsError() bool {
	return c.Error != nil
}
func (c *CompletionValue) IsNull() bool {
	return false
}
func (c *CompletionValue) IsAbrupt() bool {
	return c.Type != CompletionTypeNormal
}

func NewNormalCompletion(value Value) *CompletionValue {
	return NewCompletionValue(value)
}
func NewCompletion(t CompletionType, value Value) *CompletionValue {
	return &CompletionValue{
		Type:  t,
		Value: value,
	}
}

var UndefinedNormalCompletion = NewCompletionValueUndefined()

func NewThrowCompletion(value Value) *CompletionValue {
	return NewCompletionValueError(value)
}

var TypeErrorCompletion = NewThrowCompletion(NewStringValue("TypeError"))
