package coldmoon

type CompletionConvertable[T any] interface {
	ToCompletion() Completion[T]
}

type CompletionType int

const (
	CompletionTypeNormal CompletionType = iota
	CompletionTypeBreak
	CompletionTypeContinue
	CompletionTypeReturn
	CompletionTypeThrow
)

// 6.2.4
type Completion[T any] struct {
	err    Value
	value  T
	t      CompletionType
	target string
}

func (c Completion[T]) IsError() bool {
	return c.err != nil
}

func (c Completion[T]) IsAbrupt() bool {
	return c.t != CompletionTypeNormal || c.IsError()
}

func (c Completion[T]) Data() T {
	return c.value
}

func (c Completion[T]) Error() Value {
	return c.err
}

func (c Completion[T]) ToCompletion() Completion[T] {
	return c
}

func (c Completion[T]) ThrowTypeError(agent *Agent, msg string) {
	c.ThrowError(agent, TypeError, msg)
}

func (c Completion[T]) ThrowError(agent *Agent, errorType ExceptionType, msg string) {
	c.t = CompletionTypeThrow
	c.err = agent.ThrowException(errorType, msg)
}

// CompletionFrom returns a new Completion with the error from `other`.
func CompletionFrom[T any, U any](a Completion[T], other Completion[U]) Completion[T] {
	a.err = other.err
	return a
}

type CompletionValue = Completion[Value]
