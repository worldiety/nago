// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"testing"
	"testing/fstest"
)

func TestInferModulePath(t *testing.T) {
	fsys := fstest.MapFS{
		"scene/scene.go":             {Data: []byte("package scene\n\nimport (\n\t\"go.wdy.de/nago/pkg/gostrict/std/list\"\n\t\"example.com/app/scene/messages\"\n)\n")},
		"scene/messages/messages.go": {Data: []byte("package messages\n")},
	}
	if m, ok, err := InferModulePath(fsys); err != nil || !ok || m != "example.com/app" {
		t.Fatalf("got %q %v %v", m, ok, err)
	}

	alone := fstest.MapFS{"scene/scene.go": {Data: []byte("package scene\n\nimport \"math\"\n")}}
	if _, ok, err := InferModulePath(alone); err != nil || ok {
		t.Fatalf("a program without local imports has no module path: %v %v", ok, err)
	}

	fsys["other/other.go"] = &fstest.MapFile{Data: []byte("package other\n\nimport \"example.org/x/scene/messages\"\n")}
	if _, _, err := InferModulePath(fsys); err == nil {
		t.Fatal("contradicting imports were accepted")
	}
}
