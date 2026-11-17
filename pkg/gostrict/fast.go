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
//
// A list.List[float64] or list.List[int64], also of a named type of float64 or int64, is written and read as a whole
// with Float64s and Int64s, instead of List, its elements and End, e.g. the block of a time series, which a host
// passes to a callback of the program:
//
//	w := cb.Begin()
//	defer cb.End()
//	w.Int64s(times)
//	w.Float64s(values)
//	r, err := cb.Invoke()
//
// The program gets a list of its own, which a later change of the slice does not change; the slice, which the host
// reads, is valid until End. In the vm, a conversion allocates the same number of times for every length, see
// [ValueList].
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
	Int64() int64
	Uint64() uint64
	Float64() float64
	String() string
	Error() error
	Handle() js.Ref
	// Big returns a big.Int as its decimal.
	Big() string
	// Bytes returns a bin.Bytes, which must not change.
	Bytes() []byte
	State() Token
	Func() Callback
	// Record begins a record, whose fields follow. If the record never changes and the reader has kept its Go value,
	// see Keep, it returns the value and the record has already ended.
	Record() (kept any, ok bool)
	// List begins a list and returns the number of its elements, which follow.
	List() int
	// Map begins a map and returns the number of its entries, whose keys and values follow alternately.
	Map() int
	// Float64s reads a whole list of float64. The slice is valid until End.
	Float64s() []float64
	// Int64s reads a whole list of int64. The slice is valid until End.
	Int64s() []int64
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
	Int64(v int64)
	Uint64(v uint64)
	Float64(v float64)
	String(v string)
	Error(err error)
	Handle(r js.Ref)
	Big(decimal string)
	Bytes(b []byte)
	State(t Token)
	Record()
	List(n int)
	Map(n int)
	Variant(tag string)
	End()
	// Float64s writes a whole list of float64, whose elements it copies.
	Float64s(v []float64)
	// Int64s writes a whole list of int64, whose elements it copies.
	Int64s(v []int64)
}
