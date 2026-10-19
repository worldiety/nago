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
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/std/js"
)

func compileSrc(t *testing.T, src string) *gostrict.Bundle {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags.Text())
	}
	if _, err := b.VM(); err != nil {
		t.Fatal(err)
	}
	return b
}

// TestVariadicWithoutArgs calls variadic functions of a header package without variadic arguments, which receive
// an empty list.
func TestVariadicWithoutArgs(t *testing.T) {
	b := compileSrc(t, `package s

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func View() ui.View {
	return ui.PieChart(ui.Chart{}).Series(ui.ChartSeries("A", ui.ChartSeriesTypeLine, ui.ChartDataPoint{X: "a", Y: 1}))
}

func Empty() ui.View {
	return ui.PieChart(ui.Chart{}).Series()
}
`)
	in, err := new(vm.Engine).Load(context.Background(), b, ui.Host())
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	for _, name := range []string{"View", "Empty"} {
		if _, err := in.Call(context.Background(), name); err != nil {
			t.Fatal(name, err)
		}
	}
}

// TestUnnamedVariant passes a variant of a union of handles through the program, which never names its type.
func TestUnnamedVariant(t *testing.T) {
	b := compileSrc(t, `package s

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func Wrap(v ui.View) ui.View {
	return v
}
`)
	const tag = "go.wdy.de/nago/pkg/gostrict/lib/ui.TBox"
	// the program only passes the handle on
	box := js.NewRef("box")
	in, err := new(vm.Engine).Load(context.Background(), b, ui.Host())
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	res, err := in.Call(context.Background(), "Wrap", gostrict.Variant{Tag: tag, Fields: []any{box}})
	if err != nil {
		t.Fatal(err)
	}
	if v := res[0].(gostrict.Variant); v.Tag != tag || v.Fields[0].(js.Ref).Value() != box.Value() {
		t.Fatalf("unexpected result %v", v)
	}
}

// TestHostVars loads a program, which reads a variable of a header package, with the functions of the host, which
// the bundle names.
func TestHostVars(t *testing.T) {
	b := compileSrc(t, `package s

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func View() ui.View {
	return ui.Text("x").Font(ui.TitleMedium)
}
`)
	const name = "go.wdy.de/nago/pkg/gostrict/lib/ui.TitleMedium"
	if len(b.HostVars) != 1 || b.HostVars[0].Name != name || b.HostVars[0].Type.Kind != gostrict.ValueRecord || b.HostVars[0].JS == "" {
		t.Fatalf("unexpected variables %+v", b.HostVars)
	}
	all := ui.Host()
	host := gostrict.Funcs{}
	for _, x := range b.Externs {
		host[x.Name] = all[x.Name]
	}
	for _, v := range b.HostVars {
		host[v.Name] = all[v.Name]
	}
	in, err := new(vm.Engine).Load(context.Background(), b, host)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	if _, err := in.Call(context.Background(), "View"); err != nil {
		t.Fatal(err)
	}
}
