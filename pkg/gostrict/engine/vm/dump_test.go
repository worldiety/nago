// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"os"
	"testing"
)

// TestDump prints the bytecode of the program in $VMDUMP, e.g. VMDUMP=../../internal/examples/physics.
func TestDump(t *testing.T) {
	dir := os.Getenv("VMDUMP")
	if dir == "" {
		t.Skip()
	}
	p, err := compile(t, dir).VM()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("\n" + p.Dump())
}
