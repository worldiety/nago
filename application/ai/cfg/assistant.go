// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"fmt"
	"log/slog"
	"sync"

	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// Assistant bundles everything an application needs to put a working AI assistant on its screens, so that it
// does not have to reproduce provider lookup, model resolution, settings, caching and diagnostics itself.
//
// Obtain it from [Management.Assistant] and call [Assistant.Button].
type Assistant struct {
	useCases ai.UseCases
	sessions session.UseCases

	models modelCache

	// complained remembers which kind of problem has already been logged. A missing token must not become a
	// log line per page render, nor a banner that follows the user around the application.
	complained sync.Map
}

// AssistantOptions is what an application still has to say about its own assistant: what it is called, and
// what it can do.
//
// Everything else - which provider, which model, how long an answer may get, whether it may write, whether it
// is visible at all - is an operator decision and comes from [AssistantSettings].
type AssistantOptions struct {
	// Title is the panel heading. Optional; defaults to the provider name.
	Title string

	// Tags scope the persisted history. Use one tag for the whole application if the assistant follows the
	// user from screen to screen: a conversation started on one page and continued on another is one
	// conversation, not two. Optional.
	Tags []string

	// Agents are the selectable personas with their prompts and tools. Usually exactly one.
	Agents []uicompletion.Agent

	// MaxTurns bounds the agentic loop per question. Optional.
	MaxTurns int

	// History persists conversations and offers a restore button. Optional.
	History bool

	// FileUpload lets the user attach files to a message. Optional.
	FileUpload bool

	// AskUser lets the model ask a clarifying question mid-run. Optional.
	AskUser bool

	// DisableCurrentTime removes the built-in current_time tool, which tells the model today's date and the
	// current time. Optional; the tool is on by default.
	DisableCurrentTime bool

	// ConfirmMutations asks the user before any tool marked [completion.Tool.Mutating] runs. Strongly
	// recommended for any assistant that can change something. Optional.
	ConfirmMutations bool

	// Corner places the floating button. Optional; defaults to the bottom right.
	Corner uicompletion.Corner

	// Label is the button caption. Optional.
	Label string
}

// Button returns the floating assistant button, or nil when the assistant cannot or should not run right now:
// nobody is signed in, an operator hid it, no provider is configured, or no model could be determined.
//
// Returning nil rather than an error view is deliberate - the button lives in a decorator and appears on every
// screen, so a misconfiguration must not turn into a banner on every page. The reason is logged instead, once
// per kind, with enough detail to act on. Use [uicompletion.MissingPermissions] if you want to surface a
// missing role in your own UI.
func (a *Assistant) Button(wnd core.Window, opts AssistantOptions) core.View {
	if wnd.Subject() == nil || !wnd.Subject().Valid() {
		return nil
	}

	cfg := core.GlobalSettings[AssistantSettings](wnd)
	if cfg.Hidden {
		return nil
	}

	// Diagnose the most common cause first and by name, because it is invisible from the UI: without the
	// framework permissions the provider lookup below simply yields nothing, which looks exactly like "no
	// provider configured" and sends whoever debugs it to the wrong place.
	if missing := uicompletion.MissingPermissions(wnd.Subject()); len(missing) > 0 {
		a.complain("permissions", fmt.Sprintf(
			"the AI assistant stays hidden because the acting role lacks %v; assign the %q role to the users who should use it",
			missing, RoleAssistantUser))
		return nil
	}

	prov, comps, err := a.Provider(wnd.Subject())
	if err != nil {
		a.complain("provider", fmt.Sprintf("the AI assistant stays hidden: %v", err))
		return nil
	}

	modelID, err := a.ResolveModel(wnd.Subject(), cfg, comps)
	if err != nil {
		a.complain("model", fmt.Sprintf(
			"the AI assistant stays hidden because no model could be determined: %v; pick one in the global settings", err))
		return nil
	}

	agents := make([]uicompletion.Agent, len(opts.Agents))
	copy(agents, opts.Agents)
	for i := range agents {
		// The operator's choice wins over whatever the application hard-coded, unless the agent deliberately
		// pins its own model - a specialist agent may need a specific one.
		if agents[i].Model == "" {
			agents[i].Model = modelID
		}

		if agents[i].MaxTokens == 0 {
			agents[i].MaxTokens = cfg.MaxTokensOr(DefaultAssistantMaxTokens)
		}
	}

	button := uicompletion.ChatButton(uicompletion.ChatOptions{
		Sessions:           a.sessions,
		Completions:        comps,
		Provider:           prov,
		Title:              opts.Title,
		Tags:               opts.Tags,
		MaxTurns:           opts.MaxTurns,
		History:            opts.History,
		FileUpload:         opts.FileUpload,
		AskUser:            opts.AskUser,
		DisableCurrentTime: opts.DisableCurrentTime,
		ReadOnly:           cfg.ReadOnly,
		ConfirmMutations:   opts.ConfirmMutations,
		Agents:             agents,
	})

	button = button.Corner(opts.Corner)

	if opts.Label != "" {
		button = button.Label(opts.Label)
	}

	return button
}

// Decorate wraps a view with the assistant button, which is the usual way to make it available everywhere:
//
//	cfg.SetDecorator(func(wnd core.Window, view core.View) core.View {
//		return mgmt.Assistant.Decorate(wnd, scaffold(wnd, view), opts)
//	})
//
// Attaching it here rather than to each page means it genuinely appears on every screen, including the ones
// the framework brings along. When the assistant cannot run, the view is returned untouched.
func (a *Assistant) Decorate(wnd core.Window, view core.View, opts AssistantOptions) core.View {
	button := a.Button(wnd, opts)
	if button == nil {
		return view
	}

	return ui.VStack(view, button).FullWidth()
}

// Provider picks the first configured provider that can run a completion.
//
// There is no provider dropdown on purpose: the assistant is a fixture of the application, not a place to
// experiment. Which provider is configured is an operator decision and belongs in the administration.
//
// The ways this can fail are told apart deliberately. They were once one silent "no" whose message named the
// vault - so the first person to debug a missing assistant went looking at the secret while the actual cause
// was a role holding no framework permission. A diagnosis that points at the wrong place costs more than none.
func (a *Assistant) Provider(subject auth.Subject) (provider.Provider, completion.Completions, error) {
	configured := 0

	for p, err := range a.useCases.FindAllProvider(subject) {
		if err != nil {
			return nil, nil, fmt.Errorf("the acting role may not list AI providers (%s): %w", ai.PermFindAllProvider, err)
		}

		configured++

		if c := p.Completions(); c.IsSome() {
			return p, c.Unwrap(), nil
		}
	}

	if configured == 0 {
		return nil, nil, fmt.Errorf("no AI provider is configured; add a provider token in the vault")
	}

	return nil, nil, fmt.Errorf("%d AI provider(s) are configured, but none of them offers completions", configured)
}

// ResolveModel decides which model answers.
//
// The configured one wins and costs nothing at all. Without it the first the provider reports is used, which
// is a network call - hence the cache behind it. A failure is returned rather than swallowed: silently
// falling back turns an unreachable provider into "no model set", which again sends whoever has to fix it
// looking in the wrong place.
func (a *Assistant) ResolveModel(subject auth.Subject, cfg AssistantSettings, comps completion.Completions) (model.ID, error) {
	if cfg.Model != "" {
		return cfg.Model, nil
	}

	models, err := a.Models(subject, comps)
	if err != nil {
		return "", err
	}

	if len(models) == 0 {
		return "", fmt.Errorf("the provider reports no models")
	}

	return models[0].ID, nil
}

// Models lists what the provider offers, at most once per cache interval.
//
// It is the one place that asks. The button needs it to pick a default and the settings picker needs it to
// offer a choice, and both are rendered often enough that asking twice would be twice too many.
func (a *Assistant) Models(subject auth.Subject, comps completion.Completions) ([]model.Model, error) {
	return a.models.list(func() ([]model.Model, error) {
		var out []model.Model

		for m, err := range comps.Models(subject) {
			if err != nil {
				return nil, fmt.Errorf("the provider's model list is unreachable, which usually means the API token is wrong: %w", err)
			}

			out = append(out, m)
		}

		return out, nil
	})
}

// Forget drops the cached model list so the next caller asks the provider again. Called automatically when
// the settings or a secret change.
func (a *Assistant) Forget() {
	a.models.forget()
}

// complain logs a problem once per kind.
func (a *Assistant) complain(kind, message string) {
	if _, seen := a.complained.LoadOrStore(kind, true); seen {
		return
	}

	slog.Warn(message)
}
