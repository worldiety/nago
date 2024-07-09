package ora

import (
	"bytes"
	"fmt"
	"github.com/clarketm/json"
)

// Option is introduced because range over func can only represent at most 2 arguments. Processing
// a (T, ok, error) becomes impossible. Also, it is not correct to always use pointers for modelling or
// to use hidden error types for clear optional cases where an absent thing is never an error by definition.
// Intentionally it shares the same field layout as the stdlib [sql.Null] type.
// This also helps for performance edge cases, where you can technically express, that a value is really
// just a value and does not escape.
//
// It sports also a non-nesting custom JSON serialization, which just uses NULL as representation.
// Note that if T is a pointer type, the Option becomes invalid after unmarshalling because a valid nil pointer
// cannot be distinguished from an invalid nil pointer in JSON, but you likely should not model your domain that
// way anyway.
//
// If you already have a pointer, just use its zero value which is nil and not Option.
type Option[T any] struct {
	V     T
	Valid bool
}

// Some is a factory to create a valid option.
func Some[T any](v T) Option[T] {
	return Option[T]{
		V:     v,
		Valid: true,
	}
}

// Unwrap makes the assertion that the Option is valid and otherwise panics.
func (o Option[T]) Unwrap() T {
	if !o.Valid {
		panic(fmt.Errorf("unwrapped invalid option"))
	}

	return o.V
}

// Get returns the value or [fs.ErrNotExist].
func (o Option[T]) Get() (T, bool) {
	if o.Valid {
		return o.V, true
	}

	return o.V, false
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
