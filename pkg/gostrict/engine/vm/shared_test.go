// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"context"
	"sync"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
)

// TestSharedVariables loads many instances of a program, whose package variables the first instance initializes
// once, also with a variable of a header package, and calls them concurrently.
func TestSharedVariables(t *testing.T) {
	src := `package s

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

type Style struct {
	A, B, C, D, E, F, G, H, I string
}

var (
	squares = list.Freeze(table(100))
	style   = Style{A: "a", I: "i"}
	font    = ui.BodyLarge
)

func table(n int32) list.List[int32] {
	l := list.New[int32]()
	for i := range n {
		l.Append(i * i)
	}
	return l
}

func Sum(n int32) string {
	s := int32(0)
	for i, v := range squares.All() {
		if i >= n {
			break
		}
		s += v
	}
	return text.FromInt(s) + style.A + style.I + string(font.Size)
}
`
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags.Text())
	}
	var mu sync.Mutex
	reads := 0
	host := gostrict.Funcs{}
	for name, f := range ui.Host() {
		host[name] = f
	}
	const fontName = "go.wdy.de/nago/pkg/gostrict/lib/ui.BodyLarge"
	host[fontName] = func(args []any) []any {
		mu.Lock()
		reads++
		mu.Unlock()
		return ui.Host()[fontName](args)
	}

	e := new(vm.Engine)
	var wg sync.WaitGroup
	results := make([]string, 32)
	for i := range results {
		wg.Go(func() {
			in, err := e.Load(context.Background(), b, host)
			if err != nil {
				t.Error(err)
				return
			}
			for range 20 {
				res, err := in.Call(context.Background(), "Sum", int32(10))
				if err != nil {
					t.Error(err)
					return
				}
				results[i] = res[0].(string)
			}
		})
	}
	wg.Wait()
	for _, r := range results {
		if r != "285ai1rem" {
			t.Fatalf("unexpected results %v", results)
		}
	}
	if reads != 1 {
		t.Fatalf("the variable of the header package was read %d times", reads)
	}
	t.Log(results[0])
}
