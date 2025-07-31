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

// TSingleImagePicker is a composite component(Single Image Picker).
type TSingleImagePicker struct {
	wnd         core.Window
	setCreator  image.CreateSrcSet
	loadSrcSet  image.LoadSrcSet
	loadBestFit image.LoadBestFit
	selfID      string
	id          image.ID
	state       *core.State[image.ID]

	padding            ui.Padding
	frame              ui.Frame
	border             ui.Border
	accessibilityLabel string
	invisible          bool
}

func SingleImagePicker(wnd core.Window, setCreator image.CreateSrcSet, loadSrcSet image.LoadSrcSet, loadBestFit image.LoadBestFit, selfId string, id image.ID, state *core.State[image.ID]) TSingleImagePicker {
	return TSingleImagePicker{
		wnd:         wnd,
		setCreator:  setCreator,
		loadSrcSet:  loadSrcSet,
		loadBestFit: loadBestFit,
		selfID:      selfId,
		id:          id,
		state:       state,
	}
}

func (t TSingleImagePicker) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TSingleImagePicker) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

func (t TSingleImagePicker) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TSingleImagePicker) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

func (t TSingleImagePicker) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

func (t TSingleImagePicker) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}

func (t TSingleImagePicker) Render(ctx core.RenderContext) core.RenderNode {
	if t.setCreator == nil {
		fn, ok := core.FromContext[image.CreateSrcSet](t.wnd.Context(), "")
		if !ok {
			panic("image.CreateSrcSet not available") // TODO or better an alert.Banner?
		}

		t.setCreator = fn
	}

	if t.loadSrcSet == nil {
		fn, ok := core.FromContext[image.LoadSrcSet](t.wnd.Context(), "")
		if !ok {
			panic("image.LoadSrcSet not available") // TODO or better an alert.Banner?
		}

		t.loadSrcSet = fn
	}

	if t.loadBestFit == nil {
		fn, ok := core.FromContext[image.LoadBestFit](t.wnd.Context(), "")
		if !ok {
			panic("image.LoadSrcSet not available") // TODO or better an alert.Banner?
		}

		t.loadBestFit = fn
	}

	// empty id case
	if t.id == "" {
		return ui.HStack(
			ui.SecondaryButton(func() {
				wndImportFiles(t.wnd, t.setCreator, t.selfID, t.state)
			}).PreIcon(heroOutline.Plus).Title("Bild hinzufügen"),
		).Alignment(ui.Trailing).Render(ctx)
	}

	// the preview case
	targetWidth := http_image.EstimateWidth(t.wnd)
	uri := core.URI(http_image.NewURL(http_image.Endpoint, t.id, image.FitCover, targetWidth, targetWidth))

	return ui.Box(ui.BoxLayout{
		TopTrailing: ui.HStack(
			ui.TertiaryButton(func() {
				optSet, err := t.loadSrcSet(t.wnd.Subject(), t.id)
				if err != nil {
					alert.ShowBannerError(t.wnd, err)
					return
				}

				if optSet.IsNone() {
					alert.ShowBannerMessage(t.wnd, alert.Message{Title: "Bild SrcSet nicht gefunden", Message: "Das Bild kann nicht heruntergeladen werden, da es nicht gefunden wurde."})
					return
				}

				srcSet := optSet.Unwrap()

				rf := core.NewReaderFile(func() (io.ReadCloser, error) {
					optReader, err := t.loadBestFit(t.wnd.Subject(), t.id, image.FitNone, 0, 0)
					if err != nil {
						alert.ShowBannerError(t.wnd, err)
						return nil, err
					}

					if optReader.IsNone() {
						alert.ShowBannerMessage(t.wnd, alert.Message{Title: "Bild nicht gefunden", Message: "Das Bild kann nicht heruntergeladen werden, da es nicht gefunden wurde."})
						return nil, fmt.Errorf("bind image one not found")
					}

					return optReader.Unwrap(), nil
				})

				rf.SetName(srcSet.Name)
				rf.SetMimeType("image/*")
				t.wnd.ExportFiles(core.ExportFilesOptions{
					ID:    string(t.id) + "-download",
					Files: []core.File{rf},
				})
			}).PreIcon(heroOutline.ArrowDownTray).
				AccessibilityLabel("Bild herunterladen"),
			ui.TertiaryButton(func() {
				t.state.Set("")
				t.state.Notify()
			}).PreIcon(heroOutline.Trash).
				AccessibilityLabel("Bild entfernen"),
		),
		Center: ui.Image().URI(uri).ObjectFit(ui.FitContain).Frame(ui.Frame{Width: ui.Full, Height: ui.L256}).Border(ui.Border{}.Radius(ui.L16)),
	}).Frame(ui.Frame{Width: ui.Full, Height: ui.L256}).Render(ctx)
}
