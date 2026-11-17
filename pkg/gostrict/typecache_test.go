// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/token"
	"go/types"
	"runtime"
	"testing"
	"time"
)

// The cache computes a result once per type and forgets it, once the type is no longer used, e.g. after a program has
// been translated, so that a long running host does not keep the types of all its former programs.
func TestTypeCache(t *testing.T) {
	var c typeCache[[]string]
	computed := 0
	compute := func() []string {
		computed++
		return []string{"A", "B"}
	}

	pkg := types.NewPackage("example.com/p", "p")
	n := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "T", nil), types.Typ[types.Int32], nil)
	if got := c.get(n, compute); len(got) != 2 || computed != 1 {
		t.Fatalf("first get = %v, computed %d times", got, computed)
	}
	if got := c.get(n, compute); len(got) != 2 || computed != 1 {
		t.Fatalf("second get = %v, computed %d times", got, computed)
	}
	runtime.KeepAlive(n)

	n = nil
	deadline := time.Now().Add(5 * time.Second)
	for {
		runtime.GC()
		empty := true
		c.m.Range(func(any, any) bool {
			empty = false
			return false
		})
		if empty {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("the entry of a type, which is no longer used, is kept")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
