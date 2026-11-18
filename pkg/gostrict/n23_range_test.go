// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"errors"
	"math"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
)

// The acceptance tests of R-DEC-SCRIPT-GS-RANGE (wokoda N23, L-N23-19): the vm runs a range over a list of a scalar,
// whose body begins with the assignment of the element, with fused instructions and every other loop as before; each
// script yields the same values, results and aborts as before and as with goja. "As before" is the value Go computes
// for the same loop (n23RangeWant), which the vm of 8ae1daaee yields; the rows run in the vm and in goja.
//
// That the vm fuses shows the translated program (Program.Dump of the bundle): there is an instruction, which no vm of
// 8ae1daaee had (n23OldOps, the names of internal/vmcode/dump.go of that stand), in each function of a fused row and in
// none of the last row. The test does not know its name; the description of the change names it.

const n23RangeHostSrc = `// Package n23h ends the context of a call.
//
//gostrict:header
package n23h

// Stop ends the context of the running call.
//
//gostrict:host
func Stop() {}
`

const n23RangeSrc = `// Package n23 holds the loops of R-DEC-SCRIPT-GS-RANGE.
//
//gostrict:terminating
package n23

import (
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"llm.local/n23h"
)

func i32s(n int32) list.List[int32] {
	l := list.New[int32]()
	for k := int32(0); k < n; k++ {
		l.Append(k*7 - 3*n)
	}
	return l
}

func i64s(n int32) list.List[int64] {
	l := list.New[int64]()
	for k := int32(0); k < n; k++ {
		l.Append(int64(k)*1000003 - 5)
	}
	return l
}

func f64s(n int32) list.List[float64] {
	l := list.New[float64]()
	for k := int32(0); k < n; k++ {
		l.Append(float64(k)*0.1 + 1.0/3.0)
	}
	return l
}

func bools(n int32) list.List[bool] {
	l := list.New[bool]()
	for k := int32(0); k < n; k++ {
		l.Append(k%3 == 0)
	}
	return l
}

func Empty() int32 {
	runs := int32(0)
	for _, x := range list.New[int32]().All() {
		runs += 1 + x
	}
	return runs
}

func One() int64 {
	l := list.New[int64]()
	l.Append(42)
	s, runs := int64(0), int64(0)
	for _, x := range l.All() {
		s += x
		runs++
	}
	return s*10 + runs
}

func SumI32(n int32) int32 {
	s := int32(0)
	for _, x := range i32s(n).All() {
		s += x
	}
	return s
}

func SumI64(n int32) int64 {
	s := int64(0)
	for _, x := range i64s(n).All() {
		s += x
	}
	return s
}

func SumF64(n int32) float64 {
	s := 0.0
	for _, x := range f64s(n).All() {
		s += x
	}
	return s
}

func CountBool(n int32) int32 {
	c := int32(0)
	for _, b := range bools(n).All() {
		if b {
			c++
		}
	}
	return c
}

func Break(n int32) int64 {
	s := int64(0)
	for _, x := range i64s(n).All() {
		s += x
		break
	}
	return s
}

func Continue(n int32) int64 {
	s := int64(0)
	for i, x := range i64s(n).All() {
		if i%2 == 1 {
			continue
		}
		s += x
	}
	return s
}

func Return(n int32) int64 {
	for _, x := range i64s(n).All() {
		if x > 3000000 {
			return x
		}
	}
	return -1
}

func Nested(n int32) int64 {
	l := i64s(n)
	s := int64(0)
	for _, x := range l.All() {
		for _, y := range l.All() {
			s += x ^ y
		}
	}
	return s
}

func SetInBody(n int32) int64 {
	l := i64s(n)
	d := l.Len()
	if d == 0 {
		return 0
	}
	s, runs := int64(0), int64(0)
	for i, x := range l.All() {
		_ = l.Set((i+1)%d, x*2)
		s += x
		runs++
	}
	return s*1000 + runs
}

func AppendInBody(n int32) int64 {
	l := i64s(n)
	s, runs := int64(0), int64(0)
	for _, x := range l.All() {
		l.Append(x + 1)
		s += x
		runs++
	}
	return s*1000 + runs
}

func PopInBody(n int32) int64 {
	l := i64s(n)
	s, runs := int64(0), int64(0)
	for _, x := range l.All() {
		_, _ = l.Pop()
		s += x
		runs++
	}
	return s*1000 + runs
}

func DeleteInBody(n int32) int64 {
	l := i64s(n)
	s, runs := int64(0), int64(0)
	for _, x := range l.All() {
		_ = l.Delete(0)
		s += x
		runs++
	}
	return s*1000 + runs
}

func IndexUsed(n int32) int64 {
	s := int64(0)
	for i, x := range i64s(n).All() {
		s += int64(i) * x
	}
	return s
}

func IndexOmitted(n int32) int64 {
	s := int64(0)
	for x := range i64s(n).Values() {
		s += x
	}
	return s
}

// Stopped ends its context at the element 1000 and runs on over a large list.
func Stopped(n int32) int64 {
	s := int64(0)
	for i, x := range i64s(n).All() {
		s += x
		if i == 1000 {
			n23h.Stop()
		}
	}
	return s
}

// NotFirst ranges over the indexes only: its body does not begin with the assignment of an element.
func NotFirst(n int32) int64 {
	l := i64s(n)
	s := int64(0)
	for i := range l.All() {
		v, _ := l.Get(i)
		s += v
	}
	return s
}

func Strings(n int32) int32 {
	l := list.New[string]()
	for k := int32(0); k < n; k++ {
		l.Append("ab")
	}
	c := int32(0)
	for _, x := range l.All() {
		if x == "ab" {
			c += 2
		}
	}
	return c
}
`

// n23RangeRows are the functions of the table; fused says, whether its loops are fused.
var n23RangeRows = []struct {
	fn    string
	n     int32
	fused bool
}{
	{"Empty", 0, true}, {"One", 0, true},
	{"SumI32", 1000, true}, {"SumI64", 1000, true}, {"SumF64", 1000, true}, {"CountBool", 1000, true},
	{"Break", 1000, true}, {"Continue", 1000, true}, {"Return", 1000, true}, {"Nested", 100, true},
	{"SetInBody", 100, true}, {"AppendInBody", 100, true}, {"PopInBody", 100, true}, {"DeleteInBody", 100, true},
	{"IndexUsed", 1000, true}, {"IndexOmitted", 1000, true},
	{"NotFirst", 1000, false}, {"Strings", 1000, false},
}

// n23RangeWant is the result of a row as Go computes the same loop.
func n23RangeWant(fn string, n int32) any {
	i64s := func() []int64 {
		l := make([]int64, n)
		for k := range l {
			l[k] = int64(k)*1000003 - 5
		}
		return l
	}
	switch fn {
	case "Empty":
		return int32(0)
	case "One":
		return int64(421)
	case "SumI32":
		s := int32(0)
		for k := int32(0); k < n; k++ {
			s += k*7 - 3*n
		}
		return s
	case "SumI64", "IndexOmitted", "NotFirst":
		s := int64(0)
		for _, x := range i64s() {
			s += x
		}
		return s
	case "SumF64":
		s := 0.0
		for k := int32(0); k < n; k++ {
			s += float64(k)*0.1 + 1.0/3.0
		}
		return s
	case "CountBool":
		c := int32(0)
		for k := int32(0); k < n; k++ {
			if k%3 == 0 {
				c++
			}
		}
		return c
	case "Break":
		return i64s()[0]
	case "Continue":
		s := int64(0)
		for i, x := range i64s() {
			if i%2 == 0 {
				s += x
			}
		}
		return s
	case "Return":
		for _, x := range i64s() {
			if x > 3000000 {
				return x
			}
		}
		return int64(-1)
	case "Nested":
		l := i64s()
		s := int64(0)
		for _, x := range l {
			for _, y := range l {
				s += x ^ y
			}
		}
		return s
	case "SetInBody":
		// a range visits the elements the list has when it starts, each read when its round comes (std/list)
		l := i64s()
		s, runs := int64(0), int64(0)
		for i := 0; i < len(l); i++ {
			x := l[i]
			l[(i+1)%len(l)] = x * 2
			s += x
			runs++
		}
		return s*1000 + runs
	case "AppendInBody":
		l := i64s()
		s, runs := int64(0), int64(0)
		for i, end := 0, len(l); i < end; i++ {
			x := l[i]
			l = append(l, x+1)
			s += x
			runs++
		}
		return s*1000 + runs
	case "PopInBody":
		l := i64s()
		s, runs := int64(0), int64(0)
		for i, end := 0, len(l); i < end && i < len(l); i++ {
			x := l[i]
			l = l[:len(l)-1]
			s += x
			runs++
		}
		return s*1000 + runs
	case "DeleteInBody":
		l := i64s()
		s, runs := int64(0), int64(0)
		for i, end := 0, len(l); i < end && i < len(l); i++ {
			x := l[i]
			l = l[1:]
			s += x
			runs++
		}
		return s*1000 + runs
	case "IndexUsed":
		s := int64(0)
		for i, x := range i64s() {
			s += int64(i) * x
		}
		return s
	case "Strings":
		return 2 * n
	}
	panic(fn)
}

func n23RangeBundle(t testing.TB) *gostrict.Bundle {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/n23",
		gostrict.Module{Path: "llm.local/n23", FS: fstest.MapFS{"n23.go": {Data: []byte(n23RangeSrc)}}},
		gostrict.Module{Path: "llm.local/n23h", FS: fstest.MapFS{"n23h.go": {Data: []byte(n23RangeHostSrc)}}, Trusted: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) > 0 {
		t.Fatalf("the program does not compile:\n%s", diags.Text())
	}
	return b
}

// n23Stop is the cancel of the context of the running call, which the host function Stop calls.
var n23Stop context.CancelFunc

func n23RangeHost() gostrict.Funcs {
	return gostrict.Funcs{"llm.local/n23h.Stop": func([]any) []any {
		if n23Stop != nil {
			n23Stop()
		}
		return nil
	}}
}

// n23Same compares two results of a row, a float64 by its bits.
func n23Same(a, b any) bool {
	if x, ok := a.(float64); ok {
		y, ok := b.(float64)
		return ok && math.Float64bits(x) == math.Float64bits(y)
	}
	return a == b
}

// TestN23RangeValues: each row of the table yields in the vm and in goja the value Go computes for it, a sum bit for
// bit; the row with a context, which ends during the loop over a large list, aborts with the error of the context in
// both, and without the end of the context it yields the sum.
func TestN23RangeValues(t *testing.T) {
	b := n23RangeBundle(t)
	for _, eng := range n22Engines() {
		t.Run(eng.name, func(t *testing.T) {
			in, err := eng.e.Load(context.Background(), b, n23RangeHost())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = in.Close() }()
			for _, r := range n23RangeRows {
				var args []any
				if r.fn != "Empty" && r.fn != "One" {
					args = []any{r.n}
				}
				res, err := in.Call(context.Background(), r.fn, args...)
				if err != nil {
					t.Errorf("%s: %v", r.fn, err)
					continue
				}
				if want := n23RangeWant(r.fn, r.n); len(res) != 1 || !n23Same(res[0], want) {
					t.Errorf("%s(%d) = %v (%T), want %v (%T)", r.fn, r.n, res, res, want, want)
				}
			}
			t.Run("the context ends during a loop over a large list", func(t *testing.T) {
				const n = 300_000
				ctx, cancel := context.WithCancel(context.Background())
				n23Stop = cancel
				defer func() { n23Stop = nil; cancel() }()
				_, err := in.Call(ctx, "Stopped", int32(n))
				if err == nil || !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), context.Canceled.Error()) {
					t.Errorf("Stopped: %v, want the error of the context", err)
				}
				// counter row: without the end of the context the sum
				n23Stop = nil
				in2, err := eng.e.Load(context.Background(), b, n23RangeHost())
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = in2.Close() }()
				res, err := in2.Call(context.Background(), "Stopped", int32(n))
				want := int64(0)
				for k := int64(0); k < n; k++ {
					want += k*1000003 - 5
				}
				if err != nil || len(res) != 1 || res[0] != want {
					t.Errorf("Stopped without the end of the context: %v %v, want %d", res, err, want)
				}
			})
		})
	}
}

// n23OldOps are the names of the instructions of the vm of nago 8ae1daaee (internal/vmcode/dump.go of that stand).
var n23OldOps = strings.Fields(`nop movn movs movr movnk ldin ldkn ldks ldkr ldgn ldgs ldgr stgn stgs stgr ldxn ldxs ldxr
stxn stxs stxr ldon ldos ldor ldonk ston stos stor new addi addik mulik subi muli divi remi divu remu andi ori xori
andnoti shli shri shru negi noti eqi nei lti lei ltu leu shliv shriv shruv addl addlk subl mull divl reml divul remul
shll shrl shrul shllv shrlv shrulv negl notl eql nel ltl lel ltul leul sxt32 zxt32 l2f ul2f jnltl jnlel jneql jnnel
jnltul jnleul jeqlk jnelk addf subf mulf divf negf eqf nef ltf lef i2f u2f notb isnil eqr ner eqv concat concatn eqs nes
lts les jmp loop incloop jnltlen jmpt jmpf jnlti jnlei jneqi jnnei jnltf jnlef jeqik jneik call callnative callnativer
callhost callvalue ret llen lgetn lgets lgetr lget lgetok lappn lapps lappr lapp lnew lset box zeror closure calliface
typeis view entry`)

// n23DumpOps parses the dump of a program into the names of the instructions of each function.
func n23DumpOps(dump string) map[string]map[string]bool {
	funcs := map[string]map[string]bool{}
	var cur map[string]bool
	for _, line := range strings.Split(dump, "\n") {
		if name, ok := strings.CutPrefix(line, "func "); ok {
			name, _, _ = strings.Cut(name, " ")
			cur = map[string]bool{}
			funcs[name] = cur
			continue
		}
		f := strings.Fields(line)
		if cur != nil && len(f) >= 2 && strings.HasPrefix(line, "  ") {
			cur[f[1]] = true
		}
	}
	return funcs
}

// TestN23RangeFused: the dump of the program has an instruction, which the vm of 8ae1daaee did not have, in every
// function of a fused row and in no function of the last row (a range over the indexes only and one over a list of
// texts). Counter row: no function of the last row has an instruction of its own beside those of 8ae1daaee.
func TestN23RangeFused(t *testing.T) {
	p, err := n23RangeBundle(t).VM()
	if err != nil {
		t.Fatal(err)
	}
	funcs := n23DumpOps(p.Dump())
	old := map[string]bool{}
	for _, op := range n23OldOps {
		old[op] = true
	}
	find := func(fn string) map[string]bool {
		for name, ops := range funcs {
			if name == fn || strings.HasSuffix(name, "."+fn) {
				return ops
			}
		}
		t.Fatalf("premise: the dump has no function %s: %v", fn, slices.Sorted(func(yield func(string) bool) {
			for n := range funcs {
				if !yield(n) {
					return
				}
			}
		}))
		return nil
	}
	var common map[string]bool
	for _, r := range n23RangeRows {
		if !r.fused {
			continue
		}
		news := map[string]bool{}
		for op := range find(r.fn) {
			if !old[op] {
				news[op] = true
			}
		}
		if common == nil {
			common = news
		} else {
			for op := range common {
				if !news[op] {
					delete(common, op)
				}
			}
		}
	}
	for _, r := range n23RangeRows {
		if r.fused {
			continue
		}
		for op := range find(r.fn) {
			delete(common, op)
			if !old[op] {
				t.Logf("%s has the instruction %s, which 8ae1daaee had not", r.fn, op)
			}
		}
	}
	if len(common) == 0 {
		t.Errorf("no instruction, which the vm of 8ae1daaee had not, is in every fused loop and in no loop of the last row: the loops are not fused")
	}
	t.Logf("the instructions of the fused loops: %v", slices.Sorted(func(yield func(string) bool) {
		for op := range common {
			if !yield(op) {
				return
			}
		}
	}))
}
