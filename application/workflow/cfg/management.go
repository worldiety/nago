// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgworkflow

import (
	"fmt"
	"log/slog"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/workflow"
	uiworkflow "go.wdy.de/nago/application/workflow/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

type Management struct {
	UseCases workflow.UseCases
	Pages    uiworkflow.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	instanceStore, err := cfg.EntityStore("nago.workflow.instance")
	if err != nil {
		return Management{}, fmt.Errorf("cannot open instance store: %w", err)
	}

	eventsStore, err := cfg.EntityStore("nago.workflow.instance.event")
	if err != nil {
		return Management{}, fmt.Errorf("cannot open instance event store: %w", err)
	}

	management = Management{
		UseCases: workflow.NewUseCases(cfg.EventBus(), instanceStore, eventsStore),
		Pages: uiworkflow.Pages{
			PageWorkflow:               "admin/workflow",
			PageWorkflowInstanceEvents: "admin/workflow/instance/events",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.PageWorkflow, func(wnd core.Window) core.View {
		return uiworkflow.PageWorkflow(wnd, management.UseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.PageWorkflowInstanceEvents, func(wnd core.Window) core.View {
		return uiworkflow.PageInstanceEvents(wnd, management.UseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		var cards []admin.Card
		for wf, err := range management.UseCases.FindDeclaredWorkflows(subject) {
			if err != nil {
				slog.Error("failed to find enumerate workflow", "err", err.Error())
				continue
			}

			cards = append(cards, admin.Card{
				Title:  wf.Name,
				Text:   wf.Description,
				Target: management.Pages.PageWorkflow,
				TargetParams: core.Values{
					"id": string(wf.ID),
				},
			})
		}
		group := admin.Group{
			Title:   "Arbeitsabläufe",
			Entries: cards,
		}

		return group
	})

	cfg.AddContextValue(core.ContextValue("nago.workflows", management))

	slog.Info("installed workflow management")

	return management, nil
}
