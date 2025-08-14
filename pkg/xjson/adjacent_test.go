// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xjson

import (
	"encoding/json"
	"testing"
)

func TestNewAdjacentEnvelope(t *testing.T) {
	type Cat struct {
		Name string
	}

	type Dog struct {
		Name string
	}

	buf, err := json.Marshal(NewAdjacentEnvelope(Cat{Name: "Minka"}))
	if err != nil {
		t.Fatal(err)
	}

	var tmp AdjacentEnvelope
	if err := json.Unmarshal(buf, &tmp); err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", tmp)

	if tmp.Type != "go.wdy.de/nago/pkg/xjson.Cat" {
		t.Fatalf("expecting type got %v", tmp.Type)
	}

	if tmp.Value.(Cat).Name != "Minka" {
		t.Fatalf("expecting Minka got %v", tmp.Value)
	}
}
