// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dropdown"
	"go.wdy.de/nago/presentation/ui/markdown"
)

// statelessChat is a minimal example for the stateless completion API: a multiline text field as input and a
// markdown rich text as output. It picks the first configured provider that supports Completions (e.g. an
// Anthropic secret) and its first available model.
func statelessChat(wnd core.Window, uc ai.UseCases) core.View {
	// Collect all providers that expose stateless completions so the user can pick one.
	type provEntry struct {
		prov  provider.Provider
		comps completion.Completions
	}

	var entries []provEntry
	for p, err := range uc.FindAllProvider(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		if c := p.Completions(); c.IsSome() {
			entries = append(entries, provEntry{prov: p, comps: c.Unwrap()})
		}
	}

	if len(entries) == 0 {
		return alert.BannerError(fmt.Errorf("kein Provider mit stateless Completions gefunden – bitte ein Anthropic-Secret konfigurieren"))
	}

	selectedProvider := core.AutoState[provider.ID](wnd).Init(func() provider.ID {
		return entries[0].prov.Identity()
	})

	// Resolve the currently selected provider (fallback to the first one).
	current := entries[0]
	for _, e := range entries {
		if e.prov.Identity() == selectedProvider.Get() {
			current = e
			break
		}
	}
	prov := current.prov
	comps := current.comps

	prompt := core.AutoState[string](wnd)
	answer := core.AutoState[string](wnd)
	busy := core.AutoState[bool](wnd)
	selectedModel := core.AutoState[model.ID](wnd).Init(func() model.ID {
		for m, err := range comps.Models(wnd.Subject()) {
			if err != nil {
				return ""
			}
			return m.ID
		}
		return ""
	})

	// When the provider changes, reset the model to the first model of the new provider.
	selectedProvider.Observe(func(newValue provider.ID) {
		first := model.ID("")
		for _, e := range entries {
			if e.prov.Identity() == newValue {
				for m, err := range e.comps.Models(wnd.Subject()) {
					if err == nil {
						first = m.ID
					}
					break
				}
				break
			}
		}
		selectedModel.Set(first)
		answer.Set("")
	})

	// Build the dropdown options from the available completion providers.
	providerOptions := make([]dropdown.Option[provider.ID], 0, len(entries))
	for _, e := range entries {
		providerOptions = append(providerOptions, dropdown.Option[provider.ID]{
			Value: e.prov.Identity(),
			Label: e.prov.Name(),
		})
	}


	submit := func() {
		question := strings.TrimSpace(prompt.Get())
		if question == "" || busy.Get() {
			return
		}

		busy.Set(true)
		answer.Set("")

		xsync.Go(func() error {
			res, err := comps.Complete(wnd.Subject(), completion.Options{
				Model: selectedModel.Get(),
				Messages: []completion.Message{
					{
						Role:    completion.User,
						Content: []completion.Content{completion.Text{Text: question}},
					},
				},
			})

			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					alert.ShowBannerError(wnd, err)
				})
				return nil
			}

			var sb strings.Builder
			for _, c := range res.Message.Content {
				if t, ok := c.(completion.Text); ok {
					sb.WriteString(t.Text)
				}
			}

			wnd.Post(func() {
				answer.Set(sb.String())
				busy.Set(false)
			})

			return nil
		}, func(err error) {
			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					alert.ShowBannerError(wnd, err)
				})
			}
		})
	}

	return ui.VStack(
		ui.Text(fmt.Sprintf("Stateless Chat – %s (%s)", prov.Name(), selectedModel.Get())).Font(ui.Title),

		dropdown.Dropdown("Provider", providerOptions, selectedProvider.Get()).
			InputValue(selectedProvider).
			Disabled(busy.Get()).
			Frame(ui.Frame{}.FullWidth()),

		ui.TextField("Deine Eingabe", prompt.Get()).
			InputValue(prompt).
			Lines(6).
			FullWidth().
			Disabled(busy.Get()),

		ui.PrimaryButton(submit).
			Title("Senden").
			Enabled(!busy.Get()),

		ui.If(busy.Get(), ui.Text("… die KI denkt nach")),

		ui.If(answer.Get() != "", ui.VStack(
			markdown.RichText(answer.Get()),
		).Alignment(ui.Leading).
			FullWidth().
			BackgroundColor(ui.M2).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L16))),
	).Alignment(ui.Leading).
		Gap(ui.L16).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L16))
}

