// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"context"
	"runtime"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
)

func heap() uint64 {
	runtime.GC()
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.HeapAlloc
}

// TestMemory measures the heap of loaded instances of the same bundle: the first, which compiles the program, and
// each further one.
func TestMemory(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	const n = 200
	for _, c := range []struct {
		dir, fn string
		args    []any
	}{{"../../lib/ui/internal/cells", "View", nil}, {"../../internal/examples/money", "Invoice", []any{[]any{"1.5", "2"}, "19"}}} {
		dir := c.dir
		b := compile(t, dir)
		for _, name := range []string{"goja", "vm"} {
			var e gostrict.Engine = &gojaengine.Engine{}
			if name == "vm" {
				e = &vm.Engine{}
			}
			before := heap()
			first, err := e.Load(context.Background(), b, ui.Host())
			if err != nil {
				t.Fatal(err)
			}
			afterFirst := heap()
			ins := make([]gostrict.Instance, n)
			for i := range ins {
				if ins[i], err = e.Load(context.Background(), b, ui.Host()); err != nil {
					t.Fatal(err)
				}
				// a call allocates the stacks
				if _, err := ins[i].Call(context.Background(), c.fn, c.args...); err != nil {
					t.Fatal(err)
				}
			}
			after := heap()
			t.Logf("%-6s %-40s first %7.1f KiB, each further %7.1f KiB", name, dir, float64(afterFirst-before)/1024, float64(after-afterFirst)/n/1024)
			runtime.KeepAlive(first)
			runtime.KeepAlive(ins)
		}
	}
}
