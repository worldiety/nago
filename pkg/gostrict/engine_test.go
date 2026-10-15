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
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/physics"
)

func script(src string) gostrict.Module {
	return gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}}
}

func TestCache(t *testing.T) {
	cache := gostrict.NewCache(2)
	a := script("package s\n\nfunc F() int32 { return 1 }\n")
	b1, _, err := cache.Compile("llm.local/s", a)
	if err != nil {
		t.Fatal(err)
	}
	b2, _, _ := cache.Compile("llm.local/s", a)
	if b1 != b2 {
		t.Fatal("the same sources were compiled twice")
	}

	bad := script("package s\n\nfunc F() int { return 1 }\n")
	_, diags, _ := cache.Compile("llm.local/s", bad)
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}

	// the third program drops the least recently used one
	cache.Compile("llm.local/s", script("package s\n\nfunc F() int32 { return 3 }\n"))
	b3, _, _ := cache.Compile("llm.local/s", a)
	if b3 == b1 {
		t.Fatal("the bundle should have been dropped")
	}
}

// TestSharedRuntime loads two programs into one runtime. Each keeps its state in a value, which the host holds.
func TestSharedRuntime(t *testing.T) {
	cache := gostrict.NewCache(8)
	ctx := context.Background()
	rt := goEngine.NewRuntime()

	type program struct {
		in    gostrict.Instance
		state gostrict.Token
	}
	var programs []program
	for i, src := range []string{
		"package s\n\ntype Counter struct{ N int32 }\n\nfunc New() *Counter { return &Counter{} }\n\nfunc Next(c *Counter) int32 {\n\tc.N++\n\treturn c.N\n}\n",
		"package s\n\ntype Counter struct{ N int32 }\n\nfunc New() *Counter { return &Counter{N: 100} }\n\nfunc Next(c *Counter) int32 {\n\tc.N += 10\n\treturn c.N\n}\n",
	} {
		b, diags, err := cache.Compile("llm.local/s", script(src))
		if err != nil || len(diags) > 0 {
			t.Fatal(i, err, diags.Text())
		}
		in, err := rt.Load(ctx, b, nil)
		if err != nil {
			t.Fatal(err)
		}
		res, err := in.Call(ctx, "New")
		if err != nil {
			t.Fatal(err)
		}
		programs = append(programs, program{in: in, state: res[0].(gostrict.Token)})
	}

	for _, want := range [][2]int32{{1, 110}, {2, 120}} {
		for i, p := range programs {
			res, err := p.in.Call(ctx, "Next", p.state)
			if err != nil {
				t.Fatal(err)
			}
			if res[0] != want[i] {
				t.Fatalf("program %d returned %v, want %d", i, res[0], want[i])
			}
		}
	}

	// a token of another type is rejected, before the program sees it
	if _, err := programs[0].in.Call(ctx, "Next", gostrict.NewToken("other.State", nil)); err == nil {
		t.Fatal("a token of another type was accepted")
	}
}

func TestPoolAndClient(t *testing.T) {
	b := compile(t, "internal/examples/physics")
	if err := b.Expect(gostrict.Func{Name: "Speed", Params: []gostrict.Type{gostrict.Float64, gostrict.Float64}, Results: []gostrict.Type{gostrict.Float64}}); err != nil {
		t.Fatal(err)
	}
	if err := b.Expect(gostrict.Func{Name: "Speed", Results: []gostrict.Type{gostrict.Int32}}); err == nil {
		t.Fatal("a wrong signature was accepted")
	}

	ctx := context.Background()
	pool, err := gostrict.NewPool(ctx, goEngine, b, nil, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	client := physics.Client{Instance: pool}
	nativeClient := physics.Client{Instance: mustLoad(t, native.Engine{Exports: physics.Exports()}, b)}

	var wg sync.WaitGroup
	for i := range 16 {
		wg.Go(func() {
			x, y, onGround, jumps, err := client.Simulate(ctx, 300, int32(i+1))
			if err != nil {
				t.Error(err)
				return
			}
			nx, ny, nOnGround, nJumps, _ := nativeClient.Simulate(ctx, 300, int32(i+1))
			if !gostrict.CloseValues([]any{x, y, onGround, jumps}, []any{nx, ny, nOnGround, nJumps}, gojaengine.Tolerance) {
				t.Errorf("%d: goja %v %v %v %v, native %v %v %v %v", i, x, y, onGround, jumps, nx, ny, nOnGround, nJumps)
			}
		})
	}
	wg.Wait()
}

func TestTimeout(t *testing.T) {
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", script("package s\n\nfunc Loop() int32 {\n\tn := int32(0)\n\tfor {\n\t\tn++\n\t}\n}\n"))
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags.Text())
	}
	in := mustLoad(t, goEngine, b)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = in.Call(ctx, "Loop")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
}

func mustLoad(t *testing.T, e gostrict.Engine, b *gostrict.Bundle) gostrict.Instance {
	t.Helper()
	in, err := e.Load(context.Background(), b, nil)
	if err != nil {
		t.Fatal(err)
	}
	return in
}
