package std

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

var NotAvailable = errors.New("not available")

// Option is introduced because range over func can only represent at most 2 arguments. Processing
// a (T, ok, error) becomes impossible. Also, it is not correct to always use pointers for modelling or
// to use hidden error types for clear optional cases where an absent thing is never an error by definition.
// Intentionally it shares the same field layout as the stdlib [sql.Null] type.
// This also helps for performance edge cases, where you can technically express that a value is really
// just a value and does not escape.
//
// It sports also a non-nesting custom JSON serialization, which just uses NULL as representation.
// Note that if T is a pointer type, the Option becomes invalid after unmarshalling because a valid nil pointer
// cannot be distinguished from an invalid nil pointer in JSON, but you likely should not model your domain that
// way anyway.
//
// If you already have a pointer, just use its zero value which is nil and not Option.
type Option[T any] struct {
	V     T    // TODO this encourages broken access patterns
	Valid bool // TODO this encourages broken access patterns
}

// Some is a factory to create a valid option.
func Some[T any](v T) Option[T] {
	return Option[T]{
		V:     v,
		Valid: true,
	}
}

// None is only for better readability and equal to the zero value of Option.
func None[T any]() Option[T] {
	return Option[T]{}
}

// Unwrap makes the assertion that the Option is valid and otherwise panics. Such panic is always a programming error.
func (o Option[T]) Unwrap() T {
	if !o.Valid {
		panic(fmt.Errorf("unwrapped invalid option"))
	}

	return o.V
}

// Get returns the value or [NotAvailable].
func (o Option[T]) Get() (T, error) {
	if o.Valid {
		return o.V, nil
	} // TODO better T,bool?

	return o.V, NotAvailable
}

// UnwrapOrZero returns either the valid contained value or the default zero value of T.
func (o Option[T]) UnwrapOrZero() T {
	if o.Valid {
		return o.V
	}

	var zero T
	return zero
}

// Unpack2 is a shorthand for evaluating option and error and returns [fs.ErrNotExist] if no error and not exists,
// express that fact as oneliner.
func Unpack2[T any](opt Option[T], err error) (T, error) {
	if err != nil {
		return opt.V, err
	}

	return opt.Get()
}

// Iter allows iteration over the possibly contained value. Iter is a [iter.Seq]. This allows to apply
// any map, reduce, filter pipelines on Option.
func (o Option[T]) Iter(yield func(T) bool) {
	if o.Valid {
		yield(o.V)
	}
}

func (o *Option[T]) UnmarshalJSON(buf []byte) error {
	var zero T
	if bytes.Equal([]byte("null"), buf) {
		o.Valid = false
		o.V = zero
		return nil
	}

	err := json.Unmarshal(buf, &zero)
	if err != nil {
		return err
	}

	o.Valid = true
	o.V = zero
	return nil
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.Valid {
		return json.Marshal(o.V)
	}

	return []byte("null"), nil
}
