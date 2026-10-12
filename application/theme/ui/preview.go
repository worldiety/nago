// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uitheme

import (
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// preview sketches an application with the given colors. It uses the colors as absolute values instead of the
// theme tokens, so that it shows unsaved colors and the other mode, without affecting the running application.
func preview(wnd core.Window, c ui.Colors) core.View {
	pill := func(text string, fg, bg, border ui.Color) core.View {
		return ui.Text(text).
			Color(fg).
			Font(ui.Font{Size: ui.L14, Weight: ui.DisplayAndLabelFontWeight}).
			BackgroundColor(bg).
			Border(ui.Border{}.Radius(ui.L16).Width(ui.L2).Color(border)).
			Padding(ui.Padding{}.Horizontal(ui.L16).Vertical(ui.L4))
	}

	dot := func(color ui.Color) core.View {
		return ui.VStack().BackgroundColor(color).Border(ui.Border{}.Circle()).Frame(ui.Frame{}.Size(ui.L16, ui.L16))
	}

	banner := func(text string, fg, bg ui.Color) core.View {
		return ui.Text(text).
			Color(fg).
			Font(ui.Font{Size: ui.L14}).
			BackgroundColor(bg).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L8)).
			Frame(ui.Frame{}.FullWidth())
	}

	card := ui.VStack(
		ui.HStack(ui.Text(StrPreviewCard.Get(wnd)).Color(c.M8).Font(ui.Font{Weight: ui.DisplayAndLabelFontWeight})).
			BackgroundColor(c.M9).
			Alignment(ui.Leading).
			FullWidth().
			Padding(ui.Padding{}.Horizontal(ui.L16).Vertical(ui.L8)),
		ui.VStack(
			ui.Text(StrPreviewText.Get(wnd)).Color(c.M8),
			ui.Text(StrPreviewMuted.Get(wnd)).Color(c.M7).Font(ui.Font{Size: ui.L14}),
			ui.HStack(
				ui.HStack().BackgroundColor(c.A1).Frame(ui.Frame{Width: "60%", Height: ui.L4}),
			).BackgroundColor(c.M5).Alignment(ui.Leading).Frame(ui.Frame{Height: ui.L4}.FullWidth()),
		).BackgroundColor(c.M4).
			Alignment(ui.Leading).
			Gap(ui.L8).
			FullWidth().
			Padding(ui.Padding{}.All(ui.L16)),
		ui.HStack(
			pill(rstring.ActionCancel.Get(wnd), c.I0, "", c.I0),
			pill(rstring.ActionSave.Get(wnd), c.PrimaryButtonText, c.I0, c.I0),
		).BackgroundColor(c.M3).
			Alignment(ui.Trailing).
			Gap(ui.L8).
			FullWidth().
			Padding(ui.Padding{}.Horizontal(ui.L16).Vertical(ui.L8)),
	).FullWidth().
		Border(ui.Border{}.Radius(ui.L12))

	return ui.VStack(
		ui.HStack(
			ui.Text("App").Color(c.M8).Font(ui.Font{Weight: ui.DisplayAndLabelFontWeight}),
			ui.Spacer(),
			dot(c.A0),
		).BackgroundColor(c.M2).
			FullWidth().
			Padding(ui.Padding{}.Horizontal(ui.L16).Vertical(ui.L8)),
		ui.VStack(
			card,
			ui.HStack(dot(c.Error), dot(c.Warning), dot(c.Good), dot(c.Informative), dot(c.Disabled)).Gap(ui.L8),
			banner("Error", c.BannerErrorText, c.BannerErrorBackground),
			banner("Info", c.BannerInfoText, c.BannerInfoBackground),
		).BackgroundColor(c.M1).
			Alignment(ui.Leading).
			Gap(ui.L12).
			FullWidth().
			Padding(ui.Padding{}.All(ui.L16)),
	).Border(ui.Border{}.Radius(ui.L16).Width(ui.L1).Color(c.M5)).
		Frame(ui.Frame{MaxWidth: ui.L480}.FullWidth())
}
