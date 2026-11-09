// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
)

const concurrentHeader = `// Package hdr calls functions on another goroutine.
//
//gostrict:header
package hdr

// Later calls done after it returned on another goroutine.
//
//gostrict:host
//gostrict:calls concurrent
func Later(done func(v string) string) {}

// Hold blocks, until the test lets it go on.
//
//gostrict:host
func Hold() {}
`

const concurrentProgram = `// Package p calls a function on another goroutine.
//
//gostrict:terminating
package p

import "llm.local/p/hdr"

func Start(id string) {
	hdr.Later(func(v string) string {
		return id + v + helper()
	})
}

func helper() string {
	return "!"
}

func Busy() {
	hdr.Hold()
}

func Chain() {
	hdr.Later(again)
}

func again(v string) string {
	Chain()
	return v
}
`

func compileConcurrent(t *testing.T) (*gostrict.Program, *gostrict.Bundle) {
	t.Helper()
	prog, err := gostrict.LoadModules("llm.local/p", gostrict.Module{Path: "llm.local/p", Trusted: true, FS: fstest.MapFS{
		"p.go":     {Data: []byte(concurrentProgram)},
		"hdr/h.go": {Data: []byte(concurrentHeader)},
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
	return prog, b
}

// concurrentHost records the detached function and lets Hold block.
type concurrentHost struct {
	mu      sync.Mutex
	done    gostrict.Detached
	entered chan struct{}
	release chan struct{}
}

func newConcurrentHost() *concurrentHost {
	return &concurrentHost{entered: make(chan struct{}, 8), release: make(chan struct{}, 8)}
}

func (h *concurrentHost) funcs(t *testing.T) gostrict.Funcs {
	return gostrict.Funcs{
		"llm.local/p/hdr.Later": func(args []any) []any {
			d, err := gostrict.Detach(args[0].(gostrict.Callback))
			if err != nil {
				t.Error(err)
			}
			h.mu.Lock()
			defer h.mu.Unlock()
			h.done = d
			return nil
		},
		"llm.local/p/hdr.Hold": func(args []any) []any {
			h.entered <- struct{}{}
			<-h.release
			return nil
		},
	}
}

func TestDetachedRunsInAnotherInstanceWhileTheFirstIsBusy(t *testing.T) {
	_, b := compileConcurrent(t)
	ctx := context.Background()
	h := newConcurrentHost()
	e := &vm.Engine{}
	first, err := e.Load(ctx, b, h.funcs(t))
	if err != nil {
		t.Fatal(err)
	}
	second, err := e.Load(ctx, b, h.funcs(t))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := first.Call(ctx, "Start", "a"); err != nil {
		t.Fatal(err)
	}
	if h.done.Home() != nil {
		t.Fatal("a closure of the vm may run in any instance")
	}

	busy := make(chan error, 1)
	go func() {
		_, err := first.Call(ctx, "Busy")
		busy <- err
	}()
	<-h.entered

	res, err := h.done.CallIn(ctx, second, "b")
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "ab!" {
		t.Fatalf("unexpected result: %v", res)
	}

	h.release <- struct{}{}
	if err := <-busy; err != nil {
		t.Fatal(err)
	}

	_, other := compileConcurrent(t)
	foreign, err := e.Load(ctx, other, h.funcs(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.done.CallIn(ctx, foreign, "b"); err == nil {
		t.Fatal("a closure must not run in an instance of another bundle")
	}
}

func TestDetachedWaitsForItsRuntimeInGoja(t *testing.T) {
	_, b := compileConcurrent(t)
	ctx := context.Background()
	h := newConcurrentHost()
	e := &gojaengine.Engine{}
	in, err := e.Load(ctx, b, h.funcs(t))
	if err != nil {
		t.Fatal(err)
	}
	other, err := e.Load(ctx, b, h.funcs(t))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := in.Call(ctx, "Start", "a"); err != nil {
		t.Fatal(err)
	}
	if h.done.Home() != in {
		t.Fatal("a function of JavaScript stays in its runtime")
	}
	if _, err := h.done.CallIn(ctx, other, "b"); err == nil {
		t.Fatal("a function of JavaScript must not run in another runtime")
	}

	busy := make(chan error, 1)
	go func() {
		_, err := in.Call(ctx, "Busy")
		busy <- err
	}()
	<-h.entered

	type result struct {
		res []any
		err error
	}
	ended := make(chan result, 1)
	go func() {
		res, err := h.done.CallIn(ctx, in, "b")
		ended <- result{res, err}
	}()
	select {
	case <-ended:
		t.Fatal("the function must wait, until the running call of its runtime has ended")
	case <-time.After(50 * time.Millisecond):
	}

	h.release <- struct{}{}
	r := <-ended
	if r.err != nil || r.res[0] != "ab!" {
		t.Fatalf("unexpected result: %v %v", r.res, r.err)
	}
	if err := <-busy; err != nil {
		t.Fatal(err)
	}
}

func TestPoolRunsDetachedInAFreeInstance(t *testing.T) {
	_, b := compileConcurrent(t)
	ctx := context.Background()
	h := newConcurrentHost()
	p, err := gostrict.NewPool(ctx, &vm.Engine{}, b, h.funcs(t), 2)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	if _, err := p.Call(ctx, "Start", "a"); err != nil {
		t.Fatal(err)
	}
	res, err := p.CallDetached(ctx, h.done, "b")
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "ab!" {
		t.Fatalf("unexpected result: %v", res)
	}
}

// TestGeneratedHostLeavesConcurrentFunctions checks, that the generated adapter, which would call the instance of the
// call on another goroutine, does not exist: an engine refuses the bundle, until the host implements the function.
func TestGeneratedHostLeavesConcurrentFunctions(t *testing.T) {
	prog, err := gostrict.LoadModules("llm.local/p/hdr", gostrict.Module{Path: "llm.local/p", Trusted: true, FS: fstest.MapFS{
		"hdr/h.go": {Data: []byte(concurrentHeader)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	src, err := gostrict.Generate(prog)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(src), "hdr.Later") || !strings.Contains(string(src), "hdr.Hold") {
		t.Fatalf("unexpected generated file:\n%s", src)
	}

	_, b := compileConcurrent(t)
	if _, err := new(vm.Engine).Load(context.Background(), b, gostrict.Funcs{"llm.local/p/hdr.Hold": func([]any) []any { return nil }}); err == nil || !strings.Contains(err.Error(), "hdr.Later") {
		t.Fatalf("the engine must refuse the bundle without the function: %v", err)
	}
}
