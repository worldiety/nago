// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"log/slog"
	"slices"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dropdown"
)

// defaultMaxTokens caps the generated output tokens per turn when neither the chosen [Agent] nor
// [ChatOptions] specify a value. Anthropic requires a positive limit, so we always send one. Reasoning
// (thinking) tokens count against it, so it must leave room for thinking plus the actual answer.
const defaultMaxTokens = 32000

// defaultConversationHeight is the height of the scrollable conversation area of an embedded [Chat] when the
// caller does not override it via [ChatOptions.Height].
const defaultConversationHeight = ui.L400

// Agent bundles all agent-specific configuration. It is the single place for the system prompt, model, token
// budget and callable tools of one selectable assistant persona. Callers populate [ChatOptions.Agents] with
// one or more of these; the values are provider- and domain-agnostic on purpose.
type Agent struct {
	// ID is a stable identifier of this agent. It is used as the selected value of the agent picker; it may
	// be empty when only a single agent is configured.
	ID string

	// Name is the human-readable label shown in the agent picker.
	Name string

	// SystemPrompt is the static system/developer instruction sent on every turn. Overridden per turn by
	// SystemPromptFunc when set.
	SystemPrompt string

	// SystemPromptFunc, when set, is evaluated on every submit and its result replaces SystemPrompt for that
	// turn. Use it to embed dynamic context (e.g. the currently rendered state) into the prompt. Optional.
	SystemPromptFunc func() string

	// Model overrides the provider's default model for this agent. When empty, the first model reported by
	// the provider is used. Optional.
	Model model.ID

	// MaxTokens caps the generated output tokens per turn for this agent. When zero, [defaultMaxTokens] is
	// used. Optional.
	MaxTokens int

	// Tools are the executable tools offered to the model for this agent. Because a [completion.Tool]
	// receives the acting subject on every call (see [completion.NewSubjectTool]), the same tool values can
	// be built once at start-up and shared by every window and every user - there is no need to rebuild them
	// per turn to bind an actor.
	//
	// The built-in ask_user tool and the file-upload wiring are added automatically by [ChatOptions] flags
	// and must not be listed here. Optional.
	Tools []completion.Tool
}

// resolvePrompt returns the effective system prompt for this agent (SystemPromptFunc wins over SystemPrompt).
func (a Agent) resolvePrompt() string {
	if a.SystemPromptFunc != nil {
		return a.SystemPromptFunc()
	}
	return a.SystemPrompt
}

// ChatOptions configures an embedded [Chat] (and, by embedding, a floating [ChatButton]). Only Sessions,
// Completions and Provider are required; every other field is optional and toggles a capability.
type ChatOptions struct {
	// Sessions persists the conversation history. Required. Only used when History is set; still required so
	// the same options can drive both a persisted and a transient chat without restructuring.
	Sessions session.UseCases

	// Completions is the provider capability that runs each turn. Required.
	Completions completion.Completions

	// Provider is the provider the Completions belong to. Required. It supplies the display name, the default
	// model and (for FileUpload) the Files capability.
	Provider provider.Provider

	// Title is the panel/header title. Optional; defaults to the provider name.
	Title string

	// Tags scope the persisted history: created sessions are tagged with them and the history dialog lists
	// only sessions carrying all of these tags. Ignored when History is false. Optional.
	Tags []string

	// MaxTurns bounds the agentic loop per submit (see [completion.RunOptions.MaxTurns]). Optional.
	MaxTurns int

	// Height overrides the height of the scrollable conversation area of an embedded chat. Optional.
	Height ui.Length

	// History enables persistence: every turn is written into a [session.Session] and a history button lets
	// the user restore previous conversations. When false the chat is transient (in-session only, no restore
	// button) and the history is not kept beyond the running turn.
	History bool

	// FileUpload enables user file attachments: an upload button next to the input lets the user pick files
	// that are attached to their next message (images/PDFs uploaded to the provider and referenced by id,
	// text files inlined). It additionally wires the provider's Files capability so file-providing tools can
	// attach binaries. Both require and are ignored without a provider Files capability.
	FileUpload bool

	// AskUser hooks the built-in ask_user clarification tool into every turn, letting the model ask the user
	// a question mid-run. The run suspends until the user answered; with History the question survives
	// closing the chat, navigating away and server restarts.
	AskUser bool

	// DisableCurrentTime removes the built-in current_time tool (see [CurrentTimeTool]), which is otherwise
	// wired into every turn. Without it a model calculates with the date of its training cut-off.
	DisableCurrentTime bool

	// ReadOnly removes every tool marked [completion.Tool.Mutating] before the turn starts, so the model is
	// never even told they exist. Use it for an assistant that may look but not touch.
	//
	// This is a filter, not an instruction: a tool that is not advertised cannot be called, whatever the
	// model decides to do.
	ReadOnly bool

	// ConfirmMutations asks the user to confirm every call of a tool marked [completion.Tool.Mutating]
	// before it runs, showing the tool, its stated effect and the arguments the model chose. Declining is
	// reported back to the model as a tool error, so it can offer an alternative instead of the turn failing.
	//
	// Prefer this over instructing the model in the system prompt to ask first. The prompt is a request the
	// model may ignore; this is a gate it cannot pass.
	ConfirmMutations bool

	// Agents configures the selectable assistant personas. len==0 falls back to a single default agent (empty
	// prompt, provider default model, no tools). A picker is shown only when len>1.
	Agents []Agent
}

// effectiveAgents returns the configured agents, or a single default agent when none are configured.
func (o ChatOptions) effectiveAgents() []Agent {
	if len(o.Agents) == 0 {
		return []Agent{{}}
	}
	return o.Agents
}

// resolveTools returns the tools of the agent, with mutating ones removed when the chat is read-only.
//
// Filtering here rather than expecting the caller not to configure them means an application can offer one
// tool set and let a setting decide whether it may write, without maintaining two lists that drift apart.
func (o ChatOptions) resolveTools(agent Agent) []completion.Tool {
	if !o.ReadOnly {
		return slices.Clone(agent.Tools)
	}

	tools := make([]completion.Tool, 0, len(agent.Tools))
	for _, t := range agent.Tools {
		if t.Mutating {
			continue
		}
		tools = append(tools, t)
	}

	return tools
}

// containsMutating reports whether any of the tools changes state.
func containsMutating(tools []completion.Tool) bool {
	return slices.ContainsFunc(tools, func(t completion.Tool) bool { return t.Mutating })
}

// Chat renders an embeddable, code-configured chat view on top of the stateless completion API and the
// session use cases. See [ChatOptions] for the available capabilities.
func Chat(wnd core.Window, opts ChatOptions) core.View {
	return chatBody(wnd, opts, defaultConversationHeightOr(opts.Height))
}

func defaultConversationHeightOr(h ui.Length) ui.Length {
	if h != "" {
		return h
	}
	return defaultConversationHeight
}

// chatBody builds the actual chat body (agent picker, conversation, footer, history dialog). It is shared by
// the embedded [Chat] and the floating [ChatButton] panel, the latter passing its own height.
//
// Nothing here blocks on the user. A clarifying question or an approval suspends the run (see
// [completion.Start]); with History the suspension is persisted in the session, so the chat may be closed,
// the user may navigate away and the question is still there when the chat is opened again.
func chatBody(wnd core.Window, opts ChatOptions, height ui.Length) core.View {
	comps := opts.Completions
	prov := opts.Provider
	agentsList := opts.effectiveAgents()

	title := opts.Title
	if title == "" {
		title = prov.Name()
	}

	// resumed is the session a freshly mounted chat continues because it still waits on the user, e.g. after
	// the chat was closed to look something up. Evaluated once per mount.
	resumed := core.AutoState[*session.Session](wnd).Init(func() *session.Session {
		if !opts.History {
			return nil
		}
		return findResumable(wnd.Subject(), opts.Sessions, opts.Tags, string(prov.Identity()))
	})

	history := core.AutoState[[]completion.Message](wnd).Init(func() []completion.Message {
		if r := resumed.Get(); r != nil {
			return r.Messages
		}
		return nil
	})
	prompt := core.AutoState[string](wnd)
	busy := core.AutoState[bool](wnd)
	// sessionID is the persisted conversation the panel currently continues. Empty for a fresh chat, set on
	// the first submit (lazy create), when restoring from history or when resuming a pending question. Only
	// used when History is enabled.
	sessionID := core.AutoState[session.ID](wnd).Init(func() session.ID {
		if r := resumed.Get(); r != nil {
			return r.ID
		}
		return ""
	})
	// pending and pendingRev mirror the suspension of the current run: from the session with History, from
	// the transient run otherwise.
	pending := core.AutoState[*completion.Continuation](wnd).Init(func() *completion.Continuation {
		if r := resumed.Get(); r != nil {
			return r.Pending
		}
		return nil
	})
	pendingRev := core.AutoState[int](wnd).Init(func() int {
		if r := resumed.Get(); r != nil {
			return r.PendingRevision
		}
		return 0
	})
	showHistory := core.AutoState[bool](wnd)
	status := core.AutoState[string](wnd)
	selectedAgent := core.AutoState[string](wnd).Init(func() string { return agentsList[0].ID })

	// staged holds files the user picked but has not sent yet (only when FileUpload is enabled and the
	// provider exposes a Files capability). They are attached to the next message on submit.
	staged := core.AutoState[[]stagedFile](wnd)
	var providerFiles provider.Files
	if opts.FileUpload {
		if pf := prov.Files(); pf.IsSome() {
			providerFiles = pf.Unwrap()
		}
	}
	uploadEnabled := opts.FileUpload && providerFiles != nil

	// currentAgent resolves the selected agent, falling back to the first configured one.
	currentAgent := func() Agent {
		id := selectedAgent.Get()
		for _, a := range agentsList {
			if a.ID == id {
				return a
			}
		}
		return agentsList[0]
	}

	var fileUploader completion.FileUploader
	if opts.FileUpload {
		fileUploader = ProviderFileUploader(prov)
	}

	applySession := func(s session.Session) {
		sessionID.Set(s.ID)
		history.Set(s.Messages)
		pending.Set(s.Pending)
		pendingRev.Set(s.PendingRevision)
	}

	// turnConfig resolves everything a run of the current agent needs.
	type turnConfig struct {
		model     model.ID
		system    string
		maxTokens int
		tools     []completion.Tool
		confirm   bool
	}
	resolveTurn := func() turnConfig {
		agent := currentAgent()
		cfg := turnConfig{model: agent.Model, system: agent.resolvePrompt(), maxTokens: agent.MaxTokens}
		if cfg.model == "" {
			cfg.model = firstModelID(wnd, comps)
		}
		if cfg.maxTokens <= 0 {
			cfg.maxTokens = defaultMaxTokens
		}

		cfg.tools = opts.resolveTools(agent)
		if opts.AskUser {
			cfg.tools = append(cfg.tools, completion.NewAskUserTool())
		}
		if !opts.DisableCurrentTime {
			cfg.tools = withBuiltinTool(cfg.tools, CurrentTimeTool(wnd))
		}
		// Approval is requested only when there is something to approve, so a purely reading assistant
		// never stops to ask.
		cfg.confirm = opts.ConfirmMutations && containsMutating(cfg.tools)
		return cfg
	}

	appendOptions := func(cfg turnConfig, onProgress completion.ProgressFunc) session.AppendOptions {
		return session.AppendOptions{
			Completions:     comps,
			Model:           cfg.model,
			System:          cfg.system,
			Tools:           cfg.tools,
			MaxTokens:       cfg.maxTokens,
			MaxTurns:        opts.MaxTurns,
			OnProgress:      onProgress,
			FileUploader:    fileUploader,
			ConfirmMutating: cfg.confirm,
		}
	}

	runOptions := func(cfg turnConfig, messages []completion.Message, onProgress completion.ProgressFunc) completion.RunOptions {
		return completion.RunOptions{
			Options: completion.Options{
				Model:     cfg.model,
				System:    cfg.system,
				MaxTokens: cfg.maxTokens,
				Messages:  messages,
			},
			Tools:           cfg.tools,
			MaxTurns:        opts.MaxTurns,
			OnProgress:      onProgress,
			FileUploader:    fileUploader,
			ConfirmMutating: cfg.confirm,
		}
	}

	// turnOutcome is what a background run hands back to the UI.
	type turnOutcome struct {
		history  []completion.Message
		pending  *completion.Continuation
		revision int
		usage    completion.Usage
		// progressed tells a failed run apart from one that changed nothing (roll back the view then).
		progressed bool
	}

	// execute shows the optimistic view, runs work off the event loop while mirroring each model turn live,
	// and applies the outcome. rollback restores the input when a run failed without changing anything.
	execute := func(optimistic []completion.Message, rollback func(), work func(subject auth.Subject, onProgress completion.ProgressFunc) (turnOutcome, error)) {
		prevHistory := history.Get()
		history.Set(optimistic)
		busy.Set(true)
		status.Set(thinkingLabel(0))

		// live mirrors the growing conversation while the loop runs so each assistant turn appears the moment
		// it arrives. lastStop remembers why the latest model turn ended, so a truncated or refused final
		// answer can be pointed out. Both are only touched on the loop's goroutine.
		live := slices.Clone(optimistic)
		var lastStop completion.StopReason
		onProgress := func(p completion.Progress) {
			switch p.Phase {
			case completion.PhaseTurnStarted:
				turn := p.Turn
				wnd.Post(func() { status.Set(thinkingLabel(turn)) })
			case completion.PhaseModelResponded:
				if p.Result == nil {
					return
				}
				lastStop = p.Result.StopReason
				// A truncated turn is repeated or cleaned up by the loop; showing it would only flicker.
				if p.Result.StopReason == completion.StopMaxTokens || len(p.Result.Message.Content) == 0 {
					return
				}
				live = append(live, p.Result.Message)
				snapshot := slices.Clone(live)
				wnd.Post(func() { history.Set(snapshot) })
			case completion.PhaseToolStarted:
				name := ""
				if p.ToolCall != nil {
					name = p.ToolCall.Name
				}
				wnd.Post(func() { status.Set(toolLabel(name)) })
			}
		}

		xsync.Go(func() error {
			subject := wnd.Subject()
			out, err := work(subject, onProgress)

			wnd.Post(func() {
				busy.Set(false)
				status.Set("")
				if err != nil {
					if out.progressed {
						// Keep what already happened (tools with side effects) visible.
						history.Set(out.history)
						pending.Set(out.pending)
						pendingRev.Set(out.revision)
					} else {
						history.Set(prevHistory)
						if rollback != nil {
							rollback()
						}
					}
					alert.ShowBannerError(wnd, err)
					return
				}

				if out.pending == nil {
					showStopHint(wnd, lastStop)
				}
				u := out.usage
				slog.Info("uicompletion chat usage",
					slog.String("session", string(sessionID.Get())),
					slog.Int("input_tokens", u.InputTokens),
					slog.Int("output_tokens", u.OutputTokens),
					slog.Int("cache_read_tokens", u.CacheReadTokens),
					slog.Int("cache_write_tokens", u.CacheWriteTokens),
				)
				history.Set(out.history)
				pending.Set(out.pending)
				pendingRev.Set(out.revision)
			})
			return nil
		}, func(err error) {
			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					status.Set("")
					history.Set(prevHistory)
					if rollback != nil {
						rollback()
					}
					alert.ShowBannerError(wnd, err)
				})
			}
		})
	}

	// fromSession converts a persisted session into a turn outcome; on failure it reloads what the failed
	// run persisted.
	fromSession := func(subject auth.Subject, sid session.ID, s session.Session, err error, before int) (turnOutcome, error) {
		if err != nil {
			if reloaded, ok := reloadSession(subject, opts.Sessions, sid); ok && len(reloaded.Messages) > before {
				return turnOutcome{history: reloaded.Messages, pending: reloaded.Pending, revision: reloaded.PendingRevision, progressed: true}, err
			}
			return turnOutcome{}, err
		}
		return turnOutcome{history: s.Messages, pending: s.Pending, revision: s.PendingRevision, usage: s.Usage}, nil
	}

	fromOutcome := func(out completion.Outcome, err error) (turnOutcome, error) {
		return turnOutcome{history: out.History, pending: out.Suspended, usage: out.Result.Usage, progressed: out.Progressed}, err
	}

	submit := func() {
		question := strings.TrimSpace(prompt.Get())
		// A turn needs either text or at least one attached file.
		stagedFiles := staged.Get()
		if (question == "" && len(stagedFiles) == 0) || busy.Get() || pending.Get() != nil {
			return
		}

		cfg := resolveTurn()

		// Ensure a persisted session exists (History only). Created lazily on the first message and tagged so
		// the history dialog lists only matching conversations.
		sid := sessionID.Get()
		if opts.History && sid == "" {
			created, err := opts.Sessions.Create(wnd.Subject(), session.CreateOptions{
				Title:        title,
				Model:        cfg.model,
				System:       cfg.system,
				ProviderHint: string(prov.Identity()),
				Tags:         opts.Tags,
			})
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}
			sid = created.ID
			sessionID.Set(sid)
		}

		prevHistory := history.Get()
		// Optimistic user bubble: the typed text plus a short hint per attached file.
		optimisticText := question
		for _, sf := range stagedFiles {
			optimisticText = strings.TrimSpace(optimisticText + "\n\n📎 " + sf.Name)
		}
		optimistic := append(slices.Clone(prevHistory), completion.Message{
			Role:    completion.User,
			Content: []completion.Content{completion.Text{Text: optimisticText}},
		})
		prompt.Set("")
		staged.Set(nil)

		rollback := func() {
			if prompt.Get() == "" {
				prompt.Set(question)
			}
			staged.Set(stagedFiles)
		}

		execute(optimistic, rollback, func(subject auth.Subject, onProgress completion.ProgressFunc) (turnOutcome, error) {
			// Build the user turn content: any attached files (uploaded/inlined here on the background
			// goroutine) followed by the typed text.
			input, err := buildUploadContent(subject, providerFiles, stagedFiles)
			if err != nil {
				return turnOutcome{}, err
			}
			if question != "" {
				input = append(input, completion.Text{Text: question})
			}

			if opts.History {
				ao := appendOptions(cfg, onProgress)
				ao.Input = input
				updated, err := opts.Sessions.Append(subject, sid, ao)
				return fromSession(subject, sid, updated, err, len(prevHistory))
			}

			// Transient chat: run the agentic loop directly over the history plus the real
			// (attachment-aware) user turn, without persisting anything.
			messages := append(slices.Clone(prevHistory), completion.Message{Role: completion.User, Content: input})
			return fromOutcome(completion.Start(subject, comps, runOptions(cfg, messages, onProgress)))
		})
	}

	resolve := func(resolutions []completion.Resolution) {
		cont := pending.Get()
		if cont == nil || busy.Get() {
			return
		}
		cfg := resolveTurn()
		sid := sessionID.Get()
		rev := pendingRev.Get()
		prevHistory := history.Get()

		optimistic := slices.Clone(prevHistory)
		if answers := optimisticAnswers(cont, resolutions); len(answers.Content) > 0 {
			optimistic = append(optimistic, answers)
		}
		pending.Set(nil)
		rollback := func() {
			pending.Set(cont)
			pendingRev.Set(rev)
		}

		execute(optimistic, rollback, func(subject auth.Subject, onProgress completion.ProgressFunc) (turnOutcome, error) {
			if opts.History {
				ao := appendOptions(cfg, onProgress)
				// The model of a pending run is fixed; the session knows it.
				ao.Model = ""
				updated, err := opts.Sessions.Resolve(subject, sid, session.ResolveOptions{
					Run:         ao,
					Revision:    rev,
					Resolutions: resolutions,
				})
				return fromSession(subject, sid, updated, err, len(prevHistory))
			}

			return fromOutcome(completion.Continue(subject, comps, runOptions(cfg, prevHistory, onProgress), *cont, resolutions))
		})
	}

	dismiss := func() {
		cont := pending.Get()
		if cont == nil || busy.Get() {
			return
		}

		if !opts.History {
			dismissed, err := completion.Dismiss(history.Get(), *cont)
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}
			history.Set(dismissed)
			pending.Set(nil)
			return
		}

		updated, err := opts.Sessions.Dismiss(wnd.Subject(), sessionID.Get(), pendingRev.Get())
		if err != nil {
			alert.ShowBannerError(wnd, err)
			// The session may have moved on elsewhere (e.g. another tab); show its current state.
			if reloaded, ok := reloadSession(wnd.Subject(), opts.Sessions, sessionID.Get()); ok {
				applySession(reloaded)
			}
			return
		}
		applySession(updated)
	}

	conversation := conversationView(wnd, history.Get(),
		"Stell mir eine Frage, um die Unterhaltung zu beginnen.", height)

	var footer core.View
	// A pending decision takes precedence over the input: the run waits on it, so offering the input field
	// instead would look like the assistant had simply stopped responding.
	if cont := pending.Get(); cont != nil {
		footer = renderDecisions(wnd, cont, pendingRev.Get(), busy.Get(), resolve, dismiss)
	} else {
		busyLabel := status.Get()
		if busyLabel == "" {
			busyLabel = "… die KI arbeitet"
		}
		var chips core.View
		if uploadEnabled {
			chips = stagedChips(staged, busy.Get())
		}

		footer = ui.VStack(
			ui.If(busy.Get(), ui.Text(busyLabel).Font(ui.BodySmall)),
			ui.If(chips != nil, chips),
			ui.TextField("Nachricht", prompt.Get()).
				InputValue(prompt).
				Lines(2).
				FullWidth().
				Disabled(busy.Get()).
				KeydownEnter(submit),
			ui.HStack(
				ui.If(uploadEnabled, uploadButton(wnd, staged, busy.Get())),
				ui.Spacer(),
				ui.SecondaryButton(submit).PreIcon(icons.PaperPlane).Title("Senden").Enabled(!busy.Get()),
			).Gap(ui.L8).FullWidth().Alignment(ui.Center),
		).Gap(ui.L8).FullWidth().Alignment(ui.Leading)
	}

	// Agent picker only when more than one agent is configured.
	var picker core.View
	if len(agentsList) > 1 {
		type opt = dropdown.Option[string]
		options := make([]opt, 0, len(agentsList))
		for _, a := range agentsList {
			label := a.Name
			if label == "" {
				label = a.ID
			}
			options = append(options, opt{Value: a.ID, Label: label})
		}
		picker = dropdown.Dropdown("Agent", options, selectedAgent.Get()).
			InputValue(selectedAgent).
			Disabled(busy.Get()).
			Frame(ui.Frame{}.FullWidth())
	}

	// History restore dialog and action row (History only): browse/restore a previous conversation and start
	// a fresh one. Both are disabled while a run is in flight so we never swap the history under a running
	// loop. "Neuer Chat" just detaches from the current session (and clears the view); the next submit lazily
	// creates a new one. A pending question stays in the old session and can be picked up from the history.
	var restoreDialog core.View
	var historyActions core.View
	if opts.History {
		restoreDialog = historyDialog(wnd, opts.Sessions, opts.Tags, showHistory, func(s session.Session) {
			applySession(s)
			status.Set("")
		})

		historyActions = ui.HStack(
			ui.TertiaryButton(func() {
				showHistory.Set(true)
			}).PreIcon(icons.Clock).Title("Verlauf").Enabled(!busy.Get()),
			ui.TertiaryButton(func() {
				applySession(session.Session{})
				status.Set("")
			}).PreIcon(icons.Edit).Title("Neuer Chat").Enabled(!busy.Get() && (sessionID.Get() != "" || len(history.Get()) > 0)),
			ui.Spacer(),
		).Gap(ui.L4).FullWidth().Alignment(ui.Center)
	}

	return ui.VStack(
		ui.If(restoreDialog != nil, restoreDialog),
		ui.If(historyActions != nil, historyActions),
		ui.If(picker != nil, picker),
		conversation,
		footer,
	).Gap(ui.L8).FullWidth().Alignment(ui.Leading)
}

// findResumable returns the newest session of the subject which waits on a user decision and belongs to
// exactly this chat context: created by the subject, same provider and exactly the same tags. Sessions of
// other contexts are never picked up; the history dialog still offers them.
func findResumable(subject auth.Subject, sessions session.UseCases, tags []string, providerHint string) *session.Session {
	if sessions.FindAll == nil || subject == nil || !subject.Valid() {
		return nil
	}

	var best *session.Session
	for s, err := range sessions.FindAll(subject, session.FindAllOptions{Tags: tags}) {
		if err != nil {
			return nil
		}
		if s.Pending == nil || s.CreatedBy != subject.ID() || s.ProviderHint != providerHint || !sameTags(s.Tags, tags) {
			continue
		}
		if best == nil || s.UpdatedAt > best.UpdatedAt {
			s := s
			best = &s
		}
	}
	return best
}

// sameTags reports set equality of two tag lists.
func sameTags(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, t := range b {
		if !slices.Contains(a, t) {
			return false
		}
	}
	return true
}

// firstModelID returns the id of the first model the completion provider reports.
func firstModelID(wnd core.Window, comps completion.Completions) model.ID {
	for m, err := range comps.Models(wnd.Subject()) {
		if err != nil {
			return ""
		}
		return m.ID
	}
	return ""
}

// reloadSession loads a session, e.g. after a failed run persisted a partial trace.
func reloadSession(subject auth.Subject, sessions session.UseCases, id session.ID) (session.Session, bool) {
	if sessions.FindByID == nil || id == "" {
		return session.Session{}, false
	}
	opt, err := sessions.FindByID(subject, id)
	if err != nil || opt.IsNone() {
		return session.Session{}, false
	}
	return opt.Unwrap(), true
}

// showStopHint tells the user when the final answer did not end regularly, instead of letting it look like the
// assistant simply stopped.
func showStopHint(wnd core.Window, stop completion.StopReason) {
	switch stop {
	case completion.StopMaxTokens:
		alert.ShowBannerMessage(wnd, alert.Message{
			Title:   "Antwort abgeschnitten",
			Message: "Die Antwort hat die maximale Länge erreicht und ist unvollständig. Bitte die Frage enger fassen oder die maximale Antwortlänge erhöhen.",
			Intent:  alert.IntentWarning,
		})
	case completion.StopRefusal:
		alert.ShowBannerMessage(wnd, alert.Message{
			Title:   "Anfrage abgelehnt",
			Message: "Das Modell hat die Beantwortung dieser Anfrage abgelehnt.",
			Intent:  alert.IntentWarning,
		})
	}
}
