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
// This component allows users to select and display an avatar image,
// with optional styling, state management, and accessibility support.
type TAvatarPicker struct {
	wnd        core.Window           // reference to the application window
	setCreator image.CreateSrcSet    // function to create responsive image sources
	selfID     string                // identifier for the current user
	id         image.ID              // identifier of the avatar image
	state      *core.State[image.ID] // state holding the selected avatar ID (nil if no interactivity)
	paraphe    string                // optional initials or text fallback
	style      avatar.Style          // avatar styling options

	padding            ui.Padding // layout padding
	frame              ui.Frame   // frame defining size and layout
	border             ui.Border  // border styling for the avatar
	accessibilityLabel string     // accessibility label for screen readers
	invisible          bool       // whether the avatar is invisible
	disabled           bool
}

// AvatarPicker creates a new TAvatarPicker with the given parameters.
// By default, it initializes with a 120x120 frame size.
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

func (t TAvatarPicker) Enabled(b bool) TAvatarPicker {
	t.disabled = !b
	return t
}

func (t TAvatarPicker) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

// WithFrame updates the frame of the avatar picker using a frame transformation function.
func (t TAvatarPicker) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

// Frame sets the frame of the avatar picker directly.
func (t TAvatarPicker) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

// Border sets the border styling of the avatar picker.
func (t TAvatarPicker) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

// Visible toggles the visibility of the avatar picker.
func (t TAvatarPicker) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

// AccessibilityLabel sets the accessibility label for screen readers.
func (t TAvatarPicker) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}

// Render builds and returns the RenderNode for the TAvatarPicker.
// It displays either an avatar image (if an ID is set) or a text fallback (paraphe).
// Additionally, it shows an action button:
// - If no avatar is set: a plus button to upload/select an image
// - If an avatar is set: a trash button to remove it
// The avatar is centered inside a box layout, with the action button
// positioned at the bottom trailing corner, and the whole component
// rendered within the configured frame.
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
	if !t.disabled {
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
	}

	return ui.Box(ui.BoxLayout{

		Center:         img,
		BottomTrailing: actionBtn,
	}).Frame(t.frame).
		Render(ctx)
}
