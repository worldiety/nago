// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"slices"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
)

// The acceptance tests of typed lists at the boundary (wokoda N22, L-N22-23): a host function returns and takes
// list.List[float64] and list.List[int64] as []float64 and []int64, a host calls a function of the script with them
// through Callback.Call and through the fast writer (Writer.Float64s, Writer.Int64s) and reads them with the fast reader
// (Reader.Float64s, Reader.Int64s), without boxing each value: in the vm the number of allocations of a passing does
// not depend on the length (L-N22-27). A script sees no difference to []any, in the vm and in JavaScript (goja): the
// same values, also NaN, ±Inf, -0 and the bounds of int64, the same length, the same messages of a wrong type, and a
// list of its own, which a later block of the host does not change.

const n22HostSrc = `// Package n22h hands lists of numbers between the host and the script.
//
//gostrict:header
package n22h

import "go.wdy.de/nago/pkg/gostrict/std/list"

// Floats returns the floats of the host.
//
//gostrict:host
func Floats() list.List[float64] { return list.New[float64]() }

// Ints returns the ints of the host.
//
//gostrict:host
func Ints() list.List[int64] { return list.New[int64]() }

// Mutate changes the floats and ints of the host in place.
//
//gostrict:host
func Mutate() {}

// TakeFloats hands floats to the host, which returns their number.
//
//gostrict:host
func TakeFloats(l list.List[float64]) int32 { return 0 }

// TakeInts hands ints to the host, which returns their number.
//
//gostrict:host
func TakeInts(l list.List[int64]) int32 { return 0 }

// TakeStrings hands strings to the host, which returns their number.
//
//gostrict:host
func TakeStrings(l list.List[string]) int32 { return 0 }

// CallFloats calls f with the floats of the host and returns its result.
//
//gostrict:host
func CallFloats(f func(vals list.List[float64]) list.List[float64]) list.List[float64] { return list.New[float64]() }

// CallInts calls f with the ints of the host and returns its result.
//
//gostrict:host
func CallInts(f func(vals list.List[int64]) list.List[int64]) list.List[int64] { return list.New[int64]() }

// CallTwiceFloats calls f with the floats of the host, changes them in place, calls f with them again and returns
// the second result.
//
//gostrict:host
func CallTwiceFloats(f func(vals list.List[float64]) list.List[float64]) list.List[float64] { return list.New[float64]() }

// CallTwiceInts is CallTwiceFloats for ints.
//
//gostrict:host
func CallTwiceInts(f func(vals list.List[int64]) list.List[int64]) list.List[int64] { return list.New[int64]() }

// CountFloats calls f with the floats of the host and returns its result.
//
//gostrict:host
func CountFloats(f func(vals list.List[float64]) int32) int32 { return 0 }

// CountInts calls f with the ints of the host and returns its result.
//
//gostrict:host
func CountInts(f func(vals list.List[int64]) int32) int32 { return 0 }

// FromFloats calls f and returns the floats, which the host read from its result.
//
//gostrict:host
func FromFloats(f func() list.List[float64]) list.List[float64] { return list.New[float64]() }

// FromInts calls f and returns the ints, which the host read from its result.
//
//gostrict:host
func FromInts(f func() list.List[int64]) list.List[int64] { return list.New[int64]() }

// FromCountFloats calls f and returns the number of floats, which the host read from its result.
//
//gostrict:host
func FromCountFloats(f func() list.List[float64]) int32 { return 0 }

// FoldFloats calls f once per block of the host with the sum so far and returns the last sum.
//
//gostrict:host
func FoldFloats(acc float64, f func(acc float64, vals list.List[float64]) float64) float64 { return acc }
`

const n22ProgSrc = `// Package n22 uses lists of numbers of the host.
//
//gostrict:terminating
package n22

import (
	"math"

	"go.wdy.de/nago/pkg/gostrict/std/list"
	"llm.local/n22h"
)

// Data holds lists, which the program built once.
type Data struct {
	F list.List[float64]
	I list.List[int64]
	S list.List[string]
}

// NewData builds lists of n elements, none of which a small value, which Go boxes without an allocation.
func NewData(n int32) *Data {
	f := list.New[float64]()
	i := list.New[int64]()
	s := list.New[string]()
	for k := int32(0); k < n; k++ {
		f.Append(float64(k) + 1000.5)
		i.Append(int64(k)*1000003 + 1000)
		s.Append("s")
	}
	return &Data{F: f, I: i, S: s}
}

// describeF is what the script sees of floats: the number of elements, of NaN, of -0, of +Inf and of -Inf, and the
// sum of the others.
func describeF(l list.List[float64]) list.List[float64] {
	n, nan, negZero, posInf, negInf, sum := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	for _, v := range l.All() {
		n++
		switch {
		case math.IsNaN(v):
			nan++
		case math.IsInf(v, 1):
			posInf++
		case math.IsInf(v, -1):
			negInf++
		case v == 0 && math.Signbit(v):
			negZero++
		default:
			sum += v
		}
	}
	return list.Of(n, nan, negZero, posInf, negInf, sum)
}

// describeI is what the script sees of ints: the number of elements, the smallest, the largest and the xor of all.
func describeI(l list.List[int64]) list.List[int64] {
	n := int64(0)
	lo := int64(math.MaxInt64)
	hi := int64(math.MinInt64)
	x := int64(0)
	for _, v := range l.All() {
		n++
		lo = min(lo, v)
		hi = max(hi, v)
		x ^= v
	}
	return list.Of(n, lo, hi, x)
}

func EchoFloats() list.List[float64] { return n22h.Floats() }

func SeeFloats() list.List[float64] { return describeF(n22h.Floats()) }

func EchoInts() list.List[int64] { return n22h.Ints() }

func SeeInts() list.List[int64] { return describeI(n22h.Ints()) }

func LenFloats() int32 { return n22h.Floats().Len() }

func LenInts() int32 { return n22h.Ints().Len() }

func PassFloats() int32 { return n22h.TakeFloats(n22h.Floats()) }

func PassInts() int32 { return n22h.TakeInts(n22h.Ints()) }

func PassDataFloats(d *Data) int32 { return n22h.TakeFloats(d.F) }

func PassDataInts(d *Data) int32 { return n22h.TakeInts(d.I) }

func PassDataStrings(d *Data) int32 { return n22h.TakeStrings(d.S) }

// OwnFloats keeps the floats of the host, lets the host change its own and returns the kept ones.
func OwnFloats() list.List[float64] {
	a := n22h.Floats()
	n22h.Mutate()
	return a
}

func OwnInts() list.List[int64] {
	a := n22h.Ints()
	n22h.Mutate()
	return a
}

// SetFloats changes the first of the floats of the host in its own list.
func SetFloats() int32 {
	a := n22h.Floats()
	_ = a.Set(0, 42)
	return a.Len()
}

func SetInts() int32 {
	a := n22h.Ints()
	_ = a.Set(0, 42)
	return a.Len()
}

func CallEchoFloats() list.List[float64] {
	return n22h.CallFloats(func(v list.List[float64]) list.List[float64] { return v })
}

func CallSeeFloats() list.List[float64] { return n22h.CallFloats(describeF) }

func CallEchoInts() list.List[int64] {
	return n22h.CallInts(func(v list.List[int64]) list.List[int64] { return v })
}

func CallSeeInts() list.List[int64] { return n22h.CallInts(describeI) }

// CallSetFloats changes the first of the floats, which a callback got.
func CallSetFloats() list.List[float64] {
	return n22h.CallFloats(func(v list.List[float64]) list.List[float64] {
		_ = v.Set(0, 42)
		return v
	})
}

func CallSetInts() list.List[int64] {
	return n22h.CallInts(func(v list.List[int64]) list.List[int64] {
		_ = v.Set(0, 42)
		return v
	})
}

// CallKeepFloats keeps the list of the first call of the callback and returns it from the second, after the host
// changed its floats in place.
func CallKeepFloats() list.List[float64] {
	kept := list.New[float64]()
	calls := int32(0)
	return n22h.CallTwiceFloats(func(v list.List[float64]) list.List[float64] {
		calls++
		if calls == 1 {
			kept = v
		}
		return kept
	})
}

func CallKeepInts() list.List[int64] {
	kept := list.New[int64]()
	calls := int32(0)
	return n22h.CallTwiceInts(func(v list.List[int64]) list.List[int64] {
		calls++
		if calls == 1 {
			kept = v
		}
		return kept
	})
}

func CountFloats() int32 { return n22h.CountFloats(func(v list.List[float64]) int32 { return v.Len() }) }

func CountInts() int32 { return n22h.CountInts(func(v list.List[int64]) int32 { return v.Len() }) }

// FromFloats returns the floats of the host, which the host read back from a callback.
func FromFloats() list.List[float64] {
	return n22h.FromFloats(func() list.List[float64] { return n22h.Floats() })
}

func FromInts() list.List[int64] {
	return n22h.FromInts(func() list.List[int64] { return n22h.Ints() })
}

func FromCountFloats(d *Data) int32 {
	return n22h.FromCountFloats(func() list.List[float64] { return d.F })
}

// SumFold sums the blocks of the host, one callback per block (M22c V5).
func SumFold() float64 {
	return n22h.FoldFloats(0, func(acc float64, vals list.List[float64]) float64 {
		s := acc
		for _, v := range vals.All() {
			s += v
		}
		return s
	})
}
`

// n22Bundle compiles the program of the tests once.
func n22Bundle(t testing.TB) *gostrict.Bundle {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/n22",
		gostrict.Module{Path: "llm.local/n22", FS: fstest.MapFS{"n22.go": {Data: []byte(n22ProgSrc)}}},
		gostrict.Module{Path: "llm.local/n22h", FS: fstest.MapFS{"n22h.go": {Data: []byte(n22HostSrc)}}, Trusted: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) > 0 {
		t.Fatalf("the program does not compile:\n%s", diags.Text())
	}
	return b
}

func n22Engines() []struct {
	name string
	e    gostrict.Engine
} {
	return []struct {
		name string
		e    gostrict.Engine
	}{{"vm", &vm.Engine{}}, {"goja", &gojaengine.Engine{}}}
}

// n22Mode is how the host passes a list.
type n22Mode int

const (
	// n22Any is a []any of the elements, as before N22; the counter row of every other mode.
	n22Any n22Mode = iota
	// n22Typed is a []float64 or []int64 as a result of a host function or an argument of Callback.Call.
	n22Typed
	// n22Fast is the fast writer and reader of the vm with Writer.Float64s, Writer.Int64s, Reader.Float64s and
	// Reader.Int64s for a callback; a host function returns []any in this mode.
	n22Fast
	// n22WrongElem is a []float64 for a list.List[int64] and a []int64 for a list.List[float64].
	n22WrongElem
	// n22WrongAny is a []any of float64 for a list.List[int64] and of int64 for a list.List[float64].
	n22WrongAny
)

func (m n22Mode) String() string {
	return [...]string{"[]any", "typed", "fast", "wrong element type", "[]any of a wrong element type"}[m]
}

// n22Host is the host of the program. It keeps the floats and ints, which it hands out, and records what it got.
type n22Host struct {
	mode   n22Mode
	floats []float64
	ints   []int64
	blocks [][]float64 // the blocks of FoldFloats
	// scan, if set, yields the blocks of FoldFloats instead, e.g. from tsdb.Column.ScanF64
	scan func(fn func(block []float64) bool) error

	gotFloats []float64 // what TakeFloats or FromFloats got
	gotInts   []int64
	gotForm   string // the Go type, in which TakeFloats or TakeInts got its list
	callErr   error  // the error of the last call of a callback
}

func n22Box[T any](v []T) []any {
	res := make([]any, len(v))
	for i, x := range v {
		res[i] = x
	}
	return res
}

// n22Unbox returns the elements of a list, which crosses the boundary as a []T or a []any of T.
func n22Unbox[T any](v any) ([]T, bool) {
	switch l := v.(type) {
	case []T:
		return slices.Clone(l), true
	case []any:
		res := make([]T, len(l))
		for i, e := range l {
			x, ok := e.(T)
			if !ok {
				return nil, false
			}
			res[i] = x
		}
		return res, true
	}
	return nil, false
}

// floatsOut is a list of floats as the host returns it or passes it to a callback in its mode.
func (h *n22Host) floatsOut(v []float64) any {
	switch h.mode {
	case n22Typed:
		return v
	case n22WrongElem, n22WrongAny:
		ints := make([]int64, len(v))
		for i, x := range v {
			ints[i] = int64(x)
		}
		if h.mode == n22WrongAny {
			return n22Box(ints)
		}
		return ints
	}
	return n22Box(v)
}

func (h *n22Host) intsOut(v []int64) any {
	switch h.mode {
	case n22Typed:
		return v
	case n22WrongElem, n22WrongAny:
		floats := make([]float64, len(v))
		for i, x := range v {
			floats[i] = float64(x)
		}
		if h.mode == n22WrongAny {
			return n22Box(floats)
		}
		return floats
	}
	return n22Box(v)
}

// callFloats calls cb with floats in the mode of the host and returns the list of floats it returns, if it returns
// one.
func (h *n22Host) callFloats(cb gostrict.Callback, v []float64, readList bool) any {
	if h.mode == n22Fast {
		fc := cb.(gostrict.FastCallback)
		w := fc.Begin()
		defer fc.End()
		n22WriteFloat64s(w, v)
		r, err := fc.Invoke()
		if err != nil {
			h.callErr = err
			return nil
		}
		if readList {
			// the reader is valid until End, which comes when this function returns
			return slices.Clone(n22ReadFloat64s(r))
		}
		return r.Int32()
	}
	res, err := cb.Call(h.floatsOut(v))
	if err != nil {
		h.callErr = err
		return nil
	}
	if readList {
		l, _ := n22Unbox[float64](res[0])
		return l
	}
	return res[0]
}

func (h *n22Host) callInts(cb gostrict.Callback, v []int64, readList bool) any {
	if h.mode == n22Fast {
		fc := cb.(gostrict.FastCallback)
		w := fc.Begin()
		defer fc.End()
		n22WriteInt64s(w, v)
		r, err := fc.Invoke()
		if err != nil {
			h.callErr = err
			return nil
		}
		if readList {
			return slices.Clone(n22ReadInt64s(r))
		}
		return r.Int32()
	}
	res, err := cb.Call(h.intsOut(v))
	if err != nil {
		h.callErr = err
		return nil
	}
	if readList {
		l, _ := n22Unbox[int64](res[0])
		return l
	}
	return res[0]
}

// floatsResult is a list of floats, which a callback returned, as the result of a host function: always a []any, so
// that only the call of the callback is in the mode of the host.
func floatsResult(v any) []any {
	l, _ := v.([]float64)
	return []any{n22Box(l)}
}

func intsResult(v any) []any {
	l, _ := v.([]int64)
	return []any{n22Box(l)}
}

func (h *n22Host) funcs() gostrict.Funcs {
	const p = "llm.local/n22h."
	return gostrict.Funcs{
		p + "Floats": func([]any) []any { return []any{h.floatsOut(h.floats)} },
		p + "Ints":   func([]any) []any { return []any{h.intsOut(h.ints)} },
		p + "Mutate": func([]any) []any {
			for i := range h.floats {
				h.floats[i] = -3*h.floats[i] + 7
			}
			for i := range h.ints {
				h.ints[i] = ^h.ints[i]
			}
			return nil
		},
		p + "TakeFloats": func(args []any) []any {
			h.gotForm = fmt.Sprintf("%T", args[0])
			h.gotFloats, _ = n22Unbox[float64](args[0])
			return []any{int32(len(h.gotFloats))}
		},
		p + "TakeInts": func(args []any) []any {
			h.gotForm = fmt.Sprintf("%T", args[0])
			h.gotInts, _ = n22Unbox[int64](args[0])
			return []any{int32(len(h.gotInts))}
		},
		p + "TakeStrings": func(args []any) []any {
			l, _ := n22Unbox[string](args[0])
			return []any{int32(len(l))}
		},
		p + "CallFloats": func(args []any) []any {
			return floatsResult(h.callFloats(args[0].(gostrict.Callback), h.floats, true))
		},
		p + "CallInts": func(args []any) []any {
			return intsResult(h.callInts(args[0].(gostrict.Callback), h.ints, true))
		},
		p + "CallTwiceFloats": func(args []any) []any {
			cb := args[0].(gostrict.Callback)
			h.callFloats(cb, h.floats, true)
			for i := range h.floats {
				h.floats[i] = -3*h.floats[i] + 7
			}
			return floatsResult(h.callFloats(cb, h.floats, true))
		},
		p + "CallTwiceInts": func(args []any) []any {
			cb := args[0].(gostrict.Callback)
			h.callInts(cb, h.ints, true)
			for i := range h.ints {
				h.ints[i] = ^h.ints[i]
			}
			return intsResult(h.callInts(cb, h.ints, true))
		},
		p + "CountFloats": func(args []any) []any {
			n, _ := h.callFloats(args[0].(gostrict.Callback), h.floats, false).(int32)
			return []any{n}
		},
		p + "CountInts": func(args []any) []any {
			n, _ := h.callInts(args[0].(gostrict.Callback), h.ints, false).(int32)
			return []any{n}
		},
		p + "FromFloats": func(args []any) []any {
			h.gotFloats = h.readFloats(args[0].(gostrict.Callback))
			return []any{n22Box(h.gotFloats)}
		},
		p + "FromInts": func(args []any) []any {
			h.gotInts = h.readInts(args[0].(gostrict.Callback))
			return []any{n22Box(h.gotInts)}
		},
		p + "FromCountFloats": func(args []any) []any {
			return []any{int32(len(h.readFloats(args[0].(gostrict.Callback))))}
		},
		p + "FoldFloats": func(args []any) []any {
			acc := args[0].(float64)
			cb := args[1].(gostrict.Callback)
			fold := func(b []float64) bool {
				if h.mode == n22Fast {
					fc := cb.(gostrict.FastCallback)
					w := fc.Begin()
					defer fc.End()
					w.Float64(acc)
					n22WriteFloat64s(w, b)
					r, err := fc.Invoke()
					if err != nil {
						h.callErr = err
						return false
					}
					acc = r.Float64()
					return true
				}
				res, err := cb.Call(acc, h.floatsOut(b))
				if err != nil {
					h.callErr = err
					return false
				}
				acc = res[0].(float64)
				return true
			}
			if h.scan != nil {
				if err := h.scan(fold); err != nil {
					h.callErr = err
				}
				return []any{acc}
			}
			for _, b := range h.blocks {
				if !fold(b) {
					break
				}
			}
			return []any{acc}
		},
	}
}

// readFloats calls cb without arguments and reads the list of floats it returns: with the fast reader in the fast
// mode, else from the neutral result.
func (h *n22Host) readFloats(cb gostrict.Callback) []float64 {
	if h.mode == n22Fast {
		fc := cb.(gostrict.FastCallback)
		fc.Begin()
		defer fc.End()
		r, err := fc.Invoke()
		if err != nil {
			h.callErr = err
			return nil
		}
		return slices.Clone(n22ReadFloat64s(r))
	}
	res, err := cb.Call()
	if err != nil {
		h.callErr = err
		return nil
	}
	l, _ := n22Unbox[float64](res[0])
	return l
}

func (h *n22Host) readInts(cb gostrict.Callback) []int64 {
	if h.mode == n22Fast {
		fc := cb.(gostrict.FastCallback)
		fc.Begin()
		defer fc.End()
		r, err := fc.Invoke()
		if err != nil {
			h.callErr = err
			return nil
		}
		return slices.Clone(n22ReadInt64s(r))
	}
	res, err := cb.Call()
	if err != nil {
		h.callErr = err
		return nil
	}
	l, _ := n22Unbox[int64](res[0])
	return l
}

// The values of the tables: the special floats and the bounds of int64, in lists of the lengths 0, 1, a few and
// 10 000.
var (
	n22SpecialFloats = []float64{math.NaN(), math.Inf(1), math.Inf(-1), math.Copysign(0, -1), 0, 1.5, -2.25,
		math.MaxFloat64, -math.MaxFloat64, math.SmallestNonzeroFloat64, 0.1, 1e-300, math.Float64frombits(0x7ff8000000000abc)}
	n22SpecialInts = []int64{math.MinInt64, math.MaxInt64, 0, -1, 1, 1<<53 + 1, -(1 << 53) - 1, math.MinInt64 + 1,
		math.MaxInt64 - 1, 255, 256}
)

func n22Floats(n int) []float64 {
	v := make([]float64, n)
	for i := range v {
		if i < len(n22SpecialFloats) {
			v[i] = n22SpecialFloats[i]
		} else {
			v[i] = float64(i)*0.25 + 1000.5
		}
	}
	return v
}

func n22Ints(n int) []int64 {
	v := make([]int64, n)
	for i := range v {
		if i < len(n22SpecialInts) {
			v[i] = n22SpecialInts[i]
		} else {
			v[i] = int64(i)*1000003 - 5_000_000_000
		}
	}
	return v
}

var n22Lengths = []int{0, 1, len(n22SpecialFloats), 10_000}

// n22Load loads the bundle in e with a host of the mode and the lists of length n.
func n22Load(t testing.TB, e gostrict.Engine, b *gostrict.Bundle, mode n22Mode, n int) (*n22Host, gostrict.Instance) {
	t.Helper()
	h := &n22Host{mode: mode, floats: n22Floats(n), ints: n22Ints(n)}
	in, err := e.Load(context.Background(), b, h.funcs())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = in.Close() })
	return h, in
}

// sameFloats compares floats by their bits, NaN with NaN as gostrict.SameValues does.
func sameFloats(a, b []float64) bool {
	return len(a) == len(b) && gostrict.SameValues(n22Box(a), n22Box(b))
}

// expectFloats is what the script sees of floats in Go: the number of elements, of NaN, -0, +Inf, -Inf and the sum of
// the others, as describeF of the program.
func expectFloats(v []float64) []float64 {
	var n, nan, negZero, posInf, negInf, sum float64
	for _, x := range v {
		n++
		switch {
		case math.IsNaN(x):
			nan++
		case math.IsInf(x, 1):
			posInf++
		case math.IsInf(x, -1):
			negInf++
		case x == 0 && math.Signbit(x):
			negZero++
		default:
			sum += x
		}
	}
	return []float64{n, nan, negZero, posInf, negInf, sum}
}

func expectInts(v []int64) []int64 {
	lo, hi, x := int64(math.MaxInt64), int64(math.MinInt64), int64(0)
	for _, e := range v {
		lo, hi, x = min(lo, e), max(hi, e), x^e
	}
	return []int64{int64(len(v)), lo, hi, x}
}

// n22Call calls the function name and returns its first result as a list of T; it reports a failure and returns false,
// so that a row goes on with the next function.
func n22Call[T any](t *testing.T, in gostrict.Instance, name string, args ...any) ([]T, bool) {
	t.Helper()
	res, err := in.Call(context.Background(), name, args...)
	if err != nil {
		t.Errorf("%s: %v", name, err)
		return nil, false
	}
	l, ok := n22Unbox[T](res[0])
	if !ok {
		t.Errorf("%s returned %T", name, res[0])
	}
	return l, ok
}

// TestN22ListValues is the table of the values: each script function under each mode of the host beside the counter
// row with []any, in both engines and for each length. A script function returns what it saw (Echo*, Call*Echo*), what
// it counted (See*) or what the host read back (From*, Pass*); every row must agree with []any and with Go.
func TestN22ListValues(t *testing.T) {
	b := n22Bundle(t)
	for _, eng := range n22Engines() {
		modes := []n22Mode{n22Any, n22Typed}
		if eng.name == "vm" {
			modes = append(modes, n22Fast)
		}
		for _, n := range n22Lengths {
			wantF := n22Floats(n)
			wantI := n22Ints(n)
			for _, mode := range modes {
				t.Run(fmt.Sprintf("%s/%s/%d", eng.name, mode, n), func(t *testing.T) {
					// the host result is typed only in n22Typed, a callback in n22Typed and n22Fast
					h, in := n22Load(t, eng.e, b, mode, n)
					resultRows := mode != n22Fast
					if resultRows {
						if got, ok := n22Call[float64](t, in, "EchoFloats"); ok && !sameFloats(got, wantF) {
							t.Errorf("EchoFloats: the script returned %v, want %v", head(got), head(wantF))
						}
						if got, ok := n22Call[float64](t, in, "SeeFloats"); ok && !sameFloats(got, expectFloats(wantF)) {
							t.Errorf("SeeFloats: the script saw %v, want %v", got, expectFloats(wantF))
						}
						if got, ok := n22Call[int64](t, in, "EchoInts"); ok && !slices.Equal(got, wantI) {
							t.Errorf("EchoInts: the script returned %v, want %v", head(got), head(wantI))
						}
						if got, ok := n22Call[int64](t, in, "SeeInts"); ok && !slices.Equal(got, expectInts(wantI)) {
							t.Errorf("SeeInts: the script saw %v, want %v", got, expectInts(wantI))
						}
						// a list, which the script got from the host, goes back as an argument of a host function
						if _, err := in.Call(context.Background(), "PassFloats"); err != nil || !sameFloats(h.gotFloats, wantF) {
							t.Errorf("PassFloats: the host got %v (%s, %v), want %v", head(h.gotFloats), h.gotForm, err, head(wantF))
						}
						if _, err := in.Call(context.Background(), "PassInts"); err != nil || !slices.Equal(h.gotInts, wantI) {
							t.Errorf("PassInts: the host got %v (%s, %v), want %v", head(h.gotInts), h.gotForm, err, head(wantI))
						}
					}
					if got, ok := n22Call[float64](t, in, "CallEchoFloats"); ok && (!sameFloats(got, wantF) || h.callErr != nil) {
						t.Errorf("CallEchoFloats: the callback returned %v (%v), want %v", head(got), h.callErr, head(wantF))
					}
					if got, ok := n22Call[float64](t, in, "CallSeeFloats"); ok && !sameFloats(got, expectFloats(wantF)) {
						t.Errorf("CallSeeFloats: the callback saw %v, want %v", got, expectFloats(wantF))
					}
					if got, ok := n22Call[int64](t, in, "CallEchoInts"); ok && (!slices.Equal(got, wantI) || h.callErr != nil) {
						t.Errorf("CallEchoInts: the callback returned %v (%v), want %v", head(got), h.callErr, head(wantI))
					}
					if got, ok := n22Call[int64](t, in, "CallSeeInts"); ok && !slices.Equal(got, expectInts(wantI)) {
						t.Errorf("CallSeeInts: the callback saw %v, want %v", got, expectInts(wantI))
					}
					// the host reads a list, which a callback returns: with the fast reader in the fast mode, else from
					// the neutral result of Callback.Call
					if got, ok := n22Call[float64](t, in, "FromFloats"); ok && (!sameFloats(got, wantF) || h.callErr != nil) {
						t.Errorf("FromFloats: the host read %v (%v), want %v", head(got), h.callErr, head(wantF))
					}
					if got, ok := n22Call[int64](t, in, "FromInts"); ok && (!slices.Equal(got, wantI) || h.callErr != nil) {
						t.Errorf("FromInts: the host read %v (%v), want %v", head(got), h.callErr, head(wantI))
					}
				})
			}
		}
	}
}

// head shortens a long list for a message.
func head[T any](v []T) string {
	if len(v) > 16 {
		return fmt.Sprintf("%v … (%d elements)", v[:16], len(v))
	}
	return fmt.Sprint(v)
}

// TestN22ListOwn: the script gets a list of its own. A list, which a host function returned, keeps its values when
// the host changes its slice afterwards (OwnFloats), and a change of the script does not reach the slice of the host
// (SetFloats); a callback, which keeps the list of its first call, still sees its values in the second, after the host
// changed its slice in place (CallKeepFloats), and a change of the callback does not reach the slice of the host
// (CallSetFloats). Counter row: the same functions with []any.
func TestN22ListOwn(t *testing.T) {
	b := n22Bundle(t)
	const n = 100
	for _, eng := range n22Engines() {
		modes := []n22Mode{n22Any, n22Typed}
		if eng.name == "vm" {
			modes = append(modes, n22Fast)
		}
		for _, mode := range modes {
			t.Run(eng.name+"/"+mode.String(), func(t *testing.T) {
				wantF, wantI := n22Floats(n), n22Ints(n)
				if mode != n22Fast {
					h, in := n22Load(t, eng.e, b, mode, n)
					if got, ok := n22Call[float64](t, in, "OwnFloats"); ok && (!sameFloats(got, wantF) || sameFloats(h.floats, wantF)) {
						t.Errorf("OwnFloats: the kept list is %v and the slice of the host %v, want %v and a change", head(got), head(h.floats), head(wantF))
					}
					h, in = n22Load(t, eng.e, b, mode, n)
					if got, ok := n22Call[int64](t, in, "OwnInts"); ok && (!slices.Equal(got, wantI) || slices.Equal(h.ints, wantI)) {
						t.Errorf("OwnInts: the kept list is %v and the slice of the host %v, want %v and a change", head(got), head(h.ints), head(wantI))
					}
					h, in = n22Load(t, eng.e, b, mode, n)
					if _, err := in.Call(context.Background(), "SetFloats"); err != nil || !sameFloats(h.floats, wantF) {
						t.Errorf("SetFloats: the slice of the host is %v (%v), want %v", head(h.floats), err, head(wantF))
					}
					if _, err := in.Call(context.Background(), "SetInts"); err != nil || !slices.Equal(h.ints, wantI) {
						t.Errorf("SetInts: the slice of the host is %v (%v), want %v", head(h.ints), err, head(wantI))
					}
				}
				h, in := n22Load(t, eng.e, b, mode, n)
				// premise: the host changed its slice between the two calls
				if got, ok := n22Call[float64](t, in, "CallKeepFloats"); ok && (!sameFloats(got, wantF) || sameFloats(h.floats, wantF) || h.callErr != nil) {
					t.Errorf("CallKeepFloats: the kept list of the first call is %v and the slice of the host %v (%v), want %v and a change", head(got), head(h.floats), h.callErr, head(wantF))
				}
				if got, ok := n22Call[int64](t, in, "CallKeepInts"); ok && (!slices.Equal(got, wantI) || slices.Equal(h.ints, wantI) || h.callErr != nil) {
					t.Errorf("CallKeepInts: the kept list of the first call is %v and the slice of the host %v (%v), want %v and a change", head(got), head(h.ints), h.callErr, head(wantI))
				}
				h, in = n22Load(t, eng.e, b, mode, n)
				changed := slices.Clone(wantF)
				changed[0] = 42
				if got, ok := n22Call[float64](t, in, "CallSetFloats"); ok && (!sameFloats(got, changed) || !sameFloats(h.floats, wantF)) {
					t.Errorf("CallSetFloats: the callback returned %v and the slice of the host is %v, want %v and %v", head(got), head(h.floats), head(changed), head(wantF))
				}
				changedI := slices.Clone(wantI)
				changedI[0] = 42
				if got, ok := n22Call[int64](t, in, "CallSetInts"); ok && (!slices.Equal(got, changedI) || !slices.Equal(h.ints, wantI)) {
					t.Errorf("CallSetInts: the callback returned %v and the slice of the host is %v, want %v and %v", head(got), head(h.ints), head(changedI), head(wantI))
				}
			})
		}
	}
}

// n22WrongTypeMessages are the messages of nago f4aca80c5, before N22, by engine, script function and mode: a host
// function, which returns a []int64 for a list.List[float64] and a []float64 for a list.List[int64] (LenFloats,
// LenInts), and a host, which calls a callback with them (CountFloats, CountInts). L-N22-23 keeps them with typed
// lists. Observed by the counter run of TestN22ListWrongType against
// f4aca80c5; the vm numbers the closures of a program differently from compilation to compilation, so funcN stands for
// the number (n22Normalize).
var n22WrongTypeMessages = map[string]string{
	"vm/LenFloats/wrong element type":   "vm: value 0 of llm.local/n22h.Floats: must be a list.List[float64], got []int64",
	"vm/LenInts/wrong element type":     "vm: value 0 of llm.local/n22h.Ints: must be a list.List[int64], got []float64",
	"vm/CountFloats/wrong element type": "callback: value 0 of llm.local/n22.CountFloats.funcN: []int64 does not fit; call: a callback of the program failed: value 0 of llm.local/n22.CountFloats.funcN: []int64 does not fit",
	"vm/CountInts/wrong element type":   "callback: value 0 of llm.local/n22.CountInts.funcN: []float64 does not fit; call: a callback of the program failed: value 0 of llm.local/n22.CountInts.funcN: []float64 does not fit",

	"goja/LenFloats/wrong element type":   "javascript: GoError: value 0 of llm.local/n22h.Floats: must be a list.List[float64], got []int64",
	"goja/LenInts/wrong element type":     "javascript: GoError: value 0 of llm.local/n22h.Ints: must be a list.List[int64], got []float64",
	"goja/CountFloats/wrong element type": "callback: value 0 of callback: must be a list.List[float64], got []int64; call: a callback of the program failed: value 0 of callback: must be a list.List[float64], got []int64",
	"goja/CountInts/wrong element type":   "callback: value 0 of callback: must be a list.List[int64], got []float64; call: a callback of the program failed: value 0 of callback: must be a list.List[int64], got []float64",
}

var n22FuncNumber = regexp.MustCompile(`\.func[0-9]+`)

// n22Normalize replaces the number of a closure in a message by N.
func n22Normalize(msg string) string { return n22FuncNumber.ReplaceAllString(msg, ".funcN") }

// TestN22ListWrongType: a list of the wrong element type, a []float64 for a list.List[int64] or a []int64 for a
// list.List[float64], aborts with the same message as before N22, as a []any of the wrong element type does; counter
// row: the same call with []any of the right type succeeds.
func TestN22ListWrongType(t *testing.T) {
	b := n22Bundle(t)
	for _, eng := range n22Engines() {
		for _, fn := range []string{"LenFloats", "LenInts", "CountFloats", "CountInts"} {
			for _, mode := range []n22Mode{n22WrongElem, n22WrongAny} {
				t.Run(eng.name+"/"+fn+"/"+mode.String(), func(t *testing.T) {
					_, in := n22Load(t, eng.e, b, n22Any, 3)
					if _, err := in.Call(context.Background(), fn); err != nil {
						t.Fatalf("counter row: with []any of the right type %s fails: %v", fn, err)
					}
					h, in := n22Load(t, eng.e, b, mode, 3)
					_, err := in.Call(context.Background(), fn)
					msg := fmt.Sprint(err)
					if h.callErr != nil {
						msg = "callback: " + h.callErr.Error() + "; call: " + msg
					}
					msg = n22Normalize(msg)
					if mode == n22WrongAny {
						// a row beside: []any of a wrong element type aborts as before; its message is not part of
						// L-N22-23 (in the vm a callback reports the failed conversion of Go)
						if err == nil {
							t.Errorf("a []any of the wrong element type does not abort")
						}
						t.Logf("[]any of the wrong element type: %s", msg)
						return
					}
					want, ok := n22WrongTypeMessages[eng.name+"/"+fn+"/"+mode.String()]
					if !ok {
						t.Fatalf("no message of f4aca80c5 recorded; observed: %q", msg)
					}
					if err == nil || msg != want {
						t.Errorf("the wrong element type gives %q, want the message before N22 %q", msg, want)
					}
				})
			}
		}
	}
}

// n22Allocs is the number of allocations of a call of fn of the program, with lists of length n.
func n22Allocs(t *testing.T, e gostrict.Engine, b *gostrict.Bundle, mode n22Mode, n int, fn string, withData bool) float64 {
	t.Helper()
	h, in := n22Load(t, e, b, mode, n)
	h.blocks = [][]float64{h.floats}
	ctx := context.Background()
	var args []any
	if withData {
		res, err := in.Call(ctx, "NewData", int32(n))
		if err != nil {
			t.Fatal(err)
		}
		args = res
	}
	call := func() {
		if _, err := in.Call(ctx, fn, args...); err != nil {
			t.Fatalf("%s: %v", fn, err)
		}
	}
	call()
	if h.callErr != nil {
		t.Fatalf("%s: %v", fn, h.callErr)
	}
	return testing.AllocsPerRun(20, call)
}

// TestN22ListAllocs: the number of allocations of a passing does not depend on the length (L-N22-23), in the vm
// (L-N22-27; the JS runtime only gives the same values and messages, TestN22ListValues and TestN22ListWrongType): a call
// with lists of 10 000 elements allocates at most n22AllocSlack more than one with 10. Rows: a typed result of a host
// function (LenFloats), a list of the program as an argument of a host function (PassDataFloats), a callback with a
// typed list through Callback.Call and through the fast writer (CountFloats), and the fast reader of a list, which a
// callback returns (FromCountFloats). Counter rows show that the instrument sees a dependence on the length: a host,
// which boxes its floats into a []any for each call, as before N22, and a list.List[string] as an argument, which stays
// a []any.
func TestN22ListAllocs(t *testing.T) {
	const n22AllocSlack = 8
	b := n22Bundle(t)
	type row struct {
		name     string
		mode     n22Mode
		fn       string
		withData bool
		grows    bool // a counter row
	}
	rows := []row{
		{"typed result of a host function", n22Typed, "LenFloats", false, false},
		{"typed result of a host function, ints", n22Typed, "LenInts", false, false},
		{"counter row: a result boxed into []any per call", n22Any, "LenFloats", false, true},
		{"list of the program as an argument of a host function", n22Any, "PassDataFloats", true, false},
		{"list of the program as an argument of a host function, ints", n22Any, "PassDataInts", true, false},
		{"counter row: a list.List[string] as an argument stays []any", n22Any, "PassDataStrings", true, true},
		{"typed list through Callback.Call", n22Typed, "CountFloats", false, false},
		{"typed list through Callback.Call, ints", n22Typed, "CountInts", false, false},
		{"counter row: a callback with a list boxed into []any per call", n22Any, "CountFloats", false, true},
		{"the fast writer with Writer.Float64s", n22Fast, "CountFloats", false, false},
		{"the fast writer with Writer.Int64s", n22Fast, "CountInts", false, false},
		{"the fast reader with Reader.Float64s", n22Fast, "FromCountFloats", true, false},
	}
	e := &vm.Engine{}
	for _, r := range rows {
		t.Run("vm/"+r.name, func(t *testing.T) {
			small := n22Allocs(t, e, b, r.mode, 10, r.fn, r.withData)
			large := n22Allocs(t, e, b, r.mode, 10_000, r.fn, r.withData)
			t.Logf("%s: %.0f allocations with 10 elements, %.0f with 10 000", r.fn, small, large)
			if r.grows {
				if large-small < 5_000 {
					t.Errorf("counter row: %.0f and %.0f allocations, the instrument does not see the boxing of 10 000 elements", small, large)
				}
				return
			}
			if math.Abs(large-small) > n22AllocSlack {
				t.Errorf("%.0f allocations with 10 elements, %.0f with 10 000: the passing depends on the length", small, large)
			}
		})
	}
}
