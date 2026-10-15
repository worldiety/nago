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
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
)

const wholePrefix = "go.wdy.de/nago/pkg/gostrict/internal/whole/"

var (
	errDenied = errors.New("permission denied")
	errDepth  = errors.New("too deep")
)

// frame is a call of an entry point with the transaction of its area, which it owns or joins.
type frame struct {
	entry, area string
	tx          map[string]string
	owner       bool
}

// fakeHost is a host like wokoda: the user may not call some use cases, the frames are limited, and a frame of
// another area than the frames below it runs in a transaction of its own, which it commits, if it returns no error.
type fakeHost struct {
	data     map[string]string
	frames   []*frame
	log      []string
	denied   map[string]bool
	maxDepth int
}

func newFakeHost() *fakeHost {
	return &fakeHost{data: map[string]string{}, denied: map[string]bool{"admin/wipe.Main": true}, maxDepth: 8}
}

func short(entry string) string {
	return strings.TrimPrefix(entry, wholePrefix)
}

func (h *fakeHost) Enter(entry string) error {
	name := short(entry)
	h.log = append(h.log, "enter "+name)
	switch {
	case h.denied[name]:
		return errDenied
	case len(h.frames) >= h.maxDepth:
		return errDepth
	}
	f := &frame{entry: entry, area: strings.Split(name, "/")[0]}
	for _, below := range h.frames {
		if below.area == f.area {
			f.tx = below.tx
		}
	}
	if f.tx == nil {
		f.tx, f.owner = map[string]string{}, true
	}
	h.frames = append(h.frames, f)
	return nil
}

func (h *fakeHost) Exit(entry string, err error) {
	f := h.frames[len(h.frames)-1]
	h.frames = h.frames[:len(h.frames)-1]
	if f.entry != entry {
		panic("the frames are not nested: " + f.entry + " exits as " + entry)
	}
	state := "ok"
	if err != nil {
		state = err.Error()
	}
	h.log = append(h.log, "exit "+short(entry)+" "+state)
	if f.owner && err == nil {
		maps.Copy(h.data, f.tx)
	}
}

func (h *fakeHost) funcs() gostrict.Funcs {
	return gostrict.Funcs{
		wholePrefix + "store.Get": func(args []any) []any {
			f := h.frames[len(h.frames)-1]
			if v, ok := f.tx[args[0].(string)]; ok {
				return []any{v, true}
			}
			v, ok := h.data[args[0].(string)]
			return []any{v, ok}
		},
		wholePrefix + "store.Put": func(args []any) []any {
			h.frames[len(h.frames)-1].tx[args[0].(string)] = args[1].(string)
			return nil
		},
	}.WithAspect(h)
}

// loadWhole loads the scripts of internal/whole as one program.
func loadWhole(t *testing.T, roots ...string) *gostrict.Program {
	t.Helper()
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, r := range roots {
		paths = append(paths, wholePrefix+r)
	}
	prog, err := gostrict.LoadRoots(paths, gostrict.Module{Path: "go.wdy.de/nago", FS: os.DirFS(dir), Dir: dir, Trusted: true})
	if err != nil {
		t.Fatal(err)
	}
	if diags := gostrict.Check(prog); len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	return prog
}

var wholeRoots = []string{"crm/create", "crm/read", "crm/purge", "billing/invoice", "admin/wipe", "chain/c1", "chain/c2", "chain/c3", "chain/c4"}

// TestWholeProgram runs scripts, which call each other, as one program within the frames of a fake host.
func TestWholeProgram(t *testing.T) {
	prog := loadWhole(t, wholeRoots...)
	b, diags := gostrict.Emit(prog)
	if len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	if len(b.Exports) != len(wholeRoots) {
		t.Fatalf("expected an export of each entry point, got %v", slices.Sorted(maps.Keys(b.Exports)))
	}

	ctx := context.Background()
	if _, err := new(gojaengine.Engine).Load(ctx, b, nil); err == nil || !strings.Contains(err.Error(), gostrict.EnterFunc) {
		t.Fatalf("a whole program was loaded without an aspect: %v", err)
	}

	h := newFakeHost()
	in, err := new(gojaengine.Engine).Load(ctx, b, h.funcs())
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()

	run := func(entry string, args ...any) ([]any, error) {
		h.log = nil
		return in.Call(ctx, wholePrefix+entry, args...)
	}
	expectLog := func(want ...string) {
		t.Helper()
		if !slices.Equal(h.log, want) {
			t.Fatalf("unexpected frames\n got: %q\nwant: %q", h.log, want)
		}
		if len(h.frames) != 0 {
			t.Fatalf("%d frames are left", len(h.frames))
		}
	}

	// the frames nest, read joins the transaction of create and sees its customer, invoice commits its own
	res, err := run("crm/create.Main", "c1", "Alice", int32(100))
	if err != nil || res[0] != "Alice invoice 100" {
		t.Fatalf("create returned %v, %v", res, err)
	}
	expectLog("enter crm/create.Main", "enter crm/read.Main", "exit crm/read.Main ok", "enter billing/invoice.Main", "exit billing/invoice.Main ok", "exit crm/create.Main ok")
	if h.data["customer/c1"] != "Alice" || h.data["invoice/c1"] != "100" {
		t.Fatalf("not committed: %v", h.data)
	}

	// a failing entry point rolls back its own transaction, while its caller continues
	res, err = run("crm/create.Main", "c2", "Bob", int32(-5))
	if err != nil || res[0] != "Bob without invoice: negative amount" {
		t.Fatalf("create returned %v, %v", res, err)
	}
	if _, ok := h.data["invoice/c2"]; ok || h.data["customer/c2"] != "Bob" {
		t.Fatalf("the invoice was not rolled back: %v", h.data)
	}

	// a denied call aborts the whole call, which returns the error of the aspect, and the frames exit with it
	_, err = run("crm/purge.Main")
	if !errors.Is(err, errDenied) {
		t.Fatalf("the denial was not returned: %v", err)
	}
	expectLog("enter crm/purge.Main", "enter admin/wipe.Main", "exit crm/purge.Main aborted")
	if _, ok := h.data["purge"]; ok {
		t.Fatalf("the aborted frame was committed: %v", h.data)
	}

	// the depth of the frames is limited
	h.maxDepth = 3
	_, err = run("chain/c1.Main", int32(0))
	if !errors.Is(err, errDepth) {
		t.Fatalf("the depth was not limited: %v", err)
	}
	expectLog("enter chain/c1.Main", "enter chain/c2.Main", "enter chain/c3.Main", "enter chain/c4.Main", "exit chain/c3.Main aborted", "exit chain/c2.Main aborted", "exit chain/c1.Main aborted")
	h.maxDepth = 8
	res, err = run("chain/c1.Main", int32(0))
	if err != nil || res[0] != int32(6) {
		t.Fatalf("the chain returned %v, %v", res, err)
	}

	// the instance works after an abort
	res, err = run("crm/read.Main", "c1")
	if err != nil || res[0] != "Alice" {
		t.Fatalf("read returned %v, %v", res, err)
	}
	expectLog("enter crm/read.Main", "exit crm/read.Main ok")
}

// TestWholeProgramTargets translates a bundle for each entry point, which contains only what it reaches.
func TestWholeProgramTargets(t *testing.T) {
	prog := loadWhole(t, wholeRoots...)
	read, diags := gostrict.EmitEntries(prog, wholePrefix+"crm/read.Main")
	if len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	create, diags := gostrict.EmitEntries(prog, wholePrefix+"crm/create.Main")
	if len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	if len(read.Exports) != 1 || strings.Contains(read.JS, "invoice") || !strings.Contains(create.JS, "invoice") {
		t.Fatalf("the bundles contain the wrong code:\n%s", read.JS)
	}
	if len(read.JS) >= len(create.JS) {
		t.Fatalf("the bundle of read is not smaller: %d >= %d", len(read.JS), len(create.JS))
	}
	if _, diags := gostrict.EmitEntries(prog, wholePrefix+"crm/read.missing"); len(diags) == 0 {
		t.Fatal("an unknown entry point was accepted")
	}

	h := newFakeHost()
	h.data["customer/x"] = "Xaver"
	in, err := new(gojaengine.Engine).Load(context.Background(), read, h.funcs())
	if err != nil {
		t.Fatal(err)
	}
	res, err := in.Call(context.Background(), wholePrefix+"crm/read.Main", "x")
	if err != nil || res[0] != "Xaver" {
		t.Fatalf("read returned %v, %v", res, err)
	}
}

// TestWholeProgramEntries reports the exported functions of a root, which can not be entry points.
func TestWholeProgramEntries(t *testing.T) {
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	prog, err := gostrict.LoadRoots([]string{wholePrefix + "bad/entry"}, gostrict.Module{Path: "go.wdy.de/nago", FS: os.DirFS(dir), Dir: dir, Trusted: true})
	if err != nil {
		t.Fatal(err)
	}
	diags := gostrict.Check(prog)
	var msgs []string
	for _, d := range diags {
		if d.Code != gostrict.CodeEntry {
			t.Fatalf("unexpected diagnostic %s: %s", d.Code, d.Msg)
		}
		msgs = append(msgs, d.Msg)
	}
	if len(msgs) != 2 || !strings.Contains(msgs[0], "Identity") && !strings.Contains(msgs[1], "Identity") {
		t.Fatalf("unexpected diagnostics %q", msgs)
	}
}

// TestCompileWhole checks a whole program once and keeps a bundle for each set of entry points.
func TestCompileWhole(t *testing.T) {
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var roots []string
	for _, r := range wholeRoots {
		roots = append(roots, wholePrefix+r)
	}
	w, diags, err := gostrict.CompileWhole(roots, gostrict.Module{Path: "go.wdy.de/nago", FS: os.DirFS(dir), Dir: dir, Trusted: true})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags)
	}
	if len(w.Entries()) != len(wholeRoots) {
		t.Fatalf("unexpected entry points %v", w.Entries())
	}
	a, err := w.Bundle(wholePrefix+"crm/read.Main", wholePrefix+"crm/create.Main")
	if err != nil {
		t.Fatal(err)
	}
	b, err := w.Bundle(wholePrefix+"crm/create.Main", wholePrefix+"crm/read.Main")
	if err != nil || a != b || len(a.Exports) != 2 {
		t.Fatalf("the bundle was not kept: %v", err)
	}
	if _, err := w.Bundle(wholePrefix + "nope.Main"); err == nil {
		t.Fatal("an unknown entry point was accepted")
	}
}
