package pkg

// SliceSafeGet returns the element at the given index in the slice.
func SliceSafeGet[T any](s []T, index int) (t T) {
	if index < 0 || index >= len(s) {
		return
	}
	return s[index]
}
