// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package document

import (
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

type ToggleStyle int

const (
	TopTrailing ToggleStyle = iota
	InlineTrailing
	Clickable
)

// TEditable is a util component (Editable).
// This component allows toggling between a view mode and an edit mode.
// It is typically used when users should be able to quickly switch
// between read-only display and editing input fields or values.
type TEditable struct {
	frame           ui.Frame
	padding         ui.Padding
	border          ui.Border
	backgroundColor ui.Color
	alignment       ui.Alignment
	onView          func() core.View
	onEdit          func() core.View
	editPresented   *core.State[bool]
	style           ToggleStyle
	comment         func()
}

// Editable creates a new TEditable with view and edit callbacks.
func Editable(onView func() core.View, onEdit func() core.View) TEditable {
	return TEditable{
		onView: onView,
		onEdit: onEdit,
	}
}

// InputValue sets the edit presented state. If nil or never set, only the onView func is ever evaluated.
func (c TEditable) InputValue(editPresented *core.State[bool]) TEditable {
	c.editPresented = editPresented
	return c
}

// Style sets the toggle style of the editable component.
func (c TEditable) Style(style ToggleStyle) TEditable {
	c.style = style
	return c
}

// Comment sets the callback executed when a comment action is triggered.
func (c TEditable) Comment(action func()) TEditable {
	c.comment = action
	return c
}

// Frame sets the frame (size and layout) of the editable component.
func (c TEditable) Frame(frame ui.Frame) TEditable {
	c.frame = frame
	return c
}

// BackgroundColor sets the background color of the editable component.
func (c TEditable) BackgroundColor(color ui.Color) TEditable {
	c.backgroundColor = color
	return c
}

// Alignment sets the alignment of the editable component.
func (c TEditable) Alignment(alignment ui.Alignment) TEditable {
	c.alignment = alignment
	return c
}

// Border sets the border of the editable component.
func (c TEditable) Border(border ui.Border) TEditable {
	c.border = border
	return c
}

// Render displays the editable component using the rendering strategy
// defined by its style.
func (c TEditable) Render(ctx core.RenderContext) core.RenderNode {
	switch c.style {
	default:
		return c.renderTopTrailing(ctx)
	case InlineTrailing:
		return c.renderInlineTrailing(ctx)
	case Clickable:
		return c.renderClickable(ctx)
	}
}

// renderInlineTrailing places the toggle button inline to the right of the content.
func (c TEditable) renderInlineTrailing(ctx core.RenderContext) core.RenderNode {
	return ui.HStack(
		c.renderContent(),
		c.renderToggleButton(),
	).BackgroundColor(c.backgroundColor).
		Gap(ui.L8).
		Frame(c.frame).
		Border(c.border).
		Render(ctx)
}

// renderClickable makes the whole component clickable to enter edit mode,
// with the toggle button shown only while editing.
func (c TEditable) renderClickable(ctx core.RenderContext) core.RenderNode {
	v := ui.VStack().BackgroundColor(c.backgroundColor).
		Gap(ui.L8).
		Frame(c.frame).
		Border(c.border).(ui.TVStack)

	if c.editPresented != nil {
		if !c.editPresented.Get() {
			v = v.Action(func() {
				c.editPresented.Set(true)
			}).HoveredBackgroundColor(ui.ColorCardFooter)

			v = v.Padding(ui.Padding{}.All(ui.L8)).
				Border(ui.Border{}.Radius(ui.L8)).
				AccessibilityLabel("Klicken zum Bearbeiten").(ui.TVStack)

		} else {
			v = v.Append(ui.HStack(c.renderToggleButton()).FullWidth().Alignment(ui.Trailing))
		}
	}

	v = v.Append(c.renderContent())

	return v.Render(ctx)
}

func (c TEditable) FullWidth() TEditable {
	c.frame.Width = ui.Full
	return c
}

// renderTopTrailing places the toggle button above the content, aligned to the top right.
func (c TEditable) renderTopTrailing(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		// toolbar
		ui.IfFunc(c.editPresented != nil, func() core.View {
			return ui.HStack(c.renderToggleButton()).FullWidth().Alignment(ui.Trailing)
		}),

		c.renderContent(),
	).BackgroundColor(c.backgroundColor).
		Gap(ui.L8).
		Frame(c.frame).
		Border(c.border).
		Render(ctx)
}

// renderToggleButton returns the appropriate toggle button(s)
// depending on the current edit state.
func (c TEditable) renderToggleButton() core.View {
	if c.editPresented == nil {
		return nil
	}

	if !c.editPresented.Get() {
		v := ui.HStack(

			ui.SecondaryButton(func() {
				c.editPresented.Set(true)
				c.editPresented.Notify()
			}).PreIcon(flowbiteOutline.Edit).
				AccessibilityLabel("Bearbeiten"),
		)

		if c.comment != nil {
			v = v.Append(
				ui.TertiaryButton(c.comment).PreIcon(flowbiteOutline.Annotation).AccessibilityLabel("Kommentar hinzufügen"),
			).Gap(ui.L4)
		}

		return v
	}

	return ui.SecondaryButton(func() {
		c.editPresented.Set(false)
		c.editPresented.Notify()
	}).PreIcon(flowbiteOutline.Check).
		AccessibilityLabel("Bearbeitungsansicht schließen")
}

// renderContent returns the actual content of the component.
func (c TEditable) renderContent() core.View {
	if c.editPresented == nil && c.onView != nil {
		return c.onView()
	}

	if c.editPresented == nil {
		return nil
	}

	if c.editPresented.Get() && c.onEdit() != nil {
		return c.onEdit()
	}

	if !c.editPresented.Get() && c.onView != nil {
		return c.onView()
	}

	return nil
}
