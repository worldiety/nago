// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"log/slog"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/provider/cache"
	uiai "go.wdy.de/nago/application/ai/ui"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

var (
	StrMaintenanceAdminCardDesc = i18n.MustString("nago.ai.admin.maintenance_desc", i18n.Values{language.English: "Apply some maintenance tasks to the AI subsystem.", language.German: "Wartungsarbeiten am KI Subsystem durchführen."})
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

	cacheEnabled := true
	repoAgents, err := application.JSONRepository[agent.Agent](cfg, "nago.ai.cache.agent")
	if err != nil {
		return Management{}, err
	}

	repoConversations, err := application.JSONRepository[conversation.Conversation](cfg, "nago.ai.cache.conversation")
	if err != nil {
		return Management{}, err
	}

	repoMessages, err := application.JSONRepository[message.Message](cfg, "nago.ai.cache.message")
	if err != nil {
		return Management{}, err
	}

	repoLibraries, err := application.JSONRepository[library.Library](cfg, "nago.ai.cache.library")
	if err != nil {
		return Management{}, err
	}

	repoDocuments, err := application.JSONRepository[document.Document](cfg, "nago.ai.cache.document")
	if err != nil {
		return Management{}, err
	}

	repoModels, err := application.JSONRepository[model.Model](cfg, "nago.ai.cache.model")
	if err != nil {
		return Management{}, err
	}

	idxConvStore, err := cfg.EntityStore("nago.ai.cache.idx_conversation_message")
	if err != nil {
		return Management{}, err
	}
	idxConvMsg := data.NewCompositeIndex[conversation.ID, message.ID](idxConvStore)

	ucAI := ai.NewUseCases(cfg.EventBus(), secrets.UseCases.FindGroupSecrets, func(provider provider.Provider) (provider.Provider, error) {
		if !cacheEnabled {
			return provider, nil
		}

		prov := cache.NewProvider(
			provider,
			repoModels,
			repoLibraries,
			repoAgents,
			repoDocuments,
			repoConversations,
			repoMessages,
			idxConvMsg,
		)

		return prov, nil
	})

	management = Management{

		UseCases: ucAI,
		Pages: uiai.Pages{
			Maintenance:  "admin/ai/maintenance",
			Provider:     "admin/ai/provider",
			Library:      "admin/ai/library",
			Conversation: "admin/ai/provider/conversation",
			Chat:         "admin/ai/chat",
			Agent:        "admin/ai/agent",
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

	cfg.RootViewWithDecoration(management.Pages.Maintenance, func(wnd core.Window) core.View {
		return uiai.PageMaintenance(wnd, management.UseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.Agent, func(wnd core.Window) core.View {
		return uiai.PageAgent(wnd, management.UseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {

		grp := admin.Group{
			Title: "AI",
		}

		grp.Entries = append(grp.Entries, admin.Card{
			Title:      rstring.LabelMaintenance.Get(subject),
			Text:       StrMaintenanceAdminCardDesc.Get(subject),
			Target:     management.Pages.Maintenance,
			Permission: ai.PermClearCache,
		})

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
