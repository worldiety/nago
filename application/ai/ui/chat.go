// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

type TChat struct {
	ws           workspace.ID
	conv         *core.State[conversation.ID]
	text         *core.State[string]
	startOptions conversation.StartOptions
}

func Chat(conv *core.State[conversation.ID], text *core.State[string]) TChat {
	return TChat{conv: conv, text: text}
}

func (c TChat) StartOptions(opts conversation.StartOptions) TChat {
	c.startOptions = opts
	return c
}

func (c TChat) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()

	aiStartConv, ok := core.FromContext[conversation.Start](wnd.Context(), "")
	if !ok {
		return alert.Banner("no ai", "the ai module has not been enabled").Render(ctx)
	}

	return ui.VStack(
		ui.HStack(
			ChatField(c.text).
				Action(func() {
					if c.conv.Get() == "" {
						// create a new conversation
						tmp := c.text.Get()
						c.startOptions.Input = append(c.startOptions.Input, message.Content{Text: option.Pointer(&tmp)})
						if _, err := aiStartConv(wnd.Subject(), c.startOptions); err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						return
					}
				}),
			ui.PrimaryButton(nil).Title(rstring.ActionFileUpload.Get(wnd)).PreIcon(icons.Upload).Frame(ui.Frame{MinWidth: "12rem"}),
			ui.SecondaryButton(nil).Title(rstring.LabelHowItWorks.Get(wnd)).Frame(ui.Frame{MinWidth: "12rem"}),
		).Gap(ui.L8).
			FullWidth(),
	).FullWidth().
		Render(ctx)
}

type MessageStyle int

const (
	MessageAgent MessageStyle = iota
	MessageHuman
)

type TChatMessage struct {
	style MessageStyle
	text  string
}

func (c TChatMessage) Render(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		ui.Text(c.text),
	).
		BackgroundColor(ui.M2).
		Border(ui.Border{}.Width(ui.L1).Color(ui.M5).Radius(ui.L24)).
		Padding(ui.Padding{}.All(ui.L8)).
		Frame(ui.Frame{}.FullWidth()).
		Render(ctx)
}

type TChatField struct {
	text   *core.State[string]
	action func()
}

func ChatField(text *core.State[string]) TChatField {
	return TChatField{text: text}
}

func (c TChatField) Action(fn func()) TChatField {
	c.action = fn
	return c
}

func (c TChatField) Render(ctx core.RenderContext) core.RenderNode {
	return ui.HStack(
		ui.TextField("", c.text.String()).InputValue(c.text).Style(ui.TextFieldBasic).FullWidth(),
		ui.PrimaryButton(c.action).PreIcon(icons.ArrowRight).Frame(ui.Frame{MinWidth: ui.L40}),
	).
		BackgroundColor(ui.M2).
		Border(ui.Border{}.Width(ui.L1).Color(ui.M5).Radius(ui.L24)).
		Padding(ui.Padding{}.All(ui.L8)).
		Frame(ui.Frame{}.FullWidth()).
		Render(ctx)
}
