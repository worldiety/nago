package enum

import (
	"fmt"
)

// Error as a wrapper shall be used whenever there is a clear domain error, otherwise it can be used
// safely anywhere where an unspecified Go error is used. If there is no error, return just nil as always.
type Error[T any] interface {
	error
	Cause() T      // Cause returns the actual underlying type, probably an Enumeration.
	Unwrap() error // Unwrap returns either nil or Cause if T is also an error.
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
		return "error: " + err.Error()
	}

	return fmt.Sprintf("error: %v", e.cause)
}

// IntoErr returns a typed error which contains the given T as a cause.
// If T is an error, it can be unwrapped.
func IntoErr[T any](t T) Error[T] {
	return err[T]{
		cause: t,
	}
}
