// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

// NewAppend returns an [Append] use case.
//
// It performs a read-modify-write cycle guarded by the session's keyed lock: load the session, append the new
// user turn, run the completion against the supplied [completion.Completions] (agentic via [completion.Run]
// when tools are given, otherwise a single [completion.Completions.Complete]), append the produced messages,
// update the accumulated usage and timestamp, persist and return the updated session.
//
// The potentially long-running provider call happens while the per-session lock is held so that a concurrent
// Append on the SAME session cannot build on a stale history; operations on other sessions are unaffected.
// Callers should run Append off the UI thread.
func NewAppend(locks *locker, repo Repository) Append {
	return func(subject auth.Subject, id ID, opts AppendOptions) (Session, error) {
		if opts.Completions == nil {
			return Session{}, fmt.Errorf("session: AppendOptions.Completions must not be nil")
		}

		if len(opts.Input) == 0 {
			return Session{}, fmt.Errorf("session: AppendOptions.Input must not be empty")
		}

		defer locks.lock(id)()

		optSession, err := repo.FindByID(id)
		if err != nil {
			return Session{}, fmt.Errorf("cannot load session: %w", err)
		}

		if optSession.IsNone() {
			return Session{}, fmt.Errorf("session %q does not exist", id)
		}

		// Resource-scoped authorization: the subject needs PermAppend globally or as an instance grant.
		// A denied audit is reported like a missing session so foreign sessions are not revealed.
		if err := subject.AuditResource(Namespace, rebacInstance(id), PermAppend); err != nil {
			return Session{}, fmt.Errorf("session %q does not exist", id)
		}

		session := optSession.Unwrap()

		mdl := opts.Model
		if mdl == "" {
			mdl = session.Model
		}
		if mdl == "" {
			return Session{}, fmt.Errorf("session: no model set (neither on the session nor in AppendOptions)")
		}

		// System prompt: a per-turn override wins over the session's fixed prompt. This lets callers rebuild
		// a fresh system prompt each turn (e.g. embedding the currently rendered domain model). The override
		// is transient and never persisted on the session.
		system := session.System
		if opts.System != "" {
			system = opts.System
		}

		// Build the request history: the persisted history plus the new user turn.
		userMsg := completion.Message{Role: completion.User, Content: opts.Input}
		history := make([]completion.Message, 0, len(session.Messages)+1)
		history = append(history, session.Messages...)
		history = append(history, userMsg)

		baseOpts := completion.Options{
			Model:       mdl,
			System:      system,
			Messages:    history,
			MaxTokens:   opts.MaxTokens,
			Temperature: opts.Temperature,
		}

		if session.Pending != nil {
			return Session{}, ErrPendingDecision
		}

		if len(opts.Tools) == 0 {
			res, cerr := opts.Completions.Complete(subject, baseOpts)
			if cerr != nil {
				return Session{}, fmt.Errorf("completion failed: %w", cerr)
			}
			// A single turn: our request history plus the assistant answer.
			newHistory := history
			if len(res.Message.Content) > 0 {
				newHistory = append(newHistory, res.Message)
			}
			return saveOutcome(repo, session, mdl, completion.Outcome{Result: res, History: newHistory})
		}

		// Agentic loop: the outcome carries the full trace (starting from our history) including all
		// intermediate tool calls and tool results, and possibly a suspension on a user decision.
		out, rerr := completion.Start(subject, opts.Completions, runOptions(baseOpts, opts))
		if rerr != nil {
			return Session{}, persistFailure(repo, session, mdl, out, rerr)
		}

		return saveOutcome(repo, session, mdl, out)
	}
}

// runOptions builds the agentic run configuration of an [Append] or [Resolve].
func runOptions(base completion.Options, opts AppendOptions) completion.RunOptions {
	return completion.RunOptions{
		Options:          base,
		Tools:            opts.Tools,
		MaxTurns:         opts.MaxTurns,
		OnProgress:       opts.OnProgress,
		FileUploader:     opts.FileUploader,
		OnBeforeToolCall: opts.OnBeforeToolCall,
		ConfirmMutating:  opts.ConfirmMutating,
	}
}

// saveOutcome stores the history of a finished or suspended run.
func saveOutcome(repo Repository, session Session, mdl model.ID, out completion.Outcome) (Session, error) {
	session.Messages = out.History
	session.Model = mdl
	session.Usage = addUsage(session.Usage, out.Result.Usage)
	session.UpdatedAt = xtime.Now()
	session.Pending = out.Suspended
	if out.Suspended != nil {
		session.PendingRevision++
	}

	if err := repo.Save(session); err != nil {
		return Session{}, fmt.Errorf("cannot persist session: %w", err)
	}

	return session, nil
}

// persistFailure keeps what a failed run already did. Tools that ran may have had side effects, so the
// session must reflect them and a follow-up question builds on them. The history of a failed run never
// contains a tool_use without its tool_result.
func persistFailure(repo Repository, session Session, mdl model.ID, out completion.Outcome, runErr error) error {
	if !out.Progressed {
		return fmt.Errorf("completion run failed: %w", runErr)
	}

	session.Messages = out.History
	session.Model = mdl
	session.Pending = nil
	session.UpdatedAt = xtime.Now()
	if err := repo.Save(session); err != nil {
		return fmt.Errorf("completion run failed: %w (and cannot persist partial session: %v)", runErr, err)
	}

	return fmt.Errorf("completion run failed: %w", runErr)
}

// addUsage accumulates token usage across turns.
func addUsage(a, b completion.Usage) completion.Usage {
	return completion.Usage{
		InputTokens:      a.InputTokens + b.InputTokens,
		OutputTokens:     a.OutputTokens + b.OutputTokens,
		CacheReadTokens:  a.CacheReadTokens + b.CacheReadTokens,
		CacheWriteTokens: a.CacheWriteTokens + b.CacheWriteTokens,
	}
}
