package std

import (
	"github.com/worldiety/option"
)

type Option[T any] = option.Opt[T]

// Some is a factory to create a valid option.
func Some[T any](v T) Option[T] {
	return option.Some[T](v)
}

// None is only for better readability and equal to the zero value of Option.
func None[T any]() Option[T] {
	return Option[T]{}
}
