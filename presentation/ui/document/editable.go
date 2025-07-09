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

func (c TEditable) Style(style ToggleStyle) TEditable {
	c.style = style
	return c
}

func (c TEditable) Comment(action func()) TEditable {
	c.comment = action
	return c
}

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

func (c TEditable) renderInlineTrailing(ctx core.RenderContext) core.RenderNode {
	return ui.HStack(
		c.renderContent(),
		c.renderToggleButton(),
	).Gap(ui.L8).Render(ctx)
}

func (c TEditable) renderClickable(ctx core.RenderContext) core.RenderNode {
	v := ui.VStack().Gap(ui.L8)

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

func (c TEditable) renderTopTrailing(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		// toolbar
		ui.IfFunc(c.editPresented != nil, func() core.View {
			return ui.HStack(c.renderToggleButton()).FullWidth().Alignment(ui.Trailing)
		}),

		c.renderContent(),
	).Gap(ui.L8).Render(ctx)
}

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
