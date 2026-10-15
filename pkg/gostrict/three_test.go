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
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/spin"
	"go.wdy.de/nago/pkg/gostrict/lib/three"
)

// TestThree runs a three.js program natively and in goja with the native scene graph as host, which receives the
// handles back through goja, and compares the scenes.
func TestThree(t *testing.T) {
	b := compile(t, "internal/examples/spin")
	run := func(e gostrict.Engine, host gostrict.Funcs) string {
		in, err := e.Load(context.Background(), b, host)
		if err != nil {
			t.Fatal(err)
		}
		client := spin.Client{Instance: in}
		n, err := client.Setup(context.Background(), 16.0/9)
		if err != nil {
			t.Fatal(err)
		}
		if n != 13 {
			t.Fatalf("got %d meshes", n)
		}
		if err := client.Frame(context.Background(), 1.5); err != nil {
			t.Fatal(err)
		}
		return three.Dump()
	}

	want := run(native.Engine{Exports: spin.Exports()}, nil)
	got := run(goEngine, three.Host())
	if got != want {
		t.Fatalf("the scenes differ\nnative:\n%s\ngoja:\n%s", want, got)
	}
	if !strings.Contains(want, "Mesh pos=[") || strings.Count(want, "\n") != 16 {
		t.Fatalf("unexpected scene:\n%s", want)
	}
	t.Log("\n" + want)
}
