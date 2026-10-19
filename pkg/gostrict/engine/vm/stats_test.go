// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build vmstats

package vm_test

import (
	"context"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/gostrict/internal/vmcode"
)

// TestStats prints the executed instructions of the compute benchmarks: go test -tags vmstats -run TestStats -v.
func TestStats(t *testing.T) {
	for _, bc := range []struct {
		pkg, name string
		args      []any
	}{
		{"physics", "Simulate", []any{int32(2000), int32(45)}},
		{"ints", "Xorshift", []any{uint32(2463534242), int32(10000)}},
		{"ints", "Hash", []any{strings.Repeat("Grüße 😀", 100)}},
		{"money", "Invoice", []any{[]any{"19.99", "0.333", "1234567890.12", "7", "0.01"}, "19"}},
	} {
		in, err := engines(bc.pkg)["vm"].Load(context.Background(), compile(t, "../../internal/examples/"+bc.pkg), nil)
		if err != nil {
			t.Fatal(err)
		}
		vmcode.Stats()
		if _, err := in.Call(context.Background(), bc.name, bc.args...); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s.%s\n%s", bc.pkg, bc.name, vmcode.Stats())
	}
}
