// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"go.wdy.de/nago/application/image"
	http_image "go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/avatar"
)

func AvatarPicker(wnd core.Window, setCreator image.CreateSrcSet, selfId string, id image.ID, state *core.State[image.ID], paraphe string, style avatar.Style) ui.DecoredView {
	if setCreator == nil {
		fn, ok := core.FromContext[image.CreateSrcSet](wnd.Context(), "")
		if !ok {
			panic("image.CreateSrcSet not available")
		}

		setCreator = fn
	}

	var img core.View
	if id != "" {
		// TODO replace me with source set due to different density problem
		uri := core.URI(http_image.NewURL(http_image.Endpoint, id, image.FitCover, 120, 120))
		img = avatar.URI(uri).Size(ui.L120).Style(style)
	} else {
		img = avatar.Text(paraphe).Size(ui.L120).Style(style)
	}

	var actionBtn core.View
	if id == "" {
		actionBtn = ui.HStack(ui.ImageIcon(heroOutline.Plus).StrokeColor(ui.ColorBlack).Frame(ui.Frame{}.FullWidth())).
			Action(func() {
				wndImportFiles(wnd, setCreator, selfId, state)
			}).
			BackgroundColor(ui.ColorWhite).
			Frame(ui.Frame{}.Size(ui.L32, ui.L32)).
			Padding(ui.Padding{}.All(ui.L2)).
			Border(ui.Border{}.Width(ui.L4).Circle().Color(ui.ColorBlack))
	} else {
		actionBtn = ui.HStack(ui.ImageIcon(heroOutline.Trash).StrokeColor(ui.ColorError).Frame(ui.Frame{}.FullWidth())).
			Action(func() {
				state.Set("")
				state.Notify()
			}).
			BackgroundColor(ui.ColorWhite).
			Frame(ui.Frame{}.Size(ui.L32, ui.L32)).
			Padding(ui.Padding{}.All(ui.L2)).
			Border(ui.Border{}.Width(ui.L4).Circle().Color(ui.ColorError))
	}

	return ui.Box(ui.BoxLayout{

		Center:         img,
		BottomTrailing: actionBtn,
	}).Frame(ui.Frame{}.Size(ui.L120, ui.L120))
}
