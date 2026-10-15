// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	"go.wdy.de/nago/presentation/ui/document"
)

func TestC5Enums(t *testing.T) {
	if int(DocumentTopTrailing) != int(document.TopTrailing) {
		t.Error("DocumentTopTrailing")
	}
	if int(DocumentInlineTrailing) != int(document.InlineTrailing) {
		t.Error("DocumentInlineTrailing")
	}
	if int(DocumentClickable) != int(document.Clickable) {
		t.Error("DocumentClickable")
	}
}
