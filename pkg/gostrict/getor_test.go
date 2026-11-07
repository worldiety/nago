// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
)

// The tests of this file check the substitutes of an optional result after the contract R-SCRIPT-GS-GETOR of wokoda
// (L-N9-1): GetOr of list.List, list.Frozen and dict.Map, and the functions of std/opt for a result (T, bool) and
// (T, error), here opt.OrZero and opt.OrZeroErr, which take the call itself as their argument, e.g.
// opt.OrZero(l.Get(1)). Every row of its table runs in goja and in the vm, each with the expectation of the contract,
// not of an engine. The call is written in each form, which the engines translate differently: an inlined method of
// std (Get of a list, a frozen list and a map), a function of std (text.ParseInt), an own function and a method of an
// interface. A function with (T, bool) or (T, error), which returns a value besides false or an error, shows, that
// the result is the zero value and not the value. The rejection rows check every type, whose zero value may contain
// nil (as the checker defines it, check_total.go): the checker reports at the call, and its finding names GetOr or
// the check of ok.

// getOrBasicSrc holds the rows of GetOr with numbers.
const getOrBasicSrc = `package s

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

func ListGetOr(i int32) int32 {
	l := list.Of[int32](5, 6)
	return l.GetOr(i, 9)
}

func FrozenGetOr(i int32) int32 {
	f := list.Freeze(list.Of[int32](5, 6))
	return f.GetOr(i, 9)
}

func DictGetOr(withA bool) int32 {
	d := dict.New[string, int32]()
	d.Set("b", 1)
	if withA {
		d.Set("a", 4)
	}
	return d.GetOr("a", 9)
}
`

// getOrStructSrc changes the result of GetOr and reads the element and standard back afterwards. Each function returns
// the field of the result before the change, then the elements and standard after it.
const getOrStructSrc = `package s

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type P struct{ X int32 }

type Q struct{ In P }

func at(l list.List[P], i int32) int32 {
	if e, ok := l.Get(i); ok {
		return e.X
	}
	return -1
}

func ListStruct(i int32) (int32, int32, int32, int32) {
	l := list.Of(P{X: 1}, P{X: 2})
	standard := P{X: 9}
	v := l.GetOr(i, standard)
	got := v.X
	v.X = 100
	return got, at(l, 0), at(l, 1), standard.X
}

func FrozenStruct(i int32) (int32, int32, int32, int32) {
	f := list.Freeze(list.Of(P{X: 1}, P{X: 2}))
	standard := P{X: 9}
	v := f.GetOr(i, standard)
	got := v.X
	v.X = 100
	a, b := int32(-1), int32(-1)
	if e, ok := f.Get(0); ok {
		a = e.X
	}
	if e, ok := f.Get(1); ok {
		b = e.X
	}
	return got, a, b, standard.X
}

func DictStruct(k string) (int32, int32, int32) {
	d := dict.New[string, P]()
	d.Set("a", P{X: 1})
	standard := P{X: 9}
	v := d.GetOr(k, standard)
	got := v.X
	v.X = 100
	a := int32(-1)
	if e, ok := d.Get("a"); ok {
		a = e.X
	}
	return got, a, standard.X
}

func NestedStruct(i int32) (int32, int32, int32) {
	l := list.Of(Q{In: P{X: 1}})
	standard := Q{In: P{X: 9}}
	v := l.GetOr(i, standard)
	got := v.In.X
	v.In.X = 100
	a := int32(-1)
	if e, ok := l.Get(0); ok {
		a = e.In.X
	}
	return got, a, standard.In.X
}
`

// orZeroSrc holds the rows of opt.OrZero and opt.OrZeroErr.
const orZeroSrc = `package s

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/opt"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

type Out struct{ N float64 }

type ID float64

func ListOrZero(i int32) int32 {
	l := list.Of[int32](5, 6)
	return opt.OrZero(l.Get(i))
}

func FrozenOrZero(i int32) int32 {
	f := list.Freeze(list.Of[int32](5, 6))
	return opt.OrZero(f.Get(i))
}

func DictOrZero(k string) string {
	d := dict.New[string, string]()
	d.Set("y", "z")
	return opt.OrZero(d.Get(k))
}

func OutOrZero(i int32) (float64, bool) {
	l := list.Of(Out{N: 3})
	o := opt.OrZero(l.Get(i))
	return o.N, o == Out{}
}

func IDOrZero(i int32) float64 {
	l := list.Of[ID](1.5)
	return float64(opt.OrZero(l.Get(i)))
}

func ParseOrZero(s string) int32 {
	return opt.OrZeroErr(text.ParseInt(s))
}

// pick returns 7 also besides false.
func pick(ok bool) (int32, bool) {
	return 7, ok
}

func OwnOrZero(ok bool) int32 {
	return opt.OrZero(pick(ok))
}

// fail returns 7 also besides the error.
func fail(bad bool) (int32, error) {
	if bad {
		return 7, errors.New("bad")
	}
	return 7, nil
}

func OwnOrZeroErr(bad bool) int32 {
	return opt.OrZeroErr(fail(bad))
}

func parseID(s string) (ID, error) {
	f, err := text.ParseFloat(s)
	if err != nil {
		return 0, err
	}
	return ID(f), nil
}

func IDOrZeroErr(s string) float64 {
	return float64(opt.OrZeroErr(parseID(s)))
}

func out(bad bool) (Out, error) {
	if bad {
		return Out{N: 7}, errors.New("bad")
	}
	return Out{N: 3}, nil
}

func OutOrZeroErr(bad bool) (float64, bool) {
	o := opt.OrZeroErr(out(bad))
	return o.N, o == Out{}
}

type Src interface {
	Get(i int32) (int32, bool)
}

type arr struct{ l list.List[int32] }

func (a arr) Get(i int32) (int32, bool) {
	return a.l.Get(i)
}

func IfaceOrZero(i int32) int32 {
	var s Src = arr{l: list.Of[int32](5, 6)}
	return opt.OrZero(s.Get(i))
}
`

// getOrRow is a call of an entry and its result after the contract.
type getOrRow struct {
	name string // the row of the table of the contract, which the call belongs to
	fn   string
	args []any
	want []any
}

var getOrBasicRows = []getOrRow{
	{"l.GetOr(1, 9) with l = [5, 6]", "ListGetOr", []any{int32(1)}, []any{int32(6)}},
	{"l.GetOr(2, 9) with l = [5, 6]", "ListGetOr", []any{int32(2)}, []any{int32(9)}},
	{"l.GetOr(-1, 9) with l = [5, 6]", "ListGetOr", []any{int32(-1)}, []any{int32(9)}},
	{"l.GetOr(0, 9) with l = [5, 6]", "ListGetOr", []any{int32(0)}, []any{int32(5)}},
	{"list.Frozen: GetOr(1, 9)", "FrozenGetOr", []any{int32(1)}, []any{int32(6)}},
	{"list.Frozen: GetOr(2, 9)", "FrozenGetOr", []any{int32(2)}, []any{int32(9)}},
	{"list.Frozen: GetOr(-1, 9)", "FrozenGetOr", []any{int32(-1)}, []any{int32(9)}},
	{"list.Frozen: GetOr(0, 9)", "FrozenGetOr", []any{int32(0)}, []any{int32(5)}},
	{`d.GetOr("a", 9) with a = 4`, "DictGetOr", []any{true}, []any{int32(4)}},
	{`d.GetOr("a", 9) without a`, "DictGetOr", []any{false}, []any{int32(9)}},
}

var getOrStructRows = []getOrRow{
	// the field of the result, the elements 0 and 1 and standard after the change of the result
	{"list of structs: GetOr(1, standard), result changed", "ListStruct", []any{int32(1)}, []any{int32(2), int32(1), int32(2), int32(9)}},
	{"list of structs: GetOr(0, standard), result changed", "ListStruct", []any{int32(0)}, []any{int32(1), int32(1), int32(2), int32(9)}},
	{"list of structs: GetOr(5, standard), result changed", "ListStruct", []any{int32(5)}, []any{int32(9), int32(1), int32(2), int32(9)}},
	{"frozen list of structs: GetOr(1, standard), result changed", "FrozenStruct", []any{int32(1)}, []any{int32(2), int32(1), int32(2), int32(9)}},
	{"frozen list of structs: GetOr(5, standard), result changed", "FrozenStruct", []any{int32(5)}, []any{int32(9), int32(1), int32(2), int32(9)}},
	{`map of structs: GetOr("a", standard), result changed`, "DictStruct", []any{"a"}, []any{int32(1), int32(1), int32(9)}},
	{`map of structs: GetOr("z", standard), result changed`, "DictStruct", []any{"z"}, []any{int32(9), int32(1), int32(9)}},
	{"list of nested structs: GetOr(0, standard), inner field changed", "NestedStruct", []any{int32(0)}, []any{int32(1), int32(1), int32(9)}},
	{"list of nested structs: GetOr(5, standard), inner field changed", "NestedStruct", []any{int32(5)}, []any{int32(9), int32(1), int32(9)}},
}

var orZeroRows = []getOrRow{
	{"opt.OrZero(l.Get(1)) with l = [5, 6]", "ListOrZero", []any{int32(1)}, []any{int32(6)}},
	{"opt.OrZero(l.Get(5)) with l = [5, 6]", "ListOrZero", []any{int32(5)}, []any{int32(0)}},
	{"opt.OrZero(f.Get(1)) with a list.Frozen [5, 6]", "FrozenOrZero", []any{int32(1)}, []any{int32(6)}},
	{"opt.OrZero(f.Get(5)) with a list.Frozen [5, 6]", "FrozenOrZero", []any{int32(5)}, []any{int32(0)}},
	{`opt.OrZero(d.Get("x")) with string values, x missing`, "DictOrZero", []any{"x"}, []any{""}},
	{`opt.OrZero(d.Get("y")) with string values, y = "z"`, "DictOrZero", []any{"y"}, []any{"z"}},
	{"opt.OrZero(l.Get(5)) with elements Out{N float64}", "OutOrZero", []any{int32(5)}, []any{float64(0), true}},
	{"opt.OrZero(l.Get(0)) with elements Out{N float64}, Out{N: 3}", "OutOrZero", []any{int32(0)}, []any{float64(3), false}},
	{"opt.OrZero(l.Get(5)) with elements type ID float64", "IDOrZero", []any{int32(5)}, []any{float64(0)}},
	{"opt.OrZero(l.Get(0)) with elements type ID float64, ID(1.5)", "IDOrZero", []any{int32(0)}, []any{1.5}},
	{`opt.OrZeroErr(text.ParseInt("12"))`, "ParseOrZero", []any{"12"}, []any{int32(12)}},
	{`opt.OrZeroErr(text.ParseInt("x"))`, "ParseOrZero", []any{"x"}, []any{int32(0)}},
	{"opt.OrZero of an own function (7, true)", "OwnOrZero", []any{true}, []any{int32(7)}},
	{"opt.OrZero of an own function (7, false): the zero value, not 7", "OwnOrZero", []any{false}, []any{int32(0)}},
	{"opt.OrZeroErr of an own function (7, nil)", "OwnOrZeroErr", []any{false}, []any{int32(7)}},
	{"opt.OrZeroErr of an own function (7, error): the zero value, not 7", "OwnOrZeroErr", []any{true}, []any{int32(0)}},
	{`opt.OrZeroErr with type ID float64, "2.5"`, "IDOrZeroErr", []any{"2.5"}, []any{2.5}},
	{`opt.OrZeroErr with type ID float64, "x"`, "IDOrZeroErr", []any{"x"}, []any{float64(0)}},
	{"opt.OrZeroErr with Out{N float64}, (Out{N: 3}, nil)", "OutOrZeroErr", []any{false}, []any{float64(3), false}},
	{"opt.OrZeroErr with Out{N float64}, (Out{N: 7}, error): Out{N: 0}", "OutOrZeroErr", []any{true}, []any{float64(0), true}},
	{"opt.OrZero of a method of an interface, s.Get(1)", "IfaceOrZero", []any{int32(1)}, []any{int32(6)}},
	{"opt.OrZero of a method of an interface, s.Get(5)", "IfaceOrZero", []any{int32(5)}, []any{int32(0)}},
}

// getOrRun runs the rows in goja and in the vm. A program, which does not compile, fails all its rows.
func getOrRun(t *testing.T, src string, rows []getOrRow) {
	t.Helper()
	engines := loadTextNum(t, src)
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			for _, e := range engines {
				got := e.call(t, r.fn, r.args...)
				if !gostrict.SameValues(got, r.want) {
					t.Errorf("%s: %s(%v) = %v, want %v", e.name, r.fn, shortArgs(r.args), got, r.want)
				}
			}
		})
	}
}

// TestGetOr checks GetOr of list.List, list.Frozen and dict.Map with numbers (R-SCRIPT-GS-GETOR, rows 1 to 3).
func TestGetOr(t *testing.T) {
	getOrRun(t, getOrBasicSrc, getOrBasicRows)
}

// TestGetOrCopies changes the result of GetOr with structs: neither the element nor standard change
// (R-SCRIPT-GS-GETOR, row 4, also at a frozen list, a map and a nested struct).
func TestGetOrCopies(t *testing.T) {
	getOrRun(t, getOrStructSrc, getOrStructRows)
}

// TestOrZero checks opt.OrZero and opt.OrZeroErr with the call as their argument in each form of a call, also with
// the types type ID float64 and Out{N float64}, which the checker must accept (R-SCRIPT-GS-GETOR, rows 5 to 8 and the
// counter rows of row 9).
func TestOrZero(t *testing.T) {
	getOrRun(t, orZeroSrc, orZeroRows)
}

// getOrRejectRow is a program, which calls opt.OrZero or opt.OrZeroErr with a type, whose zero value may contain nil.
type getOrRejectRow struct {
	name string
	// decls are the declarations of the program besides the imports; the line with the call contains "opt.OrZero"
	decls string
}

// getOrRejectRows are the rows of row 9 of the table of the contract (a list, a struct with a list field, decimal.Dec,
// an interface), each with opt.OrZero and opt.OrZeroErr, and further types of the definition of the detail: a map, a
// pointer, a function, a type parameter without a core type and a struct, which holds such a struct.
var getOrRejectRows = []getOrRejectRow{
	{"OrZero with a list", `
func Run() int32 {
	d := dict.New[string, list.List[int32]]()
	v := opt.OrZero(d.Get("k"))
	return v.Len()
}`},
	{"OrZeroErr with a list", `
func load() (list.List[int32], error) {
	return list.New[int32](), nil
}

func Run() int32 {
	v := opt.OrZeroErr(load())
	return v.Len()
}`},
	{"OrZero with a struct with a list field", `
type Row struct{ L list.List[int32] }

func Run() int32 {
	l := list.Of(Row{L: list.New[int32]()})
	v := opt.OrZero(l.Get(0))
	return v.L.Len()
}`},
	{"OrZeroErr with a struct with a list field", `
type Row struct{ L list.List[int32] }

func row() (Row, error) {
	return Row{L: list.New[int32]()}, nil
}

func Run() int32 {
	v := opt.OrZeroErr(row())
	return v.L.Len()
}`},
	{"OrZero with decimal.Dec", `
func Run() string {
	d := opt.OrZero(decimal.Parse("1.5"))
	return decimal.String(d)
}`},
	{"OrZeroErr with decimal.Dec", `
func dec(s string) (decimal.Dec, error) {
	if d, ok := decimal.Parse(s); ok {
		return d, nil
	}
	return decimal.Zero(), errors.New("no number")
}

func Run() string {
	d := opt.OrZeroErr(dec("1.5"))
	return decimal.String(d)
}`},
	{"OrZero with an interface", `
type Shape interface{ Area() float64 }

type Sq struct{ S float64 }

func (q Sq) Area() float64 { return q.S * q.S }

func Run() float64 {
	l := list.Of[Shape](Sq{S: 2})
	s := opt.OrZero(l.Get(0))
	return s.Area()
}`},
	{"OrZeroErr with an interface", `
type Shape interface{ Area() float64 }

type Sq struct{ S float64 }

func (q Sq) Area() float64 { return q.S * q.S }

func shape() (Shape, error) {
	return Sq{S: 2}, nil
}

func Run() float64 {
	s := opt.OrZeroErr(shape())
	return s.Area()
}`},
	{"OrZero with a map", `
func m() (dict.Map[string, int32], bool) {
	return dict.New[string, int32](), true
}

func Run() int32 {
	d := opt.OrZero(m())
	return d.Len()
}`},
	{"OrZero with a pointer", `
type P struct{ X int32 }

func Run() int32 {
	l := list.Of(&P{X: 1})
	p := opt.OrZero(l.Get(0))
	return p.X
}`},
	{"OrZeroErr with a function", `
func pickF() (func() int32, error) {
	return func() int32 { return 1 }, nil
}

func Run() int32 {
	f := opt.OrZeroErr(pickF())
	return f()
}`},
	{"OrZero with a type parameter without a core type", `
func first[T any](l list.List[T]) T {
	return opt.OrZero(l.Get(0))
}

func Run() int32 {
	return first(list.Of[int32](1))
}`},
	{"OrZero with a struct, which holds a struct with a list field", `
type Row struct{ L list.List[int32] }

type Wrap struct {
	N float64
	R Row
}

func Run() float64 {
	l := list.Of(Wrap{N: 1, R: Row{L: list.New[int32]()}})
	w := opt.OrZero(l.Get(0))
	return w.N
}`},
}

// getOrRejectSrc is the program of the declarations with the imports, which they use.
func getOrRejectSrc(decls string) string {
	var imp strings.Builder
	for _, p := range []string{"errors", "go.wdy.de/nago/pkg/gostrict/std/decimal", "go.wdy.de/nago/pkg/gostrict/std/dict",
		"go.wdy.de/nago/pkg/gostrict/std/list", "go.wdy.de/nago/pkg/gostrict/std/opt"} {
		name := p[strings.LastIndex(p, "/")+1:]
		if strings.Contains(decls, name+".") {
			imp.WriteString("\t\"" + p + "\"\n")
		}
	}
	return "package s\n\nimport (\n" + imp.String() + ")\n" + decls + "\n"
}

// getOrNames tells, whether the text of a finding names GetOr or the check of ok (or of the error).
var getOrNames = regexp.MustCompile(`GetOr|if \w+, \w+ :=`)

// TestOrZeroRejected checks, that the checker rejects opt.OrZero and opt.OrZeroErr for each type, whose zero value may
// contain nil, at the call and with a finding, which names GetOr or the check of ok (R-SCRIPT-GS-GETOR, row 9). The
// counter rows, type ID float64 and Out{N float64}, compile and run in TestOrZero; here the frame of the rows, both
// calls with Out instead of a rejected type, compiles. Once observed, not checked here (20916441c, before N9): each
// program of a row compiled without any finding with a local generic function in place of opt.OrZero.
func TestOrZeroRejected(t *testing.T) {
	check := func(t *testing.T, src string) gostrict.Diagnostics {
		t.Helper()
		prog, err := gostrict.LoadModules("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
		if err != nil {
			t.Fatal(err)
		}
		return gostrict.Check(prog)
	}

	// the frame of the rows compiles with a type without nil
	frame := getOrRejectSrc(`
type Out struct{ N float64 }

func Run() float64 {
	l := list.Of(Out{N: 1})
	o := opt.OrZero(l.Get(0))
	return o.N
}

func RunErr() float64 {
	o := opt.OrZeroErr(out())
	return o.N
}

func out() (Out, error) {
	return Out{N: 2}, nil
}`)
	t.Run("counter row: the frame with Out{N float64}", func(t *testing.T) {
		if diags := check(t, frame); len(diags) > 0 {
			t.Fatalf("the frame with Out does not compile:\n%s", diags.Text())
		}
	})

	for _, r := range getOrRejectRows {
		t.Run(r.name, func(t *testing.T) {
			src := getOrRejectSrc(r.decls)
			line := 0
			for i, l := range strings.Split(src, "\n") {
				if strings.Contains(l, "opt.OrZero") {
					if line != 0 {
						t.Fatalf("the row calls opt twice")
					}
					line = i + 1
				}
			}
			diags := check(t, src)
			var at []string
			found := false
			for _, d := range diags {
				if d.Line != line || d.File != "llm.local/s/s.go" {
					continue
				}
				at = append(at, fmt.Sprintf("%s: %s / %s / %s", d.Code, d.Msg, d.Why, d.Fix))
				if d.Code == gostrict.CodeTypeError || d.Code == gostrict.CodeInternal {
					continue
				}
				if getOrNames.MatchString(d.Msg + "\n" + d.Why + "\n" + d.Fix) {
					found = true
				}
			}
			if !found {
				t.Errorf("no finding at line %d, which names GetOr or the check of ok; at the line: %q\nall:\n%s\n%s", line, at, diags.Text(), src)
			}
		})
	}
}
