// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
)

// The tests of this file check the fixes of GS033 after the contract R-SCRIPT-GS-OPTIONAL of wokoda (L-N11-4) and
// the functions of std, which return the zero value beside false or an error. A fix is applied with
// [gostrict.ApplyEdits]; the fixed program has no finding, its line of the binding has the form of the contract, and
// it runs in goja and in the vm. For a type without nil it gives the same values as the program before the fix: that
// program is translated with [gostrict.Emit] without the check, which rejects it now, so the values before are
// observed in the same engines. The form of each fix, as the detail writes it, is also written by hand: it compiles
// without a finding and gives the same values, so that a run against nago 48d826d81 shows, that the form translates
// there. Whether the fix itself translates under nago 48d826d81, can not be shown in one suite: nago 48d826d81 gives
// no fix. For a type with nil (only GetOr) there is no value before; the fixed program gives an empty list, an empty
// map resp. 0 for each field. The counter rows have GS033 and no fix at all.

// optEngines are the engines of the rows, see diff_test.go.
var optEngines = []struct {
	name string
	e    gostrict.Engine
}{{"goja", goEngine}, {"vm", vmEngine}}

// optCall loads the bundle into each engine and calls Run with args.
func optCall(t *testing.T, b *gostrict.Bundle, args []any) map[string][]any {
	t.Helper()
	res := map[string][]any{}
	for _, e := range optEngines {
		in := mustLoad(t, e.e, b)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		got, err := in.Call(ctx, "Run", args...)
		cancel()
		_ = in.Close()
		if err != nil {
			t.Fatalf("%s: Run(%v): %v", e.name, args, err)
		}
		res[e.name] = got
	}
	return res
}

// optRunChecked compiles the program with the check, which must not find anything, and calls Run.
func optRunChecked(t *testing.T, src string, args []any) map[string][]any {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", optModule(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) > 0 {
		t.Fatalf("the program has findings:\n%s\n%s", diags.Text(), optNumbered(src))
	}
	return optCall(t, b, args)
}

// optRunRaw translates the program without the check and calls Run: the values of a program, which the check
// rejects since N11, in the same engines.
func optRunRaw(t *testing.T, src string, args []any) map[string][]any {
	t.Helper()
	prog, err := gostrict.LoadModules("llm.local/s", optModule(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Diagnostics) > 0 {
		t.Fatalf("the program is not valid Go:\n%s", prog.Diagnostics.Text())
	}
	b, diags := gostrict.Emit(prog)
	if len(diags) > 0 {
		t.Fatalf("the program does not translate without the check:\n%s\n%s", diags.Text(), optNumbered(src))
	}
	return optCall(t, b, args)
}

// optFixRow is a fix of the table.
type optFixRow struct {
	name string
	// src is the program; the lines, at which GS033 may stand, are marked "// at"
	src string
	// form matches the line of the binding after the fix
	form string
	// hand is the program after the fix in the form of the detail, written by hand, or "" for a type with nil
	hand string
	// calls are the arguments of Run
	calls [][]any
	// want are the results of Run after the fix for a type with nil, one per call
	want [][]any
}

// optFixDecls are the declarations of the rows.
const optFixDecls = `
type P struct {
	X int32
	S string
}

type S struct{ F int32 }

func find(i int32) (int32, bool) {
	if i == 0 {
		return 5, true
	}
	return 0, false
}

func load(s string) (int32, error) {
	if s == "" {
		return 0, errors.New("empty")
	}
	return 3, nil
}

func pick(i int32) (P, bool) {
	if i == 0 {
		return P{X: 1, S: "a"}, true
	}
	return P{}, false
}

func pos(i int32) (int32, bool) {
	ok := i > 0
	if ok {
		return i, true
	}
	return 0, ok
}
`

var optFixRows = []optFixRow{
	{name: "GetOr at list.List, int32", form: `elem := l\.GetOr\(i, .+\)`, src: `
func Run(i int32) int32 {
	l := list.Of[int32](5, 6)
	elem, _ := l.Get(i) // at
	return elem // at
}`, hand: `
func Run(i int32) int32 {
	l := list.Of[int32](5, 6)
	elem := l.GetOr(i, 0)
	return elem
}`, calls: [][]any{{int32(0)}, {int32(1)}, {int32(2)}, {int32(-1)}}},
	{name: "GetOr at list.Frozen, int32", form: `elem := f\.GetOr\(i, .+\)`, src: `
func Run(i int32) int32 {
	f := list.Freeze(list.Of[int32](5, 6))
	elem, _ := f.Get(i) // at
	return elem // at
}`, hand: `
func Run(i int32) int32 {
	f := list.Freeze(list.Of[int32](5, 6))
	elem := f.GetOr(i, 0)
	return elem
}`, calls: [][]any{{int32(1)}, {int32(2)}}},
	{name: "GetOr at dict.Map, int32", form: `elem := d\.GetOr\(k, .+\)`, src: `
func Run(k string) int32 {
	d := dict.New[string, int32]()
	d.Set("a", 4)
	elem, _ := d.Get(k) // at
	return elem // at
}`, hand: `
func Run(k string) int32 {
	d := dict.New[string, int32]()
	d.Set("a", 4)
	elem := d.GetOr(k, 0)
	return elem
}`, calls: [][]any{{"a"}, {"b"}}},
	{name: "GetOr at list.List, string", form: `elem := l\.GetOr\(i, .+\)`, src: `
func Run(i int32) string {
	l := list.Of("x", "y")
	elem, _ := l.Get(i) // at
	return elem // at
}`, hand: `
func Run(i int32) string {
	l := list.Of("x", "y")
	elem := l.GetOr(i, "")
	return elem
}`, calls: [][]any{{int32(1)}, {int32(2)}}},
	{name: "GetOr at list.List, opt.Option", form: `elem := l\.GetOr\(i, .+\)`, src: `
func Run(i int32) (int32, bool) {
	l := list.Of(opt.Some[int32](3))
	elem, _ := l.Get(i) // at
	return elem.Or(-1), elem.IsSome() // at
}`, hand: `
func Run(i int32) (int32, bool) {
	l := list.Of(opt.Some[int32](3))
	elem := l.GetOr(i, opt.None[int32]())
	return elem.Or(-1), elem.IsSome()
}`, calls: [][]any{{int32(0)}, {int32(1)}}},
	{name: "GetOr at list.List, a struct without lists", form: `elem := l\.GetOr\(i, .+\)`, src: optFixDecls + `
func Run(i int32) (int32, string) {
	l := list.Of(P{X: 1, S: "a"})
	elem, _ := l.Get(i) // at
	return elem.X, elem.S // at
}`, hand: optFixDecls + `
func Run(i int32) (int32, string) {
	l := list.Of(P{X: 1, S: "a"})
	elem := l.GetOr(i, P{})
	return elem.X, elem.S
}`, calls: [][]any{{int32(0)}, {int32(1)}}},
	{name: "opt.OrZero of list.Pop", form: `elem := opt\.OrZero\(l\.Pop\(\)\)`, src: `
func Run(n int32) int32 {
	l := list.New[int32]()
	for range n {
		l.Append(7)
	}
	elem, _ := l.Pop() // at
	return elem // at
}`, hand: `
func Run(n int32) int32 {
	l := list.New[int32]()
	for range n {
		l.Append(7)
	}
	elem := opt.OrZero(l.Pop())
	return elem
}`, calls: [][]any{{int32(0)}, {int32(1)}}},
	{name: "opt.OrZeroErr of text.ParseInt", form: `n := opt\.OrZeroErr\(text\.ParseInt\(s\)\)`, src: `
func Run(s string) int32 {
	n, _ := text.ParseInt(s) // at
	return n // at
}`, hand: `
func Run(s string) int32 {
	n := opt.OrZeroErr(text.ParseInt(s))
	return n
}`, calls: [][]any{{"12"}, {"x"}, {""}, {"1x"}, {"99999999999"}, {"-"}}},
	{name: "a binding with = to a field: GetOr", form: `s\.F = l\.GetOr\(i, .+\)`, src: optFixDecls + `
func Run(i int32) int32 {
	l := list.Of[int32](5, 6)
	s := S{F: 9}
	s.F, _ = l.Get(i) // at
	return s.F // at
}`, hand: optFixDecls + `
func Run(i int32) int32 {
	l := list.Of[int32](5, 6)
	s := S{F: 9}
	s.F = l.GetOr(i, 0)
	return s.F
}`, calls: [][]any{{int32(1)}, {int32(2)}}},
	{name: "a binding with = to a variable: opt.OrZeroErr", form: `x = opt\.OrZeroErr\(text\.ParseInt\(s\)\)`, src: `
func Run(s string) int32 {
	x := int32(9)
	x, _ = text.ParseInt(s) // at
	return x // at
}`, hand: `
func Run(s string) int32 {
	x := int32(9)
	x = opt.OrZeroErr(text.ParseInt(s))
	return x
}`, calls: [][]any{{"12"}, {"x"}}},
	{name: "an own function with return 0, false: opt.OrZero", form: `elem := opt\.OrZero\(find\(i\)\)`, src: optFixDecls + `
func Run(i int32) int32 {
	elem, _ := find(i) // at
	return elem // at
}`, hand: optFixDecls + `
func Run(i int32) int32 {
	elem := opt.OrZero(find(i))
	return elem
}`, calls: [][]any{{int32(0)}, {int32(1)}}},
	{name: "an own function with return 0, errors.New(…): opt.OrZeroErr", form: `elem := opt\.OrZeroErr\(load\(s\)\)`, src: optFixDecls + `
func Run(s string) int32 {
	elem, _ := load(s) // at
	return elem // at
}`, hand: optFixDecls + `
func Run(s string) int32 {
	elem := opt.OrZeroErr(load(s))
	return elem
}`, calls: [][]any{{"a"}, {""}}},
	{name: "an own function with return P{}, false: opt.OrZero", form: `elem := opt\.OrZero\(pick\(i\)\)`, src: optFixDecls + `
func Run(i int32) (int32, string) {
	elem, _ := pick(i) // at
	return elem.X, elem.S // at
}`, hand: optFixDecls + `
func Run(i int32) (int32, string) {
	elem := opt.OrZero(pick(i))
	return elem.X, elem.S
}`, calls: [][]any{{int32(0)}, {int32(1)}}},
	{name: "an own function with return 0, ok: opt.OrZero", form: `elem := opt\.OrZero\(pos\(i\)\)`, src: optFixDecls + `
func Run(i int32) int32 {
	elem, _ := pos(i) // at
	return elem // at
}`, hand: optFixDecls + `
func Run(i int32) int32 {
	elem := opt.OrZero(pos(i))
	return elem
}`, calls: [][]any{{int32(3)}, {int32(-3)}}},

	// a type with nil: the zero value, which the fix writes, has an empty list, an empty map resp. 0 in each field
	{name: "GetOr for a struct with list.List, list.Frozen, dict.Map, big.Int and decimal.Dec", form: `elem := l\.GetOr\(i, .+\)`, src: `
type In struct{ L list.List[int32] }

type Rec struct {
	N  int32
	L  list.List[int32]
	F  list.Frozen[int32]
	M  dict.Map[string, int32]
	B  big.Int
	D  decimal.Dec
	In In
}

func Run(i int32) (int32, int32, int32, int32, int32, int32, int32) {
	m := dict.New[string, int32]()
	m.Set("a", 1)
	l := list.Of(Rec{N: 1, L: list.Of[int32](1, 2), F: list.Freeze(list.Of[int32](3)), M: m, B: big.FromInt32(4), D: decimal.FromInt32(5), In: In{L: list.Of[int32](6)}})
	elem, _ := l.Get(i) // at
	return elem.N, elem.L.Len(), elem.F.Len(), elem.M.Len(), big.Sign(elem.B), decimal.Sign(elem.D), elem.In.L.Len() // at
}`, calls: [][]any{{int32(0)}, {int32(5)}}, want: [][]any{
		{int32(1), int32(2), int32(1), int32(1), int32(1), int32(1), int32(1)},
		{int32(0), int32(0), int32(0), int32(0), int32(0), int32(0), int32(0)},
	}},
}

// TestOptionalFix checks each fix of the table for the forms GetOr at list.List, list.Frozen and dict.Map,
// opt.OrZero, opt.OrZeroErr, a binding with = to a field and to a variable and own functions, whose returns give the
// zero value as a literal beside false or an error, and the fix of GetOr for a struct, whose fields contain nil.
func TestOptionalFix(t *testing.T) {
	for _, r := range optFixRows {
		t.Run(r.name, func(t *testing.T) {
			src := optProgram(r.src)
			at := optMarked(src, "at")
			// the form of the fix by hand: it compiles and gives the values of the program before the fix
			var before []map[string][]any
			if r.hand != "" {
				for _, args := range r.calls {
					raw := optRunRaw(t, src, args)
					before = append(before, raw)
					hand := optRunChecked(t, optProgram(r.hand), args)
					for _, e := range optEngines {
						if !gostrict.SameValues(hand[e.name], raw[e.name]) {
							t.Errorf("premise, %s: Run(%v) of the form by hand = %v, before the fix %v", e.name, args, hand[e.name], raw[e.name])
						}
					}
				}
			}

			diags := optCheck(t, src)
			found := false
			for _, d := range diags {
				if d.Code != gostrict.CodeGuard || d.File != optFile {
					t.Errorf("a finding besides GS033: %s:%d %s %s", d.File, d.Line, d.Code, d.Msg)
				}
				found = found || d.Code == gostrict.CodeGuard && slices.Contains(at, d.Line)
			}
			if !found {
				t.Fatalf("no GS033 at the lines %v:\n%s\n%s", at, diags.Text(), optNumbered(src))
			}
			edits := diags.Edits(optFile)
			if len(edits) == 0 {
				t.Fatalf("no fix:\n%s", diags.Text())
			}
			fixed, err := gostrict.ApplyEdits([]byte(src), edits)
			if err != nil {
				t.Fatal(err)
			}
			if ds := optCheck(t, string(fixed)); len(ds) > 0 {
				t.Fatalf("the fixed program has findings:\n%s\n%s", ds.Text(), optNumbered(string(fixed)))
			}
			form := regexp.MustCompile(`(?m)^\s*` + r.form + `\s*$`)
			if !form.Match(fixed) {
				t.Errorf("no line of the fixed program has the form %s:\n%s", r.form, optNumbered(string(fixed)))
			}
			for i, args := range r.calls {
				after := optRunChecked(t, string(fixed), args)
				for _, e := range optEngines {
					want := []any(nil)
					if r.hand != "" {
						want = before[i][e.name]
					} else {
						want = r.want[i]
					}
					if !gostrict.SameValues(after[e.name], want) {
						t.Errorf("%s: Run(%v) after the fix = %v, want %v\n%s", e.name, args, after[e.name], want, optNumbered(string(fixed)))
					}
				}
			}
		})
	}
}

// optNoFixRows are the counter rows of the fix: each is GS033 without a fix.
var optNoFixRows = []optRow{
	{name: "a function value", src: optFixDecls + `
func Run(i int32) int32 {
	f := find
	elem, _ := f(i) // at
	return elem // at
}`},
	{name: "a method of an interface", src: `
type Finder interface {
	Find(i int32) (int32, bool)
}

type Box struct{ l list.List[int32] }

func (b Box) Find(i int32) (int32, bool) {
	return b.l.Get(i)
}

func Run(i int32) int32 {
	var f Finder = Box{l: list.Of[int32](5)}
	elem, _ := f.Find(i) // at
	return elem // at
}`},
	{name: "a method of a type parameter", src: `
type Finder interface {
	Find(i int32) (int32, bool)
}

type Box struct{ l list.List[int32] }

func (b Box) Find(i int32) (int32, bool) {
	return b.l.Get(i)
}

func first[F Finder](f F, i int32) int32 {
	elem, _ := f.Find(i) // at
	return elem // at
}

func Run(i int32) int32 {
	return first(Box{l: list.Of[int32](5)}, i)
}`},
	{name: "an own function with return 7, false", src: `
func find(i int32) (int32, bool) {
	if i == 0 {
		return 5, true
	}
	return 7, false
}

func Run(i int32) int32 {
	elem, _ := find(i) // at
	return elem // at
}`},
	{name: "an own function with return 5, err", src: `
func load(s string) (int32, error) {
	n, err := text.ParseInt(s)
	if err != nil {
		return 5, err
	}
	return n, nil
}

func Run(s string) int32 {
	elem, _ := load(s) // at
	return elem // at
}`},
	{name: "an own function with return g()", src: `
func find(i int32) (int32, bool) {
	return lib.Find(i)
}

func Run(i int32) int32 {
	elem, _ := find(i) // at
	return elem // at
}`},
	{name: "a type with nil in opt.OrZero: list.Pop of a list of lists", src: `
func Run() int32 {
	ll := list.New[list.List[int32]]()
	elem, _ := ll.Pop() // at
	return elem.Len() // at
}`},
	{name: "a field with a pointer", src: `
type P struct{ X int32 }

type Q struct{ P *P }

func Run(i int32) int32 {
	l := list.Of(Q{P: &P{X: 1}})
	elem, _ := l.Get(i) // at
	return elem.P.X // at
}`},
	{name: "a field with msg.Out", src: `
type Q struct{ O msg.Out[int32] }

func Run(i int32) int32 {
	l := list.Of(Q{O: msg.New[int32]()})
	elem, _ := l.Get(i) // at
	return elem.O.Len() // at
}`},
	{name: "a field with a function", src: `
type Q struct{ F func() int32 }

func Run(i int32) int32 {
	l := list.Of(Q{F: func() int32 { return 1 }})
	elem, _ := l.Get(i) // at
	return elem.F() // at
}`},
	{name: "a field with an interface", src: `
type Shape interface{ Area() float64 }

type Sq struct{ S float64 }

func (q Sq) Area() float64 { return q.S * q.S }

type Q struct{ S Shape }

func Run(i int32) float64 {
	l := list.Of(Q{S: Sq{S: 2}})
	elem, _ := l.Get(i) // at
	return elem.S.Area() // at
}`},
	{name: "a field, which can not be set outside of its package", src: `
func Run(i int32) int32 {
	l := list.Of(lib.NewBag())
	elem, _ := l.Get(i) // at
	return elem.Len() // at
}`},
	{name: "three results", src: `
func split(n int32) (int32, int32, bool) {
	if n < 0 {
		return 0, 0, false
	}
	return n / 2, n % 2, true
}

func Run(n int32) int32 {
	a, b, _ := split(n) // at
	return a + b // at
}`},
	{name: "reading before the guard", src: `
func Run(i int32) int32 {
	l := list.Of[int32](5, 6)
	elem, ok := l.Get(i) // at
	s := elem + 1 // at
	if !ok {
		return 0
	}
	return s
}`},
}

// TestOptionalNoFix checks the counter rows of the fixes: a function value, a method of an interface and of a type
// parameter, own functions with return 7, false, return 5, err and return g(), a type with nil in opt.OrZero, a field
// with a pointer, msg.Out, a function and an interface, a field, which can not be set, three results and reading
// before the guard are GS033 without a fix. Invoke is checked in wokoda.
func TestOptionalNoFix(t *testing.T) {
	for _, r := range optNoFixRows {
		t.Run(r.name, func(t *testing.T) {
			src := optProgram(r.src)
			at := optMarked(src, "at")
			diags := optCheck(t, src)
			found := false
			for _, d := range diags {
				if d.Code != gostrict.CodeGuard || d.File != optFile {
					t.Errorf("a finding besides GS033: %s:%d %s %s", d.File, d.Line, d.Code, d.Msg)
				}
				found = found || d.Code == gostrict.CodeGuard && slices.Contains(at, d.Line)
			}
			if !found {
				t.Errorf("no GS033 at the lines %v:\n%s\n%s", at, diags.Text(), optNumbered(src))
			}
			if e := diags.Edits(optFile); len(e) > 0 {
				t.Errorf("a fix, where the contract gives none: %+v", e)
			}
		})
	}
}

// optMarkRow is a call of a function of std, which returns false or an error.
type optMarkRow struct {
	name string
	// typ is the type of the value, decl the declarations, call the call, which fails, zero the zero value of typ
	typ, decls, call string
	zero             any
	// required tells, that the contract names the function as marked: the fix to opt.OrZero resp. opt.OrZeroErr
	// must exist
	required bool
	// nilType tells, that the type contains nil: there must not be a fix to opt.OrZero resp. opt.OrZeroErr
	nilType bool
}

var optMarkRows = []optMarkRow{
	{name: `text.ParseInt("x")`, typ: "int32", call: `text.ParseInt("x")`, zero: int32(0), required: true},
	{name: `text.ParseInt("")`, typ: "int32", call: `text.ParseInt("")`, zero: int32(0), required: true},
	{name: `text.ParseInt("12x")`, typ: "int32", call: `text.ParseInt("12x")`, zero: int32(0), required: true},
	{name: `text.ParseInt("99999999999")`, typ: "int32", call: `text.ParseInt("99999999999")`, zero: int32(0), required: true},
	{name: `text.ParseInt("-99999999999")`, typ: "int32", call: `text.ParseInt("-99999999999")`, zero: int32(0), required: true},
	{name: "list.Pop of an empty list of int32", typ: "int32", decls: "\tl := list.New[int32]()\n", call: "l.Pop()", zero: int32(0), required: true},
	{name: "list.Pop of an empty list of string", typ: "string", decls: "\tl := list.New[string]()\n", call: "l.Pop()", zero: "", required: true},
	{name: `text.ParseFloat("x")`, typ: "float64", call: `text.ParseFloat("x")`, zero: float64(0)},
	{name: `text.ParseFloat("1e999")`, typ: "float64", call: `text.ParseFloat("1e999")`, zero: float64(0)},
	{name: `text.ParseFloat("-1e999")`, typ: "float64", call: `text.ParseFloat("-1e999")`, zero: float64(0)},
	{name: "big.Int32 of 10^12", typ: "int32", call: "big.Int32(big.Pow10(12))", zero: int32(0)},
	{name: "big.Int32 of -10^12", typ: "int32", call: "big.Int32(big.Neg(big.Pow10(12)))", zero: int32(0)},
	{name: "opt.Option.Get of None", typ: "int32", decls: "\to := opt.None[int32]()\n", call: "o.Get()", zero: int32(0)},
	{name: `big.Parse("x")`, typ: "big.Int", call: `big.Parse("x")`, nilType: true},
	{name: "big.Quo by 0", typ: "big.Int", call: "big.Quo(big.FromInt32(1), big.Zero())", nilType: true},
	{name: `decimal.Parse("x")`, typ: "decimal.Dec", call: `decimal.Parse("x")`, nilType: true},
}

// TestOptionalMarked checks the functions of std, which the fix to opt.OrZero resp. opt.OrZeroErr may name: a fix
// exists for text.ParseInt and list.Pop, which the detail names, and wherever a fix exists, the function gives the
// zero value beside false resp. an error, observed in goja and in the vm in the program before the fix. A function,
// whose type contains nil, has no fix to opt.OrZero (R-SCRIPT-GS-GETOR).
func TestOptionalMarked(t *testing.T) {
	for _, r := range optMarkRows {
		t.Run(r.name, func(t *testing.T) {
			read := "v"
			switch r.typ {
			case "big.Int":
				read = "big.String(v)"
			case "decimal.Dec":
				read = "decimal.String(v)"
			}
			ret := map[string]string{"big.Int": "string", "decimal.Dec": "string"}[r.typ]
			if ret == "" {
				ret = r.typ
			}
			src := optProgram(`
func Run() ` + ret + ` {
` + r.decls + `	v, _ := ` + r.call + ` // at
	return ` + read + ` // at
}`)
			diags := optCheck(t, src)
			var fix []gostrict.Edit
			for _, d := range diags {
				if d.Code == gostrict.CodeGuard {
					fix = append(fix, d.Edits...)
				}
			}
			fixed := ""
			if len(fix) > 0 {
				b, err := gostrict.ApplyEdits([]byte(src), fix)
				if err != nil {
					t.Fatal(err)
				}
				fixed = string(b)
			}
			orZero := strings.Contains(fixed, "opt.OrZero")
			if r.nilType {
				if orZero {
					t.Errorf("a fix to opt.OrZero for a type with nil:\n%s", fixed)
				}
				return
			}
			if !r.required && !orZero {
				t.Skipf("%s is not marked (no fix to opt.OrZero)", r.call)
			}
			raw := optRunRaw(t, src, nil)
			for _, e := range optEngines {
				if !gostrict.SameValues(raw[e.name], []any{r.zero}) {
					t.Errorf("%s: %s gives %v beside false resp. the error, want the zero value %v", e.name, r.call, raw[e.name], r.zero)
				}
			}
			if r.required && !orZero {
				t.Errorf("no fix to opt.OrZero resp. opt.OrZeroErr for %s, which the detail names as marked:\n%s", r.call, diags.Text())
			}
		})
	}
}
