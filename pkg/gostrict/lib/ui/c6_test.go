// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	"go.wdy.de/nago/presentation/ui/editor"
	"go.wdy.de/nago/presentation/ui/navsplitview"
)

// TestEnumsC6 compares the constants of the enums of batch c6 with the constants of nago.
func TestEnumsC6(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"NavsplitviewKindSidebar", int64(NavsplitviewKindSidebar), int64(navsplitview.KindSidebar)},
		{"NavsplitviewKindContent", int64(NavsplitviewKindContent), int64(navsplitview.KindContent)},
		{"NavsplitviewKindDetail", int64(NavsplitviewKindDetail), int64(navsplitview.KindDetail)},
		{"EditorContentFull", int64(EditorContentFull), int64(editor.ContentFull)},
		{"EditorContentPage", int64(EditorContentPage), int64(editor.ContentPage)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s = %d, want %d", e.name, e.got, e.want)
		}
	}
}
