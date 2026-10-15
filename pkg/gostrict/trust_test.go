// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
)

const evilHeader = `// Package evil binds arbitrary JavaScript.
//
//gostrict:header
package evil

//gostrict:js () => { document.cookie; return 1; }
func Steal() int32 { return 1 }
`

const evilMain = `package app

import "llm.local/app/evil"

func Run() int32 { return evil.Steal() }
`

// TestUntrustedHeader loads a program, which declares its own header package, from a module, which is not
// trusted, e.g. one which a language model has written. Its header package must be rejected.
func TestUntrustedHeader(t *testing.T) {
	fsys := fstest.MapFS{
		"main.go":      {Data: []byte(evilMain)},
		"evil/evil.go": {Data: []byte(evilHeader)},
	}

	compile := func(trusted bool) gostrict.Diagnostics {
		prog, err := gostrict.LoadModules("llm.local/app", gostrict.Module{Path: "llm.local/app", FS: fsys, Trusted: trusted})
		if err != nil {
			t.Fatal(err)
		}
		return gostrict.Check(prog)
	}

	diags := compile(false)
	found := false
	for _, d := range diags {
		if d.Code == gostrict.CodeUntrusted && d.Line == 3 && strings.HasSuffix(d.File, "evil/evil.go") {
			found = true
		}
	}
	if !found {
		t.Fatalf("the untrusted header package was not reported:\n%s", diags.Text())
	}

	if diags := compile(true); len(diags) > 0 {
		t.Fatalf("the trusted header package was reported:\n%s", diags.Text())
	}
}

// TestUntrustedShadow checks, that an untrusted module can not replace a package of the libraries.
func TestUntrustedShadow(t *testing.T) {
	fake := gostrict.Module{Path: "go.wdy.de/nago/pkg/gostrict/lib", FS: fstest.MapFS{
		"three/three.go": {Data: []byte("//gostrict:header\npackage three\n\n//gostrict:js () => 1\nfunc Evil() int32 { return 1 }\n")},
	}}
	app := gostrict.Module{Path: "llm.local/app", FS: fstest.MapFS{
		"main.go": {Data: []byte("package app\n\nimport \"go.wdy.de/nago/pkg/gostrict/lib/three\"\n\nfunc Run() int32 { return three.Evil() }\n")},
	}}
	prog, err := gostrict.LoadModules("llm.local/app", app, fake)
	if err != nil {
		t.Fatal(err)
	}
	diags := gostrict.Check(prog)
	if len(diags) == 0 || !strings.Contains(diags.Text(), "Evil") {
		t.Fatalf("the fake library was used:\n%s", diags.Text())
	}
}

// TestUntrustedLibrary checks, that an untrusted program may use the built-in libraries.
func TestUntrustedLibrary(t *testing.T) {
	app := gostrict.Module{Path: "llm.local/app", FS: fstest.MapFS{
		"main.go": {Data: []byte(`package app

import "go.wdy.de/nago/pkg/gostrict/lib/three"

var scene = three.NewScene()

func Setup() {
	three.SetBackground(scene, 0x101010)
}
`)},
	}}
	prog, err := gostrict.LoadModules("llm.local/app", app)
	if err != nil {
		t.Fatal(err)
	}
	if diags := gostrict.Check(prog); len(diags) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diags.Text())
	}
}
