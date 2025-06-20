// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package pdf

import (
	_ "embed"
	"fmt"
	"testing"
)

//go:embed example.pdf
var pdf []byte

func Test_parsePDF(t *testing.T) {
	obj := parsePDF(pdf)
	fmt.Printf("%#v\n", obj)
	if v, _ := obj.Get("Address 1 Text Box"); v.String() != "Nordseestraße" {
		t.Fatalf("unexpected value: %v", v)
	}

	if v, _ := obj.Get("Given Name Text Box"); v.String() != "Torben Äöü" {
		t.Fatalf("unexpected value: %v", v)
	}
}
