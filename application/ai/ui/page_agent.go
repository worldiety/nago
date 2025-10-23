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
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrPrompt       = i18n.MustString("nago.ai.admin.prompt", i18n.Values{language.English: "System Prompt", language.German: "Anweisung"})
	StrModel        = i18n.MustString("nago.ai.admin.model", i18n.Values{language.English: "Model", language.German: "Modell"})
	StrTemperature  = i18n.MustString("nago.ai.admin.temperature", i18n.Values{language.English: "Temperature", language.German: "Zufälligkeit"})
	StrCapabilities = i18n.MustString("nago.ai.admin.capabilities", i18n.Values{language.English: "Capabilities", language.German: "Fähigkeiten"})

	StrFunctions = i18n.MustString("nago.ai.admin.functions", i18n.Values{language.English: "Functions", language.German: "Funktionen"})
)

func PageAgent(wnd core.Window, ucWS workspace.UseCases, ucAgents agent.UseCases) core.View {
	optWS, err := ucWS.FindByID(wnd.Subject(), workspace.ID(wnd.Values()["workspace"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optWS.IsNone() {
		return alert.BannerError(fmt.Errorf("workspace not found: %w", os.ErrNotExist))
	}

	ws := optWS.Unwrap()

	optAg, err := ucAgents.FindByID(wnd.Subject(), agent.ID(wnd.Values()["agent"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optAg.IsNone() {
		return alert.BannerError(fmt.Errorf("agent not found: %w", os.ErrNotExist))
	}

	ag := optAg.Unwrap()

	return viewEditAgent(wnd, ucWS, ucAgents, ws, ag)
}

type AgentForm struct {
	ID           agent.ID           `visible:"false"`
	Name         string             `label:"nago.common.label.name"`
	Description  string             `label:"nago.common.label.description" lines:"3"`
	Prompt       string             `label:"nago.ai.admin.prompt" lines:"5"`
	Model        agent.Model        `label:"nago.ai.admin.model" source:"nago.ai.agent.models"`
	Libraries    []library.ID       `label:"nago.ai.admin.libraries" visible:"false"`
	Capabilities []agent.Capability `label:"nago.ai.admin.capabilities" source:"nago.ai.agent.capabilities"`
	Temperature  agent.Temperature  `label:"nago.ai.admin.temperature"`
	System       bool               `visible:"false"`
}

func (f AgentForm) IntoAgent() agent.Agent {
	return agent.Agent{
		ID:           f.ID,
		Name:         f.Name,
		Description:  f.Description,
		Prompt:       f.Prompt,
		Model:        f.Model,
		Libraries:    f.Libraries,
		Capabilities: f.Capabilities,
		Temperature:  f.Temperature,
		System:       f.System,
	}
}

func viewEditAgent(wnd core.Window, ucWS workspace.UseCases, ucAgents agent.UseCases, ws workspace.Workspace, ag agent.Agent) core.View {
	agentForm := core.AutoState[AgentForm](wnd).Init(func() AgentForm {
		return AgentForm{
			ID:           ag.ID,
			Name:         ag.Name,
			Description:  ag.Description,
			Prompt:       ag.Prompt,
			Model:        ag.Model,
			Libraries:    ag.Libraries,
			Capabilities: ag.Capabilities,
			Temperature:  ag.Temperature,
			System:       ag.System,
		}
	})

	return ui.VStack(
		ui.H1(StrAgent.Get(wnd)),

		breadcrumb.Breadcrumbs(
			ui.TertiaryButton(func() {
				wnd.Navigation().BackwardTo("admin/ai/workspaces", wnd.Values())
			}).Title(StrWorkspaces.Get(wnd)),
			ui.TertiaryButton(func() {
				wnd.Navigation().BackwardTo("admin/ai/workspace", wnd.Values())
			}).Title(ws.Name),
		).ClampLeading(),

		ui.Space(ui.L16),

		form.Auto(form.AutoOptions{}, agentForm),
		ui.HLine(),
		ui.HStack(
			ui.SecondaryButton(func() {
				wnd.Navigation().BackwardTo("admin/ai/workspace", wnd.Values())
			}).Title(rstring.ActionCancel.Get(wnd)),
			ui.PrimaryButton(func() {
				if err := ucAgents.Update(wnd.Subject(), agentForm.Get().IntoAgent()); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				wnd.Navigation().BackwardTo("admin/ai/workspace", wnd.Values())
			}).Title(rstring.ActionSave.Get(wnd)),
		).Alignment(ui.Trailing).FullWidth().Gap(ui.L8),
	).Alignment(ui.Leading).Frame(ui.Frame{MaxWidth: ui.L880, Width: ui.Full})
}
