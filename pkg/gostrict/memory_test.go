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
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
)

// TestLoadFromMemory compiles a program with several files and packages, which exists only in memory, like one
// written by a language model in a running application.
func TestLoadFromMemory(t *testing.T) {
	fsys := fstest.MapFS{
		"main.go": {Data: []byte(`package app

import "llm.local/app/geom"

var origin = geom.Vec{X: 1, Y: 2}

func Run() float64 {
	return scaled(3).Len() + origin.X
}
`)},
		"scale.go": {Data: []byte(`package app

import "llm.local/app/geom"

func scaled(f float64) geom.Vec {
	return geom.Vec{X: 3, Y: 4}.Scale(f)
}
`)},
		"geom/geom.go": {Data: []byte(`package geom

import "math"

type Vec struct{ X, Y float64 }

func (v Vec) Scale(f float64) Vec { return Vec{v.X * f, v.Y * f} }

func (v Vec) Len() float64 { return math.Hypot(v.X, v.Y) }
`)},
		"bad/bad.go": {Data: []byte(`package bad

func Count(s string) int {
	return len(s)
}
`)},
	}

	prog, err := gostrict.LoadModules("llm.local/app", gostrict.Module{Path: "llm.local/app", FS: fsys})
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
	in, err := new(gojaengine.Engine).Load(context.Background(), b, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := in.Call(context.Background(), "Run")
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != 16.0 {
		t.Fatalf("got %v", res[0])
	}

	prog, err = gostrict.LoadModules("llm.local/app/bad", gostrict.Module{Path: "llm.local/app", FS: fsys})
	if err != nil {
		t.Fatal(err)
	}
	diags = gostrict.Check(prog)
	text := diags.Text()
	for _, want := range []string{"llm.local/app/bad/bad.go:3:22: GS006", "return len(s)", "llm.local/app/bad/bad.go:4:9: GS008"} {
		if !strings.Contains(text, want) {
			t.Errorf("diagnostics lack %q:\n%s", want, text)
		}
	}
}
