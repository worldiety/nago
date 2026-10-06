// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package logging

import (
	"strings"
	"testing"
)

func TestSecretHidesTheID(t *testing.T) {
	id := "24f675b60edffcad29162d0be4eba80253b4982e1b43496f96c588babf16b4aa"
	got := Secret(id)

	if got == "" || strings.Contains(id, strings.TrimPrefix(got, "#")) || got != Secret(id) || got == Secret(id+"x") {
		t.Fatalf("expected a stable fingerprint which does not reveal the id, got %q", got)
	}

	if Secret("") != "" {
		t.Fatal("an empty id has no fingerprint")
	}
}
