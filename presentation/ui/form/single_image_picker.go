// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"fmt"
	"go.wdy.de/nago/application/image"
	http_image "go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"io"
)

func SingleImagePicker(wnd core.Window, setCreator image.CreateSrcSet, loadSrcSet image.LoadSrcSet, loadBestFit image.LoadBestFit, selfId string, id image.ID, state *core.State[image.ID]) ui.DecoredView {
	if setCreator == nil {
		fn, ok := core.SystemService[image.CreateSrcSet](wnd.Application())
		if !ok {
			panic("image.CreateSrcSet not available") // TODO or better an alert.Banner?
		}

		setCreator = fn
	}

	if loadSrcSet == nil {
		fn, ok := core.SystemService[image.LoadSrcSet](wnd.Application())
		if !ok {
			panic("image.LoadSrcSet not available") // TODO or better an alert.Banner?
		}

		loadSrcSet = fn
	}

	if loadBestFit == nil {
		fn, ok := core.SystemService[image.LoadBestFit](wnd.Application())
		if !ok {
			panic("image.LoadSrcSet not available") // TODO or better an alert.Banner?
		}

		loadBestFit = fn
	}

	// empty id case
	if id == "" {
		return ui.HStack(
			ui.SecondaryButton(func() {
				wndImportFiles(wnd, setCreator, selfId, state)
			}).PreIcon(heroOutline.Plus).Title("Bild hinzufügen"),
		).Alignment(ui.Trailing)
	}

	// the preview case
	targetWidth := http_image.EstimateWidth(wnd)
	uri := core.URI(http_image.NewURL(http_image.Endpoint, id, image.FitCover, targetWidth, targetWidth))

	return ui.Box(ui.BoxLayout{
		TopTrailing: ui.HStack(
			ui.TertiaryButton(func() {
				optSet, err := loadSrcSet(wnd.Subject(), id)
				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				if optSet.IsNone() {
					alert.ShowBannerMessage(wnd, alert.Message{Title: "Bild SrcSet nicht gefunden", Message: "Das Bild kann nicht heruntergeladen werden, da es nicht gefunden wurde."})
					return
				}

				srcSet := optSet.Unwrap()

				rf := core.NewReaderFile(func() (io.ReadCloser, error) {
					optReader, err := loadBestFit(wnd.Subject(), id, image.FitCover, 0, 0)
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return nil, err
					}

					if optReader.IsNone() {
						alert.ShowBannerMessage(wnd, alert.Message{Title: "Bild nicht gefunden", Message: "Das Bild kann nicht heruntergeladen werden, da es nicht gefunden wurde."})
						return nil, fmt.Errorf("bind image one not found")
					}

					return optReader.Unwrap(), nil
				})

				rf.SetName(srcSet.Name)
				rf.SetMimeType("image/*")
				wnd.ExportFiles(core.ExportFilesOptions{
					ID:    string(id) + "-download",
					Files: []core.File{rf},
				})
			}).PreIcon(heroOutline.ArrowDownTray).
				AccessibilityLabel("Bild herunterladen"),
			ui.TertiaryButton(func() {
				state.Set("")
				state.Notify()
			}).PreIcon(heroOutline.Trash).
				AccessibilityLabel("Bild entfernen"),
		),
		Center: ui.Image().URI(uri).ObjectFit(ui.FitContain).Frame(ui.Frame{Width: ui.Full, Height: ui.L256}).Border(ui.Border{}.Radius(ui.L16)),
	}).Frame(ui.Frame{Width: ui.Full, Height: ui.L256})
}
