// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"iter"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

// OutgoingFilter narrows the result of [FindOutgoingIDs]. The zero value matches all mails.
type OutgoingFilter struct {
	Status []Status // empty matches all
	Server string   // empty matches all
	Since  time.Time
	// Stuck matches mails, which are not yet sent and queued before the given duration.
	Stuck time.Duration
}

func (f OutgoingFilter) matches(o Outgoing, now time.Time) bool {
	if len(f.Status) > 0 && !slices.Contains(f.Status, o.Status) {
		return false
	}

	if f.Server != "" && o.ServerName != f.Server {
		return false
	}

	if !f.Since.IsZero() && o.QueuedAt.Before(f.Since) {
		return false
	}

	if f.Stuck > 0 {
		if o.done() || now.Sub(o.QueuedAt) < f.Stuck {
			return false
		}
	}

	return true
}

// FindOutgoingIDs returns the identifiers of all matching mails, the latest queued first.
type FindOutgoingIDs func(subject auth.Subject, filter OutgoingFilter) iter.Seq2[ID, error]
type FindOutgoingByID func(subject auth.Subject, id ID) (option.Opt[Outgoing], error)
type DeleteOutgoingByID func(subject auth.Subject, id ID) error

// RetryOutgoing puts the given mails back into the queue. The attempt history is kept.
type RetryOutgoing func(subject auth.Subject, ids ...ID) error

// ResendOutgoing creates a new queued copy of the given mail and returns its identifier.
type ResendOutgoing func(subject auth.Subject, id ID) (ID, error)

func NewFindOutgoingIDs(repo Repository) FindOutgoingIDs {
	return func(subject auth.Subject, filter OutgoingFilter) iter.Seq2[ID, error] {
		return func(yield func(ID, error) bool) {
			if err := subject.Audit(PermOutgoingFindAll); err != nil {
				yield("", err)
				return
			}

			type entry struct {
				id       ID
				queuedAt time.Time
			}

			now := time.Now()
			var tmp []entry
			for o, err := range repo.All() {
				if err != nil {
					yield("", err)
					return
				}

				if filter.matches(o, now) {
					tmp = append(tmp, entry{o.ID, o.QueuedAt})
				}
			}

			slices.SortFunc(tmp, func(a, b entry) int {
				if c := b.queuedAt.Compare(a.queuedAt); c != 0 {
					return c
				}

				return strings.Compare(string(a.id), string(b.id))
			})

			for _, e := range tmp {
				if !yield(e.id, nil) {
					return
				}
			}
		}
	}
}

func NewFindOutgoingByID(repo Repository) FindOutgoingByID {
	return func(subject auth.Subject, id ID) (option.Opt[Outgoing], error) {
		if err := subject.Audit(PermOutgoingFindByID); err != nil {
			return option.None[Outgoing](), err
		}

		return repo.FindByID(id)
	}
}

func NewDeleteOutgoingByID(mutex *sync.Mutex, repo Repository) DeleteOutgoingByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermOutgoingDeleteByID); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		return repo.DeleteByID(id)
	}
}

func NewRetryOutgoing(mutex *sync.Mutex, repo Repository) RetryOutgoing {
	return func(subject auth.Subject, ids ...ID) error {
		if err := subject.Audit(PermOutgoingRetry); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		for _, id := range ids {
			optOut, err := repo.FindByID(id)
			if err != nil {
				return err
			}

			if optOut.IsNone() {
				continue
			}

			out := optOut.Unwrap()
			if out.Status == StatusSendSuccess {
				continue // use resend instead
			}

			if out.Status == StatusSuppressed {
				out.BypassSpamGuard = true // explicitly requested by an administrator
			}

			out.Status = StatusQueued
			out.NextAttemptAt = time.Time{}
			if err := repo.Save(out); err != nil {
				return err
			}
		}

		return nil
	}
}

func NewResendOutgoing(mutex *sync.Mutex, repo Repository) ResendOutgoing {
	return func(subject auth.Subject, id ID) (ID, error) {
		if err := subject.Audit(PermOutgoingResend); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optOut, err := repo.FindByID(id)
		if err != nil {
			return "", err
		}

		if optOut.IsNone() {
			return "", os.ErrNotExist
		}

		src := optOut.Unwrap()
		out := Outgoing{
			ID:       data.RandIdent[ID](),
			Mail:     src.Mail,
			Subject:  src.Subject,
			Receiver: src.Receiver,
			Status:   StatusQueued,
			QueuedAt: time.Now(),
		}

		if err := repo.Save(out); err != nil {
			return "", err
		}

		return out.ID, nil
	}
}
