// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui_test

import (
	"context"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
)

func compileView(t *testing.T, src string) *gostrict.Bundle {
	t.Helper()
	prog, err := gostrict.LoadModules("llm.local/v", gostrict.Module{Path: "llm.local/v", FS: fstest.MapFS{"v.go": {Data: []byte(src)}}})
	if err != nil {
		t.Fatal(err)
	}
	if diags := gostrict.Check(prog); len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	b, diags := gostrict.Emit(prog)
	if len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	return b
}

// TestBundleBindsOnlyUsedFunctions checks, that a view binds only the functions of lib/ui, which it calls, and not
// the methods of every variant of ui.View, which the program never calls through the interface.
func TestBundleBindsOnlyUsedFunctions(t *testing.T) {
	b := compileView(t, `package v

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func Render() ui.View {
	return ui.VStack(ui.Text("hello").Font(ui.Title)).Gap(ui.L8)
}
`)
	if len(b.Externs) > 10 {
		var names []string
		for _, x := range b.Externs {
			names = append(names, x.Name)
		}
		t.Fatalf("the bundle binds %d functions: %v", len(b.Externs), names)
	}
}

// TestChartWithoutLabels calls a chart without its variadic labels.
func TestChartWithoutLabels(t *testing.T) {
	b := compileView(t, `package v

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func Render() ui.View {
	return ui.VStack(ui.LineChart(ui.Chart{}), ui.BarChart(ui.Chart{}, "Jan"), ui.PieChart(ui.Chart{}))
}
`)
	in, err := new(gojaengine.Engine).Load(context.Background(), b, ui.Host())
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	if _, err := in.Call(context.Background(), "Render"); err != nil {
		t.Fatal(err)
	}
}
