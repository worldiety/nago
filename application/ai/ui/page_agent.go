// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"
	"iter"
	"os"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrAgentUpdated      = i18n.MustString("nago.ai.admin.agent.updated", i18n.Values{language.English: "Agent updated", language.German: "Agent aktualisiert"})
	StrAgentUpdatedDesc  = i18n.MustString("nago.ai.admin.agent.updated_desc", i18n.Values{language.English: "The Agent has been updated.", language.German: "Der Agent wurde erfolgreich aktualisiert."})
	StrAgentTemperature  = i18n.MustString("nago.ai.admin.agent.temperature", i18n.Values{language.English: "Temperature", language.German: "Reproduzierbarkeit"})
	StrAgentInstructions = i18n.MustString("nago.ai.admin.agent.instructions", i18n.Values{language.English: "Instructions", language.German: "Instruktionen"})
)

func PageAgent(wnd core.Window, uc ai.UseCases) core.View {
	optProv, err := uc.FindProviderByID(wnd.Subject(), provider.ID(wnd.Values()["provider"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optProv.IsNone() {
		return alert.BannerError(fmt.Errorf("provider not found: %s: %w", wnd.Values()["provider"], os.ErrNotExist))
	}

	prov := optProv.Unwrap()
	optAgents := prov.Agents()
	if optAgents.IsNone() {
		return alert.BannerError(fmt.Errorf("provider does not support agents: %s", wnd.Values()["provider"]))
	}

	agents := optAgents.Unwrap()
	agentID := agent.ID(wnd.Values()["agent"])
	ag := core.AutoState[agent.Agent](wnd).AsyncInit(func() agent.Agent {
		optAg, err := agents.FindByID(wnd.Subject(), agentID)
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return agent.Agent{}
		}

		if optAg.IsNone() {
			alert.ShowBannerError(wnd, fmt.Errorf("agent not found: %s.%s: %w", prov.Identity(), agentID, os.ErrNotExist))
			return agent.Agent{}
		}

		return optAg.Unwrap()
	})

	return ui.VStack(
		breadcrumb.Breadcrumbs(
			ui.TertiaryButton(func() {
				wnd.Navigation().BackwardTo("admin/ai/provider", wnd.Values())
			}).Title(StrLibraries.Get(wnd)),
			ui.TertiaryButton(nil).Title(StrLibrary.Get(wnd)),
		),
		ui.H1(StrLibrary.Get(wnd)),
		ui.IfFunc(ag.Valid(), func() core.View {
			return formAgentSettings(wnd, prov, agents, ag)
		}),
	).Alignment(ui.Leading).Frame(ui.Frame{}.Larger())
}

func formAgentSettings(wnd core.Window, prov provider.Provider, agents provider.Agents, ag *core.State[agent.Agent]) core.View {
	type EditForm struct {
		Name         string            `label:"nago.common.label.name"`
		Description  string            `label:"nago.common.label.description" lines:"3"`
		Instructions string            `label:"nago.ai.admin.agent.instructions" lines:"5"`
		Temperature  agent.Temperature `label:"nago.ai.admin.agent.temperature"`
		Libraries    []library.ID      `label:"nago.ai.admin.libraries" source:"prov-libs"`
	}

	ctx := core.WithContext(wnd.Context(), core.ContextValue("prov-libs", form.AnyUseCaseList(func(subject auth.Subject) iter.Seq2[library.Library, error] {
		if prov.Libraries().IsSome() {
			return prov.Libraries().Unwrap().All(subject)
		}

		return func(yield func(library.Library, error) bool) {}
	})))

	cfg := core.AutoState[EditForm](wnd).Init(func() EditForm {
		return EditForm{
			Name:         ag.Get().Name,
			Description:  ag.Get().Description,
			Temperature:  ag.Get().Temperature,
			Instructions: ag.Get().Instructions,
			Libraries:    ag.Get().Libraries,
		}
	})

	return ui.VStack(
		form.Card(
			form.Auto(form.AutoOptions{Context: ctx}, cfg).FullWidth(),
		),
		ui.HLine(),
		ui.HStack(ui.SecondaryButton(func() {
			frm := cfg.Get()
			info, err := agents.Agent(ag.Get().ID).Update(wnd.Subject(), agent.UpdateOptions{
				Name:         option.Some(frm.Name),
				Description:  option.Some(frm.Description),
				Temperature:  option.Some(frm.Temperature),
				Instructions: option.Some(frm.Instructions),
				Libraries:    option.Some(frm.Libraries),
			})
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}
			ag.Set(info)
			alert.ShowBannerMessage(wnd, alert.Message{
				Title:   StrAgentUpdated.Get(wnd),
				Message: StrAgentUpdatedDesc.Get(wnd),
				Intent:  alert.IntentOk,
			})
		}).Title(rstring.ActionSave.Get(wnd))).FullWidth().Alignment(ui.Trailing),
	).FullWidth()
}
