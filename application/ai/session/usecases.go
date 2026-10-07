// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package session models a persistable, provider-independent chat session on top of the stateless
// [completion] API.
//
// Because the completion API is stateless and the provider stores nothing, a session stores the full, rich
// [completion.Message] history locally in a [data.Repository]. Because the history is embedded verbatim (including Text, Media, ToolCall,
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
	"context"
	"errors"
	"iter"

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
	// It is intentionally a plain string, not a provider.ID: this package is a leaf on top of [completion] and
	// must not import the provider package. provider already imports the leaf packages (completion, file) to
	// aggregate their capabilities, so importing provider here would invert that layering and risk an import
	// cycle. A caller resolves the hint back to a provider at runtime (e.g. to preselect the
	// matching provider when continuing a session). Optional.
	ProviderHint string `json:"providerHint,omitempty"`

	// Tags are opaque, free-form labels used to classify and filter sessions (see [FindAllOptions.Tags]).
	// They are provider- and domain-agnostic on purpose; a caller decides on their meaning. Typical uses are
	// binding a session to an application context (e.g. "ctx:invoice/42"), marking user-pinned chats, or
	// grouping by feature area. Optional.
	Tags []string `json:"tags,omitempty"`

	// Messages is the full, ordered, lossless history including tool calls and tool results. It is exactly
	// the slice that would be fed back into [completion.Completions.Complete] to continue the session.
	Messages []completion.Message `json:"messages,omitempty"`

	// Usage accumulates the token usage reported across all turns of the session: every completion of every
	// run, not only the last turn of each.
	Usage completion.Usage `json:"usage,omitempty"`

	// SubUsage accumulates the token usage of the sub-agents this session delegated to (see [NewSubRunner]).
	// It is booked with the next save of this session after a sub-agent finished, so the usage of a background
	// task still running is not included yet. The child sessions carry their own Usage as well.
	SubUsage completion.Usage `json:"subUsage,omitzero"`

	// ParentID is set on a child session, which holds the transcript of a sub-agent the parent session
	// delegated a task to (see [NewSubRunner]). Child sessions are hidden from [FindAll] unless
	// [FindAllOptions.IncludeChildren] is set, and deleted together with their parent.
	ParentID ID `json:"parentId,omitempty"`

	// ParentCallID is the id of the tool call of the parent session which started this child session.
	ParentCallID string `json:"parentCallId,omitempty"`

	// Pending is set while the last run waits on a user decision (a clarifying question or the approval of a
	// mutating tool). Messages then ends with the assistant turn carrying the pending calls; resolve them with
	// [Resolve] or [Dismiss]. A session with Pending rejects [Append].
	Pending *completion.Continuation `json:"pending,omitempty"`

	// PendingRevision increases with every new suspension. [Resolve] and [Dismiss] must name it, so a stale or
	// duplicate answer (e.g. from a second tab) never lands on a newer question.
	PendingRevision int `json:"pendingRevision,omitempty"`

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

	// Tags optionally classifies the session (see [Session.Tags]). Used later to filter via
	// [FindAllOptions.Tags].
	Tags []string

	// Input is an optional first user turn. When set, it is stored as the initial history entry but NOT yet
	// completed - call [Append] to obtain an assistant answer. Leave empty to create an empty session.
	Input []completion.Content

	// ParentID makes the new session a child session of the given one, see [Session.ParentID]. Optional.
	ParentID ID

	// ParentCallID records the tool call of the parent which started the child session. Optional.
	ParentCallID string
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

	// System overrides the session's system/developer prompt for this (and only this) turn. Optional; when
	// empty [Session.System] is used. This lets a caller rebuild a fresh system prompt per turn (e.g. to
	// embed the currently rendered, compact domain model) instead of relying on the session's fixed one. The
	// override is not persisted on the session.
	System string

	// Tools are the executable tools offered to the model for this turn. When non-empty, [Append] drives the
	// full agentic loop via [completion.Run]; otherwise a single [completion.Completions.Complete] is used,
	// unless Agentic is set. Tools are never persisted.
	Tools []completion.Tool

	// Agentic drives the agentic loop even without Tools, so that a truncated answer is continued, a turn
	// with reasoning only is asked for its answer and the history is compacted on overflow. Sub-agents use
	// it, see [NewSubRunner]. Optional.
	Agentic bool

	// FileUploader is required only when [Tools] contains file-providing tools (see
	// [completion.NewOpenFileTool]); it uploads a file to the active provider so it can be attached to the
	// conversation by id. Wire it to provider.Files().Put. Runtime-only; not persisted. Optional.
	FileUploader completion.FileUploader

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

	// OnBeforeToolCall is forwarded to [completion.Run] and may refuse individual tool calls, e.g. to ask
	// the user for confirmation before a mutating tool runs. Ignored without tools. Optional.
	OnBeforeToolCall completion.BeforeToolCallFunc

	// ConfirmMutating suspends the run before every mutating tool call until the user approved it via
	// [Resolve] (see [completion.RunOptions.ConfirmMutating]). Ignored without tools. Optional.
	ConfirmMutating bool

	// ConfirmMarked suspends the run only before calls of tools marked [completion.Tool.RequiresApproval] (see
	// [completion.RunOptions.ConfirmMarked]). Ignored without tools. Optional.
	ConfirmMarked bool

	// Context bounds the run: cancelling it aborts the in-flight provider request and stops the loop (see
	// [completion.RunOptions.Context]). What the run did until then is persisted, and the use case returns an
	// error satisfying errors.Is(err, context.Canceled). Nil means [context.Background]. Optional.
	Context context.Context

	// OnUsage is forwarded to [completion.RunOptions.OnUsage]. Optional.
	OnUsage func(completion.Usage)

	// BeforeFinish is forwarded to [completion.RunOptions.BeforeFinish], e.g. to join background tasks with
	// [completion.TaskGroup.BeforeFinish]. Ignored without tools. Optional.
	BeforeFinish func(ctx context.Context) string
}

// ResolveOptions carries the user's decisions on a pending run plus the runtime dependencies to continue it.
type ResolveOptions struct {
	// Run configures the continued run exactly like an [Append]. Its Input is ignored and Model must be empty
	// or equal the session model, because a pending run cannot move to another model.
	Run AppendOptions

	// Revision must equal [Session.PendingRevision].
	Revision int

	// Resolutions decide every pending call exactly once.
	Resolutions []completion.Resolution
}

// ErrNoPendingDecision is returned by [Resolve] and [Dismiss] when the session does not wait on the user (any
// more), or the revision does not match, e.g. because the question was already answered elsewhere.
var ErrNoPendingDecision = errors.New("session has no matching pending decision")

// ErrPendingDecision is returned by [Append] while the session waits on a user decision.
var ErrPendingDecision = errors.New("session waits on a pending decision")

// Create persists a new, optionally pre-seeded [Session].
type Create func(subject auth.Subject, opts CreateOptions) (Session, error)

// FindByID returns the session with the given id if it exists and the subject may read it.
type FindByID func(subject auth.Subject, id ID) (option.Opt[Session], error)

// FindAllOptions filters the sessions yielded by [FindAll]. The zero value applies no filter and yields every
// session the subject may see.
type FindAllOptions struct {
	// Tags, when non-empty, restricts the result to sessions carrying ALL of the given tags (set/AND
	// semantics). An empty slice means "no tag filter". This is how a caller scopes sessions to an
	// application context, e.g. Tags: []string{"ctx:invoice/42"}.
	Tags []string

	// IncludeChildren also yields child sessions (see [Session.ParentID]). They hold the transcripts of
	// sub-agents and are hidden by default, because nobody continues them.
	IncludeChildren bool
}

// FindAll yields the sessions the subject may see (per ReBAC), optionally narrowed by [FindAllOptions].
type FindAll func(subject auth.Subject, opts FindAllOptions) iter.Seq2[Session, error]

// Append adds a user turn, runs the completion (optionally agentic) against the supplied provider capability,
// appends the produced messages to the history and persists the updated session, which it returns.
type Append func(subject auth.Subject, id ID, opts AppendOptions) (Session, error)

// Resolve answers the pending decisions of a suspended run and continues it. The run may suspend again.
type Resolve func(subject auth.Subject, id ID, opts ResolveOptions) (Session, error)

// Dismiss closes the pending decisions without an answer and without asking the model, e.g. when the user moves
// on to another topic. Already executed calls keep their results. Background tasks of the session (see
// [UseCases.Tasks]) are cancelled.
type Dismiss func(subject auth.Subject, id ID, revision int) (Session, error)

// Rename changes the human-readable title of a session.
type Rename func(subject auth.Subject, id ID, title string) error

// Delete removes a session and its embedded history, together with its child sessions (see [Session.ParentID]).
// Background tasks of the session (see [UseCases.Tasks]) are cancelled.
type Delete func(subject auth.Subject, id ID) error

// UseCases bundles all session use cases. Construct it with [NewUseCases].
type UseCases struct {
	Create   Create
	FindByID FindByID
	FindAll  FindAll
	Append   Append
	Resolve  Resolve
	Dismiss  Dismiss
	Rename   Rename
	Delete   Delete

	// Tasks keeps the background tasks of running conversations, keyed by session id (see
	// [completion.NewTaskTools]). [Dismiss] and [Delete] cancel the tasks of their session.
	Tasks *completion.TaskRegistry

	// subUsage collects the usage of sub-agents until their parent session is saved next, see
	// [Session.SubUsage].
	subUsage *usageLedger
	// repo and locks let [NewSubRunner] keep the task of a sub-agent which failed before its first turn.
	repo  Repository
	locks *locker
}

// NewUseCases wires the session use cases against the given repository and ReBAC database.
//
// Mutating operations on an existing session ([Append], [Rename], [Delete]) serialize per session id via a
// keyed lock, so a long-running [Append] only blocks other
// operations on the same session, never on unrelated ones. [Create] needs no lock because it works on a
// freshly generated, collision-free id.
//
// Authorization is resource-scoped: every use case audits via [auth.Subject.AuditResource] against the
// session's ReBAC instance, so a subject may act either through a global permission (e.g. an IAM group that
// is allowed to create/list sessions) or through an instance grant. [Create] writes such an instance grant
// for the creator, so users can see and continue only their own sessions unless additionally granted global
// access.
// Option configures [NewUseCases].
type Option func(o *useCaseOptions)

type useCaseOptions struct {
	onDelete []func(id ID) error
}

// OnDelete is invoked for every session right before it is deleted, also for each of its child sessions, e.g. to
// release the files it uploaded to a provider. An error aborts the deletion. Several hooks run in the order of
// their options.
func OnDelete(fn func(id ID) error) Option {
	return func(o *useCaseOptions) {
		o.onDelete = append(o.onDelete, fn)
	}
}

// deleteHook runs all hooks of [OnDelete], nil without any.
func (o useCaseOptions) deleteHook() func(id ID) error {
	if len(o.onDelete) == 0 {
		return nil
	}

	return func(id ID) error {
		for _, fn := range o.onDelete {
			if err := fn(id); err != nil {
				return err
			}
		}

		return nil
	}
}

func NewUseCases(repo Repository, rdb *rebac.DB, opts ...Option) UseCases {
	var locks locker
	tasks := completion.NewTaskRegistry()
	ledger := &usageLedger{}
	var options useCaseOptions
	for _, opt := range opts {
		opt(&options)
	}

	return UseCases{
		Create:   NewCreate(repo, rdb),
		FindByID: NewFindByID(repo),
		FindAll:  NewFindAll(repo),
		Append:   NewAppend(&locks, repo, ledger),
		Resolve:  NewResolve(&locks, repo, ledger),
		Dismiss:  NewDismiss(&locks, repo, ledger, tasks),
		Rename:   NewRename(&locks, repo),
		Delete:   NewDelete(&locks, repo, rdb, tasks, ledger, options.deleteHook()),
		Tasks:    tasks,
		subUsage: ledger,
		repo:     repo,
		locks:    &locks,
	}
}
