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
	"go.wdy.de/nago/pkg/gostrict/internal/examples/errs"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/generics"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/host"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/hostuse"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/ints"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/iters"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/physics"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/texts"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/values"
)

type diffCase struct {
	pkg  string
	name string
	args []any
}

// goEngine is shared by the tests, so that they use its cache of compiled programs.
var goEngine = &gojaengine.Engine{}

// exports are the generated functions of the examples for the native engine.
var exports = map[string]gostrict.Funcs{
	"errs":     errs.Exports(),
	"generics": generics.Exports(),
	"hostuse":  hostuse.Exports(),
	"ints":     ints.Exports(),
	"iters":    iters.Exports(),
	"physics":  physics.Exports(),
	"texts":    texts.Exports(),
	"values":   values.Exports(),
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
	if calls == 0 {
		t.Fatal("the host functions were not called")
	}

	mismatches, err := gojaengine.CheckBindings(context.Background(), b, host.Host(), 200)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mismatches {
		t.Error(m)
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
