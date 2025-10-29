// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"log/slog"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/ai"
	uiai "go.wdy.de/nago/application/ai/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

type Management struct {
	UseCases ai.UseCases
	Pages    uiai.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	secrets, err := cfg.SecretManagement()
	if err != nil {
		return Management{}, err
	}

	ucAI := ai.NewUseCases(cfg.EventBus(), secrets.UseCases.FindGroupSecrets)

	management = Management{

		UseCases: ucAI,
		Pages: uiai.Pages{
			Workspaces:   "admin/ai/workspaces",
			Agents:       "admin/ai/workspace",
			Agent:        "admin/ai/workspace/agent",
			Provider:     "admin/ai/provider",
			Library:      "admin/ai/library",
			Conversation: "admin/ai/provider/conversation",
			Chat:         "admin/ai/chat",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.Provider, func(wnd core.Window) core.View {
		return uiai.PageProvider(wnd, management.UseCases)
	})
	cfg.RootViewWithDecoration(management.Pages.Library, func(wnd core.Window) core.View {
		return uiai.PageLibrary(wnd, management.UseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.Conversation, func(wnd core.Window) core.View {
		return uiai.PageConversation(wnd, management.UseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.Chat, func(wnd core.Window) core.View {
		return uiai.PageChat(wnd, management.UseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {

		grp := admin.Group{
			Title: "AI",
		}

		for provider, err := range ucAI.FindAllProvider(user.SU()) {
			if err != nil {
				slog.Error("failed to find provider", "err", err.Error())
				continue
			}

			grp.Entries = append(grp.Entries, admin.Card{
				Title:        provider.Name(),
				Text:         provider.Description(),
				Target:       management.Pages.Provider,
				TargetParams: core.Values{"provider": string(provider.Identity())},
			})
		}

		return grp
	})

	cfg.AddContextValue(core.ContextValue("nago.ai", management))

	cfg.AddContextValue(core.ContextValue("", management.UseCases.FindProviderByID))
	cfg.AddContextValue(core.ContextValue("", management.UseCases.FindProviderByName))

	slog.Info("installed AI module")
	return management, nil
}
