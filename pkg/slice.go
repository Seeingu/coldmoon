package pkg

import "iter"

// SliceSafeGet returns the element at the given index in the slice.
func SliceSafeGet[T any](s []T, index int) (t T) {
	if index < 0 || index >= len(s) {
		return
	}
	return s[index]
}

// IterAll returns all items in the iterator.
func IterAll[T any](i iter.Seq[T]) (items []T) {
	for item := range i {
		items = append(items, item)
	}
	return
}
