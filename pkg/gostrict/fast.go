// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import "go.wdy.de/nago/pkg/gostrict/std/js"

// FastCallback is a [Callback], which passes its values without converting them into neutral values, e.g. a function
// of a program in engine/vm. The adapters, which [Generate] writes, use it, if a callback implements it:
//
//	w := cb.Begin()
//	defer cb.End()
//	w.Int32(row)
//	w.Int32(col)
//	r, err := cb.Invoke()
//	if err != nil { ... }
//	cell := readCell(r)
//
// The values are written and read in the order of the parameters and results, a record in the order of the fields
// of its [Record], like the neutral values. Begin locks the instance like Call, End releases it, also after a
// failure. The reader and writer are valid until End.
type FastCallback interface {
	Callback
	// Begin begins a call and returns the writer of the arguments.
	Begin() Writer
	// Invoke calls the function and returns the reader of the results. An error is reported by the instance once
	// more, like the error of Call.
	Invoke() (Reader, error)
	// End ends the call.
	End()
}

// Reader reads the values of a call one after another, see [FastCallback].
type Reader interface {
	Bool() bool
	Int32() int32
	Uint32() uint32
	Float64() float64
	String() string
	Error() error
	Handle() js.Ref
	// Big returns a big.Int as its decimal.
	Big() string
	State() Token
	Func() Callback
	// Record begins a record, whose fields follow. If the record never changes and the reader has kept its Go value,
	// see Keep, it returns the value and the record has already ended.
	Record() (kept any, ok bool)
	// List begins a list and returns the number of its elements, which follow.
	List() int
	// Map begins a map and returns the number of its entries, whose keys and values follow alternately.
	Map() int
	// Variant begins a variant of a union and returns its full name, its fields or its handle follow.
	Variant() string
	// End ends a record, list, map or variant. It reports for a record, whether the reader would keep its Go value.
	End() bool
	// Keep keeps the Go value of the record, which has just ended, see Record.
	Keep(v any)
}

// Writer writes the values of a call one after another, see [FastCallback] and [Reader].
type Writer interface {
	Bool(v bool)
	Int32(v int32)
	Uint32(v uint32)
	Float64(v float64)
	String(v string)
	Error(err error)
	Handle(r js.Ref)
	Big(decimal string)
	State(t Token)
	Record()
	List(n int)
	Map(n int)
	Variant(tag string)
	End()
}
