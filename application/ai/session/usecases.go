// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package session models a persistable, provider-independent chat session on top of the stateless
// [completion] API.
//
// In contrast to [conversation.Conversation], which delegates storage to the provider (Mistral/OpenAI
// "conversations"; Anthropic stores nothing), a session stores the full, rich [completion.Message] history
// locally in a [data.Repository]. Because the history is embedded verbatim (including Text, Media, ToolCall,
// ToolResult and Thinking blocks, made JSON-safe by application/ai/completion/json.go), a session works with
// ANY provider that exposes [completion.Completions] - the caller simply passes the desired Completions to
// [Append] at runtime.
//
// The executable tools of an agentic run are intentionally NOT persisted (Go functions are not
// serializable). Instead they are supplied per call via [AppendOptions.Tools]; when present, [Append] drives
// the agentic loop through [completion.Run], otherwise it performs a single [completion.Completions.Complete]
// turn. The resulting tool_call/tool_result blocks are persisted losslessly as part of the history.
package session

import (
	"iter"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

// Namespace is the ReBAC namespace of session resources. It must equal the repository/store name so that
// [auth.Subject.AuditResource] and the ReBAC editor address the same instances. See the wiring in
// application/ai/cfg where the store is opened under this name and the static rules are registered.
const Namespace rebac.Namespace = "nago.ai.session"

// ID uniquely identifies a [Session] within the local repository.
type ID string

// Session is a locally persisted chat, owning the complete stateless [completion] history.
type Session struct {
	ID ID `json:"id,omitempty"`

	// Title is a human-readable label for the session (e.g. shown in a session list). Optional.
	Title string `json:"title,omitempty"`

	// Model is the model the session runs against. It is applied to every completion request issued by
	// [Append] unless overridden there.
	Model model.ID `json:"model,omitempty"`

	// System is the stable system/developer prompt of the session. It is sent on every [Append] so the
	// instruction stays constant across turns (which also benefits provider-side prompt caching).
	System string `json:"system,omitempty"`

	// ProviderHint records which provider originally served this session (its provider.ID as a string).
	//
	// It is intentionally a plain string, not a provider.ID: this package is a leaf on top of [completion]
	// and, like [conversation.Conversation] (which also stores only agent.ID/model.ID), it must not import
	// the provider package. provider already imports the leaf packages (completion, conversation, message,
	// model, ...) to aggregate their capabilities, so importing provider here would invert that layering and
	// risk an import cycle. A caller resolves the hint back to a provider at runtime (e.g. to preselect the
	// matching provider when continuing a session). Optional.
	ProviderHint string `json:"providerHint,omitempty"`

	// Messages is the full, ordered, lossless history including tool calls and tool results. It is exactly
	// the slice that would be fed back into [completion.Completions.Complete] to continue the session.
	Messages []completion.Message `json:"messages,omitempty"`

	// Usage accumulates the token usage reported across all turns of the session.
	Usage completion.Usage `json:"usage,omitempty"`

	CreatedAt xtime.UnixMilliseconds `json:"createdAt,omitempty"`
	CreatedBy user.ID                `json:"createdBy,omitempty"`
	UpdatedAt xtime.UnixMilliseconds `json:"updatedAt,omitempty"`
}

func (s Session) Identity() ID {
	return s.ID
}

// String returns a human-readable label for the session, preferring its title and otherwise a short preview
// of the first user message. It is used e.g. as the instance label in the ReBAC editor.
func (s Session) String() string {
	if s.Title != "" {
		return s.Title
	}

	for _, msg := range s.Messages {
		if msg.Role != completion.User {
			continue
		}
		for _, c := range msg.Content {
			if t, ok := c.(completion.Text); ok && t.Text != "" {
				preview := t.Text
				if r := []rune(preview); len(r) > 48 {
					preview = string(r[:48]) + "…"
				}
				return preview
			}
		}
	}

	return "Session " + string(s.ID)
}

// Repository persists [Session] aggregates locally.
type Repository data.Repository[Session, ID]

// CreateOptions configures [Create].
type CreateOptions struct {
	// Title is an optional human-readable label.
	Title string

	// Model the session runs against. Required to later run completions, but may be empty at creation time
	// and set on the first [Append] via [AppendOptions.Model].
	Model model.ID

	// System is the optional stable system prompt of the session.
	System string

	// ProviderHint optionally records the originating provider (see [Session.ProviderHint]).
	ProviderHint string

	// Input is an optional first user turn. When set, it is stored as the initial history entry but NOT yet
	// completed - call [Append] to obtain an assistant answer. Leave empty to create an empty session.
	Input []completion.Content
}

// AppendOptions carries a new user turn plus the runtime-only dependencies required to produce an assistant
// answer. None of these are persisted.
type AppendOptions struct {
	// Completions is the provider capability that actually runs the turn. Required.
	Completions completion.Completions

	// Input is the new user content appended before the model is asked to respond. Required and non-empty.
	Input []completion.Content

	// Model overrides [Session.Model] for this (and only this) turn. Optional; when empty the session model
	// is used.
	Model model.ID

	// Tools are the executable tools offered to the model for this turn. When non-empty, [Append] drives the
	// full agentic loop via [completion.Run]; otherwise a single [completion.Completions.Complete] is used.
	// Tools are never persisted.
	Tools []completion.Tool

	// MaxTokens caps the generated output tokens for this turn. Optional.
	MaxTokens int

	// Temperature overrides the sampling temperature for this turn. Optional.
	Temperature option.Opt[float64]

	// OnProgress is forwarded to [completion.Run] for agentic runs so a caller can observe tool execution.
	// Ignored when no tools are supplied. Optional.
	OnProgress completion.ProgressFunc

	// MaxTurns bounds the agentic loop (see [completion.RunOptions.MaxTurns]). Ignored without tools.
	// Optional.
	MaxTurns int
}

// Create persists a new, optionally pre-seeded [Session].
type Create func(subject auth.Subject, opts CreateOptions) (Session, error)

// FindByID returns the session with the given id if it exists and the subject may read it.
type FindByID func(subject auth.Subject, id ID) (option.Opt[Session], error)

// FindAll yields all sessions owned by (visible to) the subject.
type FindAll func(subject auth.Subject) iter.Seq2[Session, error]

// Append adds a user turn, runs the completion (optionally agentic) against the supplied provider capability,
// appends the produced messages to the history and persists the updated session, which it returns.
type Append func(subject auth.Subject, id ID, opts AppendOptions) (Session, error)

// Rename changes the human-readable title of a session.
type Rename func(subject auth.Subject, id ID, title string) error

// Delete removes a session and its embedded history.
type Delete func(subject auth.Subject, id ID) error

// UseCases bundles all session use cases. Construct it with [NewUseCases].
type UseCases struct {
	Create   Create
	FindByID FindByID
	FindAll  FindAll
	Append   Append
	Rename   Rename
	Delete   Delete
}

// NewUseCases wires the session use cases against the given repository and ReBAC database. A single shared
// mutex serializes mutating operations so concurrent [Append] calls on the same repository cannot interleave
// read-modify-write cycles.
//
// Authorization is resource-scoped: every use case audits via [auth.Subject.AuditResource] against the
// session's ReBAC instance, so a subject may act either through a global permission (e.g. an IAM group that
// is allowed to create/list sessions) or through an instance grant. [Create] writes such an instance grant
// for the creator, so users can see and continue only their own sessions unless additionally granted global
// access.
func NewUseCases(repo Repository, rdb *rebac.DB) UseCases {
	var mutex sync.Mutex

	return UseCases{
		Create:   NewCreate(&mutex, repo, rdb),
		FindByID: NewFindByID(repo),
		FindAll:  NewFindAll(repo),
		Append:   NewAppend(&mutex, repo),
		Rename:   NewRename(&mutex, repo),
		Delete:   NewDelete(&mutex, repo, rdb),
	}
}
