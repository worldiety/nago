// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"iter"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/ai/agent"
	uiai "go.wdy.de/nago/application/ai/ui"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrAdminAIWorkspacesTitle = i18n.MustString("nago.ai.admin.workspaces_title", i18n.Values{language.English: "AI Workspaces", language.German: "KI Workspaces"})
	StrAdminAIWorkspacesDesc  = i18n.MustString("nago.ai.admin.workspaces_desc", i18n.Values{language.English: "Manage the available AI workspaces. A workspace defines agents, libraries and prompts.", language.German: "Verwalte die verfügbaren KI Workspaces, die insbesondere KI Agenten, Bibliotheksquellen und Anweisungen enthalten."})
)

type Management struct {
	AgentUseCases     agent.UseCases
	WorkspaceUseCases workspace.UseCases
	Pages             uiai.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	repoWorkspace, err := application.JSONRepository[workspace.Workspace](cfg, "nago.ai.workspace")
	if err != nil {
		return Management{}, err
	}

	repoAgents, err := application.JSONRepository[agent.Agent](cfg, "nago.ai.agent")
	if err != nil {
		return Management{}, err
	}

	management = Management{
		WorkspaceUseCases: workspace.NewUseCases(repoWorkspace, repoAgents),
		Pages: uiai.Pages{
			Workspaces: "admin/ai/workspaces",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.Workspaces, func(wnd core.Window) core.View {
		return uiai.PageWorkspaces(wnd, management.WorkspaceUseCases)
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
	return management, nil
}
