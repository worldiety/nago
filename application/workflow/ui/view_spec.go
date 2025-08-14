// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiworkflow

import (
	"fmt"

	"go.wdy.de/nago/application/workflow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
)

func specPage(wnd core.Window, uc workflow.UseCases, id workflow.ID) core.View {
	specSvg := core.AutoState[core.SVG](wnd)

	opts := core.AutoState[workflow.RenderOptions](wnd).Init(func() workflow.RenderOptions {
		return workflow.RenderOptions{
			Language: wnd.Locale(),
		}
	})
	showOpts := core.AutoState[bool](wnd).Observe(func(newValue bool) {
		specSvg.Set(nil)
	})

	var svg core.SVG
	if len(specSvg.Get()) != 0 {
		svg = specSvg.Get()
	} else {
		tmp, err := uc.Render(wnd.Subject(), id, opts.Get())
		if err != nil {
			return alert.BannerError(err)
		}

		specSvg.Set(tmp)
		svg = tmp

	}

	return ui.VStack(
		alert.Dialog("Einstellungen", form.Auto[workflow.RenderOptions](form.AutoOptions{}, opts), showOpts, alert.Closeable(), alert.Close(func() (close bool) {
			return true
		})),
		ui.HStack(
			ui.SecondaryButton(func() {
				showOpts.Set(true)
				showOpts.Notify()
			}).Title("Ansicht"),
			ui.SecondaryButton(func() {
				wnd.ExportFiles(core.ExportFilesOptions{
					Files: []core.File{
						core.MemFile{
							Filename:     fmt.Sprintf("workflow-%s.svg", id),
							MimeTypeHint: "image/svg+xml",
							Bytes:        specSvg.Get(),
						},
					},
				})
			}).Title("Download SVG"),
		).Gap(ui.L8),
		ui.Image().
			Embed(svg).
			Frame(ui.Frame{}.FullWidth()),
	).BackgroundColor(ui.ColorWhite).
		Alignment(ui.Trailing).
		Padding(ui.Padding{}.All(ui.L16)).
		Border(ui.Border{}.Radius(ui.L16))
}
