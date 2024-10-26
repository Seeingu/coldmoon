package pkg

import "reflect"

// FuncEqual is a function that check if two functions are equal
func FuncEqual[F any](a, b F) bool {
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}
