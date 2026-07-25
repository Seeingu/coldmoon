package pkg

type Queue[T any] struct {
	items []T
}

func (s *Queue[T]) Data() []T {
	return s.items
}

// Enqueue adds an item to the end of the queue
func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

// Dequeue removes and returns the item from the front of the queue
func (q *Queue[T]) Dequeue() T {
	if len(q.items) == 0 {
		panic("queue is empty")
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

// Peek returns the item at the front of the queue without removing it
func (q *Queue[T]) Peek() interface{} {
	if len(q.items) == 0 {
		return nil
	}
	return q.items[0]
}

// IsEmpty checks if the queue is empty
func (q *Queue[T]) IsEmpty() bool {
	return len(q.items) == 0
}

// Size returns the number of items in the queue
func (q *Queue[T]) Size() int {
	return len(q.items)
}
