// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

type TChats struct {
	selected *core.State[conversation.ID]
	frame    ui.Frame
}

func Chats(selected *core.State[conversation.ID]) TChats {
	return TChats{selected: selected}
}

func (c TChats) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()

	aiFindAll, ok := core.FromContext[conversation.FindAll](wnd.Context(), "")
	if !ok {
		return alert.Banner("no ai", "the ai module has not been enabled").Render(ctx)
	}

	return ui.VStack(
		ui.Each2(aiFindAll(wnd.Subject()), func(conv conversation.Conversation, err error) core.View {
			if err != nil {
				return alert.BannerError(err)
			}

			if conv.ID == c.selected.Get() {
				return ui.SecondaryButton(nil).Title(c.title(conv)).Frame(ui.Frame{}.FullWidth())
			}

			return ui.TertiaryButton(func() {
				c.selected.Set(conv.ID)
				c.selected.Notify()
			}).Title(c.title(conv)).Frame(ui.Frame{}.FullWidth())
		})...,
	).Frame(c.frame).Render(ctx)
}

func (c TChats) Frame(frame ui.Frame) TChats {
	c.frame = frame
	return c
}

func (c TChats) title(conv conversation.Conversation) string {
	if conv.Name != "" {
		return conv.Name
	}

	return string(conv.ID)
}
