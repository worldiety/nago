// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"context"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/errs"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/ints"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/money"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/physics"
)

var exports = map[string]gostrict.Funcs{
	"money":   native.Reflect(map[string]any{"Calc": money.Calc, "Factorial": money.Factorial, "Ints": money.Ints, "Invoke": money.Invoice, "Invoice": money.Invoice, "Limits": money.Limits, "Parse": money.Parse, "Round": money.Round}),
	"errs":    native.Reflect(map[string]any{"Average": errs.Average, "Run": errs.Run}),
	"ints":    native.Reflect(map[string]any{"Bits": ints.Bits, "Conv": ints.Conv, "Div": ints.Div, "DivU": ints.DivU, "Hash": ints.Hash, "Wrap": ints.Wrap, "Xorshift": ints.Xorshift}),
	"physics": native.Reflect(map[string]any{"Simulate": physics.Simulate, "Speed": physics.Speed}),
}

var bundles = map[string]*gostrict.Bundle{}

func compile(t testing.TB, dir string) *gostrict.Bundle {
	t.Helper()
	if b, ok := bundles[dir]; ok {
		return b
	}
	b, diags, err := gostrict.Compile(dir)
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags.Text())
	}
	bundles[dir] = b
	return b
}

func call(t testing.TB, e gostrict.Engine, b *gostrict.Bundle, host gostrict.Funcs, name string, args []any) []any {
	t.Helper()
	in, err := e.Load(context.Background(), b, host)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	res, err := in.Call(context.Background(), name, args...)
	if err != nil {
		t.Fatal(err)
	}
	return res
}
