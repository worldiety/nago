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
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/alias"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/arena"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/draw"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/errs"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/generics"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/host"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/hostuse"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/ints"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/iters"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/money"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/orders"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/physics"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/shapes"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/spin"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/texts"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/trees"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/values"
	"go.wdy.de/nago/pkg/gostrict/std/big"
)

type diffCase struct {
	pkg  string
	name string
	args []any
}

// goEngine is shared by the tests, so that they use its cache of compiled programs.
var goEngine = &gojaengine.Engine{}

// exports are the exported functions of the examples for the native engine.
var exports = map[string]gostrict.Funcs{
	"alias":    native.Reflect(map[string]any{"Order": alias.Order, "Run": alias.Run}),
	"arena":    native.Reflect(map[string]any{"Frame": arena.Frame, "NewClient": arena.NewClient, "OnHit": arena.OnHit, "OnWorld": arena.OnWorld, "Status": arena.Status}),
	"draw":     native.Reflect(map[string]any{"Frame": draw.Frame, "New": draw.New}),
	"errs":     native.Reflect(map[string]any{"Average": errs.Average, "Run": errs.Run}),
	"generics": native.Reflect(map[string]any{"Run": generics.Run, "SumFloats": generics.SumFloats}),
	"hostuse":  native.Reflect(map[string]any{"Callbacks": hostuse.Callbacks, "Records": hostuse.Records, "Run": hostuse.Run}),
	"ints":     native.Reflect(map[string]any{"Bits": ints.Bits, "Conv": ints.Conv, "Div": ints.Div, "DivU": ints.DivU, "Hash": ints.Hash, "Wrap": ints.Wrap, "Xorshift": ints.Xorshift}),
	"iters":    native.Reflect(map[string]any{"FirstOver": iters.FirstOver, "Live": iters.Live, "Nested": iters.Nested, "Sum": iters.Sum, "Switch": iters.Switch}),
	"money":    native.Reflect(map[string]any{"Calc": money.Calc, "Factorial": money.Factorial, "Ints": money.Ints, "Invoice": money.Invoice, "Limits": money.Limits, "Parse": money.Parse, "Round": money.Round}),
	"orders":   native.Reflect(map[string]any{"NewGame": orders.NewGame, "Snapshot": orders.Snapshot, "Step": orders.Step, "Summarize": orders.Summarize}),
	"physics":  native.Reflect(map[string]any{"Simulate": physics.Simulate, "Speed": physics.Speed}),
	"shapes":   native.Reflect(map[string]any{"Area": shapes.Area, "ColorName": shapes.ColorName, "Largest": shapes.Largest, "NewCanvas": shapes.NewCanvas, "OnEvent": shapes.OnEvent, "Scale": shapes.Scale, "Summary": shapes.Summary}),
	"spin":     native.Reflect(map[string]any{"Frame": spin.Frame, "Orbit": spin.Orbit, "Setup": spin.Setup}),
	"texts":    native.Reflect(map[string]any{"Misc": texts.Misc, "Order": texts.Order, "Parse": texts.Parse, "Reverse": texts.Reverse, "Stats": texts.Stats, "Words": texts.Words}),
	"trees":    native.Reflect(map[string]any{"Run": trees.Run, "Values": trees.Values}),
	"values":   native.Reflect(map[string]any{"Run": values.Run}),
}

// nativeEngine returns the native engine of an example.
func nativeEngine(pkg string) native.Engine {
	return native.Engine{Exports: exports[pkg]}
}

// TestDifferential runs each case natively and translated in goja, both behind the same engine interface, and
// compares the results, floats up to the tolerance.
func TestDifferential(t *testing.T) {
	cases := []diffCase{
		{"generics", "Run", nil},
		{"generics", "SumFloats", []any{0.1, 0.2, 0.3}},
		{"physics", "Simulate", []any{int32(600), int32(0)}},
		{"physics", "Simulate", []any{int32(600), int32(45)}},
		{"physics", "Speed", []any{3.0, -4.0}},
		{"texts", "Reverse", []any{"Grüße 😀!"}},
		{"texts", "Stats", []any{"Möhre 😀 ö"}},
		{"texts", "Words", []any{"  der Hund und der Hase und der  Igel "}},
		{"texts", "Order", []any{"\uFFFD", "😀"}},
		{"texts", "Order", []any{"abc", "abd"}},
		{"texts", "Parse", []any{"-2147483648"}},
		{"texts", "Parse", []any{"2147483648"}},
		{"texts", "Parse", []any{"12x"}},
		{"texts", "Parse", []any{"1073741824"}},
		{"texts", "Misc", []any{"\u0085 ab😀a \uFEFF"}},
		{"ints", "Hash", []any{"Grüße 😀"}},
		{"ints", "Wrap", []any{int32(math.MaxInt32), int32(7)}},
		{"ints", "Wrap", []any{int32(math.MinInt32), int32(-1)}},
		{"ints", "Div", []any{int32(-7), int32(2)}},
		{"ints", "Div", []any{int32(math.MinInt32), int32(-1)}},
		{"ints", "Div", []any{int32(5), int32(0)}},
		{"ints", "DivU", []any{uint32(4000000000), uint32(7)}},
		{"ints", "Bits", []any{uint32(0xDEADBEEF)}},
		{"ints", "Conv", []any{-2.5}},
		{"ints", "Conv", []any{math.NaN()}},
		{"ints", "Conv", []any{1e12}},
		{"ints", "Conv", []any{math.Inf(-1)}},
		{"ints", "Xorshift", []any{uint32(2463534242), int32(1000)}},
		{"values", "Run", nil},
		{"iters", "FirstOver", []any{int32(20), int32(7)}},
		{"iters", "FirstOver", []any{int32(4), int32(7)}},
		{"iters", "Sum", []any{int32(30)}},
		{"iters", "Nested", []any{int32(10)}},
		{"iters", "Switch", []any{"hallo-welt-xyz"}},
		{"iters", "Live", nil},
		{"errs", "Run", []any{int32(40)}},
		{"errs", "Run", []any{int32(0)}},
		{"errs", "Run", []any{int32(-3)}},
		{"trees", "Run", []any{int32(0)}},
		{"trees", "Run", []any{int32(200)}},
		{"trees", "Values", nil},
		{"hostuse", "Callbacks", []any{int32(5), 100.0}},
		{"hostuse", "Callbacks", []any{int32(5), 3.0}},
		{"alias", "Run", nil},
		{"alias", "Order", nil},
		{"money", "Ints", []any{"123456789012345678901234567890", "-987654321"}},
		{"money", "Ints", []any{"-7", "2"}},
		{"money", "Ints", []any{"5", "0"}},
		{"money", "Factorial", []any{int32(0)}},
		{"money", "Factorial", []any{int32(40)}},
		{"money", "Parse", []any{"-0012.3400"}},
		{"money", "Parse", []any{"+7"}},
		{"money", "Parse", []any{"1e3"}},
		{"money", "Parse", []any{".5"}},
		{"money", "Parse", []any{"5."}},
		{"money", "Parse", []any{"--1"}},
		{"money", "Parse", []any{"1.2.3"}},
		{"money", "Parse", []any{" 1"}},
		{"money", "Parse", []any{""}},
		{"money", "Calc", []any{"10", "3", int32(4), int32(0)}},
		{"money", "Calc", []any{"-10", "3", int32(2), int32(1)}},
		{"money", "Calc", []any{"0.1", "0.2", int32(10), int32(2)}},
		{"money", "Calc", []any{"1.005", "0.001", int32(0), int32(3)}},
		{"money", "Calc", []any{"1", "0", int32(2), int32(0)}},
		{"money", "Calc", []any{"1.00", "1", int32(-3), int32(0)}},
		{"money", "Round", []any{[]any{"2.5", "-2.5", "3.5", "-3.5", "2.51", "-2.49", "0.004", "-0.005", "12.345", "7"}, int32(0)}},
		{"money", "Round", []any{[]any{"12.345", "-12.345", "12.355", "0.005", "x"}, int32(2)}},
		{"money", "Invoice", []any{[]any{"19.99", "0.333", "1234567890.12"}, "19"}},
		{"money", "Invoice", []any{[]any{"0.10", "0.10", "0.10"}, "7"}},
		{"money", "Limits", []any{"-2147483648", uint32(4294967295), int32(-5)}},
		{"money", "Limits", []any{"2147483648", uint32(0), int32(30)}},
		{"money", "Limits", []any{"9007199254740993", uint32(7), int32(5000)}},
		{"money", "Limits", []any{"-" + strings.Repeat("9", 400), uint32(1), int32(0)}},
		{"money", "Limits", []any{"18446744073709553664", uint32(1), int32(0)}},
		{"money", "Limits", []any{"-" + strings.Repeat("7", 300), uint32(1), int32(0)}},
	}

	ctx := context.Background()
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s.%s%v", tc.pkg, tc.name, tc.args), func(t *testing.T) {
			b := compile(t, "internal/examples/"+tc.pkg)
			want := call(t, native.Engine{Exports: exports[tc.pkg]}, b, nil, tc.name, tc.args)
			got := call(t, goEngine, b, nil, tc.name, tc.args)
			if !gostrict.CloseValues(want, got, gojaengine.Tolerance) {
				t.Fatalf("the engines differ\n  native: %#v\n  goja:   %#v", want, got)
			}
			t.Logf("%v", want)
		})
	}
	_ = ctx
}

// TestHost calls the native functions of a header package in goja, checks, that they agree with their bindings,
// and runs the program with the bindings.
func TestHost(t *testing.T) {
	b := compile(t, "internal/examples/hostuse")

	calls := 0
	h := gostrict.Funcs{}
	for name, f := range host.Host() {
		h[name] = func(args []any) []any {
			calls++
			return f(args)
		}
	}

	for _, args := range [][]any{{3.0, 4.0}, {3.0, 0.0}, {200.0, math.Copysign(0, -1)}} {
		want, _ := hostuse.Run(args[0].(float64), args[1].(float64))
		viaHost := call(t, goEngine, b, h, "Run", args)
		viaJS := call(t, goEngine, b, nil, "Run", args)
		if viaHost[0] != want || !gostrict.SameValues(viaHost, viaJS) {
			t.Fatalf("%v: host %v, bindings %v, want %v", args, viaHost, viaJS, want)
		}
	}
	for _, d := range []float64{0, 1.5, -2} {
		w1, w2, w3 := hostuse.Records(d)
		want := []any{w1, w2, w3}
		viaHost := call(t, goEngine, b, h, "Records", []any{d})
		viaJS := call(t, goEngine, b, nil, "Records", []any{d})
		if !gostrict.SameValues(viaHost, want) || !gostrict.SameValues(viaJS, want) {
			t.Fatalf("records %v: host %v, bindings %v, want %v", d, viaHost, viaJS, want)
		}
	}
	if calls == 0 {
		t.Fatal("the host functions were not called")
	}

	// the callbacks are called during the call and the kept one later, between two calls
	var kept gostrict.Callback
	keep := "go.wdy.de/nago/pkg/gostrict/internal/examples/host.Keep"
	withKeep := gostrict.Funcs{}
	for name, f := range h {
		withKeep[name] = f
	}
	withKeep[keep] = func(args []any) []any {
		kept = args[0].(gostrict.Callback)
		return h[keep](args)
	}
	in, err := goEngine.Load(context.Background(), b, withKeep)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	for _, limit := range []float64{100, 5} {
		w1, w2, w3 := hostuse.Callbacks(6, limit)
		rects, err := gostrict.ToNeutral(w1)
		if err != nil {
			t.Fatal(err)
		}
		want := []any{rects, w2, w3}
		viaHost, err := in.Call(context.Background(), "Callbacks", int32(6), limit)
		if err != nil {
			t.Fatal(err)
		}
		viaJS := call(t, goEngine, b, nil, "Callbacks", []any{int32(6), limit})
		if !gostrict.SameValues(viaHost, want) || !gostrict.SameValues(viaJS, want) {
			t.Fatalf("callbacks %v: host %v, bindings %v, want %v", limit, viaHost, viaJS, want)
		}
	}
	if got, err := kept.Call(1.5); err != nil || got[0] != 9.0 {
		t.Fatalf("the kept callback returned %v, %v", got, err)
	}

	// an error of a callback is reported by the next call
	if _, err := kept.Call("x"); err == nil {
		t.Fatal("a callback took a wrong argument")
	}
	if _, err := in.Call(context.Background(), "Callbacks", int32(1), 1.0); err == nil || !strings.Contains(err.Error(), "callback") {
		t.Fatalf("the error of the callback was not reported: %v", err)
	}
	if _, err := in.Call(context.Background(), "Callbacks", int32(1), 1.0); err != nil {
		t.Fatalf("the error of the callback was reported twice: %v", err)
	}

	mismatches, err := gojaengine.CheckBindings(context.Background(), b, host.Host(), 200)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mismatches {
		t.Error(m)
	}
}

// TestBigBindings checks the bindings of std/big, which the money example uses, against math/big.
func TestBigBindings(t *testing.T) {
	b := compile(t, "internal/examples/money")
	mismatches, err := gojaengine.CheckBindings(context.Background(), b, big.Host(), 500)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mismatches {
		t.Error(m)
	}
	checked := 0
	for _, x := range b.Externs {
		if _, ok := big.Host()[x.Name]; ok {
			checked++
		}
	}
	if checked < 10 {
		t.Fatalf("only %d functions of std/big were checked", checked)
	}
}

// TestBrokenBinding expects the check of the bindings to find, that Math.round is not math.Round.
func TestBrokenBinding(t *testing.T) {
	b := compile(t, "testdata/headers/brokenuse")
	mismatches, err := gojaengine.CheckBindings(context.Background(), b, gostrict.Funcs{
		"go.wdy.de/nago/pkg/gostrict/testdata/headers/broken.Round": func(args []any) []any {
			return []any{math.Round(args[0].(float64))}
		},
	}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(mismatches) == 0 {
		t.Fatal("the broken binding was not found")
	}
	t.Logf("found %d mismatches, e.g. %s", len(mismatches), mismatches[0])
}

func call(t *testing.T, e gostrict.Engine, b *gostrict.Bundle, host gostrict.Funcs, name string, args []any) []any {
	t.Helper()
	ctx := context.Background()
	in, err := e.Load(ctx, b, host)
	if err != nil {
		t.Fatalf("%v\n%s", err, numbered(b.JS))
	}
	defer in.Close()
	res, err := in.Call(ctx, name, args...)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

var (
	bundlesMu sync.Mutex
	bundles   = map[string]*gostrict.Bundle{}
)

func compile(t *testing.T, dir string) *gostrict.Bundle {
	t.Helper()
	bundlesMu.Lock()
	defer bundlesMu.Unlock()
	if b, ok := bundles[dir]; ok {
		return b
	}

	b, diags, err := gostrict.Compile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) > 0 {
		t.Fatalf("%s is not in the subset:\n%s", dir, diags.Text())
	}
	bundles[dir] = b
	return b
}

func numbered(js string) string {
	var sb strings.Builder
	for i, l := range strings.Split(js, "\n") {
		fmt.Fprintf(&sb, "%4d %s\n", i+1, l)
	}
	return sb.String()
}
