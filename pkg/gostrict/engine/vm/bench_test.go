// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"context"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
)

func engines(pkg string) map[string]gostrict.Engine {
	return map[string]gostrict.Engine{
		"native": native.Engine{Exports: exports[pkg]},
		"goja":   &gojaengine.Engine{},
		"vm":     &vm.Engine{},
	}
}

// BenchmarkCompute runs examples, which compute much and cross the boundary rarely.
func BenchmarkCompute(b *testing.B) {
	for _, bc := range []struct {
		pkg, name string
		args      []any
	}{
		{"physics", "Simulate", []any{int32(10000), int32(45)}},
		{"ints", "Xorshift", []any{uint32(2463534242), int32(100000)}},
		{"ints", "Hash", []any{strings.Repeat("Grüße 😀", 1000)}},
		{"money", "Factorial", []any{int32(300)}},
		{"money", "Invoice", []any{[]any{"19.99", "0.333", "1234567890.12", "7", "0.01"}, "19"}},
	} {
		for _, name := range []string{"native", "goja", "vm"} {
			b.Run(bc.pkg+"."+bc.name+"/"+name, func(b *testing.B) {
				in, err := engines(bc.pkg)[name].Load(context.Background(), compile(b, "../../internal/examples/"+bc.pkg), nil)
				if err != nil {
					b.Fatal(err)
				}
				b.ReportAllocs()
				for b.Loop() {
					if _, err := in.Call(context.Background(), bc.name, bc.args...); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
