// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package csv

import (
	_ "embed"
	"fmt"
	"testing"
)

//go:embed example.csv
var csvf []byte

func Test_parseCSV(t *testing.T) {
	array, err := parseCSV(csvf)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%#v\n", array)
	/*	if v := obj["Address 1 Text Box"]; v != "Nordseestraße" {
			t.Fatalf("unexpected value: %v", v)
		}

		if v := obj["Given Name Text Box"]; v != "Torben Äöü" {
			t.Fatalf("unexpected value: %v", v)
		}*/
}
