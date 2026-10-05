// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
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

	// models holds a *modelCache per provider.ID.
	models sync.Map

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

	// Confirmation decides whether the user approves each call of a tool marked [completion.Tool.Mutating]
	// before it runs, or only the calls of tools marked [completion.Tool.RequiresApproval]. Optional; the zero
	// value follows the operator's [AssistantSettings.SkipConfirmation].
	Confirmation Confirmation

	// Delegation lets the model hand independent tasks to sub-agents which work on them in parallel, see
	// [uicompletion.DelegationOptions]. Sub-agents are read-only whenever changes must be confirmed (see
	// Confirmation) or the operator's read-only setting is active. Optional; nil disables it.
	Delegation *uicompletion.DelegationOptions

	// Corner places the floating button. Optional; defaults to the bottom right.
	Corner uicompletion.Corner

	// Label is the button caption. Optional.
	Label string
}

// Confirmation decides whether the user approves changes of the assistant, see [AssistantOptions.Confirmation].
type Confirmation int

const (
	// ConfirmationGlobal follows the operator's [AssistantSettings.SkipConfirmation] and
	// [AssistantSettings.ConfirmMarkedOnly], which ask before every change by default.
	ConfirmationGlobal Confirmation = iota
	// ConfirmationAlways asks before every change, regardless of the global settings.
	ConfirmationAlways
	// ConfirmationNever runs changes without asking, regardless of the global settings.
	ConfirmationNever
	// ConfirmationMarked asks only before calls of tools marked [completion.Tool.RequiresApproval], e.g. deleting
	// data or granting rights, regardless of the global settings. Other changes run without asking.
	ConfirmationMarked
)

// confirm resolves the effective behaviour.
func (c Confirmation) confirm(cfg AssistantSettings) bool {
	switch c {
	case ConfirmationAlways:
		return true
	case ConfirmationNever, ConfirmationMarked: // the latter confirms only marked tools, see confirmMarked
		return false
	default:
		return !cfg.SkipConfirmation && !cfg.ConfirmMarkedOnly
	}
}

// confirmMarked resolves whether only the marked tools are confirmed.
func (c Confirmation) confirmMarked(cfg AssistantSettings) bool {
	switch c {
	case ConfirmationMarked:
		return true
	case ConfirmationGlobal:
		return !cfg.SkipConfirmation && cfg.ConfirmMarkedOnly
	default:
		return false
	}
}

var (
	// ErrAssistantNotSignedIn is returned by [Assistant.ChatOptions], if nobody is signed in.
	ErrAssistantNotSignedIn = errors.New("ai: nobody is signed in")
	// ErrAssistantHidden is returned by [Assistant.ChatOptions], if an operator hid the assistant, see
	// [AssistantSettings.Hidden].
	ErrAssistantHidden = errors.New("ai: an operator hid the assistant")
)

// assistantProblem is a misconfiguration, which [Assistant.Button] logs once per kind.
type assistantProblem struct {
	kind string
	err  error
}

func (p assistantProblem) Error() string {
	return p.err.Error()
}

func (p assistantProblem) Unwrap() error {
	return p.err
}

// ChatOptions resolves the chat options exactly like [Assistant.Button] does, so that a chat embedded into a page
// by [uicompletion.Chat] runs with the same provider, model and safety settings as the floating one, including
// the operator's [AssistantSettings.ReadOnly]. [AssistantOptions.Corner] and [AssistantOptions.Label] only apply
// to the button.
//
// It fails with [ErrAssistantNotSignedIn] or [ErrAssistantHidden], or with an error naming the missing
// permissions or the reason why no provider and model could be determined.
func (a *Assistant) ChatOptions(wnd core.Window, opts AssistantOptions) (uicompletion.ChatOptions, error) {
	if wnd.Subject() == nil || !wnd.Subject().Valid() {
		return uicompletion.ChatOptions{}, ErrAssistantNotSignedIn
	}

	cfg := core.GlobalSettings[AssistantSettings](wnd)
	if cfg.Hidden {
		return uicompletion.ChatOptions{}, ErrAssistantHidden
	}

	// Diagnose the most common cause first and by name, because it is invisible from the UI: without the
	// framework permissions the provider lookup below simply yields nothing, which looks exactly like "no
	// provider configured" and sends whoever debugs it to the wrong place.
	if missing := uicompletion.MissingPermissions(wnd.Subject()); len(missing) > 0 {
		return uicompletion.ChatOptions{}, assistantProblem{kind: "permissions", err: fmt.Errorf(
			"the AI assistant stays hidden because the acting role lacks %v; assign the %q role to the users who should use it",
			missing, RoleAssistantUser)}
	}

	chosen, modelID, err := a.Resolve(wnd.Subject(), cfg)
	if err != nil {
		return uicompletion.ChatOptions{}, assistantProblem{kind: "model", err: fmt.Errorf(
			"the AI assistant stays hidden because no provider and model could be determined: %w; check the vault and the global settings", err)}
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

	return uicompletion.ChatOptions{
		Sessions:           a.sessions,
		Completions:        chosen.Completions,
		Provider:           chosen.Provider,
		Title:              opts.Title,
		Tags:               opts.Tags,
		MaxTurns:           opts.MaxTurns,
		History:            opts.History,
		FileUpload:         opts.FileUpload,
		AskUser:            opts.AskUser,
		DisableCurrentTime: opts.DisableCurrentTime,
		ReadOnly:           cfg.ReadOnly,
		ConfirmMutations:   opts.Confirmation.confirm(cfg),
		ConfirmMarked:      opts.Confirmation.confirmMarked(cfg),
		Agents:             agents,
		Delegation:         opts.Delegation,
	}, nil
}

// Button returns the floating assistant button, or nil when the assistant cannot or should not run right now:
// nobody is signed in, an operator hid it, no provider is configured, or no model could be determined.
//
// Returning nil rather than an error view is deliberate - the button lives in a decorator and appears on every
// screen, so a misconfiguration must not turn into a banner on every page. The reason is logged instead, once
// per kind, with enough detail to act on. Use [uicompletion.MissingPermissions] if you want to surface a
// missing role in your own UI. See [Assistant.ChatOptions] to embed the chat into a page instead.
func (a *Assistant) Button(wnd core.Window, opts AssistantOptions) core.View {
	chatOpts, err := a.ChatOptions(wnd, opts)
	if err != nil {
		var problem assistantProblem
		if errors.As(err, &problem) {
			a.complain(problem.kind, problem.Error())
		}

		return nil
	}

	button := uicompletion.ChatButton(chatOpts)
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
//
// The button floats above the page and takes no space. The view is stretched to the full width, so a page
// looks the same with and without the assistant, whether the scaffold wraps the assistant or the other way
// round.
func (a *Assistant) Decorate(wnd core.Window, view core.View, opts AssistantOptions) core.View {
	button := a.Button(wnd, opts)
	if button == nil {
		return view
	}

	// A stack centers its children by default, which would shrink a scaffold to the width of its content.
	return ui.VStack(view, button).FullWidth().Alignment(ui.Stretch)
}

// Candidate is a configured provider which can run a completion.
type Candidate struct {
	Provider    provider.Provider
	Completions completion.Completions
}

// Providers lists the configured providers which can run a completion, sorted by name.
//
// The ways this can fail are told apart deliberately. They were once one silent "no" whose message named the
// vault - so the first person to debug a missing assistant went looking at the secret while the actual cause
// was a role holding no framework permission. A diagnosis that points at the wrong place costs more than none.
func (a *Assistant) Providers(subject auth.Subject) ([]Candidate, error) {
	configured := 0
	var res []Candidate

	for p, err := range a.useCases.FindAllProvider(subject) {
		if err != nil {
			return nil, fmt.Errorf("the acting role may not list AI providers (%s): %w", ai.PermFindAllProvider, err)
		}

		configured++

		if c := p.Completions(); c.IsSome() {
			res = append(res, Candidate{Provider: p, Completions: c.Unwrap()})
		}
	}

	if len(res) > 0 {
		return res, nil
	}

	if configured == 0 {
		return nil, fmt.Errorf("no AI provider is configured; add a provider token in the vault")
	}

	return nil, fmt.Errorf("%d AI provider(s) are configured, but none of them offers completions", configured)
}

// Resolve decides which provider and which model answer.
//
// There is no provider dropdown in the chat on purpose: the assistant is a fixture of the application, not a
// place to experiment. The operator picks provider and model together in [AssistantSettings.Model]:
//   - A choice naming a configured provider costs nothing at all.
//   - A bare model id, stored before providers could be chosen, is never replaced, because an API may accept
//     ids its model list does not show, e.g. aliases. It goes to the first provider offering it, otherwise to
//     the first provider, just like before providers could be chosen.
//   - Without a choice, or if the chosen provider is gone, the first provider with its first model is used.
//
// Looking at the offered models is a network call, hence the cache behind [Assistant.Models]. A failure is
// returned rather than swallowed: silently falling back turns an unreachable provider into "no model set",
// which again sends whoever has to fix it looking in the wrong place.
func (a *Assistant) Resolve(subject auth.Subject, cfg AssistantSettings) (Candidate, model.ID, error) {
	candidates, err := a.Providers(subject)
	if err != nil {
		return Candidate{}, "", err
	}

	byID := func(id provider.ID) (Candidate, bool) {
		for _, c := range candidates {
			if c.Provider.Identity() == id {
				return c, true
			}
		}

		return Candidate{}, false
	}

	if pid, mid, ok := cfg.Model.split(func(id provider.ID) bool { _, ok := byID(id); return ok }); ok {
		c, _ := byID(pid)
		return c, mid, nil
	}

	if cfg.Model != "" && !cfg.Model.namesProvider() {
		// A bare model id, stored before a provider could be chosen. It is never replaced, because the API may
		// accept ids its model list does not show, e.g. aliases. Only the provider offering it is looked up.
		wanted := model.ID(cfg.Model)
		for _, c := range candidates {
			models, err := a.Models(subject, c)
			if err != nil {
				continue
			}

			if slices.ContainsFunc(models, func(m model.Model) bool { return m.ID == wanted }) {
				return c, wanted, nil
			}
		}

		// like before providers could be chosen, the first provider is asked
		return candidates[0], wanted, nil
	}

	if cfg.Model != "" {
		a.complain("choice", fmt.Sprintf("the provider of the configured assistant model %q is gone, the default is used instead; pick one in the global settings", cfg.Model))
	}

	models, err := a.Models(subject, candidates[0])
	if err != nil {
		return Candidate{}, "", err
	}

	if len(models) == 0 {
		return Candidate{}, "", fmt.Errorf("the provider %q reports no models", candidates[0].Provider.Name())
	}

	return candidates[0], models[0].ID, nil
}

// Models lists what the provider offers, at most once per cache interval and provider.
//
// It is the one place that asks. The button needs it to pick a default and the settings picker needs it to
// offer a choice, and both are rendered often enough that asking twice would be twice too many.
func (a *Assistant) Models(subject auth.Subject, c Candidate) ([]model.Model, error) {
	cache, _ := a.models.LoadOrStore(c.Provider.Identity(), &modelCache{})

	return cache.(*modelCache).list(func() ([]model.Model, error) {
		var out []model.Model

		for m, err := range c.Completions.Models(subject) {
			if err != nil {
				return nil, fmt.Errorf("the model list of %q is unreachable, which usually means the API token is wrong: %w", c.Provider.Name(), err)
			}

			out = append(out, m)
		}

		return out, nil
	})
}

// Forget drops the cached model lists so the next caller asks the providers again. Called automatically when
// the settings or a secret change.
func (a *Assistant) Forget() {
	a.models.Clear()
}

// complain logs a problem once per kind.
func (a *Assistant) complain(kind, message string) {
	if _, seen := a.complained.LoadOrStore(kind, true); seen {
		return
	}

	slog.Warn(message)
}
