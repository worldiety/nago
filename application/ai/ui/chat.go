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
	"go.wdy.de/nago/pkg/events"
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
	padding      ui.Padding
}

func Chat(conv *core.State[conversation.ID], text *core.State[string]) TChat {
	return TChat{conv: conv, text: text}
}

func (c TChat) Padding(padding ui.Padding) TChat {
	c.padding = padding
	return c
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

	aiFindMessages, ok := core.FromContext[conversation.FindMessages](wnd.Context(), "")
	if !ok {
		return alert.Banner("no ai", "the ai module has not been enabled").Render(ctx)
	}

	aiAppendMessage, ok := core.FromContext[conversation.Append](wnd.Context(), "")
	if !ok {
		return alert.Banner("no ai", "the ai module has not been enabled").Render(ctx)
	}

	bus, ok := core.FromContext[events.EventBus](wnd.Context(), "")
	if !ok {
		panic("no event-bus")
	}

	pleaseWaitPresented := core.DerivedState[bool](c.conv, "-pw-presented")

	ctx.Window().AddDestroyObserver(events.SubscribeFor(bus, func(evt conversation.AgentAppended) {
		pleaseWaitPresented.Set(false)
		// clear the prompt
		c.text.Set("")
	}))

	return ui.VStack(
		ui.Each2(aiFindMessages(wnd.Subject(), c.conv.Get()), func(msg message.Message, err error) core.View {
			if err != nil {
				return alert.BannerError(err)
			}

			stack := ui.VStack()
			style := MessageAgent
			align := ui.Leading
			if msg.CreatedBy != "" {
				style = MessageHuman
				align = ui.Trailing
			}

			for content := range msg.Inputs.All() {
				switch {
				case content.Text.IsSome():
					stack = stack.Append(ChatMessage().Markdown(content.Text.Unwrap()).Style(style))
				}
			}

			return stack.Alignment(align).FullWidth()
		})...,
	).Append(
		ui.If(pleaseWaitPresented.Get(),
			ui.HStack(
				ChatMessage().Icon(icons.DotsHorizontal).Style(MessageAgent),
			).FullWidth().Alignment(ui.Leading),
		),
	).
		Append(
			ui.HStack(
				ChatField(c.text).
					Enabled(!pleaseWaitPresented.Get()).
					Action(func() {
						pleaseWaitPresented.Set(true)
						if c.conv.Get() == "" {
							// create a new conversation
							tmp := c.text.Get()
							if c.startOptions.Name == "" {
								c.startOptions.Name = tmp
							}

							c.startOptions.Input = append(c.startOptions.Input, message.Content{Text: option.Pointer(&tmp)})
							cid, err := aiStartConv(wnd.Subject(), c.startOptions)
							if err != nil {
								alert.ShowBannerError(wnd, err)
								return
							}

							c.conv.Update(cid)
						} else {
							// append to existing conversation
							tmp := c.text.Get()
							_, err := aiAppendMessage(wnd.Subject(), conversation.AppendOptions{
								Conversation: c.conv.Get(),
								Input: []message.Content{
									{Text: option.Pointer(&tmp)},
								},
								CloudStore: c.startOptions.CloudStore,
							})

							if err != nil {
								alert.ShowBannerError(wnd, err)
								return
							}
						}

					}),
				ui.PrimaryButton(nil).Title(rstring.ActionFileUpload.Get(wnd)).PreIcon(icons.Upload).Frame(ui.Frame{MinWidth: "12rem"}),
				ui.SecondaryButton(nil).Title(rstring.LabelHowItWorks.Get(wnd)).Frame(ui.Frame{MinWidth: "12rem"}),
			).Gap(ui.L8).
				FullWidth(),
		).
		Append(ui.VStack().ID("end-of-chat")).
		FullWidth().
		Gap(ui.L16).
		Padding(c.padding).
		Render(ctx)
}
