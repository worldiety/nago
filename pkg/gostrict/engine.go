// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"context"
	"fmt"
	"math"
	"strings"

	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// ValueKind is the kind of a value, which crosses the boundary of an engine: an argument or result of an exported
// function of the main package or of a function of a header package. Every engine represents these values
// exactly, also -0 and NaN, and without serialization.
type ValueKind int

const (
	ValueInvalid ValueKind = iota
	ValueBool
	ValueInt32
	ValueUint32
	ValueFloat64
	ValueString
	// ValueError is an error, which may be nil. Its message crosses the boundary.
	ValueError
	// ValueHandle is a handle of a header package, see package std/js, whose value in Go is a js.Ref.
	ValueHandle
	// ValueList is a list.List, whose value in Go is a []any of its elements.
	ValueList
	// ValueRecord is a struct of the program, whose fields cross the boundary, e.g. the input of a use case. Its
	// value in Go is a []any of its fields in the order of the [Record].
	ValueRecord
	// ValueState is a pointer to a struct of the program, which holds its mutable state. The host keeps it as an
	// opaque [Token] and passes it back, e.g. in each frame of an animation.
	ValueState
	// ValueOut is a msg.Out of messages, which a function returns and the host sends to the backend, see
	// [Session]. Its value in Go is a []any of records.
	ValueOut
	// ValueMap is a dict.Map, whose value in Go is a []any of its entries in their order, each a []any of the key
	// and the value.
	ValueMap
	// ValueUnion is a sealed interface of the program, whose variants are records, see [Union]. Its value in Go is
	// a [Variant].
	ValueUnion
	// ValueBig is a big.Int of package std/big, an integer of any size, whose value in Go is its decimal string.
	ValueBig
	// ValueFunc is a function of the program, which a function of a header package takes, e.g. the callback, which
	// renders a cell of a table. Its value in Go is a [Callback], its signature is [Type.Func].
	ValueFunc
)

func (k ValueKind) String() string {
	switch k {
	case ValueBool:
		return "bool"
	case ValueInt32:
		return "int32"
	case ValueUint32:
		return "uint32"
	case ValueFloat64:
		return "float64"
	case ValueString:
		return "string"
	case ValueError:
		return "error"
	case ValueHandle:
		return "handle"
	case ValueList:
		return "list"
	case ValueRecord:
		return "record"
	case ValueState:
		return "state"
	case ValueOut:
		return "out"
	case ValueMap:
		return "map"
	case ValueUnion:
		return "union"
	case ValueBig:
		return "big.Int"
	case ValueFunc:
		return "func"
	}
	return "invalid"
}

// Type is the type of a value, which crosses the boundary.
type Type struct {
	Kind ValueKind
	// Name is the full name of a record or state, e.g. "example.com/app.Input".
	Name string
	// Elem is the type of the elements of a list or the values of a map.
	Elem *Type
	// Key is the type of the keys of a map.
	Key *Type
	// Func is the signature of a function, without a name.
	Func *Func
}

// The types of the basic values.
var (
	Bool    = Type{Kind: ValueBool}
	Int32   = Type{Kind: ValueInt32}
	Uint32  = Type{Kind: ValueUint32}
	Float64 = Type{Kind: ValueFloat64}
	String  = Type{Kind: ValueString}
	Error   = Type{Kind: ValueError}
)

func (t Type) String() string {
	switch t.Kind {
	case ValueList:
		return "list.List[" + t.Elem.String() + "]"
	case ValueOut:
		return "msg.Out[" + t.Elem.String() + "]"
	case ValueMap:
		return "dict.Map[" + t.Key.String() + ", " + t.Elem.String() + "]"
	case ValueRecord, ValueUnion:
		return t.Name
	case ValueState:
		return "*" + t.Name
	case ValueFunc:
		return t.Func.String()
	}
	return t.Kind.String()
}

// Equal reports, whether both types are the same.
func (t Type) Equal(o Type) bool {
	if t.Kind != o.Kind || t.Name != o.Name {
		return false
	}
	return equalPtr(t.Elem, o.Elem) && equalPtr(t.Key, o.Key) && equalFunc(t.Func, o.Func)
}

func equalFunc(a, b *Func) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Name == b.Name && equalTypes(a.Params, b.Params) && equalTypes(a.Results, b.Results)
}

func equalTypes(a, b []Type) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}

func equalPtr(a, b *Type) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

// Func is the signature of a function, which crosses the boundary of an engine.
type Func struct {
	// Name is the name of an exported function, e.g. "Run", or the full name of a function of a header package,
	// e.g. "example.com/host.Mix".
	Name    string
	Params  []Type
	Results []Type
}

func (f Func) String() string {
	join := func(ts []Type) string {
		var parts []string
		for _, t := range ts {
			parts = append(parts, t.String())
		}
		return strings.Join(parts, ", ")
	}
	res := join(f.Results)
	if len(f.Results) > 1 {
		res = "(" + res + ")"
	}
	if res != "" {
		res = " " + res
	}
	name := f.Name
	if name != "" {
		name = " " + name
	}
	return "func" + name + "(" + join(f.Params) + ")" + res
}

// Record describes a struct, which crosses the boundary as a record.
type Record struct {
	Name   string
	Fields []Field
	// Immutable is set, if the program never changes a value of the record, also not through a list or map of
	// it. Then the translated program shares its values, e.g. a style, which many cells refer to, and an engine may
	// convert a value once.
	Immutable bool
}

// Field is a field of a record.
type Field struct {
	// Name is the name in Go, e.g. for a form, which an application generates for the input of a use case.
	Name string
	// JS is the name of the property in JavaScript.
	JS   string
	Type Type
}

// Union describes a sealed interface, which crosses the boundary: its variants are the records, which implement
// it in its package.
type Union struct {
	Name string
	// Variants are the full names of the records.
	Variants []string
	// Handles is set, if the variants are handles of a header package, e.g. the views of lib/ui, which cross the
	// boundary from the program to the host. The value of such a variant is a [Variant] with the js.Ref of the
	// handle as its only field.
	Handles bool
}

// Variant is the value of a union: the full name of its record and the fields of the record.
type Variant struct {
	Tag    string
	Fields []any
}

// Token is the opaque value of a state, see [ValueState]. It refers to the state in the engine, which created it,
// and is only valid there.
type Token struct {
	typ string
	v   any
}

// NewToken returns a token of the state v of the type with the full name typ. Only an engine creates tokens.
func NewToken(typ string, v any) Token {
	return Token{typ: typ, v: v}
}

// Type returns the full name of the type of the state.
func (t Token) Type() string {
	return t.typ
}

// Value returns, what the token refers to in its engine.
func (t Token) Value() any {
	return t.v
}

// State is a token of a state of the type T, which a generated client returns and accepts, so that a host can not
// mix up the states of different types.
type State[T any] struct {
	Token Token
}

// CheckArgs reports, whether the values have the types.
func CheckArgs(name string, types []Type, values []any) error {
	if len(values) != len(types) {
		return fmt.Errorf("%s needs %d values, got %d", name, len(types), len(values))
	}
	for i, t := range types {
		if err := checkValue(t, values[i]); err != nil {
			return fmt.Errorf("value %d of %s: %w", i, name, err)
		}
	}
	return nil
}

func checkValue(t Type, v any) error {
	ok := false
	switch t.Kind {
	case ValueBool:
		_, ok = v.(bool)
	case ValueInt32:
		_, ok = v.(int32)
	case ValueUint32:
		_, ok = v.(uint32)
	case ValueFloat64:
		_, ok = v.(float64)
	case ValueString:
		_, ok = v.(string)
	case ValueError:
		_, ok = v.(error)
		ok = ok || v == nil
	case ValueHandle:
		_, ok = v.(js.Ref)
	case ValueList, ValueOut:
		var list []any
		if list, ok = v.([]any); ok {
			for i, e := range list {
				if err := checkValue(*t.Elem, e); err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
			}
		}
	case ValueMap:
		var entries []any
		if entries, ok = v.([]any); ok {
			for i, e := range entries {
				kv, isPair := e.([]any)
				if !isPair || len(kv) != 2 {
					return fmt.Errorf("entry %d must be a []any of the key and the value", i)
				}
				if err := checkValue(*t.Key, kv[0]); err != nil {
					return fmt.Errorf("key %d: %w", i, err)
				}
				if err := checkValue(*t.Elem, kv[1]); err != nil {
					return fmt.Errorf("value %d: %w", i, err)
				}
			}
		}
	case ValueUnion:
		_, ok = v.(Variant)
	case ValueBig:
		var s string
		if s, ok = v.(string); ok && !isDecimal(s) {
			return fmt.Errorf("%q is no decimal integer", s)
		}
	case ValueFunc:
		_, ok = v.(Callback)
	case ValueRecord:
		// the fields are checked by the engine, which knows the record
		_, ok = v.([]any)
	case ValueState:
		var tok Token
		if tok, ok = v.(Token); ok && tok.typ != t.Name {
			return fmt.Errorf("must be a state of %s, got one of %s", t.Name, tok.typ)
		}
	}
	if !ok {
		return fmt.Errorf("must be a %s, got %T", t, v)
	}
	return nil
}

// SameValues compares values exactly: floats by their bits, so -0 differs from 0 and NaN equals NaN, errors by
// their messages and lists and records by their elements.
func SameValues(a, b []any) bool {
	return compareValues(a, b, func(x, y float64) bool {
		return math.Float64bits(x) == math.Float64bits(y) || (math.IsNaN(x) && math.IsNaN(y))
	})
}

// CloseValues compares values like [SameValues], but floats only up to the relative tolerance, e.g. 1e-9. The
// native engine may compute floats more precisely or differently than JavaScript, e.g. by fusing a
// multiplication and an addition, and so may other hardware. NaN equals NaN and infinities must be equal.
func CloseValues(a, b []any, tolerance float64) bool {
	return compareValues(a, b, func(x, y float64) bool {
		switch {
		case math.IsNaN(x) || math.IsNaN(y):
			return math.IsNaN(x) == math.IsNaN(y)
		case math.IsInf(x, 0) || math.IsInf(y, 0):
			return x == y
		}
		return math.Abs(x-y) <= tolerance*max(math.Abs(x), math.Abs(y))
	})
}

func compareValues(a, b []any, floats func(x, y float64) bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		switch x := a[i].(type) {
		case float64:
			y, ok := b[i].(float64)
			if !ok || !floats(x, y) {
				return false
			}
		case error:
			y, ok := b[i].(error)
			if !ok || x.Error() != y.Error() {
				return false
			}
		case []any:
			y, ok := b[i].([]any)
			if !ok || !compareValues(x, y, floats) {
				return false
			}
		case Variant:
			y, ok := b[i].(Variant)
			if !ok || x.Tag != y.Tag || !compareValues(x.Fields, y.Fields, floats) {
				return false
			}
		case Token:
			// tokens of different engines refer to different objects, so only their types compare
			y, ok := b[i].(Token)
			if !ok || x.typ != y.typ {
				return false
			}
		default:
			if a[i] != b[i] {
				return false
			}
		}
	}
	return true
}

// Fn is a function, which takes and returns values of the types of its [Func].
type Fn func(args []any) []any

// Callback is the value of a [ValueFunc]: a function of the program, which the host calls with values of the types of
// its signature, see [Type.Func], e.g. a host function, which renders a table, calls it for each visible cell. A
// callback is valid as long as its instance. The host calls it on the goroutine, which uses the instance, e.g. the
// event loop of a window: during a call of the instance, e.g. in the host function itself, or between two calls,
// e.g. when the window renders the table later.
//
// A checked program does not fail, but an engine may, e.g. if the instance is closed. Then Call returns the error
// and the instance returns it once more from its next call, so that the host learns about it, even if a generated
// adapter, which turns the callback into a Go function, has to drop it and returns zero values.
type Callback interface {
	Call(args ...any) ([]any, error)
}

// ErrorOf returns the error of the neutral value of an error, which may be nil.
func ErrorOf(v any) error {
	err, _ := v.(error)
	return err
}

// CallbackFunc is a [Callback] of a Go function, e.g. for a test.
type CallbackFunc func(args ...any) ([]any, error)

// Call calls the function.
func (f CallbackFunc) Call(args ...any) ([]any, error) {
	return f(args...)
}

// Funcs maps functions by name, see [Func.Name]. The generator writes the functions of a package as Funcs, see
// [Generate].
type Funcs map[string]Fn

// Engine executes bundles. The compiler does not depend on an engine, so an application chooses one, e.g. goja on
// a server, a browser, or the native code of the package in a test.
type Engine interface {
	// Load prepares the bundle. The host implements functions of header packages by their full name and replaces
	// their JavaScript bindings. A missing function uses its binding.
	Load(ctx context.Context, b *Bundle, host Funcs) (Instance, error)
}

// Instance is a loaded bundle.
type Instance interface {
	// Call calls an exported function of the main package with values of its types, see [Bundle.Exports].
	Call(ctx context.Context, name string, args ...any) ([]any, error)
	Close() error
}

// isDecimal tells, whether s is a decimal integer with an optional sign, the neutral value of a big.Int.
func isDecimal(s string) bool {
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
