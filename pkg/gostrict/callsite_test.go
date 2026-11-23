// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

// The acceptance tests of gostrict.CallSite (wokoda N30, R-DEC-SCRIPT-ABORT-POSITION-NAGO, L-N30-1, the rows "nago: …"
// of the table in R-SCRIPT-ABORT-POSITION), written without knowledge of the implementation. A host function asks
// CallSite while it runs and records the answer; the expected line is the line of the marker //at:<tag> in the source,
// so that the source can change without the numbers of the test.
//
// Each counter row is chosen so that without the rule another site or none would appear:
//   - the file: the same call in y.go instead of x.go;
//   - the position of the name: a selector over two lines, the third method of a chain over three lines (the start of
//     the expression is the first line), a function value and a method value (the line of the call of the value is a
//     later one);
//   - the held index: during a callback the site of the call within it, after the callback again the site of the call,
//     which runs the callback (both for Callback.Call and for a fast call), and no site in $enter of a call after an
//     abort and in $exit of the outermost entry point after a gostrict.ReturnError in a nested one.
//
// Binding (the contract does not fix it): the file is the import path of the package and the name of the file, like
// in Bundle.Positions and gostrict.ErrorPos, e.g. llm.local/p/x.go.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// csSite is an answer of gostrict.CallSite.
type csSite struct {
	file string
	line int
	ok   bool
}

func (s csSite) String() string {
	if !s.ok {
		return fmt.Sprintf("none (%q, %d)", s.file, s.line)
	}
	return fmt.Sprintf("%s:%d", s.file, s.line)
}

// csAsk is the only call of gostrict.CallSite in this file, so that the file compiles except for it, until it exists.
func csAsk(in gostrict.Instance) csSite {
	file, line, ok := gostrict.CallSite(in)
	return csSite{file: file, line: line, ok: ok}
}

// errCsAbort is the error, with which the host function Fail aborts the call.
var errCsAbort = errors.New("aborted by the host at its call site")

// csHeader is the header of the tests: each function records the answer of CallSite under its tag.
const csHeader = `// Package hdr asks the host for the site of each call.
//
//gostrict:header
package hdr

import "go.wdy.de/nago/pkg/gostrict/std/js"

// Probe records the site of its call.
//
//gostrict:host
func Probe(tag string) {}

// Fail records the site of its call and aborts the call.
//
//gostrict:host
func Fail(tag string) {}

// Reject records the site of its call and ends its entry point with an error.
//
//gostrict:host
func Reject(tag string) {}

// Each records the site of its call, calls f and records the site again.
//
//gostrict:host
//gostrict:calls now
func Each(tag string, f func(i int32)) {}

// EachFast is Each with a fast call of f.
//
//gostrict:host
//gostrict:calls now
func EachFast(tag string, f func(i int32)) {}

// T is a handle.
type T struct{ ref js.Ref }

// New records the site of its call and returns a handle.
//
//gostrict:host
func New(tag string) T { return T{} }

// A records the site of its call.
//
//gostrict:host
func (t T) A(tag string) T { return t }

// B records the site of its call.
//
//gostrict:host
func (t T) B(tag string) T { return t }
`

// csX and csY are the files of the program p.
const csX = `// Package p calls the functions of hdr.
package p

import "llm.local/p/hdr"

func Plain() {
	hdr.Probe("plain") //at:plain
}

func Selector() {
	hdr.
		Probe("selector") //at:selector
}

func Chain() {
	_ = hdr.New("c1").
		A("c2").
		B("c3") //at:c3
}

func FuncValue() {
	f := hdr.Probe //at:funcvalue
	hdr.Probe("between")
	f("funcvalue")
}

func MethodValue() {
	t := hdr.New("m0")
	g := t.A //at:methodvalue
	hdr.Probe("between")
	_ = g("methodvalue")
}

func Callback() {
	hdr.Each("each", func(i int32) { //at:each
		hdr.Probe("each.inner") //at:each.inner
	})
}

func CallbackFast() {
	hdr.EachFast("fast", func(i int32) { //at:fast
		hdr.Probe("fast.inner") //at:fast.inner
	})
}

func Abort() {
	hdr.Fail("abort") //at:abort
}
`

const csY = `package p

import "llm.local/p/hdr"

func Other() {
	hdr.Probe("other") //at:other
}
`

// csLine is the line of the marker //at:<tag> in src.
func csLine(t *testing.T, src, tag string) int {
	t.Helper()
	for i, l := range strings.Split(src, "\n") {
		if strings.HasSuffix(strings.TrimSpace(l), "//at:"+tag) {
			return i + 1
		}
	}
	t.Fatalf("premise: no marker //at:%s in the source", tag)
	return 0
}

// csHost is the host of the program: it records the answers of CallSite by tag, in instance in and, while other is
// set, also in the instance other.
type csHost struct {
	t     *testing.T
	mu    sync.Mutex
	in    gostrict.Instance
	other gostrict.Instance
	seen  map[string]csSite
}

func newCsHost(t *testing.T) *csHost {
	return &csHost{t: t, seen: map[string]csSite{}}
}

func (h *csHost) record(tag string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seen[tag] = csAsk(h.in)
	if h.other != nil {
		h.seen[tag+"@other"] = csAsk(h.other)
	}
}

func (h *csHost) get(tag string) (csSite, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.seen[tag]
	return s, ok
}

func (h *csHost) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seen = map[string]csSite{}
}

func (h *csHost) funcs(prefix string) gostrict.Funcs {
	probe := func(args []any) []any {
		h.record(args[0].(string))
		return nil
	}
	return gostrict.Funcs{
		prefix + "hdr.Probe": probe,
		prefix + "hdr.Fail": func(args []any) []any {
			h.record(args[0].(string))
			gostrict.Abort(errCsAbort)
			return nil
		},
		prefix + "hdr.Reject": func(args []any) []any {
			h.record(args[0].(string))
			gostrict.ReturnError(errors.New("rejected " + args[0].(string)))
			return nil
		},
		prefix + "hdr.Each": func(args []any) []any {
			tag := args[0].(string)
			h.record(tag + ".before")
			if _, err := args[1].(gostrict.Callback).Call(int32(1)); err != nil {
				h.t.Errorf("the callback of %s failed: %v", tag, err)
			}
			h.record(tag + ".after")
			return nil
		},
		prefix + "hdr.EachFast": func(args []any) []any {
			tag := args[0].(string)
			h.record(tag + ".before")
			fc, ok := args[1].(gostrict.FastCallback)
			if !ok {
				// premise: the row tests the fast call, not Call once more
				h.t.Errorf("premise: the callback of %s is no gostrict.FastCallback: %T", tag, args[1])
				return nil
			}
			w := fc.Begin()
			w.Int32(1)
			_, err := fc.Invoke()
			fc.End()
			if err != nil {
				h.t.Errorf("the fast call of %s failed: %v", tag, err)
			}
			h.record(tag + ".after")
			return nil
		},
		prefix + "hdr.New": func(args []any) []any {
			h.record(args[0].(string))
			return []any{js.NewRef(args[0].(string))}
		},
		prefix + "hdr.T.A": func(args []any) []any {
			h.record(args[1].(string))
			return []any{args[0]}
		},
		prefix + "hdr.T.B": func(args []any) []any {
			h.record(args[1].(string))
			return []any{args[0]}
		},
	}
}

// csCompile compiles the program p with the files x.go and y.go and the header hdr.
func csCompile(t *testing.T) *gostrict.Bundle {
	t.Helper()
	prog, err := gostrict.LoadModules("llm.local/p", gostrict.Module{Path: "llm.local/p", Trusted: true, FS: fstest.MapFS{
		"x.go":     {Data: []byte(csX)},
		"y.go":     {Data: []byte(csY)},
		"hdr/h.go": {Data: []byte(csHeader)},
	}})
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
	return b
}

// TestCallSite: each row calls an export of p in the vm and compares the site, which a host function saw, with the
// marker in the source; rows and counter rows stand next to each other (see the head of the file).
func TestCallSite(t *testing.T) {
	b := csCompile(t)
	ctx := context.Background()
	h := newCsHost(t)
	in, err := vmEngine.Load(ctx, b, h.funcs("llm.local/p/"))
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	h.in = in
	other, err := vmEngine.Load(ctx, b, newCsHost(t).funcs("llm.local/p/"))
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()

	const x, y = "llm.local/p/x.go", "llm.local/p/y.go"
	for _, c := range []struct {
		name, export string
		other        bool   // asks CallSite of another instance, too (tag@other)
		tag          string // the tag of the recorded answer
		file         string // "" expects no site
		src, marker  string
	}{
		{name: "a host function asks in the call from x.go", export: "Plain", tag: "plain", file: x, src: csX, marker: "plain"},
		{name: "counter row: the same call in y.go", export: "Other", tag: "other", file: y, src: csY, marker: "other"},
		{name: "another instance of the same bundle, in which no host function runs, has no site", export: "Plain", other: true, tag: "plain@other"},
		{name: "counter row: the own instance at the same moment", export: "Plain", other: true, tag: "plain", file: x, src: csX, marker: "plain"},
		{name: "a selector over two lines: the line of the name", export: "Selector", tag: "selector", file: x, src: csX, marker: "selector"},
		{name: "the third method of a chain over three lines: the line of the third", export: "Chain", tag: "c3", file: x, src: csX, marker: "c3"},
		{name: "a function value: the line, where the value is formed, not that of its call", export: "FuncValue", tag: "funcvalue", file: x, src: csX, marker: "funcvalue"},
		{name: "a method value: the line, where the value is formed, not that of its call", export: "MethodValue", tag: "methodvalue", file: x, src: csX, marker: "methodvalue"},
		{name: "during a callback: the site of the host function within it", export: "Callback", tag: "each.inner", file: x, src: csX, marker: "each.inner"},
		{name: "counter row: after the callback again the site of the host function, which called it", export: "Callback", tag: "each.after", file: x, src: csX, marker: "each"},
		{name: "during a fast call: the site of the host function within it", export: "CallbackFast", tag: "fast.inner", file: x, src: csX, marker: "fast.inner"},
		{name: "counter row: after the fast call again the site of the host function, which called it", export: "CallbackFast", tag: "fast.after", file: x, src: csX, marker: "fast"},
	} {
		t.Run(c.name, func(t *testing.T) {
			h.reset()
			if c.other {
				h.other = other
				defer func() { h.other = nil }()
			}
			if _, err := in.Call(ctx, c.export); err != nil {
				t.Fatalf("%s: %v", c.export, err)
			}
			got, seen := h.get(c.tag)
			if !seen {
				t.Fatalf("premise: no host function recorded %s", c.tag)
			}
			if c.file == "" {
				if got.ok {
					t.Errorf("site %v, want none", got)
				}
				return
			}
			want := csSite{file: c.file, line: csLine(t, c.src, c.marker), ok: true}
			if got != want {
				t.Errorf("site %v, want %v", got, want)
			}
		})
	}

	t.Run("without a running host function: none after the call", func(t *testing.T) {
		h.reset()
		if _, err := in.Call(ctx, "Plain"); err != nil {
			t.Fatal(err)
		}
		if s, _ := h.get("plain"); !s.ok {
			t.Fatalf("premise: the host function saw no site during the call: %v", s)
		}
		if s := csAsk(in); s.ok {
			t.Errorf("after the call: site %v, want none", s)
		}
	})

	t.Run("errors.Is on the error of an abort is unchanged true", func(t *testing.T) {
		h.reset()
		_, err := in.Call(ctx, "Abort")
		if s, _ := h.get("abort"); !s.ok || s.line != csLine(t, csX, "abort") {
			t.Errorf("premise: Fail saw the site %v, want x.go:%d", s, csLine(t, csX, "abort"))
		}
		if !errors.Is(err, errCsAbort) {
			t.Errorf("errors.Is(%v, the error of Abort) is false", err)
		}
		if err != errCsAbort {
			t.Errorf("the call returned %#v, not the error of Abort itself", err)
		}
		if err != nil && err.Error() != errCsAbort.Error() {
			t.Errorf("Error() is %q, want %q", err.Error(), errCsAbort.Error())
		}
		if s := csAsk(in); s.ok {
			t.Errorf("after the abort: site %v, want none", s)
		}
	})
}

// csWholeHeader is the header of the whole program w.
const csWholeHeader = `// Package hdr asks the host for the site of each call.
//
//gostrict:header
package hdr

// Probe records the site of its call.
//
//gostrict:host
func Probe(tag string) {}

// Fail records the site of its call and aborts the call.
//
//gostrict:host
func Fail(tag string) {}

// Reject records the site of its call and ends its entry point with an error.
//
//gostrict:host
func Reject(tag string) {}
`

// csWholeA is the outer entry point: abort aborts in a, reject lets b end with gostrict.ReturnError, after that a
// calls Probe.
const csWholeA = `// Package a is the outer entry point.
package a

import (
	"llm.local/w/b"
	"llm.local/w/hdr"
)

func Main(mode string) (string, error) {
	if mode == "abort" {
		hdr.Fail("abort") //at:abort
	}
	_, err := b.Main(mode)
	hdr.Probe("after") //at:after
	if err != nil {
		return "value " + err.Error(), nil
	}
	return "ok", nil
}
`

const csWholeB = `// Package b is the nested entry point.
package b

import "llm.local/w/hdr"

func Main(mode string) (string, error) {
	if mode == "reject" {
		hdr.Reject("reject") //at:reject
	}
	return "b", nil
}
`

// csAspect records the answers of CallSite in $enter and $exit of each entry point, by "enter <entry>" and
// "exit <entry>".
type csAspect struct {
	h    *csHost
	seen map[string]csSite
}

func (a *csAspect) Enter(entry string) error {
	a.seen["enter "+entry] = csAsk(a.h.in)
	return nil
}

func (a *csAspect) Exit(entry string, err error) {
	a.seen["exit "+entry] = csAsk(a.h.in)
}

// frame is the answer for the frame, e.g. "enter llm.local/w/a.Main".
func (a *csAspect) frame(t *testing.T, frame string) csSite {
	t.Helper()
	s, ok := a.seen[frame]
	if !ok {
		t.Fatalf("premise: no %s in %v", frame, a.seen)
	}
	return s
}

// TestCallSiteWhole: a whole program with an aspect in the vm. $enter of an outermost call has no site, also in the
// second call of the same instance after an abort in a host function (counter row: the host function saw its site
// before it aborted). After gostrict.ReturnError in a nested entry point a host function in the outer one sees its own
// site, and $exit of the outermost entry point has none (counter row: Reject saw its site before it ended b).
func TestCallSiteWhole(t *testing.T) {
	prog, err := gostrict.LoadRoots([]string{"llm.local/w/a", "llm.local/w/b"}, gostrict.Module{Path: "llm.local/w", Trusted: true, FS: fstest.MapFS{
		"a/a.go":   {Data: []byte(csWholeA)},
		"b/b.go":   {Data: []byte(csWholeB)},
		"hdr/h.go": {Data: []byte(csWholeHeader)},
	}})
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
	ctx := context.Background()
	h := newCsHost(t)
	asp := &csAspect{h: h, seen: map[string]csSite{}}
	in, err := new(vm.Engine).Load(ctx, b, h.funcs("llm.local/w/").WithAspect(asp))
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	h.in = in
	const a = "llm.local/w/a.Main"
	const fileA, fileB = "llm.local/w/a/a.go", "llm.local/w/b/b.go"
	site := func(src, marker, file string) csSite {
		return csSite{file: file, line: csLine(t, src, marker), ok: true}
	}

	t.Run("the first call: $enter of the outermost entry point has no site", func(t *testing.T) {
		asp.seen = map[string]csSite{}
		h.reset()
		_, err := in.Call(ctx, a, "abort")
		if !errors.Is(err, errCsAbort) || err != errCsAbort {
			t.Errorf("the abort returned %#v, want the error of Abort itself", err)
		}
		if s, _ := h.get("abort"); s != site(csWholeA, "abort", fileA) {
			t.Errorf("premise: Fail saw %v, want %v", s, site(csWholeA, "abort", fileA))
		}
		if got := asp.frame(t, "enter "+a); got.ok {
			t.Errorf("$enter of the first call: %v, want none", got)
		}
	})
	t.Run("the second call of the same instance after the abort: $enter has no site", func(t *testing.T) {
		asp.seen = map[string]csSite{}
		res, err := in.Call(ctx, a, "ok")
		if err != nil || len(res) == 0 || res[0] != "ok" {
			t.Fatalf("the second call: %v, %v", res, err)
		}
		if got := asp.frame(t, "enter "+a); got.ok {
			t.Errorf("$enter of the second call: %v, want none (the abort in the first call was at %v)", got, site(csWholeA, "abort", fileA))
		}
	})
	t.Run("ReturnError in a nested entry point: the next host function sees its site, $exit of the outermost none", func(t *testing.T) {
		asp.seen = map[string]csSite{}
		h.reset()
		res, err := in.Call(ctx, a, "reject")
		if err != nil || len(res) == 0 || res[0] != "value rejected reject" {
			t.Fatalf("premise: b did not end with the error as a value of a: %v, %v", res, err)
		}
		if s, _ := h.get("reject"); s != site(csWholeB, "reject", fileB) {
			t.Errorf("premise: Reject saw %v, want %v", s, site(csWholeB, "reject", fileB))
		}
		if s, _ := h.get("after"); s != site(csWholeA, "after", fileA) {
			t.Errorf("the host function after the ReturnError saw %v, want %v", s, site(csWholeA, "after", fileA))
		}
		if got := asp.frame(t, "exit "+a); got.ok {
			t.Errorf("$exit of the outermost entry point: %v, want none (Reject was at %v)", got, site(csWholeB, "reject", fileB))
		}
	})
}

// TestCallSiteOtherEngine: goja runs the same program and the same host function, but reports no site; a statement,
// not a counter row.
func TestCallSiteOtherEngine(t *testing.T) {
	b := csCompile(t)
	ctx := context.Background()
	h := newCsHost(t)
	in, err := new(gojaengine.Engine).Load(ctx, b, h.funcs("llm.local/p/"))
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	h.in = in
	if _, err := in.Call(ctx, "Plain"); err != nil {
		t.Fatal(err)
	}
	s, seen := h.get("plain")
	if !seen {
		t.Fatal("premise: the host function did not run in goja")
	}
	if s.ok {
		t.Errorf("goja reports the site %v, want none", s)
	}
}
