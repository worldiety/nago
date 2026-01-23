// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xerrors

import "sync"

func WithFields(msg string, args ...string) error {
	if len(args)%2 != 0 {
		args = append(args, "!MISSING_VALUE")
	}

	f := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		f[args[i]] = args[i+1]
	}

	return ErrorWithFields{msg, f}
}

type ErrorWithFields struct {
	Message string
	Fields  map[string]string // name -> value
}

func (e ErrorWithFields) UnwrapFields() any {
	return e.Fields
}

func (e ErrorWithFields) Error() string {
	return e.Message
}

// FieldBuilder is a helper to build a validation error map of fields.
type FieldBuilder struct {
	mutex  sync.Mutex
	fields map[string]string
}

// Add inserts another Field/Message validation error tuple.
// If field already exists, the new message is appended to the existing one with a new line as separator.
// Adding is thread safe.
func (b *FieldBuilder) Add(field, msg string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.fields == nil {
		b.fields = map[string]string{}
	}

	if v, ok := b.fields[field]; ok {
		b.fields[field] = v + "\n" + msg
	} else {
		b.fields[field] = msg
	}
}

// Error returns the accumulated validation errors as error or nil if nothing has been added.
func (b *FieldBuilder) Error() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.fields) == 0 {
		return nil
	}

	return ErrorWithFields{
		Message: "field validation failed",
		Fields:  b.fields,
	}
}
