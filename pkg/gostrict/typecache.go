// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/types"
	"runtime"
	"sync"
	"weak"
)

// typeCache memoizes a result, which only depends on a named type, e.g. the variants of a sealed interface. Finding
// it searches the whole scope of the package of the type, which is expensive for a big package, whose signatures
// and switches name the same type again and again, e.g. ui.View of lib/ui, so that the search would grow with the
// uses times the size of the package. go/types creates a named type once per type check, so the type identifies the
// result.
//
// The keys are weak and the values must not refer to a type, e.g. names and constant values only: then the entries of
// a program, which is no longer used, are collected together with its types.
type typeCache[V any] struct {
	m sync.Map // weak.Pointer[types.Named] -> V
}

// get returns the result of n, which compute returns the first time.
func (c *typeCache[V]) get(n *types.Named, compute func() V) V {
	key := weak.Make(n)
	if v, ok := c.m.Load(key); ok {
		return v.(V)
	}
	v := compute()
	if _, loaded := c.m.LoadOrStore(key, v); !loaded {
		runtime.AddCleanup(n, func(key weak.Pointer[types.Named]) { c.m.Delete(key) }, key)
	}
	return v
}
