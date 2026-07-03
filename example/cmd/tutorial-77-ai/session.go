// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dropdown"
	"go.wdy.de/nago/presentation/ui/markdown"
)

// sessionChat demonstrates persistable chat sessions on top of the stateless completion API. In contrast to
// statelessChat (which forgets everything after each request), this view stores the full history in a
// [session.Session] via [session.UseCases]. Reloading the same session id restores the whole conversation
// from the repository - it survives navigation and restarts.
func sessionChat(wnd core.Window, uc ai.UseCases, sessions session.UseCases) core.View {
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
		return alert.BannerError(fmt.Errorf("kein Provider mit stateless Completions gefunden – bitte ein Secret konfigurieren"))
	}

	selectedProvider := core.AutoState[provider.ID](wnd).Init(func() provider.ID {
		return entries[0].prov.Identity()
	})

	// completionsFor resolves the provider entry for a provider id, falling back to the first entry.
	completionsFor := func(pid provider.ID) provEntry {
		for _, e := range entries {
			if e.prov.Identity() == pid {
				return e
			}
		}
		return entries[0]
	}

	// The currently opened session id. Empty means "no session yet". This state is bound to the session
	// dropdown, so selecting an entry there opens (continues) that persisted session.
	currentSession := core.AutoState[session.ID](wnd)
	prompt := core.AutoState[string](wnd)
	busy := core.AutoState[bool](wnd)
	// reloadTick is bumped to force a re-read of the session history from the repository after each append.
	reloadTick := core.AutoState[int](wnd)

	// If a session is currently open, load it once up front: it is the single source of truth for which
	// provider and model are in effect (an open session pins both). Only when no session is open do the
	// provider/model dropdowns drive the selection. Reading reloadTick here makes this reload after every
	// append/create.
	_ = reloadTick.Get()
	var openSession option.Opt[session.Session]
	if sid := currentSession.Get(); sid != "" {
		optSession, err := sessions.FindByID(wnd.Subject(), sid)
		if err != nil {
			return alert.BannerError(err)
		}
		openSession = optSession
	}

	// Resolve the effective provider: the open session's provider wins over the dropdown.
	effectiveProviderID := selectedProvider.Get()
	if openSession.IsSome() && openSession.Unwrap().ProviderHint != "" {
		effectiveProviderID = provider.ID(openSession.Unwrap().ProviderHint)
	}
	current := completionsFor(effectiveProviderID)
	prov := current.prov
	comps := current.comps

	selectedModel := core.AutoState[model.ID](wnd).Init(func() model.ID {
		for m, err := range comps.Models(wnd.Subject()) {
			if err != nil {
				return ""
			}
			return m.ID
		}
		return ""
	})

	// Effective model: the open session's model wins over the dropdown.
	effectiveModel := selectedModel.Get()
	if openSession.IsSome() && openSession.Unwrap().Model != "" {
		effectiveModel = openSession.Unwrap().Model
	}

	// Deliberately NO reset observers on selectedProvider/selectedModel here.
	//
	// Earlier versions reset currentSession from within those observers. That is unreliable with the
	// controlled-state model: restoring a session programmatically calls selectedModel.Set(...), and on the
	// next frontend roundtrip (e.g. when typing into prompt) the frontend replays selectedModel's changed
	// value through its observer - which then wrongly cleared the just-selected session. Instead, consistency
	// between the running provider/model and the open session is enforced declaratively in submit().

	// createSession creates a fresh, empty session and opens it. It uses the currently effective provider and
	// model (what the title shows), so the new session always starts from a valid provider/model pair.
	createSession := func() {
		s, err := sessions.Create(wnd.Subject(), session.CreateOptions{
			Title:        "Chat",
			Model:        effectiveModel,
			System:       "You are a helpful assistant. Keep answers concise.",
			ProviderHint: string(prov.Identity()),
		})
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return
		}
		currentSession.Set(s.ID)
		reloadTick.Invalidate()
		slog.Info("created new session", "id", s.ID, "provider", prov.Identity(), "model", effectiveModel)
	}

	submit := func() {
		question := strings.TrimSpace(prompt.Get())
		if question == "" || busy.Get() {
			return
		}

		// Lazily create a session on the first message.
		if currentSession.Get() == "" {
			createSession()
			if currentSession.Get() == "" {
				return // creation failed, banner already shown
			}
		}

		sid := currentSession.Get()
		busy.Set(true)
		prompt.Set("")

		xsync.Go(func() error {
			_, err := sessions.Append(wnd.Subject(), sid, session.AppendOptions{
				Completions: comps,
				Model:       effectiveModel,
				Input:       []completion.Content{completion.Text{Text: question}},
				MaxTokens:   1024,
			})
			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					alert.ShowBannerError(wnd, err)
				})
				return nil
			}

			wnd.Post(func() {
				busy.Set(false)
				reloadTick.Invalidate()
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

	// The persisted history of the currently opened session (already loaded above into openSession).
	var history []completion.Message
	if openSession.IsSome() {
		history = openSession.Unwrap().Messages
	}

	// Build the session dropdown so the user can switch between and continue existing sessions. The empty
	// option represents "no session / new chat".
	sessionOptions := []dropdown.Option[session.ID]{
		{Value: "", Label: "– neue Unterhaltung –"},
	}
	for s, err := range sessions.FindAll(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}
		sessionOptions = append(sessionOptions, dropdown.Option[session.ID]{
			Value: s.ID,
			Label: sessionLabel(s),
		})
	}

	// Build the provider dropdown.
	providerOptions := make([]dropdown.Option[provider.ID], 0, len(entries))
	for _, e := range entries {
		providerOptions = append(providerOptions, dropdown.Option[provider.ID]{
			Value: e.prov.Identity(),
			Label: e.prov.Name(),
		})
	}

	// Build the model dropdown for the currently selected provider.
	var modelOptions []dropdown.Option[model.ID]
	for m, err := range comps.Models(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}
		label := m.Name
		if label == "" {
			label = string(m.ID)
		}
		modelOptions = append(modelOptions, dropdown.Option[model.ID]{
			Value: m.ID,
			Label: label,
		})
	}

	sessionOpen := openSession.IsSome()

	return ui.VStack(
		ui.Text(fmt.Sprintf("Persistente Session – %s (%s)", prov.Name(), effectiveModel)).Font(ui.Title),

		ui.HStack(
			dropdown.Dropdown("Session", sessionOptions, currentSession.Get()).
				InputValue(currentSession).
				Disabled(busy.Get()),

			// While a session is open it pins provider and model, so these dropdowns are read-only mirrors of
			// the session's configuration. Use "Neue Session" (or the empty session option) to choose a
			// different provider/model.
			dropdown.Dropdown("Provider", providerOptions, effectiveProviderID).
				InputValue(selectedProvider).
				Disabled(busy.Get() || sessionOpen),

			dropdown.Dropdown("Modell", modelOptions, effectiveModel).
				InputValue(selectedModel).
				Disabled(busy.Get() || sessionOpen),

			ui.SecondaryButton(createSession).Title("Neue Session"),
		).Gap(ui.L8).FullWidth().Alignment(ui.Leading),

		ui.If(sessionOpen,
			ui.Text(fmt.Sprintf("Session-ID: %s  (im Dropdown wieder auswählbar)", currentSession.Get())).
				Font(ui.Small)),

		// Render the full persisted conversation.
		renderHistory(history),

		ui.TextField("Deine Eingabe", prompt.Get()).
			InputValue(prompt).
			Lines(4).
			FullWidth().
			Disabled(busy.Get()),

		ui.PrimaryButton(submit).
			Title("Senden").
			Enabled(!busy.Get()),

		ui.If(busy.Get(), ui.Text("… die KI denkt nach")),
	).Alignment(ui.Leading).
		Gap(ui.L16).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L16))
}

// sessionLabel builds a concise, human-readable label for a session dropdown entry: a preview of the first
// user message (or the title) plus a short id suffix to keep entries distinguishable.
func sessionLabel(s session.Session) string {
	preview := ""
	for _, msg := range s.Messages {
		if msg.Role != completion.User {
			continue
		}
		for _, c := range msg.Content {
			if t, ok := c.(completion.Text); ok {
				preview = t.Text
				break
			}
		}
		if preview != "" {
			break
		}
	}

	preview = strings.TrimSpace(strings.ReplaceAll(preview, "\n", " "))
	if len([]rune(preview)) > 40 {
		preview = string([]rune(preview)[:40]) + "…"
	}

	short := string(s.ID)
	if len(short) > 8 {
		short = short[:8]
	}

	switch {
	case preview != "":
		return fmt.Sprintf("%s (%s)", preview, short)
	case s.Title != "":
		return fmt.Sprintf("%s (%s)", s.Title, short)
	default:
		return fmt.Sprintf("Session %s", short)
	}
}

// renderHistory renders a persisted completion history as a simple chat transcript.
func renderHistory(history []completion.Message) core.View {
	if len(history) == 0 {
		return ui.Text("Noch keine Nachrichten in dieser Session.").Font(ui.Small)
	}

	var bubbles []core.View
	for _, msg := range history {
		var sb strings.Builder
		for _, c := range msg.Content {
			switch v := c.(type) {
			case completion.Text:
				sb.WriteString(v.Text)
			case completion.Thinking:
				sb.WriteString("_(denkt: " + v.Text + ")_")
			case completion.ToolCall:
				fmt.Fprintf(&sb, "`ruft Tool %s(%s)`", v.Name, string(v.Arguments))
			case completion.ToolResult:
				for _, ic := range v.Content {
					if t, ok := ic.(completion.Text); ok {
						fmt.Fprintf(&sb, "`Tool-Ergebnis: %s`", t.Text)
					}
				}
			}
		}

		bg := ui.M2
		role := "Du"
		if msg.Role == completion.Assistant {
			bg = ui.M3
			role = "KI"
		}

		bubbles = append(bubbles, ui.VStack(
			ui.Text(role).Font(ui.Small),
			markdown.RichText(sb.String()),
		).Alignment(ui.Leading).
			FullWidth().
			BackgroundColor(bg).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L8)))
	}

	return ui.VStack(bubbles...).
		Gap(ui.L8).
		FullWidth().
		Alignment(ui.Leading)
}
