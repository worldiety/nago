// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package std

// LocalizedError provides a for the current context already translated error message.
// See [NewLocalizedError].
type LocalizedError interface {
	error
	Unwrap() error
	Title() string
	Description() string
	WithError(cause error) LocalizedError
}

type localizedError struct {
	title string
	desc  string
	cause error
}

func (e localizedError) Error() string {
	return e.desc
}

func (e localizedError) Title() string {
	return e.title
}

func (e localizedError) Description() string {
	return e.desc
}

func (e localizedError) Unwrap() error {
	return e.cause
}

func (e localizedError) WithError(cause error) LocalizedError {
	e.cause = cause
	return e
}

func NewLocalizedError(title, desc string) LocalizedError {
	return localizedError{
		title: title,
		desc:  desc,
	}
}

type Error string

func (e Error) Error() string {
	return string(e)
}

// ExpectZero panics if the given value is not equal to zero. This
// is also useful to assert a nil error.
func ExpectZero[T comparable](value T) {
	var zero T
	if zero != value {
		panic(value)
	}
}

// Must asserts that the tupel of (T,error) does not contain an error and
// otherwise panics.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
