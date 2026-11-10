// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"fmt"
	"testing"

	"go.wdy.de/nago/pkg/gostrict/engine/vm"
)

// TestMaxBuffer checks, that a buffer of the vm has no limit, unless the host sets one, and that a limit of the host
// keeps the length within it and fails loudly at Freeze.
func TestMaxBuffer(t *testing.T) {
	b := compile(t, "../../internal/examples/binary")
	cases := []struct {
		name  string
		limit int
		n     int64
		// size is the length of the buffer, err the error of Grow
		size int64
		err  string
	}{
		{"no limit", 0, 3000, 3000, "<nil>"},
		{"within the limit", 100, 100, 100, "<nil>"},
		{"beyond the limit", 100, 101, 100, "bin: the buffer exceeds its limit of 100 bytes"},
		{"no limit for the same length", 0, 101, 101, "<nil>"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := call(t, &vm.Engine{MaxBuffer: c.limit}, b, nil, "Grow", []any{c.n})
			if res[0] != c.size || res[1] != c.size || fmt.Sprint(res[6]) != c.err {
				t.Fatalf("got %v", res)
			}
		})
	}
}
