// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"
	"os"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

var (
	StrHowCanXHelp = i18n.MustVarString("nago.ai.chat.help_x", i18n.Values{language.English: "How can {name} help you today?", language.German: "Wie kann {name} Dir heute weiterhelfen?"})
)

func PageChat(wnd core.Window, uc ai.UseCases) core.View {
	var prov provider.Provider

	if provName := wnd.Values()["provider-name"]; provName != "" {
		optProv, err := uc.FindProviderByName(wnd.Subject(), provName)
		if err != nil {
			return alert.BannerError(err)
		}

		if optProv.IsNone() {
			return alert.BannerError(fmt.Errorf("provider by name %s not found: %w", provName, os.ErrNotExist))
		}

		prov = optProv.Unwrap()
	}

	if prov == nil && wnd.Values()["provider"] != "" {
		pid := provider.ID(wnd.Values()["provider"])
		optProv, err := uc.FindProviderByID(wnd.Subject(), pid)
		if err != nil {
			return alert.BannerError(err)
		}

		if optProv.IsNone() {
			return alert.BannerError(fmt.Errorf("provider by id %s not found: %w", pid, os.ErrNotExist))
		}

		prov = optProv.Unwrap()
	}

	if prov == nil {
		for p, err := range uc.FindAllProvider(wnd.Subject()) {
			if err != nil {
				return alert.BannerError(err)
			}

			prov = p
			//slog.Info("no ai provider specified for chat, selected first one", "provider", p.Identity())
			break
		}

		if prov == nil {
			return alert.BannerError(fmt.Errorf("not ai provider found, check secret and permissions: %w", os.ErrNotExist))
		}
	}

	prompt := core.AutoState[string](wnd)
	conv := core.AutoState[conversation.ID](wnd).Init(func() conversation.ID {
		return conversation.ID(wnd.Values()["conversation"])
	})

	const innerFullHeight = "calc(100vh - 6rem - 1px)" // remember: this only works with NoFooter option for the according page path

	return ui.HStack(
		ui.ScrollView(
			Chats(prov, conv).Frame(ui.Frame{Width: ui.Full, MinHeight: innerFullHeight}),
		).Axis(ui.ScrollViewAxisVertical).Frame(ui.Frame{Height: ui.Full, Width: ui.L320}),

		ui.ScrollView(
			Chat(prov, conv, prompt).
				StartOptions(StartOptions{
					AgentName:  wnd.Values()["agent-name"],
					Agent:      agent.ID(wnd.Values()["agent"]),
					CloudStore: true,
				}).
				Teaser(teaser(wnd, prov)).
				Padding(ui.Padding{}.All(ui.L16)).
				Frame(ui.Frame{Width: ui.Full, Height: ui.Full, MinHeight: innerFullHeight}),
		).ScrollToView("end-of-chat", ui.ScrollAnimationSmooth).
			Axis(ui.ScrollViewAxisVertical).Frame(ui.Frame{Height: ui.Full, Width: ui.Full}),
	).Alignment(ui.Top).Frame(ui.Frame{Width: ui.Full, Height: innerFullHeight})
}

func teaser(wnd core.Window, prov provider.Provider) core.View {
	return ui.VStack(
		ui.Text(StrHowCanXHelp.Get(wnd, i18n.String("name", prov.Name()))).
			Font(ui.DisplayLarge).
			Frame(ui.Frame{MaxWidth: "33%"}),
	).FullWidth().Alignment(ui.Leading)
}
