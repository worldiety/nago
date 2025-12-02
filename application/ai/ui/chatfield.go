// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

type TChatField struct {
	text     *core.State[string]
	action   func()
	disabled bool
}

func ChatField(text *core.State[string]) TChatField {
	return TChatField{text: text}
}

func (c TChatField) Action(fn func()) TChatField {
	c.action = fn
	return c
}

func (c TChatField) Enabled(enabled bool) TChatField {
	c.disabled = !enabled
	return c
}

func (c TChatField) Render(ctx core.RenderContext) core.RenderNode {
	return ui.HStack(
		ui.TextField("", c.text.String()).
			ID("ai-user-prompt").
			InputValue(c.text).
			Style(ui.TextFieldBasic).
			KeydownEnter(c.action).
			Lines(1).
			FullWidth().
			Disabled(c.disabled),

		ui.PrimaryButton(c.action).
			PreIcon(icons.ArrowRight).
			Frame(ui.Frame{MinWidth: ui.L40}).
			Disabled(c.disabled).
			ID("start-chat-button"),
	).
		BackgroundColor(ui.M2).
		Border(ui.Border{}.Width(ui.L1).Color(ui.M5).Radius(ui.L24)).
		Padding(ui.Padding{}.All(ui.L8)).
		Frame(ui.Frame{}.FullWidth()).
		Render(ctx)
}
