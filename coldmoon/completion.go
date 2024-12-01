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
type Completion[T any] interface {
	IsError() bool
	IsUndefined() bool
	IsNull() bool
	IsAbrupt() bool
	Data() T
}

type completionDefaultImpl[T any] struct {
	Completion[T]
	Type        CompletionType
	isUndefined bool
	isNull      bool
	err         Value
	data        T
}

var _ Completion[any] = &completionDefaultImpl[any]{}

func (c *completionDefaultImpl[T]) IsError() bool {
	return c.Type == CompletionTypeThrow
}
func (c *completionDefaultImpl[T]) IsUndefined() bool {
	return c.isUndefined
}
func (c *completionDefaultImpl[T]) IsNull() bool {
	return c.isNull
}
func (c *completionDefaultImpl[T]) IsAbrupt() bool {
	return c.Type != CompletionTypeNormal
}
func (c *completionDefaultImpl[T]) Data() T {
	return c.data
}

type completionNormalArgs[T any] struct {
	isUndefined bool
	isNull      bool
	data        T
}

func newCompletionNormal[T any](args completionNormalArgs[T]) *completionDefaultImpl[T] {
	c := completionDefaultImpl[T]{
		Type:        CompletionTypeNormal,
		isUndefined: args.isUndefined,
		isNull:      args.isNull,
		data:        args.data,
	}
	return &c
}
func newCompletionError[T any](err Value) *completionDefaultImpl[T] {
	return &completionDefaultImpl[T]{
		Type: CompletionTypeThrow,
		err:  err,
	}
}

// MARK: - PropertyDescriptor

type CompletionPropertyDescriptor completionDefaultImpl[*PropertyDescriptor]

func NewCompletionPropertyDescriptorUndefined() *CompletionPropertyDescriptor {
	c := newCompletionNormal(completionNormalArgs[*PropertyDescriptor]{
		isUndefined: true,
	})
	return (*CompletionPropertyDescriptor)(c)
}
func NewCompletionPropertyDescriptor(desc *PropertyDescriptor) *CompletionPropertyDescriptor {
	c := newCompletionNormal(completionNormalArgs[*PropertyDescriptor]{
		data: desc,
	})
	return (*CompletionPropertyDescriptor)(c)
}

// MARK: - Object

type CompletionObject completionDefaultImpl[ObjectType]

func NewCompletionObject(obj ObjectType) *CompletionObject {
	c := newCompletionNormal(completionNormalArgs[ObjectType]{
		data: obj,
	})
	return (*CompletionObject)(c)
}
func NewCompletionObjectError(err Value) *CompletionObject {
	c := newCompletionError[ObjectType](err)
	return (*CompletionObject)(c)
}
func NewCompletionObjectNull() *CompletionObject {
	return &CompletionObject{
		Type:   CompletionTypeNormal,
		isNull: true,
	}
}

// MARK: - Value

type CompletionValue struct {
	Completion[Value]
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
	return c.Type != CompletionTypeNormal && c.Type != CompletionTypeReturn
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
