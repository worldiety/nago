package data

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// deprecated
// Option efficiently wraps an arbitrary (value) type and tells if it is available or not.
// It serializes either as "null" or the value without introducing a 'box' in the json structure.
type Option[T any] struct {
	value T
	ok    bool
}

func (o Option[T]) IsNone() bool {
	return !o.ok
}

func (o Option[T]) Ok() bool {
	return o.ok
}

func (o Option[T]) Unwrap() T {
	if !o.ok {
		panic(fmt.Errorf("expected a value of %T but got none", o.value))
	}

	return o.value
}

func (o Option[T]) UnwrapOr(defaultValue T) T {
	if !o.ok {
		return defaultValue
	}

	return o.value
}

func (o *Option[T]) UnmarshalJSON(buf []byte) error {
	var zero T
	if bytes.Equal([]byte("null"), buf) {
		o.ok = false
		o.value = zero
		return nil
	}

	err := json.Unmarshal(buf, &zero)
	if err != nil {
		return err
	}

	o.ok = true
	o.value = zero
	return nil
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.ok {
		return json.Marshal(o.value)
	}

	return []byte("null"), nil
}

func None[T any]() Option[T] {
	return Option[T]{}
}

func Some[T any](v T) Option[T] {
	return Option[T]{
		ok:    true,
		value: v,
	}
}
