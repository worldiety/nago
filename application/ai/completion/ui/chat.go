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
// [ChatOptions] specify a value. Anthropic requires a positive limit, so we always send one.
const defaultMaxTokens = 4096

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

	// Tools returns the executable tools offered to the model for this agent, resolved per turn against the
	// acting subject. The built-in ask_user tool and the file-upload wiring are added automatically by
	// [ChatOptions] flags and must not be returned here. Optional.
	Tools func(subject auth.Subject) []completion.Tool
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
	// a question mid-run and block until answered (analogous to FileUpload).
	AskUser bool

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
func chatBody(wnd core.Window, opts ChatOptions, height ui.Length) core.View {
	comps := opts.Completions
	prov := opts.Provider
	agentsList := opts.effectiveAgents()

	title := opts.Title
	if title == "" {
		title = prov.Name()
	}

	history := core.AutoState[[]completion.Message](wnd)
	prompt := core.AutoState[string](wnd)
	busy := core.AutoState[bool](wnd)
	// sessionID is the persisted conversation the panel currently continues. Empty for a fresh chat, set on
	// the first submit (lazy create) or when restoring from history. Only used when History is enabled.
	sessionID := core.AutoState[session.ID](wnd)
	showHistory := core.AutoState[bool](wnd)
	status := core.AutoState[string](wnd)
	ask := core.AutoState[*pendingAsk](wnd)
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

	submit := func() {
		question := strings.TrimSpace(prompt.Get())
		// A turn needs either text or at least one attached file.
		stagedFiles := staged.Get()
		if (question == "" && len(stagedFiles) == 0) || busy.Get() {
			return
		}

		agent := currentAgent()

		modelID := agent.Model
		if modelID == "" {
			modelID = firstModelID(wnd, comps)
		}

		maxTokens := agent.MaxTokens
		if maxTokens <= 0 {
			maxTokens = defaultMaxTokens
		}

		system := agent.resolvePrompt()

		// Ensure a persisted session exists (History only). Created lazily on the first message and tagged so
		// the history dialog lists only matching conversations.
		sid := sessionID.Get()
		if opts.History && sid == "" {
			created, err := opts.Sessions.Create(wnd.Subject(), session.CreateOptions{
				Title:        title,
				Model:        modelID,
				System:       system,
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

		// prevHistory is the last consistent view we roll back to should the turn fail.
		prevHistory := history.Get()
		// Optimistic user bubble: the typed text plus a short hint per attached file.
		optimisticText := question
		for _, sf := range stagedFiles {
			optimisticText = strings.TrimSpace(optimisticText + "\n\n📎 " + sf.Name)
		}
		msgs := append(slices.Clone(prevHistory), completion.Message{
			Role:    completion.User,
			Content: []completion.Content{completion.Text{Text: optimisticText}},
		})
		history.Set(msgs)
		prompt.Set("")
		staged.Set(nil)
		busy.Set(true)
		status.Set(thinkingLabel(0))

		// live mirrors the growing conversation while the loop runs so each assistant turn appears the moment
		// it arrives. Every UI mutation is marshalled back onto the event loop via wnd.Post.
		live := slices.Clone(msgs)
		onProgress := func(p completion.Progress) {
			switch p.Phase {
			case completion.PhaseTurnStarted:
				turn := p.Turn
				wnd.Post(func() { status.Set(thinkingLabel(turn)) })
			case completion.PhaseModelResponded:
				if p.Result == nil {
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

			// Build the user turn content: any attached files (uploaded/inlined here on the background
			// goroutine) followed by the typed text. On failure we roll the optimistic view back.
			inputContent, uploadErr := buildUploadContent(subject, providerFiles, stagedFiles)
			if uploadErr != nil {
				wnd.Post(func() {
					busy.Set(false)
					status.Set("")
					history.Set(prevHistory)
					if prompt.Get() == "" {
						prompt.Set(question)
					}
					staged.Set(stagedFiles)
					alert.ShowBannerError(wnd, uploadErr)
				})
				return nil
			}
			if question != "" {
				inputContent = append(inputContent, completion.Text{Text: question})
			}

			var tools []completion.Tool
			if agent.Tools != nil {
				tools = agent.Tools(subject)
			}
			if opts.AskUser {
				tools = append(tools, askUserTool(wnd, ask))
			}

			if opts.History {
				updated, err := opts.Sessions.Append(subject, sid, session.AppendOptions{
					Completions:  comps,
					Input:        inputContent,
					Model:        modelID,
					System:       system,
					Tools:        tools,
					MaxTokens:    maxTokens,
					MaxTurns:     opts.MaxTurns,
					OnProgress:   onProgress,
					FileUploader: fileUploader,
				})

				wnd.Post(func() {
					busy.Set(false)
					status.Set("")
					ask.Set(nil)
					if err != nil {
						history.Set(prevHistory)
						if prompt.Get() == "" {
							prompt.Set(question)
						}
						staged.Set(stagedFiles)
						alert.ShowBannerError(wnd, err)
						return
					}
					u := updated.Usage
					slog.Info("uicompletion chat usage",
						slog.String("session", string(sid)),
						slog.String("model", string(updated.Model)),
						slog.Int("input_tokens", u.InputTokens),
						slog.Int("output_tokens", u.OutputTokens),
						slog.Int("cache_read_tokens", u.CacheReadTokens),
						slog.Int("cache_write_tokens", u.CacheWriteTokens),
					)
					history.Set(updated.Messages)
				})
				return nil
			}

			// Transient chat: run the agentic loop directly over the history plus the real (attachment-aware)
			// user turn, without persisting anything.
			runMessages := append(slices.Clone(prevHistory), completion.Message{
				Role:    completion.User,
				Content: inputContent,
			})
			_, newHistory, err := completion.Run(subject, comps, completion.RunOptions{
				Options: completion.Options{
					Model:     modelID,
					System:    system,
					MaxTokens: maxTokens,
					Messages:  runMessages,
				},
				Tools:        tools,
				MaxTurns:     opts.MaxTurns,
				OnProgress:   onProgress,
				FileUploader: fileUploader,
			})

			wnd.Post(func() {
				busy.Set(false)
				status.Set("")
				ask.Set(nil)
				if err != nil {
					history.Set(prevHistory)
					if prompt.Get() == "" {
						prompt.Set(question)
					}
					staged.Set(stagedFiles)
					alert.ShowBannerError(wnd, err)
					return
				}
				history.Set(newHistory)
			})
			return nil
		}, func(err error) {
			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					status.Set("")
					history.Set(prevHistory)
					if prompt.Get() == "" {
						prompt.Set(question)
					}
					staged.Set(stagedFiles)
					alert.ShowBannerError(wnd, err)
				})
			}
		})
	}

	conversation := conversationView(history.Get(),
		"Stell mir eine Frage, um die Unterhaltung zu beginnen.", height)

	var footer core.View
	if pa := ask.Get(); pa != nil {
		footer = renderAsk(wnd, ask, pa)
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
	// creates a new one.
	var restoreDialog core.View
	var historyActions core.View
	if opts.History {
		restoreDialog = historyDialog(wnd, opts.Sessions, opts.Tags, showHistory, func(s session.Session) {
			sessionID.Set(s.ID)
			history.Set(s.Messages)
			status.Set("")
			ask.Set(nil)
		})

		historyActions = ui.HStack(
			ui.TertiaryButton(func() {
				showHistory.Set(true)
			}).PreIcon(icons.Clock).Title("Verlauf").Enabled(!busy.Get()),
			ui.TertiaryButton(func() {
				sessionID.Set("")
				history.Set(nil)
				status.Set("")
				ask.Set(nil)
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
