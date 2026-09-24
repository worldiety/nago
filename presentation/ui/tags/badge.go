// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tags

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// StatusBadge renders a small label with the theme text color on the neutral background, so that the
// contrast is sufficient in light and dark mode. The semantic color (e.g. [ui.SG0], [ui.SE0]) is only used
// for a leading dot. If dot is empty, the badge is neutral without dot.
//
// Prefer this over [ColoredTextPill] for status information, because the latter always renders black text
// on the given color, which has a poor contrast for dark colors like [ui.SV0] or [ui.ST0].
func StatusBadge(dot ui.Color, text string) core.View {
	return ui.HStack(
		ui.IfFunc(dot != "", func() core.View {
			return ui.VStack().
				BackgroundColor(dot).
				Border(ui.Border{}.Circle()).
				Frame(ui.Frame{}.Size(ui.L8, ui.L8))
		}),
		ui.Text(text).Font(ui.BodySmall).Color(ui.M8),
	).Gap(ui.L4).
		BackgroundColor(ui.M1).
		Border(ui.Border{}.Radius(ui.L16).Color(ui.M5).Width(ui.L1)).
		Padding(ui.Padding{}.Horizontal(ui.L8).Vertical(ui.L2))
}
