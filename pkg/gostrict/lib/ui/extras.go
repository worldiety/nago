// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
)

// TBackground is a background of images and gradients, like ui.Background of nago, see TStack.Background.
type TBackground struct{ ref js.Ref }

// Background returns an empty background, like ui.Background{} of nago.
//
//gostrict:host
func Background() TBackground {
	return TBackground{ref: wrap(nui.Background{})}
}

// Fit sets, how the images fit into the view.
//
//gostrict:host
func (b TBackground) Fit(fit ObjectFit) TBackground {
	return TBackground{ref: wrap(get[nui.Background](b.ref).Fit(nui.ObjectFit(fit)))}
}

// AppendURI appends the image of the URI.
//
//gostrict:host
func (b TBackground) AppendURI(uri URI) TBackground {
	return TBackground{ref: wrap(get[nui.Background](b.ref).AppendURI(core.URI(uri)))}
}

// AppendLinearGradient appends a linear gradient of the colors.
//
//gostrict:host
func (b TBackground) AppendLinearGradient(colors ...Color) TBackground {
	return TBackground{ref: wrap(get[nui.Background](b.ref).AppendLinearGradient(colorsOf(colors)...))}
}

// Background sets the background.
//
//gostrict:host
func (c TStack) Background(bg TBackground) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Background(get[nui.Background](bg.ref)))}
}

// ImageID is the identifier of an image of the image store of nago.
type ImageID string

// Embed shows the SVG, like TImage.Embed of nago, which takes the bytes of any image.
//
//gostrict:host
func (c TImage) Embed(svg SVG) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).Embed([]byte(svg)))}
}

// EmbedAdaptive shows the SVG of the light or the dark mode.
//
//gostrict:host
func (c TImage) EmbedAdaptive(light, dark SVG) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).EmbedAdaptive([]byte(light), []byte(dark)))}
}

// Adaptive shows the image of the store of the light or the dark mode.
//
//gostrict:host
func (c TImage) Adaptive(light, dark ImageID) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).Adaptive(imageIDOf(light), imageIDOf(dark)))}
}

// IndicesState is a list of indices of the window, e.g. of the picked rows of a picker.
type IndicesState struct{ ref js.Ref }

// IndicesStateOf returns the state with the id of the window, which is empty at first.
//
//gostrict:host
func IndicesStateOf(wnd Window, id string) IndicesState {
	return IndicesState{ref: wrap(core.StateOf[[]int](wnd.nago(), id))}
}

// Get returns the indices.
//
//gostrict:host
func (s IndicesState) Get() list.List[int32] {
	return indicesOf(s.nago().Get())
}

// Set sets the indices and renders the window again.
//
//gostrict:host
func (s IndicesState) Set(v list.List[int32]) {
	s.nago().Set(intsOf(v))
}

// Observe calls fn, when the user changes the indices.
//
//gostrict:host
func (s IndicesState) Observe(fn func(v list.List[int32])) IndicesState {
	return IndicesState{ref: wrap(s.nago().Observe(func(v []int) { fn(indicesOf(v)) }))}
}

// TPicker lets the user pick rows of the program in a dialog, like picker.TPicker of nago, see PickerFromSlice.
type TPicker struct{ ref js.Ref }

func (TPicker) isView() {}

// PickerFromSlice returns a picker of rows, which the program keeps, like picker.Picker of nago, which takes the
// values themselves: the picker shows the rows by their index from 0 to rows-1, see Stringer and ItemRenderer, and
// writes the indices of the picked rows into the state.
//
//gostrict:host
func PickerFromSlice(label string, rows int32, selected IndicesState) TPicker {
	return TPicker{ref: wrap(picker.Picker(label, rowsUpTo(rows), selected.nago()))}
}

// DialogPresented returns the state, whether the dialog is shown.
//
//gostrict:host
func (c TPicker) DialogPresented() BoolState {
	return BoolState{ref: wrap(get[picker.TPicker[int]](c.ref).DialogPresented())}
}

// Title sets the title of the dialog.
//
//gostrict:host
func (c TPicker) Title(title string) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).Title(title))}
}

// SupportingText sets the text below the picker.
//
//gostrict:host
func (c TPicker) SupportingText(text string) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).SupportingText(text))}
}

// ErrorText sets the error below the picker.
//
//gostrict:host
func (c TPicker) ErrorText(text string) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).ErrorText(text))}
}

// MultiSelect lets the user pick several rows.
//
//gostrict:host
func (c TPicker) MultiSelect(mv bool) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).MultiSelect(mv))}
}

// SelectAllSupported offers to pick all rows at once.
//
//gostrict:host
func (c TPicker) SelectAllSupported(flag bool) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).SelectAllSupported(flag))}
}

// QuickFilterSupported offers a filter of the rows.
//
//gostrict:host
func (c TPicker) QuickFilterSupported(flag bool) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).QuickFilterSupported(flag))}
}

// Disabled disables the picker.
//
//gostrict:host
func (c TPicker) Disabled(disabled bool) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).Disabled(disabled))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TPicker) FullWidth() TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).FullWidth())}
}

// DetailView shows the view in the dialog.
//
//gostrict:host
func (c TPicker) DetailView(detailView View) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).DetailView(viewOf(detailView)))}
}

// WithDialogPresented binds, whether the dialog is shown, to the state.
//
//gostrict:host
func (c TPicker) WithDialogPresented(state BoolState) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).WithDialogPresented(state.nago()))}
}

// DialogOptions sets the options of the dialog.
//
//gostrict:host
func (c TPicker) DialogOptions(opts ...AlertOption) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).DialogOptions(alertOptionsOf(opts)...))}
}

// Stringer sets the text of a row, which the dialog and the filter use.
//
//gostrict:host
func (c TPicker) Stringer(fn func(row int32) string) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).Stringer(func(i int) string { return fn(int32(i)) }))}
}

// ItemRenderer sets the view of a row in the dialog.
//
//gostrict:host
func (c TPicker) ItemRenderer(fn func(row int32) View) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).ItemRenderer(func(i int) core.View { return viewOf(fn(int32(i))) }))}
}

// ItemPickedRenderer sets the view of the picked rows in the field.
//
//gostrict:host
func (c TPicker) ItemPickedRenderer(fn func(rows list.List[int32]) View) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).ItemPickedRenderer(func(rows []int) core.View { return viewOf(fn(indicesOf(rows))) }))}
}

// ItemRenderer2 sets the view of a row in the dialog with its checkbox.
//
//gostrict:host
func (c TPicker) ItemRenderer2(fn func(wnd Window, row int32, checked BoolState) View) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).ItemRenderer2(func(wnd core.Window, i int, checked *core.State[bool]) core.View {
		return viewOf(fn(Window{ref: wrap(wnd)}, int32(i), BoolState{ref: wrap(checked)}))
	}))}
}

// Padding sets the padding.
//
//gostrict:host
func (c TPicker) Padding(padding Padding) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).Padding(padding.nago()).(picker.TPicker[int]))}
}

// Frame sets the size.
//
//gostrict:host
func (c TPicker) Frame(frame Frame) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).Frame(frame.nago()).(picker.TPicker[int]))}
}

// WithFrame changes the size with fn, which gets the current one.
//
//gostrict:host
func (c TPicker) WithFrame(fn func(Frame) Frame) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(picker.TPicker[int]))}
}

// Border sets the border.
//
//gostrict:host
func (c TPicker) Border(border Border) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).Border(border.nago()).(picker.TPicker[int]))}
}

// Visible shows or hides the picker.
//
//gostrict:host
func (c TPicker) Visible(visible bool) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).Visible(visible).(picker.TPicker[int]))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TPicker) AccessibilityLabel(label string) TPicker {
	return TPicker{ref: wrap(get[picker.TPicker[int]](c.ref).AccessibilityLabel(label).(picker.TPicker[int]))}
}
