// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidataimport

import (
	"fmt"
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/dataimport"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/pager"
	"os"
	"strings"
)

func ViewPage(wnd core.Window, imp importer.Importer, stage dataimport.Staging, ucImp dataimport.UseCases, pageIdx *core.State[int], page data.Page[dataimport.Entry]) core.View {

	var statusWidth ui.Length
	var columns int
	var widths []ui.Length
	if wnd.Info().SizeClass <= core.SizeClassSmall {
		columns = 1
		statusWidth = ui.L160
	} else {
		columns = len(imp.Configuration().PreviewMappings) + 1
		for range columns - 1 {
			widths = append(widths, "1fr")
		}

		widths = append(widths, "0.1fr")
		statusWidth = ui.L320
	}

	displayUser, ok := core.SystemService[user.DisplayName](wnd.Application())
	if !ok {
		return alert.BannerError(os.ErrNotExist)
	}

	return ui.VStack(
		ui.VStack(

			ui.HStack(

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

				// status separation area
				ui.Space(ui.L16),
				ui.Grid(

					ui.GridCell(ui.Text("Status")),
					ui.GridCell(ui.Text("Datum")),
					ui.GridCell(ui.Text("Nutzer")),
				).
					Columns(3).
					Widths("1fr", "1fr", ui.L64).
					BackgroundColor(ui.ColorCardTop).
					Padding(ui.Padding{}.All(ui.L16)).
					Frame(ui.Frame{Width: statusWidth, MinWidth: statusWidth}).
					Border(ui.Border{}.Radius(ui.L16)),
			).FullWidth(),
		).
			Append(
				ui.ForEach(page.Items, func(e dataimport.Entry) core.View {
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

					lastModUser := displayUser(e.LastModBy)

					return ui.HStack(

						// the actual entry
						ui.HStack(
							ui.Grid(
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
						).
							FullWidth().
							Action(func() {
								wnd.Navigation().ForwardTo("admin/data/entry", core.Values{"importer": string(imp.Identity()), "entry": string(e.ID), "stage": string(stage.ID)})
							}).
							HoveredBackgroundColor(ui.ColorCardTop).
							BackgroundColor(ui.ColorCardBody).
							Padding(ui.Padding{}.All(ui.L16)).
							Border(ui.Border{}.Radius(ui.L16)),

						// the meta data status
						ui.Space(ui.L16),
						statusGrid(wnd, e, statusWidth, lastModUser),
					).FullWidth().Alignment(ui.Stretch)
				})...,
			).Gap(ui.L8).
			FullWidth().
			Alignment(ui.Leading),

		ui.Space(ui.L8),

		ui.HStack(
			ui.Text(fmt.Sprintf("%d-%d von %d Einträgen", page.PageIdx*page.PageSize+1, min(page.PageIdx*page.PageSize+page.PageSize, int(page.Total)), page.Total)),
			ui.Spacer(),
			pager.Pager(pageIdx).Count(page.PageCount),
		).
			FullWidth().
			BackgroundColor(ui.ColorCardFooter).
			Padding(ui.Padding{}.All(ui.L8)).
			Border(ui.Border{}.Radius(ui.L16)),
	).FullWidth()
}

func statusGrid(wnd core.Window, e dataimport.Entry, statusWidth ui.Length, lastModUser user.Compact) core.View {
	var columns int
	var widths []ui.Length
	if wnd.Info().SizeClass <= core.SizeClassSmall {
		columns = 1
		widths = []ui.Length{"1fr"}
	} else {
		columns = 3
		widths = []ui.Length{"1fr", "1fr", ui.L64}
	}

	if e.ImportedError != "" {
		return ui.HStack(
			statusView(e),
		).BackgroundColor(ui.ColorCardBody).
			Padding(ui.Padding{}.All(ui.L16)).
			Frame(ui.Frame{Width: statusWidth, MinWidth: statusWidth}).
			Border(ui.Border{}.Radius(ui.L16))
	}

	return ui.Grid(
		ui.GridCell(statusView(e)),
		ui.GridCell(ui.HStack(ui.Text(mostSignificantDate(e)))),
		ui.GridCell(avatar.TextOrImage(lastModUser.Displayname, lastModUser.Avatar)),
	).
		Columns(columns).
		Gap(ui.L8).
		Widths(widths...).
		BackgroundColor(ui.ColorCardBody).
		Padding(ui.Padding{}.All(ui.L16)).
		Frame(ui.Frame{Width: statusWidth, MinWidth: statusWidth}).
		Border(ui.Border{}.Radius(ui.L16))
}

func mostSignificantDate(e dataimport.Entry) string {
	const format = "02.01.2006 15:04"
	if e.Imported && !e.ImportedAt.IsZero() {
		return e.ImportedAt.Format(format)
	}

	if e.LastModAt.IsZero() {
		return ""
	}

	return e.LastModAt.Format(format)
}

func statusView(e dataimport.Entry) core.View {
	var statusText string
	var statusTextColor ui.Color
	var ico core.SVG
	switch {
	case e.Imported:
		statusText = "Importiert"
		statusTextColor = ui.ColorSemanticGood
		ico = flowbiteSolid.FloppyDisk
	case e.Confirmed:
		statusText = "Bestätigt"
		ico = flowbiteSolid.CheckCircle
	case e.Ignored:
		statusText = "Abgelehnt"
		statusTextColor = ui.ColorSemanticWarn
		ico = flowbiteSolid.CloseCircle
	case e.ImportedError != "":
		statusText = "Fehler: " + e.ImportedError
		statusTextColor = ui.ColorError
		ico = flowbiteOutline.ExclamationCircle
	default:
		ico = flowbiteOutline.CheckCircle
		statusText = "Ungeprüft"
	}

	return ui.HStack(ui.ImageIcon(ico), ui.Text(statusText)).TextColor(statusTextColor).Gap(ui.L4)
}
