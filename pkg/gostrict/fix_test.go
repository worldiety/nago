// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
)

// TestFix applies the mechanical fixes of the diagnostics and checks the result again.
func TestFix(t *testing.T) {
	src := `package s

import "go.wdy.de/nago/pkg/gostrict/std/list"

func Count(s string, l list.List[float64], n int) int {
	total := 0
	for i := range 3 {
		total += i
	}
	for i := 0; i < n; i++ {
		total += i
	}
	return total + len(s) + len(l)
}
`
	check := func(src string) gostrict.Diagnostics {
		prog, err := gostrict.LoadModules("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
		if err != nil {
			t.Fatal(err)
		}
		return gostrict.Check(prog)
	}

	for round := range 3 {
		diags := check(src)
		if len(diags) == 0 {
			break
		}
		edits := diags.Edits("llm.local/s/s.go")
		if len(edits) == 0 {
			t.Fatalf("round %d: no fixes for\n%s", round, diags.Text())
		}
		fixed, err := gostrict.ApplyEdits([]byte(src), edits)
		if err != nil {
			t.Fatal(err)
		}
		src = string(fixed)
	}

	if diags := check(src); len(diags) > 0 {
		t.Fatalf("the fixed program has diagnostics:\n%s\n%s", src, diags.Text())
	}
	for _, want := range []string{"n int32) int32", "text.RuneCount(s)", "l.Len()", "total := int32(0)", "range int32(3)", `import "go.wdy.de/nago/pkg/gostrict/std/text"`} {
		if !strings.Contains(src, want) {
			t.Fatalf("the fixed program lacks %q:\n%s", want, src)
		}
	}
}

// TestSemantics compares programs, whose translation once differed from Go, natively and in goja.
func TestSemantics(t *testing.T) {
	src := `package s

import (
	"iter"
	"math"

	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

type P struct{ X int32 }

func pair(n int32) (int32, int32) { return 0, n }

// Swap assigns a pointer and a field of the former pointer in one statement.
func Swap() (int32, int32) {
	p, q := &P{X: 1}, &P{X: 2}
	p, p.X = q, 5
	return p.X, q.X
}

func seq(n int32) iter.Seq[int32] {
	return func(yield func(int32) bool) {
		for i := range n {
			if !yield(i) {
				return
			}
		}
	}
}

// Labels leaves ranges over iterators with labeled break and continue.
func Labels() string {
	out := ""
outer:
	for i := range int32(4) {
		for j := range seq(4) {
			if j > i {
				continue outer
			}
			if i == 3 {
				break outer
			}
			out += text.FromInt(i) + text.FromInt(j) + " "
		}
	}
	out += "|"
rows:
	for i := range seq(3) {
		for j := range seq(3) {
			for k := range seq(3) {
				if k == 1 {
					continue rows
				}
				if i == 2 {
					break rows
				}
				out += text.FromInt(i) + text.FromInt(j) + text.FromInt(k) + " "
			}
		}
	}
	return out
}

// Expr is a sum type, whose variant Var has a pointer receiver.
type Expr interface{ isExpr() }

type Num struct{ V float64 }

type Add struct{ L, R Expr }

type Var struct {
	Name  string
	Value float64
}

func (Num) isExpr()  {}
func (Add) isExpr()  {}
func (*Var) isExpr() {}

func eval(e Expr) float64 {
	switch e := e.(type) {
	case Num:
		return e.V
	case Add:
		return eval(e.L) + eval(e.R)
	case *Var:
		return e.Value
	}
	return 0
}

// Pointers evaluates an expression, whose variable changes through its pointer.
func Pointers() (float64, float64) {
	x := &Var{Name: "x", Value: 2}
	e := Add{L: Num{V: 1}, R: Add{L: x, R: x}}
	before := eval(e)
	x.Value = 10
	return before, eval(e)
}

// Result is a generic sum type.
type Result[T any] interface{ isResult() }

type Ok[T any] struct{ V T }

type Fail[T any] struct{ Msg string }

func (Ok[T]) isResult()   {}
func (Fail[T]) isResult() {}

func describe[T any](r Result[T], show func(T) string) string {
	switch r := r.(type) {
	case Ok[T]:
		return "ok " + show(r.V)
	case Fail[T]:
		return "fail " + r.Msg
	}
	return ""
}

func parse(s string) Result[int32] {
	n, err := text.ParseInt(s)
	if err != nil {
		return Fail[int32]{Msg: s}
	}
	return Ok[int32]{V: n}
}

// Generic switches over a generic sum type.
func Generic() string {
	return describe(parse("42"), text.FromInt) + ", " + describe(parse("x"), text.FromInt) + ", " + describe[string](Ok[string]{V: "s"}, func(s string) string { return s })
}

// Shape is a sum type, whose variants are all comparable.
type Shape interface{ isShape() }

type Circle struct{ R float64 }

type Group struct {
	Name  string
	First Shape
}

func (Circle) isShape() {}
func (Group) isShape()  {}

// Compare compares values of a sum type like Go.
func Compare() string {
	a := Shape(Circle{R: 1})
	b := Shape(Group{Name: "g", First: Circle{R: 1}})
	c := Shape(Group{Name: "g", First: Circle{R: 1}})
	out := ""
	for _, ok := range list.Of(a == Circle{R: 1}, a != b, b == c, Group{Name: "g", First: a} == b, b == Group{Name: "g", First: Circle{R: 2}}).All() {
		if ok {
			out += "t"
		} else {
			out += "f"
		}
	}
	return out
}

type Outer struct{ In P }

func readAfter(in P, f func()) int32 {
	f()
	return in.X
}

func readPtr(in P, o *Outer) int32 {
	o.In.X = 9
	return in.X
}

func (p P) twice() int32 { return p.X * 2 }

// Aliases passes structs to functions, which only read them, while something else changes the original.
func Aliases() (int32, int32, int32, int32) {
	v := P{X: 1}
	a := readAfter(v, func() { v.X = 2 })
	o := &Outer{In: P{X: 1}}
	b := readPtr(o.In, o)
	w := P{X: 3}
	c := w.twice()
	d := readPtr(w, &Outer{In: w})
	return a, b, c, d
}

// Infinite uses the functions of math with an int parameter, whose constant argument is allowed.
func Infinite(x float64) (bool, bool, float64) {
	return math.IsInf(x, 0), math.IsInf(x, -1), math.Inf(1)
}

// Closures captures the variables of a loop, whose init calls a function with two results.
func Closures() string {
	fs := list.New[func() int32]()
	for i, n := pair(3); i < n; i++ {
		fs.Append(func() int32 { return i })
	}
	out := ""
	for _, f := range fs.All() {
		out += text.FromInt(f())
	}
	return out
}
`
	cache := gostrict.NewCache(1)
	b, diags, err := cache.Compile("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags)
	}
	in := mustLoad(t, goEngine, b)
	res, err := in.Call(context.Background(), "Swap")
	if err != nil {
		t.Fatal(err)
	}
	// Go: p becomes q, then the field of the former p is set, so q keeps 2
	if !gostrict.SameValues(res, []any{int32(2), int32(2)}) {
		t.Fatalf("Swap: %v", res)
	}
	res, err = in.Call(context.Background(), "Closures")
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "012" {
		t.Fatalf("Closures: %v", res)
	}
	res, err = in.Call(context.Background(), "Pointers")
	if err != nil {
		t.Fatal(err)
	}
	if !gostrict.SameValues(res, []any{5.0, 21.0}) {
		t.Fatalf("Pointers: %v", res)
	}
	res, err = in.Call(context.Background(), "Generic")
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "ok 42, fail x, ok s" {
		t.Fatalf("Generic: %v", res)
	}
	res, err = in.Call(context.Background(), "Compare")
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "ttttf" {
		t.Fatalf("Compare: %v", res)
	}
	res, err = in.Call(context.Background(), "Aliases")
	if err != nil {
		t.Fatal(err)
	}
	if !gostrict.SameValues(res, []any{int32(1), int32(1), int32(6), int32(3)}) {
		t.Fatalf("Aliases: %v", res)
	}
	res, err = in.Call(context.Background(), "Infinite", math.Inf(-1))
	if err != nil {
		t.Fatal(err)
	}
	if !gostrict.SameValues(res, []any{true, true, math.Inf(1)}) {
		t.Fatalf("Infinite: %v", res)
	}
	res, err = in.Call(context.Background(), "Labels")
	if err != nil {
		t.Fatal(err)
	}
	// computed by Go
	if want := "00 10 11 20 21 22 |000 100 "; res[0] != want {
		t.Fatalf("Labels: %q, want %q", res[0], want)
	}
}

// TestTakeOver compares the states of two versions of a program.
func TestTakeOver(t *testing.T) {
	compile := func(src string) *gostrict.Bundle {
		b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
		if err != nil || len(diags) > 0 {
			t.Fatal(err, diags)
		}
		return b
	}
	const state = "llm.local/s.Game"
	v1 := compile("package s\n\ntype Game struct{ X float64 }\n\nfunc New() *Game { return &Game{} }\n")
	v2 := compile("package s\n\ntype Game struct{ X float64 }\n\nfunc New() *Game { return &Game{X: 1} }\n\nfunc Step(g *Game) { g.X++ }\n")
	v3 := compile("package s\n\ntype Game struct{ X, Y float64 }\n\nfunc New() *Game { return &Game{} }\n")
	v4 := compile("package s\n\ntype Game struct{ F func() }\n\nfunc New() *Game { return &Game{F: func() {}} }\n")
	if v1.Hash == v2.Hash || !v2.CanTakeOver(v1, state) {
		t.Fatal("the same state must be taken over by a new version")
	}
	if v3.CanTakeOver(v1, state) {
		t.Fatal("a changed state must not be taken over")
	}
	if v4.States[state] != "" || v4.CanTakeOver(v4, state) {
		t.Fatal("a state with a function must not be taken over")
	}
}

// TestSourceMap checks, that the marks are removed and that the source map points to the statements.
func TestSourceMap(t *testing.T) {
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte("package s\n\nfunc F(a int32) int32 {\n\tb := a + 1\n\treturn b * 2\n}\n")}}})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags)
	}
	if strings.ContainsRune(b.JS, 0) {
		t.Fatal("the bundle contains marks")
	}
	var sm struct {
		Version        int      `json:"version"`
		Sources        []string `json:"sources"`
		SourcesContent []string `json:"sourcesContent"`
		Mappings       string   `json:"mappings"`
	}
	if err := json.Unmarshal(b.SourceMap, &sm); err != nil {
		t.Fatal(err)
	}
	if sm.Version != 3 || len(sm.Sources) != 1 || !strings.HasSuffix(sm.Sources[0], "s.go") || !strings.Contains(sm.SourcesContent[0], "return b * 2") {
		t.Fatalf("unexpected source map %s", b.SourceMap)
	}

	// decode the mappings: the line of "return" must map to line 5 of s.go
	lines := strings.Split(b.JS, "\n")
	srcLine := 0
	found := false
	for i, seg := range strings.Split(sm.Mappings, ";") {
		if seg == "" {
			continue
		}
		v := decodeVLQ(seg)
		srcLine += v[2]
		if strings.Contains(lines[i], "return") && strings.Contains(lines[i], "imul(b, 2)") {
			found = srcLine+1 == 5
		}
	}
	if !found {
		t.Fatalf("the return statement does not map to line 5:\n%s\n%s", b.JS, sm.Mappings)
	}
}

func decodeVLQ(s string) []int {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var res []int
	v, shift := 0, 0
	for _, c := range s {
		d := strings.IndexRune(chars, c)
		v |= (d & 31) << shift
		if d&32 != 0 {
			shift += 5
			continue
		}
		n := v >> 1
		if v&1 == 1 {
			n = -n
		}
		res = append(res, n)
		v, shift = 0, 0
	}
	return res
}
