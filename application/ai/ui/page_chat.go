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
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/markdown"
	"golang.org/x/text/language"
)

var (
	StrHowCanXHelp       = i18n.MustVarString("nago.ai.chat.help_x", i18n.Values{language.English: "How can {name} help you today?", language.German: "Wie kann {name} Dir heute weiterhelfen?"})
	StrHowItWorksDetails = i18n.MustString("nago.ai.chat.how_it_works_details", i18n.Values{
		language.English: `## 1. Describe the issue and ask your question.
Write down what the issue is in your own words – the more specific you are, the more accurate the answer will be.

## 2. Upload any relevant documents.
Images, presentations, tables, photos, etc. It is best to use digitally generated PDFs, not scans. Vary your strategies when uploading to get the best results.

## 3. Wait for the result.
The AI analyses your request in seconds and provides you with the best context-specific advice on how to proceed.

**Tip**: You can ask questions, add details or continue the chat at any time – the AI stays on topic.

Translated with DeepL.com (free version)`,

		language.German: `## 1. Sachverhalt beschreiben und Frage stellen.
Schreib in deinen Worten, worum es geht – je konkreter du Inhalte beschreibst, desto präziser die Antwort.

## 2. Dokument(e) hochladen, wenn vorhanden.
Bilder, Präsentationen, Tabellen, Fotos, etc. Verwende am besten digital erzeugte PDFs, keine Scans. Variiere deine Strategien beim Hochladen, um das beste Ergebnis zu erhalten.

## 3. Ergebnis abwarten.
Die KI analysiert dein Anliegen in Sekunden und liefert dir die im Kontext besten Hinweise wie du weiter machen kannst.

**Tipp**: Du kannst jederzeit nachfragen, Details ergänzen oder den Chatverlauf fortsetzen – die KI bleibt im Thema.`,
	})
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

	helpPresented := core.AutoState[bool](wnd)
	const innerFullHeight = "calc(100vh - 6rem - 1px)" // remember: this only works with NoFooter option for the according page path
	large := wnd.Info().SizeClass >= core.SizeClassMedium
	return ui.HStack(
		dialogHelp(wnd, helpPresented),
		ui.If(
			large,
			ui.ScrollView(
				Chats(prov, conv).Frame(ui.Frame{Width: ui.Full, MinHeight: innerFullHeight}),
			).Axis(ui.ScrollViewAxisVertical).Frame(ui.Frame{Height: ui.Full, Width: ui.L320}),
		),

		ui.ScrollView(
			Chat(prov, conv, prompt).
				StartOptions(StartOptions{
					AgentName:  wnd.Values()["agent-name"],
					Agent:      agent.ID(wnd.Values()["agent"]),
					CloudStore: true,
				}).
				Teaser(teaser(wnd, prov)).
				More(
					ui.SecondaryButton(func() {
						helpPresented.Set(true)
					}).Title(rstring.LabelHowItWorks.Get(wnd)).Frame(ui.Frame{MinWidth: "12rem"}),
				).
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

func dialogHelp(wnd core.Window, presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	body := markdown.Render(markdown.Options{RichText: true, TrimParagraph: true}, []byte(StrHowItWorksDetails.Get(wnd)))
	return alert.Dialog(rstring.LabelHowItWorks.Get(wnd), body, presented, alert.Ok(), alert.Larger())
}
