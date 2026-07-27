package pkg

import (
	"fmt"
	"reflect"
	"runtime"
)

// FuncEqual is a function that check if two functions are equal
func FuncEqual[F any](a, b F) bool {
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

// RecoverFromFunc is a function that recovers from a function panic
func RecoverFromFunc[T any](f func() T) (r T, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in try", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	return f(), nil
}

// GetFunctionName
// see: https://stackoverflow.com/questions/7052693/how-to-get-the-name-of-a-function-in-go
func GetFunctionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
