// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"slices"

	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

type TChatUploads struct {
	files *core.State[[]file.File]
}

func ChatUploads(files *core.State[[]file.File]) TChatUploads {
	return TChatUploads{files: files}
}

func (c TChatUploads) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	return ui.HStack(
		ui.ForEach(c.files.Get(), func(f file.File) core.View {
			return c.filePill(wnd, f)
		})...,
	).
		Wrap(true).
		Alignment(ui.BottomLeading).
		Gap(ui.L8).
		FullWidth().
		Render(ctx)
}

func (c TChatUploads) filePill(wnd core.Window, f file.File) core.View {
	return ui.HStack(
		ui.Text(xstrings.EllipsisEnd(f.Name, 30)),
		ui.Spacer(),
		ui.VLine().Frame(ui.Frame{Height: "2.5rem"}).Padding(ui.Padding{}),
		ui.TertiaryButton(func() {
			c.files.Set(slices.DeleteFunc(c.files.Get(), func(other file.File) bool {
				return f.ID == other.ID
			}))
			c.files.Invalidate()

		}).PreIcon(icons.Close).AccessibilityLabel(rstring.ActionDelete.Get(wnd)),
	).BackgroundColor(ui.M2).
		Gap(ui.L8).
		Border(ui.Border{}.Width(ui.L1).Color(ui.M5).Radius(ui.L24)).
		Frame(ui.Frame{MinWidth: "20.5rem", MaxWidth: "20.5rem"}).
		Padding(ui.Padding{}.All(ui.L8)).
		AccessibilityLabel(f.Name)
}
