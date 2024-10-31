package pkg

type Stack[T any] struct {
	data []T
}

// Push a value onto the stack
func (s *Stack[T]) Push(v T) {
	s.data = append(s.data, v)
}

// Pop a value from the stack
func (s *Stack[T]) Pop() T {
	v := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return v
}

// Peek at the top value of the stack
func (s *Stack[T]) Peek() T {
	return s.data[len(s.data)-1]
}

// IsEmpty checks if the stack is empty
func (s *Stack[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// NewStack creates a new stack
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}
