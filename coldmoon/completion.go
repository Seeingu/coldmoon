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
	Error() Value
}

type completionDefaultImpl[T any] struct {
	Completion[T]
	Type        CompletionType
	isUndefined bool
	isNull      bool
	err         Value
	data        T
}

var _ Completion[any] = completionDefaultImpl[any]{}

func (c completionDefaultImpl[T]) IsError() bool {
	return c.Type == CompletionTypeThrow
}

func (c completionDefaultImpl[T]) IsUndefined() bool {
	return c.isUndefined
}

func (c completionDefaultImpl[T]) IsNull() bool {
	return c.isNull
}

func (c completionDefaultImpl[T]) IsAbrupt() bool {
	return c.Type != CompletionTypeNormal
}

func (c completionDefaultImpl[T]) Data() T {
	return c.data
}

func (c completionDefaultImpl[T]) Error() Value {
	return c.err
}

type completionNormalArgs[T any] struct {
	isUndefined bool
	isNull      bool
	data        T
}

func newCompletionNormal[T any](args completionNormalArgs[T]) completionDefaultImpl[T] {
	c := completionDefaultImpl[T]{
		Type:        CompletionTypeNormal,
		isUndefined: args.isUndefined,
		isNull:      args.isNull,
		data:        args.data,
	}
	return c
}

func newCompletionError[T any](err Value) completionDefaultImpl[T] {
	return completionDefaultImpl[T]{
		Type: CompletionTypeThrow,
		err:  err,
	}
}

// MARK: - PropertyDescriptor

type CompletionPropertyDescriptor completionDefaultImpl[*PropertyDescriptor]

func NewCompletionPropertyDescriptorUndefined() CompletionPropertyDescriptor {
	c := newCompletionNormal(completionNormalArgs[*PropertyDescriptor]{
		isUndefined: true,
	})
	return CompletionPropertyDescriptor(c)
}

func NewCompletionPropertyDescriptor(desc *PropertyDescriptor) CompletionPropertyDescriptor {
	c := newCompletionNormal(completionNormalArgs[*PropertyDescriptor]{
		data: desc,
	})
	return CompletionPropertyDescriptor(c)
}

// MARK: - Object

type CompletionObject struct {
	completionDefaultImpl[ObjectType]
}

func NewCompletionObject(obj ObjectType) CompletionObject {
	c := newCompletionNormal(completionNormalArgs[ObjectType]{
		data: obj,
	})
	return CompletionObject{c}
}

func NewCompletionObjectError(err Value) CompletionObject {
	c := newCompletionError[ObjectType](err)
	return CompletionObject{c}
}

func NewCompletionObjectNull() CompletionObject {
	c := newCompletionNormal(completionNormalArgs[ObjectType]{
		isNull: true,
	})
	return CompletionObject{c}
}

// MARK: - Module

type CompletionModule struct {
	completionDefaultImpl[*ModuleRecord]
}

func NewCompletionModule(module *ModuleRecord) CompletionModule {
	c := newCompletionNormal(completionNormalArgs[*ModuleRecord]{
		data: module,
	})
	return CompletionModule{c}
}

func NewCompletionModuleError(err Value) CompletionModule {
	c := newCompletionError[*ModuleRecord](err)
	return CompletionModule{c}
}

// MARK: - Value

type CompletionValue struct {
	completionDefaultImpl[Value]
}

func NewCompletionValue(value Value) CompletionValue {
	c := newCompletionNormal(completionNormalArgs[Value]{
		data: value,
	})
	return CompletionValue{c}
}

func NewCompletionReturnValue(value Value) CompletionValue {
	c := completionDefaultImpl[Value]{
		Type: CompletionTypeReturn,
		data: value,
	}
	return CompletionValue{c}
}

func NewCompletionValueError(err Value) CompletionValue {
	c := newCompletionError[Value](err)
	return CompletionValue{c}
}

func NewCompletionValueUndefined() CompletionValue {
	c := newCompletionNormal(completionNormalArgs[Value]{
		isUndefined: true,
	})
	return CompletionValue{c}
}

func NewNormalCompletion[T any](value T) Completion[T] {
	return NewCompletion[T](CompletionTypeNormal, value)
}

func NewReturnCompletion[T any](value T) Completion[T] {
	return NewCompletion[T](CompletionTypeReturn, value)
}

func NewThrowCompletion[T any](err Value) Completion[T] {
	return completionDefaultImpl[T]{
		Type: CompletionTypeThrow,
		err:  err,
	}
}

func NewCompletion[T any](t CompletionType, value T) Completion[T] {
	return completionDefaultImpl[T]{
		Type: t,
		data: value,
	}
}

func NewCompletionUndefined[T any]() Completion[T] {
	return completionDefaultImpl[T]{
		Type:        CompletionTypeNormal,
		isUndefined: true,
	}
}

var UndefinedNormalCompletion = NewCompletionValueUndefined()
