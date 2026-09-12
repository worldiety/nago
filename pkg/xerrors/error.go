// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xerrors

import (
	"errors"
	"maps"
	"slices"
	"strings"
	"sync"
)

// PathSeparator separates the segments of a nested field path, e.g. "Issuer.Seat.Country".
// Keys in [ErrorWithFields.Fields] are either a flat struct field name ("Country") or such a
// path. Consumers like the auto form match the full path of the rendered field first and fall
// back to the bare field name for backwards compatibility.
const PathSeparator = "."

const (
	// maxUnwrapDepth bounds how deep the traversal follows a chain of wrapped errors.
	maxUnwrapDepth = 100

	// maxUnwrapNodes bounds the total number of visited nodes. A depth cap alone is not
	// enough: an Unwrap() []error node that returns the same children repeatedly makes the
	// walk branch exponentially, so a tree of depth 100 can hide 2^100 visits.
	maxUnwrapNodes = 10_000
)

// budget is a shared node counter for one traversal.
type budget struct{ left int }

func (b *budget) spend() bool {
	if b.left <= 0 {
		return false
	}

	b.left--
	return true
}

// WithFields creates an [ErrorWithFields] from a message and an even number of key/value
// arguments. Keys are field names or field paths, see [PathSeparator].
func WithFields(msg string, args ...string) error {
	if len(args)%2 != 0 {
		args = append(args, "!MISSING_VALUE")
	}

	f := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		f[args[i]] = args[i+1]
	}

	return ErrorWithFields{Message: msg, Fields: f}
}

// ErrorWithFields carries validation messages bound to individual model fields. It may
// additionally wrap a Cause, so that a field-bound validation error and an underlying
// technical error can travel together without either being lost.
type ErrorWithFields struct {
	Message string
	Fields  map[string]string // name or path -> message
	Cause   error             // optional underlying error, may be nil
}

func (e ErrorWithFields) UnwrapFields() any {
	return e.Fields
}

// Unwrap returns the optional underlying cause so that errors.Is and errors.As can
// traverse through this error.
func (e ErrorWithFields) Unwrap() error {
	return e.Cause
}

func (e ErrorWithFields) Error() string {
	msg := e.Message
	if msg == "" && len(e.Fields) > 0 {
		// No deliberate message was set, so render a technical summary for logs. This is
		// deterministic on purpose: a Go map dump has random key order and is unusable both
		// in a log and as input to a language model.
		msg = describeFields(e.Fields)
	}

	switch {
	case msg == "" && e.Cause == nil:
		return "field validation failed"
	case msg == "":
		return e.Cause.Error()
	case e.Cause == nil:
		return msg
	default:
		return msg + ": " + e.Cause.Error()
	}
}

// describeFields renders a field map deterministically for technical output.
func describeFields(fields map[string]string) string {
	var sb strings.Builder
	sb.WriteString("field validation failed: ")
	for i, k := range slices.Sorted(maps.Keys(fields)) {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(strings.ReplaceAll(fields[k], "\n", ", "))
	}

	return sb.String()
}

// Residual returns the part of the error tree which is not field bound, or nil if every leaf
// was an [ErrorWithFields] carrying a purely technical message.
//
// Use it together with [Collect] to split an error into the half that belongs on the form
// fields and the half that still needs a banner. Without this split a join of a validation
// error and an infrastructure error loses the infrastructure half, because the validation
// half makes the whole thing look handled.
func Residual(err error) error {
	return residual(err, 0, &budget{left: maxUnwrapNodes})
}

func residual(e error, depth int, b *budget) error {
	if e == nil || depth > maxUnwrapDepth || !b.spend() {
		return nil
	}

	switch x := e.(type) {
	case ErrorWithFields:
		// A deliberate message, set via FieldBuilder.SetMessage, is meant for the user and
		// must survive. An auto generated technical message must not.
		if x.Message != "" {
			return e
		}

		return residual(x.Cause, depth+1, b)

	case interface{ Unwrap() []error }:
		var keep []error
		for _, sub := range x.Unwrap() {
			if r := residual(sub, depth+1, b); r != nil {
				keep = append(keep, r)
			}
		}

		switch len(keep) {
		case 0:
			return nil
		case 1:
			return keep[0]
		default:
			return errors.Join(keep...)
		}

	case interface{ Unwrap() error }:
		if residual(x.Unwrap(), depth+1, b) == nil {
			// the wrapper only added context to field messages which are shown on the fields
			return nil
		}

		// keep the whole wrapper so that its message and errors.Is both still work
		return e

	default:
		return e
	}
}

// Scoped returns a copy whose field keys are all prefixed with the given prefix, so that the
// result of validating a nested model can be attached to the enclosing model. Scoping
// "Country" with "Issuer.Seat" yields "Issuer.Seat.Country". An empty prefix returns the
// receiver unchanged.
func (e ErrorWithFields) Scoped(prefix string) ErrorWithFields {
	if prefix == "" || len(e.Fields) == 0 {
		return e
	}

	fields := make(map[string]string, len(e.Fields))
	for k, v := range e.Fields {
		fields[prefix+PathSeparator+k] = v
	}

	e.Fields = fields
	return e
}

// Collect walks the entire error tree, including multi-error nodes produced by errors.Join,
// and merges every [ErrorWithFields] it finds into a single value. This is different from
// errors.AsType, which stops at the first match and therefore silently drops the field
// messages of all siblings.
//
// Messages for the same key are joined with a newline, in the deterministic order in which
// they are encountered. The returned bool reports whether at least one [ErrorWithFields] was
// found. The Cause of the result is the original err, so no information is lost.
func Collect(err error) (ErrorWithFields, bool) {
	if err == nil {
		return ErrorWithFields{}, false
	}

	var (
		fields   map[string]string
		messages []string
		found    bool
	)

	b := &budget{left: maxUnwrapNodes}

	var walk func(error, int)
	walk = func(e error, depth int) {
		// Guard against pathological or self referencing error graphs. A map based identity
		// guard is not possible here because error values such as ErrorWithFields contain a
		// map and are therefore not hashable, so bound the work instead.
		if e == nil || depth > maxUnwrapDepth || !b.spend() {
			return
		}

		if ewf, ok := e.(ErrorWithFields); ok {
			found = true
			if ewf.Message != "" {
				messages = append(messages, ewf.Message)
			}

			for _, k := range slices.Sorted(maps.Keys(ewf.Fields)) {
				v := ewf.Fields[k]
				if fields == nil {
					fields = map[string]string{}
				}

				prev, ok := fields[k]
				switch {
				case !ok || prev == "":
					fields[k] = v
				case !slices.Contains(strings.Split(prev, "\n"), v):
					// compare whole messages, not substrings: "empty" is a substring of
					// "must not be empty" yet a different statement
					fields[k] = prev + "\n" + v
				}
			}
		}

		switch x := e.(type) {
		case interface{ Unwrap() error }:
			walk(x.Unwrap(), depth+1)
		case interface{ Unwrap() []error }:
			for _, sub := range x.Unwrap() {
				walk(sub, depth+1)
			}
		}
	}

	walk(err, 0)

	if !found {
		return ErrorWithFields{}, false
	}

	return ErrorWithFields{
		Message: strings.Join(messages, "; "),
		Fields:  fields,
		Cause:   err,
	}, true
}

type Errors = FieldBuilder

// FieldBuilder is a helper to build a validation error map of fields.
//
// The zero value is a ready to use root builder. A builder returned by [FieldBuilder.Nested]
// is a view onto its root: it owns no state, prefixes every added field name and forwards all
// operations to the root, so that a single [FieldBuilder.Error] collects everything.
type FieldBuilder struct {
	mutex   sync.Mutex
	fields  map[string]string
	message string

	// root and prefix are only set for builders created by Nested. A nested builder never
	// touches its own mutex or fields.
	root   *FieldBuilder
	prefix string
}

// Add inserts another Field/Message validation error tuple.
// If field already exists, the new message is appended to the existing one with a new line as separator.
// Adding is thread safe.
func (b *FieldBuilder) Add(field, msg string) {
	if b.root != nil {
		b.root.Add(b.prefix+PathSeparator+field, msg)
		return
	}

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

// SetMessage sets an additional non field bound message which describes the validation failure
// as a whole. It is used as the [ErrorWithFields.Message] and is rendered as a banner by
// consumers which cannot attach it to any field.
func (b *FieldBuilder) SetMessage(msg string) {
	if b.root != nil {
		b.root.SetMessage(msg)
		return
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.message = msg
}

// Nested returns a builder which writes into the same underlying field map but prefixes every
// added field name with the given prefix. Use it to let a nested model validate itself while
// still reporting into the path namespace of the enclosing model. Nesting composes, so
// Nested("Issuer").Nested("Seat").Add("Country", ...) yields "Issuer.Seat.Country".
//
//	issuer := errs.Nested("Issuer")
//	issuer.Add("Country", "must not be empty") // -> "Issuer.Country"
func (b *FieldBuilder) Nested(prefix string) *FieldBuilder {
	if prefix == "" {
		return b
	}

	if b.root != nil {
		return &FieldBuilder{root: b.root, prefix: b.prefix + PathSeparator + prefix}
	}

	return &FieldBuilder{root: b, prefix: prefix}
}

func (b *FieldBuilder) Has() bool {
	if b.root != nil {
		return b.root.Has()
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	return len(b.fields) > 0
}

// Error returns the accumulated validation errors as error or nil if nothing has been added.
// The message is rendered deterministically from the sorted field keys, so it is stable
// between runs and safe to show in logs or to hand to a language model.
func (b *FieldBuilder) Error() error {
	if b.root != nil {
		return b.root.Error()
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.fields) == 0 {
		return nil
	}

	fields := make(map[string]string, len(b.fields))
	for k, v := range b.fields {
		fields[k] = v
	}

	// Message stays empty unless SetMessage was called, so that consumers can tell a
	// deliberate user facing message apart from the generated technical summary.
	return ErrorWithFields{
		Message: b.message,
		Fields:  fields,
	}
}
