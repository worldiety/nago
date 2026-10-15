// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

//go:generate go test -run TestGenerated -update .

// TestGenerated generates the functions of the header packages with -update and checks, that they are up to date.
func TestGenerated(t *testing.T) {
	dirs, err := filepath.Glob("internal/examples/*")
	if err != nil {
		t.Fatal(err)
	}
	libs, _ := filepath.Glob("lib/*")
	dirs = append(dirs, libs...)
	// std/big is natively implemented as well, so that its bindings are checked against math/big
	dirs = append(dirs, "std/big")
	for _, dir := range dirs {
		prog, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		if diags := Check(prog); len(diags) > 0 {
			t.Fatal(diags.Text())
		}
		if prog.Main.Kind != KindHeader {
			// a program needs no generated code
			if _, err := os.Stat(filepath.Join(dir, GeneratedFile)); err == nil {
				t.Errorf("%s is no header package, but has a generated file", dir)
			}
			continue
		}
		src, err := Generate(prog)
		if err != nil {
			t.Fatal(err)
		}

		file := filepath.Join(dir, GeneratedFile)
		if *update {
			if err := os.WriteFile(file, src, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		old, err := os.ReadFile(file)
		if err != nil || !bytes.Equal(old, src) {
			t.Errorf("%s is out of date, run go generate in pkg/gostrict", file)
		}
	}
}
