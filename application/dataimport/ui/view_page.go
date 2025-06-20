// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidataimport

import (
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/dataimport"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/pager"
	"strings"
)

func ViewPage(wnd core.Window, imp importer.Importer, stage dataimport.Staging, ucImp dataimport.UseCases, pageIdx *core.State[int], page dataimport.FilterEntriesPage) core.View {

	var columns int
	var widths []ui.Length
	if wnd.Info().SizeClass <= core.SizeClassSmall {
		columns = 1
	} else {
		columns = len(imp.Configuration().PreviewMappings) + 1
		for range columns - 1 {
			widths = append(widths, "1fr")
		}

		widths = append(widths, "0.1fr")
	}

	return ui.VStack(
		ui.VStack(
			ui.Grid(
				ui.ForEach(imp.Configuration().PreviewMappings, func(t importer.PreviewMapping) ui.TGridCell {
					return ui.GridCell(ui.Text(t.Name))
				})...,
			).Append(ui.GridCell(ui.Text(" "))).
				FullWidth().
				Columns(columns).
				Widths(widths...).
				BackgroundColor(ui.ColorCardTop).
				Padding(ui.Padding{}.All(ui.L16)).
				Border(ui.Border{}.Radius(ui.L16)),
		).
			Append(
				ui.ForEach(page.Entries, func(e dataimport.Entry) core.View {
					var values []string

				NextMapping:
					for _, mapping := range imp.Configuration().PreviewMappings {
						obj := e.Transform(stage.Transformation)

						for _, keyword := range mapping.Keywords {
							if strings.HasPrefix(keyword, "/") {
								val, err := jsonptr.Eval(obj, keyword)
								if err != nil {
									//slog.Error("failed to evaluate json pointer from importer preview mapping", "err", err.Error(), "keyword", keyword)
									continue
								}

								if v := val.String(); !val.Bool() && v != "" {
									values = append(values, v)
									continue NextMapping
								}
							} else {
								for k, v := range obj.All() {
									if strings.Contains(strings.ToLower(k), keyword) {
										if val := v.String(); !v.Bool() && val != "" {
											values = append(values, val)
											continue NextMapping
										}
									}
								}
							}

						}

						values = append(values, " ")
						continue NextMapping
					}

					return ui.VStack(ui.Grid(
						ui.ForEach(values, func(t string) ui.TGridCell {
							return ui.GridCell(ui.VStack(ui.Text(t)).Alignment(ui.Leading))
						})...,
					).Append(
						ui.GridCell(ui.HStack(
							ui.ImageIcon(flowbiteOutline.ChevronRight),
						).Alignment(ui.Trailing)),
					).
						FullWidth().
						Columns(columns).
						Widths(widths...),
					).FullWidth().
						Action(func() {
							wnd.Navigation().ForwardTo("admin/data/entry", core.Values{"importer": string(imp.Identity()), "entry": string(e.ID), "stage": string(stage.ID)})
						}).
						HoveredBackgroundColor(ui.ColorCardTop).
						BackgroundColor(ui.ColorCardBody).
						Padding(ui.Padding{}.All(ui.L16)).
						Border(ui.Border{}.Radius(ui.L16))
				})...,
			).Gap(ui.L8).FullWidth().Alignment(ui.Leading),

		ui.Space(ui.L8),

		ui.HStack(pager.Pager(pageIdx).Count(page.PageCount)).
			FullWidth().
			BackgroundColor(ui.ColorCardFooter).
			Padding(ui.Padding{}.All(ui.L8)).
			Border(ui.Border{}.Radius(ui.L16)),
	).FullWidth()
}
