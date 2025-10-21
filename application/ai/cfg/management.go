// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"iter"
	"log/slog"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/provider/mistralai"
	uiai "go.wdy.de/nago/application/ai/ui"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrAdminAIWorkspacesTitle = i18n.MustString("nago.ai.admin.workspaces_title", i18n.Values{language.English: "AI Workspaces", language.German: "KI Workspaces"})
	StrAdminAIWorkspacesDesc  = i18n.MustString("nago.ai.admin.workspaces_desc", i18n.Values{language.English: "Manage the available AI workspaces. A workspace defines agents, libraries and prompts.", language.German: "Verwalte die verfügbaren KI Workspaces, die insbesondere KI Agenten, Bibliotheksquellen und Anweisungen enthalten."})
)

type Management struct {
	AgentUseCases        agent.UseCases
	WorkspaceUseCases    workspace.UseCases
	ConversationUseCases conversation.UseCases
	Pages                uiai.Pages
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

	repoWorkspace, err := application.JSONRepository[workspace.Workspace](cfg, "nago.ai.workspace")
	if err != nil {
		return Management{}, err
	}

	repoAgents, err := application.JSONRepository[agent.Agent](cfg, "nago.ai.agent")
	if err != nil {
		return Management{}, err
	}

	syncAgentRepo, err := application.JSONRepository[mistralai.SynchronizedAgent](cfg, "nago.ai.provider.mistralai.agents")
	if err != nil {
		return Management{}, err
	}

	syncConvRepo, err := application.JSONRepository[mistralai.SynchronizedConversation](cfg, "nago.ai.provider.mistralai.conversations")
	if err != nil {
		return Management{}, err
	}

	repoConversation, err := application.JSONRepository[conversation.Conversation](cfg, "nago.ai.conversation")
	if err != nil {
		return Management{}, err
	}

	repoMessages, err := application.JSONRepository[message.Message](cfg, "nago.ai.message")
	if err != nil {
		return Management{}, err
	}

	storeIdxConvMsg, err := cfg.EntityStore("nago.ai.conversation_message_idx")
	if err != nil {
		return Management{}, err
	}
	idxConvMsg := data.NewCompositeIndex[conversation.ID, message.ID](storeIdxConvMsg)

	management = Management{
		WorkspaceUseCases:    workspace.NewUseCases(cfg.EventBus(), repoWorkspace, repoAgents),
		AgentUseCases:        agent.NewUseCases(cfg.EventBus(), repoAgents),
		ConversationUseCases: conversation.NewUseCases(cfg.EventBus(), repoConversation, repoWorkspace, repoAgents, repoMessages, idxConvMsg),
		Pages: uiai.Pages{
			Workspaces: "admin/ai/workspaces",
			Agents:     "admin/ai/workspace",
			Agent:      "admin/ai/workspace/agent",
		},
	}

	ucMistral := mistralai.NewUseCases(
		cfg.EventBus(),
		repoWorkspace.Name(),
		syncAgentRepo,
		syncConvRepo,
		secrets.UseCases.Match,
		management.ConversationUseCases.FindAll,
		management.WorkspaceUseCases.FindWorkspacesByPlatform,
		management.AgentUseCases.FindByID,
	)

	xsync.Go(func() error {
		slog.Info("sync Mistral AI workspaces")
		if err := ucMistral.Sync(user.SU()); err != nil {
			slog.Error("failed to sync Mistral AI workspaces", "err", err.Error())
		} else {
			slog.Info("sync Mistral AI workspaces successful")
		}

		return nil
	}, nil)

	cfg.RootViewWithDecoration(management.Pages.Workspaces, func(wnd core.Window) core.View {
		return uiai.PageWorkspaces(wnd, management.WorkspaceUseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.Agents, func(wnd core.Window) core.View {
		return uiai.PageWorkspace(wnd, management.WorkspaceUseCases, management.AgentUseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.Agent, func(wnd core.Window) core.View {
		return uiai.PageAgent(wnd, management.WorkspaceUseCases, management.AgentUseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		if !subject.HasPermission(workspace.PermCreate) {
			return admin.Group{}
		}

		return admin.Group{
			Title: "AI",
			Entries: []admin.Card{
				{
					Title:  StrAdminAIWorkspacesTitle.Get(subject),
					Text:   StrAdminAIWorkspacesDesc.Get(subject),
					Target: management.Pages.Workspaces,
				},
			},
		}
	})

	cfg.AddContextValue(core.ContextValue("nago.ai", management))
	cfg.AddContextValue(core.ContextValue("nago.ai.platforms", form.AnyUseCaseList[workspace.Platform, workspace.Platform](func(subject auth.Subject) iter.Seq2[workspace.Platform, error] {
		return xslices.ValuesWithError([]workspace.Platform{workspace.OpenAI, workspace.MistralAI}, nil)
	})))

	cfg.AddContextValue(core.ContextValue("nago.ai.agent.capabilities", form.AnyUseCaseList[agent.Capability, agent.Capability](func(subject auth.Subject) iter.Seq2[agent.Capability, error] {
		return xslices.ValuesWithError(agent.Capabilities.Clone(), nil)
	})))

	cfg.AddContextValue(core.ContextValue("nago.ai.agent.models", form.AnyUseCaseList[agent.Model, agent.Model](func(subject auth.Subject) iter.Seq2[agent.Model, error] {
		return xslices.ValuesWithError(agent.Models.Clone(), nil)
	})))

	cfg.AddContextValue(core.ContextValue("", management.ConversationUseCases.Start))

	slog.Info("installed AI module")
	return management, nil
}
