package enum

import (
	"fmt"
)

type Error[T any] interface {
	error
	Cause() T
	Unwrap() error
}

type err[T any] struct {
	cause T
}

func (e err[T]) Cause() T {
	return e.cause
}

func (e err[T]) Unwrap() error {
	box := any(e.cause)
	for {
		if uwrap, ok := box.(interface{ Unwrap() any }); ok {
			box = uwrap.Unwrap()
		} else {
			break
		}
	}

	if err, ok := box.(error); ok {
		return err
	}

	return nil
}

func (e err[T]) Error() string {
	if err := e.Unwrap(); err != nil {
		return "enum error: " + err.Error()
	}

	return fmt.Sprintf("%T: %v", e.cause, e.cause)
}

// IntoErr returns a typed error which contains the given T as a cause.
// If T is an error, it can be unwrapped.
func IntoErr[T any](t T) Error[T] {
	return err[T]{
		cause: t,
	}
}
