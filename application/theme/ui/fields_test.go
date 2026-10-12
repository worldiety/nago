// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uitheme

import (
	"testing"

	"go.wdy.de/nago/application/theme"
)

// Every color of ui.Colors must be editable, either as a base color or in the advanced mode.
func TestEveryColorIsEditable(t *testing.T) {
	editable := map[string]int{"M0": 1, "I0": 1, "A0": 1}
	for _, grp := range colorGroups {
		for _, f := range grp.fields {
			editable[f.key]++
		}
	}

	for _, key := range theme.ColorKeys() {
		if editable[key] != 1 {
			t.Errorf("color %s is editable %d times", key, editable[key])
		}
		delete(editable, key)
	}

	for key := range editable {
		t.Errorf("unknown color %s", key)
	}
}
