// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

// usageLedger collects the usage of sub-agents per parent session until the parent is saved next. It exists so
// a sub-agent never has to write its parent: the parent is locked for its whole run while a synchronous
// delegation waits, and a background task must not contend with the conversation it belongs to.
type usageLedger struct {
	mu      sync.Mutex
	pending map[ID]completion.Usage
}

func (l *usageLedger) add(id ID, u completion.Usage) {
	if l == nil || u.IsZero() {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.pending == nil {
		l.pending = map[ID]completion.Usage{}
	}
	l.pending[id] = l.pending[id].Add(u)
}

func (l *usageLedger) drain(id ID) completion.Usage {
	if l == nil {
		return completion.Usage{}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	u := l.pending[id]
	delete(l.pending, id)
	return u
}

// NewSubRunner returns a [completion.SubRunner] which persists every sub-agent as a child session of parentID
// (see [Session.ParentID]) and runs it through [UseCases.Append], so its transcript can be inspected later.
// The child is created for the acting subject, who becomes its owner exactly like for [Create].
//
// The usage of each sub-agent is added to [Session.SubUsage] of the parent with the parent's next save; the
// parent itself is never written by a sub-agent.
func NewSubRunner(uc UseCases, parentID ID) completion.SubRunner {
	return func(ctx context.Context, subject auth.Subject, req completion.SubRunRequest) (completion.SubRunResult, error) {
		// a sub-agent cancelled before it started must not leave an empty child session behind
		if err := ctx.Err(); err != nil {
			return completion.SubRunResult{}, err
		}

		child, err := uc.Create(subject, CreateOptions{
			Title:        req.Title,
			Model:        req.Options.Model,
			System:       req.Options.System,
			ParentID:     parentID,
			ParentCallID: req.ParentCallID,
		})
		if err != nil {
			return completion.SubRunResult{}, fmt.Errorf("cannot create child session: %w", err)
		}

		res := completion.SubRunResult{SessionID: string(child.ID)}

		// The usage of every single completion is booked as it happens, so a failed or cancelled sub-agent is
		// accounted as well.
		var mu sync.Mutex
		onUsage := func(u completion.Usage) {
			mu.Lock()
			res.Usage = res.Usage.Add(u)
			mu.Unlock()
			uc.subUsage.add(parentID, u)
		}

		updated, err := uc.Append(subject, child.ID, AppendOptions{
			Agentic:          true,
			Completions:      req.Completions,
			Input:            req.Input,
			Model:            req.Options.Model,
			System:           req.Options.System,
			Tools:            req.Tools,
			FileUploader:     req.FileUploader,
			MaxTokens:        req.Options.MaxTokens,
			Temperature:      req.Options.Temperature,
			OnProgress:       req.OnProgress,
			MaxTurns:         req.MaxTurns,
			OnBeforeToolCall: req.OnBeforeToolCall,
			Context:          ctx,
			OnUsage:          onUsage,
		})

		mu.Lock()
		defer mu.Unlock()

		if err != nil {
			if reloaded, ok, _ := findChild(uc, subject, child.ID); ok {
				res.History = reloaded.Messages

				switch {
				case len(reloaded.Messages) > 0:
					// the run kept what it did
				case ctx.Err() != nil:
					// stopped or timed out during the first request: an empty child has nothing to inspect
					if uc.Delete != nil {
						if derr := uc.Delete(subject, child.ID); derr == nil {
							res.SessionID = ""
						}
					}
				case uc.repo != nil && uc.locks != nil:
					// Failed before its first turn: at least the task stays inspectable. The child is written under
					// its lock and only if it still exists, so a deletion meanwhile is not undone.
					if release, lerr := uc.locks.lock(ctx, child.ID); lerr == nil {
						if cur, ferr := uc.repo.FindByID(child.ID); ferr == nil && cur.IsSome() {
							c := cur.Unwrap()
							c.Messages = []completion.Message{{Role: completion.User, Content: req.Input}}
							c.UpdatedAt = xtime.Now()
							if serr := uc.repo.Save(c); serr == nil {
								res.History = c.Messages
							}
						}
						release()
					}
				}
			}
			return res, err
		}

		res.History = updated.Messages

		if updated.Pending != nil {
			// Leave a valid, closed transcript behind rather than a question nobody will ever answer.
			if dismissed, derr := uc.Dismiss(subject, child.ID, updated.PendingRevision); derr == nil {
				res.History = dismissed.Messages
			}
			return res, completion.ErrSubRunSuspended
		}

		res.Answer = completion.FinalAnswer(updated.Messages)
		return res, nil
	}
}

// findChild loads a child session for the acting subject.
func findChild(uc UseCases, subject auth.Subject, id ID) (Session, bool, error) {
	if uc.FindByID == nil {
		return Session{}, false, errors.New("no FindByID use case")
	}

	opt, err := uc.FindByID(subject, id)
	if err != nil || opt.IsNone() {
		return Session{}, false, err
	}

	return opt.Unwrap(), true, nil
}
