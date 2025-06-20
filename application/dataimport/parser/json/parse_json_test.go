// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package json

import (
	"bytes"
	"context"
	_ "embed"
	"go.wdy.de/nago/application/dataimport/parser"
	"go.wdy.de/nago/pkg/xslices"
	"testing"
)

//go:embed object.json
var dataObj []byte

//go:embed array.json
var array []byte

//go:embed jsonl.txt
var jsonl []byte

func Test_parseJSON(t *testing.T) {
	p := NewParser()
	slice, err := xslices.Collect2(p.Parse(context.Background(), bytes.NewReader(dataObj), parser.Options{}))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(slice)
	if n := len(slice); n != 1 {
		t.Fatalf("expect 1 element, got %d", n)
	}

	slice, err = xslices.Collect2(p.Parse(context.Background(), bytes.NewReader(array), parser.Options{}))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(slice)
	if n := len(slice); n != 2 {
		t.Fatalf("expect 2 element, got %d", n)
	}

	slice, err = xslices.Collect2(p.Parse(context.Background(), bytes.NewReader(jsonl), parser.Options{}))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(slice)
	if n := len(slice); n != 3 {
		t.Fatalf("expect 3 element, got %d", n)
	}
}
