// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"
	"sync"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

// NewAppend returns an [Append] use case.
//
// It performs a read-modify-write cycle guarded by the shared mutex: load the session, append the new user
// turn, run the completion against the supplied [completion.Completions] (agentic via [completion.Run] when
// tools are given, otherwise a single [completion.Completions.Complete]), append the produced messages, update
// the accumulated usage and timestamp, persist and return the updated session.
//
// Note that the potentially long-running provider call happens while the mutex is held so that a concurrent
// Append on the same session cannot build on a stale history. Callers should run Append off the UI thread.
func NewAppend(mutex *sync.Mutex, repo Repository) Append {
	return func(subject auth.Subject, id ID, opts AppendOptions) (Session, error) {
		if opts.Completions == nil {
			return Session{}, fmt.Errorf("session: AppendOptions.Completions must not be nil")
		}

		if len(opts.Input) == 0 {
			return Session{}, fmt.Errorf("session: AppendOptions.Input must not be empty")
		}

		mutex.Lock()
		defer mutex.Unlock()

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

		// Build the request history: the persisted history plus the new user turn.
		userMsg := completion.Message{Role: completion.User, Content: opts.Input}
		history := make([]completion.Message, 0, len(session.Messages)+1)
		history = append(history, session.Messages...)
		history = append(history, userMsg)

		baseOpts := completion.Options{
			Model:       mdl,
			System:      session.System,
			Messages:    history,
			MaxTokens:   opts.MaxTokens,
			Temperature: opts.Temperature,
		}

		var (
			result     completion.Result
			newHistory []completion.Message
		)

		if len(opts.Tools) > 0 {
			// Agentic loop: completion.Run returns the full trace (starting from our history) including all
			// intermediate tool calls and tool results.
			res, runHistory, rerr := completion.Run(subject, opts.Completions, completion.RunOptions{
				Options:    baseOpts,
				Tools:      opts.Tools,
				MaxTurns:   opts.MaxTurns,
				OnProgress: opts.OnProgress,
			})
			if rerr != nil {
				return Session{}, fmt.Errorf("completion run failed: %w", rerr)
			}
			result = res
			newHistory = runHistory
		} else {
			res, cerr := opts.Completions.Complete(subject, baseOpts)
			if cerr != nil {
				return Session{}, fmt.Errorf("completion failed: %w", cerr)
			}
			result = res
			// A single turn: our request history plus the assistant answer.
			newHistory = append(history, res.Message)
		}

		session.Messages = newHistory
		session.Model = mdl
		session.Usage = addUsage(session.Usage, result.Usage)
		session.UpdatedAt = xtime.Now()

		if err := repo.Save(session); err != nil {
			return Session{}, fmt.Errorf("cannot persist session: %w", err)
		}

		return session, nil
	}
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
