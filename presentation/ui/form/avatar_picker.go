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

// TAvatarPicker is a composite component(Avatar Picker).
type TAvatarPicker struct {
	wnd        core.Window
	setCreator image.CreateSrcSet
	selfID     string
	id         image.ID
	state      *core.State[image.ID]
	paraphe    string
	style      avatar.Style

	padding            ui.Padding
	frame              ui.Frame
	border             ui.Border
	accessibilityLabel string
	invisible          bool
}

func AvatarPicker(wnd core.Window, setCreator image.CreateSrcSet, selfId string, id image.ID, state *core.State[image.ID], paraphe string, style avatar.Style) TAvatarPicker {
	return TAvatarPicker{
		wnd:        wnd,
		setCreator: setCreator,
		selfID:     selfId,
		id:         id,
		state:      state,
		paraphe:    paraphe,
		style:      style,
		frame:      ui.Frame{}.Size(ui.L120, ui.L120),
	}
}

func (t TAvatarPicker) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TAvatarPicker) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

func (t TAvatarPicker) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TAvatarPicker) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

func (t TAvatarPicker) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

func (t TAvatarPicker) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}

func (t TAvatarPicker) Render(ctx core.RenderContext) core.RenderNode {
	if t.setCreator == nil {
		fn, ok := core.FromContext[image.CreateSrcSet](t.wnd.Context(), "")
		if !ok {
			panic("image.CreateSrcSet not available")
		}

		t.setCreator = fn
	}

	var img core.View
	if t.id != "" {
		// TODO replace me with source set due to different density problem
		uri := core.URI(http_image.NewURL(http_image.Endpoint, t.id, image.FitCover, 120, 120))
		img = avatar.URI(uri).Size(ui.L120).Style(t.style)
	} else {
		img = avatar.Text(t.paraphe).Size(ui.L120).Style(t.style)
	}

	var actionBtn core.View
	if t.id == "" {
		actionBtn = ui.HStack(ui.ImageIcon(heroOutline.Plus).StrokeColor(ui.ColorBlack).Frame(ui.Frame{}.FullWidth())).
			Action(func() {
				wndImportFiles(t.wnd, t.setCreator, t.selfID, t.state)
			}).
			BackgroundColor(ui.ColorWhite).
			Frame(ui.Frame{}.Size(ui.L32, ui.L32)).
			Padding(ui.Padding{}.All(ui.L2)).
			Border(ui.Border{}.Width(ui.L4).Circle().Color(ui.ColorBlack))
	} else {
		actionBtn = ui.HStack(ui.ImageIcon(heroOutline.Trash).StrokeColor(ui.ColorError).Frame(ui.Frame{}.FullWidth())).
			Action(func() {
				t.state.Set("")
				t.state.Notify()
			}).
			BackgroundColor(ui.ColorWhite).
			Frame(ui.Frame{}.Size(ui.L32, ui.L32)).
			Padding(ui.Padding{}.All(ui.L2)).
			Border(ui.Border{}.Width(ui.L4).Circle().Color(ui.ColorError))
	}

	return ui.Box(ui.BoxLayout{

		Center:         img,
		BottomTrailing: actionBtn,
	}).Frame(t.frame).
		Render(ctx)
}
