// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
)

var wantRe = regexp.MustCompile(`// want ((?:GS\d{3} ?)+)`)

// TestDiagnostics compares the diagnostics with the // want comments of the files in testdata/bad.
func TestDiagnostics(t *testing.T) {
	dirs, err := filepath.Glob("testdata/bad/*")
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			want := map[string]bool{}
			files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
			more, _ := filepath.Glob(filepath.Join(dir, "*", "*.go"))
			files = append(files, more...)
			for _, f := range files {
				abs, _ := filepath.Abs(f)
				buf, err := os.ReadFile(f)
				if err != nil {
					t.Fatal(err)
				}
				for i, line := range strings.Split(string(buf), "\n") {
					if m := wantRe.FindStringSubmatch(line); m != nil {
						for _, code := range strings.Fields(m[1]) {
							want[fmt.Sprintf("%s:%d %s", abs, i+1, code)] = true
						}
					}
				}
			}

			prog, err := gostrict.Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			diags := gostrict.Check(prog)

			got := map[string]bool{}
			for _, d := range diags {
				got[fmt.Sprintf("%s:%d %s", d.File, d.Line, d.Code)] = true
			}

			var missing, unexpected []string
			for k := range want {
				if !got[k] {
					missing = append(missing, k)
				}
			}
			for k := range got {
				if !want[k] {
					unexpected = append(unexpected, k)
				}
			}
			slices.Sort(missing)
			slices.Sort(unexpected)
			if len(missing) > 0 || len(unexpected) > 0 {
				t.Fatalf("missing:\n  %s\nunexpected:\n  %s\n\n%s", strings.Join(missing, "\n  "), strings.Join(unexpected, "\n  "), diags.Text())
			}
		})
	}
}
