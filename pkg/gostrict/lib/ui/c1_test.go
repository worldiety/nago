// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

// TestEnumsC1 compares the constants of the enums of batch c1 with the constants of nago.
func TestEnumsC1(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"AlertIntentError", int64(AlertIntentError), int64(alert.IntentError)},
		{"AlertIntentOk", int64(AlertIntentOk), int64(alert.IntentOk)},
		{"AlertIntentWarning", int64(AlertIntentWarning), int64(alert.IntentWarning)},
		{"AlertIntentSuccess", int64(AlertIntentSuccess), int64(alert.IntentSuccess)},
		{"AvatarCircle", int64(AvatarCircle), int64(avatar.Circle)},
		{"AvatarRounded", int64(AvatarRounded), int64(avatar.Rounded)},
		{"CardTitleLarge", int64(CardTitleLarge), int64(cardlayout.TitleLarge)},
		{"CardTitleCompact", int64(CardTitleCompact), int64(cardlayout.TitleCompact)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s = %d, want %d", e.name, e.got, e.want)
		}
	}
}

// TestIconsC1 compares the icons of the banners with the icons of nago.
func TestIconsC1(t *testing.T) {
	if AlertIconOK != alert.IconOK || AlertIconError != alert.IconError || AlertIconWarning != alert.IconWarning || AlertIconSuccess != alert.IconSuccess {
		t.Error("the icons of the banners differ from nago")
	}
}
