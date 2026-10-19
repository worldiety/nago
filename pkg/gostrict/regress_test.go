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
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
)

// The tests of this file reproduce defects, which a downstream application has found.

// TestCheckAfterLoadErrors checks programs, whose type information is incomplete: Check must report, not panic.
func TestCheckAfterLoadErrors(t *testing.T) {
	t.Run("forbidden function", func(t *testing.T) {
		prog, err := gostrict.LoadModules("llm.local/s", script(`package s

import "strings"

func Main() (string, error) {
	return strings.Repeat("a", 3), nil
}
`))
		if err != nil {
			t.Fatal(err)
		}
		prog.Terminating = true
		diags := gostrict.Check(prog)
		noInternal(t, diags)
		if !strings.Contains(diags.Text(), string(gostrict.CodeStdFunc)) && !strings.Contains(diags.Text(), string(gostrict.CodeImport)) {
			t.Fatalf("the forbidden function was not reported:\n%s", diags.Text())
		}
	})

	t.Run("import cycle", func(t *testing.T) {
		fsys := fstest.MapFS{
			"a/basis/basis.go": {Data: []byte("package basis\n\nimport \"llm.local/app/a/oben\"\n\nfunc F() int32 { return oben.G() }\n")},
			"a/oben/oben.go":   {Data: []byte("package oben\n\nimport \"llm.local/app/b/mitte\"\n\nfunc G() int32 { return mitte.H() }\n")},
			"b/mitte/mitte.go": {Data: []byte("package mitte\n\nimport \"llm.local/app/a/basis\"\n\nfunc H() int32 {\n\tf := basis.F\n\treturn f()\n}\n")},
		}
		prog, err := gostrict.LoadRoots([]string{"llm.local/app/a/basis"}, gostrict.Module{Path: "llm.local/app", FS: fsys})
		if err != nil {
			t.Fatal(err)
		}
		prog.Terminating = true
		diags := gostrict.Check(prog)
		noInternal(t, diags)
		if len(diags) == 0 {
			t.Fatal("the import cycle was not reported")
		}
	})
}

func noInternal(t *testing.T, diags gostrict.Diagnostics) {
	t.Helper()
	for _, d := range diags {
		if d.Code == gostrict.CodeInternal {
			t.Fatalf("the checker failed:\n%s", diags.Text())
		}
	}
}

// TestPairToField assigns a result (T, bool) with a nilable T to a field.
func TestPairToField(t *testing.T) {
	prog, err := gostrict.LoadModules("llm.local/s", script(`package s

import (
	"go.wdy.de/nago/pkg/gostrict/std/decimal"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Row struct {
	Max decimal.Dec
}

func Max(l list.List[decimal.Dec]) Row {
	r := Row{}
	ok := false
	r.Max, ok = l.Get(0)
	_ = ok
	return r
}
`))
	if err != nil {
		t.Fatal(err)
	}
	diags := gostrict.Check(prog)
	noInternal(t, diags)
	found := false
	for _, d := range diags {
		found = found || d.Code == gostrict.CodeGuard && d.Line == 15
	}
	if !found {
		t.Fatalf("expected a guard at the assignment:\n%s", diags.Text())
	}
}

// TestRangeOverCallResult ranges over a list, which a call returns, in the terminating profile.
func TestRangeOverCallResult(t *testing.T) {
	prog, err := gostrict.LoadModules("llm.local/s", script(`package s

import "go.wdy.de/nago/pkg/gostrict/std/list"

func items() list.List[int32] {
	return list.Of[int32](1, 2, 3)
}

func Sum() int32 {
	s := int32(0)
	for v := range items().Values() {
		s += v
	}
	return s
}

func Grow(l list.List[int32]) int32 {
	alias := l
	for v := range alias.Values() {
		l.Append(v)
	}
	return l.Len()
}
`))
	if err != nil {
		t.Fatal(err)
	}
	prog.Terminating = true
	if diags := gostrict.Check(prog); len(diags) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diags.Text())
	}
	b, diags := gostrict.Emit(prog)
	if len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	in, err := goEngine.Load(context.Background(), b, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := in.Call(context.Background(), "Sum")
	if err != nil || res[0] != int32(6) {
		t.Fatalf("Sum returned %v, %v", res, err)
	}
	// the range sees the elements, which the list had, when the loop started, like a range over a slice
	res, err = in.Call(context.Background(), "Grow", []any{int32(1), int32(2)})
	if err != nil || res[0] != int32(4) {
		t.Fatalf("Grow returned %v, %v", res, err)
	}
}

const variadicHeader = `//gostrict:header
package h

//gostrict:js (s, rest) => s + ":" + rest.length + ":" + rest.join(",")
func Join(s string, rest ...string) string { return s }

type Rec struct {
	A int32
}

//gostrict:js (r, rest) => r.A + rest.length
func Count(r Rec, rest ...int32) int32 { return r.A }
`

// TestVariadicWithoutArguments calls a variadic function of a header package without variadic arguments.
func TestVariadicWithoutArguments(t *testing.T) {
	fsys := fstest.MapFS{
		"h/h.go": {Data: []byte(variadicHeader)},
		"main.go": {Data: []byte(`package app

import "llm.local/app/h"

func None() string { return h.Join("SELECT") }

func Some() string { return h.Join("SELECT", "a", "b") }

func Record() int32 { return h.Count(h.Rec{A: 5}) + h.Count(h.Rec{A: 1}, 7) }

func three() (string, string, string) { return "x", "y", "z" }

func Tuple() string { return h.Join(three()) }
`)},
	}
	prog, err := gostrict.LoadModules("llm.local/app", gostrict.Module{Path: "llm.local/app", FS: fsys, Trusted: true})
	if err != nil {
		t.Fatal(err)
	}
	if diags := gostrict.Check(prog); len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	b, diags := gostrict.Emit(prog)
	if len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	in, err := goEngine.Load(context.Background(), b, nil)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]any{"None": "SELECT:0:", "Some": "SELECT:2:a,b", "Record": int32(7), "Tuple": "x:2:y,z"} {
		res, err := in.Call(context.Background(), name)
		if err != nil || res[0] != want {
			t.Fatalf("%s returned %v, %v, want %v", name, res, err, want)
		}
	}
}

// TestRecursionAlongImports links the dynamic calls of a package only to the packages, which are connected to it by
// imports: a and b form a cycle through their interfaces only, when a third package passes values between them.
func TestRecursionAlongImports(t *testing.T) {
	fsys := fstest.MapFS{
		"a/a.go": {Data: []byte(`package a

type N interface{ N() int32 }

type T struct{}

func (t T) M() int32 { return call(t) }

func (t T) N() int32 { return 1 }

func call(n N) int32 { return n.N() }

func Main() int32 { return T{}.M() }
`)},
		"b/b.go": {Data: []byte(`package b

type M interface{ M() int32 }

type U struct{}

func (u U) N() int32 { return call(u) }

func (u U) M() int32 { return 2 }

func call(m M) int32 { return m.M() }

func Main() int32 { return U{}.N() }
`)},
		"c/c.go": {Data: []byte(`package c

import (
	"llm.local/app/a"
	"llm.local/app/b"
)

func Main() int32 { return a.T{}.M() + b.U{}.N() }
`)},
	}
	check := func(roots ...string) gostrict.Diagnostics {
		var paths []string
		for _, r := range roots {
			paths = append(paths, "llm.local/app/"+r)
		}
		prog, err := gostrict.LoadRoots(paths, gostrict.Module{Path: "llm.local/app", FS: fsys})
		if err != nil {
			t.Fatal(err)
		}
		prog.Terminating = true
		return gostrict.Check(prog)
	}
	for _, roots := range [][]string{{"a"}, {"b"}, {"a", "b"}} {
		if diags := check(roots...); len(diags) > 0 {
			t.Fatalf("%v: unexpected diagnostics:\n%s", roots, diags.Text())
		}
	}
	// c connects a and b, so their methods may call each other
	diags := check("a", "b", "c")
	if !strings.Contains(diags.Text(), "may call itself") {
		t.Fatalf("the recursion through c was not reported:\n%s", diags.Text())
	}
}

// TestRecursionThroughSort reports a recursion through the comparison of list.SortFunc.
func TestRecursionThroughSort(t *testing.T) {
	prog, err := gostrict.LoadModules("llm.local/s", script(`package s

import "go.wdy.de/nago/pkg/gostrict/std/list"

func cmp(a, b int32) int32 {
	Sort(list.Of[int32](b, a))
	return a - b
}

func Sort(l list.List[int32]) {
	list.SortFunc(l, cmp)
}
`))
	if err != nil {
		t.Fatal(err)
	}
	prog.Terminating = true
	if diags := gostrict.Check(prog); !strings.Contains(diags.Text(), "may call itself") {
		t.Fatalf("the recursion was not reported:\n%s", diags.Text())
	}
}

func compileScript(t *testing.T, src string) *gostrict.Bundle {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", script(src))
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags.Text())
	}
	return b
}

// TestCancelAfterCall cancels the context of each call, after it has returned, which must not reach the next call.
func TestCancelAfterCall(t *testing.T) {
	b := compileScript(t, "package s\n\nfunc F(n int32) int32 { return n + 1 }\n")
	in, err := new(gojaengine.Engine).Load(context.Background(), b, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 500 {
		ctx, cancel := context.WithCancel(context.Background())
		res, err := in.Call(ctx, "F", int32(i))
		cancel()
		if err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
		if res[0] != int32(i+1) {
			t.Fatalf("call %d returned %v, %v", i, res, err)
		}
	}
}

// TestStackExhausted returns an error of the package gostrict for a recursion, which exhausts the stack.
func TestStackExhausted(t *testing.T) {
	b := compileScript(t, "package s\n\nfunc F(n int32) int32 { return F(n+1) + 1 }\n")
	in, err := (&gojaengine.Engine{MaxCallStackSize: 100}).Load(context.Background(), b, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.Call(context.Background(), "F", int32(0)); !errors.Is(err, gostrict.ErrStackExhausted) {
		t.Fatalf("expected ErrStackExhausted, got %v", err)
	}
	// the instance works afterwards
	if _, err := in.Call(context.Background(), "F", int32(0)); !errors.Is(err, gostrict.ErrStackExhausted) {
		t.Fatalf("expected ErrStackExhausted again, got %v", err)
	}
}

type requestKey struct{}

// TestPoolSlots binds the functions of the host to each instance of a pool and passes the context of the call.
func TestPoolSlots(t *testing.T) {
	fsys := fstest.MapFS{
		"h/h.go":  {Data: []byte("//gostrict:header\npackage h\n\n//gostrict:host\nfunc Who() string { return \"\" }\n")},
		"main.go": {Data: []byte("package app\n\nimport \"llm.local/app/h\"\n\nfunc Run() string { return h.Who() }\n")},
	}
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/app", gostrict.Module{Path: "llm.local/app", FS: fsys, Trusted: true})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags.Text())
	}
	pool, err := gostrict.NewPoolFunc(context.Background(), goEngine, b, 2, func(s *gostrict.Slot) gostrict.Funcs {
		return gostrict.Funcs{"llm.local/app/h.Who": func([]any) []any {
			req, _ := s.Context().Value(requestKey{}).(string)
			return []any{strconv.Itoa(s.Index) + ":" + req}
		}}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for i := range 4 {
		req := "r" + strconv.Itoa(i)
		res, err := pool.Call(context.WithValue(context.Background(), requestKey{}, req), "Run")
		if err != nil {
			t.Fatal(err)
		}
		got := res[0].(string)
		if !strings.HasSuffix(got, ":"+req) || (got[0] != '0' && got[0] != '1') {
			t.Fatalf("unexpected slot and request %q for %s", got, req)
		}
	}
}
