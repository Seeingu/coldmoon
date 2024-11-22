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
