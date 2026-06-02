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
	"go.wdy.de/nago/presentation/ui/markdown"
)

// statelessChat is a minimal example for the stateless completion API: a multiline text field as input and a
// markdown rich text as output. It picks the first configured provider that supports Completions (e.g. an
// Anthropic secret) and its first available model.
func statelessChat(wnd core.Window, uc ai.UseCases) core.View {
	var prov provider.Provider
	var comps completion.Completions

	for p, err := range uc.FindAllProvider(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		if c := p.Completions(); c.IsSome() {
			prov = p
			comps = c.Unwrap()
			break
		}
	}

	if comps == nil {
		return alert.BannerError(fmt.Errorf("kein Provider mit stateless Completions gefunden – bitte ein Anthropic-Secret konfigurieren"))
	}

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

