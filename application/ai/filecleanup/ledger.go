// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package filecleanup deletes the files, which a conversation uploaded to an AI provider, together with the
// conversation. A provider like Anthropic keeps an uploaded file until it is deleted, and it may contain personal
// data. Every owned upload is recorded in a persistent ledger. Releasing the owner, e.g. by deleting the session,
// makes its files due, and a worker deletes them at the provider and retries until it succeeds.
package filecleanup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
)

const (
	// TransientLifetime is the time after which a file of the [file.TransientOwner] is deleted.
	TransientLifetime = 24 * time.Hour

	// maxBackoff bounds the time between two attempts to delete a file.
	maxBackoff = 24 * time.Hour

	// complainAfter is the number of failed attempts, after which a failure is logged as an error.
	complainAfter = 5

	// pollInterval is the time between two passes of the worker, without a release in the meantime.
	pollInterval = time.Minute
)

// Entry records a file which has been uploaded to a provider on behalf of an owner.
type Entry struct {
	ID        string      `json:"id"`
	Provider  provider.ID `json:"provider"`
	File      file.ID     `json:"file"`
	Owner     file.Owner  `json:"owner"`
	CreatedAt time.Time   `json:"createdAt"`

	// DueAt is the time from which the file is to be deleted. It is zero, as long as the owner uses the file.
	DueAt time.Time `json:"dueAt,omitzero"`

	Attempts  int    `json:"attempts,omitzero"`
	LastError string `json:"lastError,omitempty"`
}

func (e Entry) Identity() string {
	return e.ID
}

func entryID(prov provider.ID, id file.ID) string {
	return string(prov) + "/" + string(id)
}

type Repository = data.Repository[Entry, string]

// FindProvider resolves a provider by its identity, see [provider.Provider.Identity].
type FindProvider func(id provider.ID) (std.Option[provider.Provider], error)

// Ledger records owned uploads and deletes the files of released owners.
type Ledger struct {
	repo         Repository
	findProvider FindProvider
	subject      func() auth.Subject
	now          func() time.Time
	mutex        sync.Mutex // serializes the read-modify-write of entries
	wake         chan struct{}
}

// New creates a ledger. The subject is used to delete files at the provider. Call [Ledger.Run] to start the
// worker.
func New(repo Repository, findProvider FindProvider, subject func() auth.Subject) *Ledger {
	return &Ledger{
		repo:         repo,
		findProvider: findProvider,
		subject:      subject,
		now:          time.Now,
		wake:         make(chan struct{}, 1),
	}
}

// Decorate wraps the Files capability of the provider, so that every upload with an [file.CreateOptions.Owner] is
// recorded. It fits the decorator of [go.wdy.de/nago/application/ai.NewUseCases].
func (l *Ledger) Decorate(prov provider.Provider) (provider.Provider, error) {
	optFiles := prov.Files()
	if optFiles.IsNone() {
		return prov, nil
	}

	return ownedProvider{Provider: prov, files: ownedFiles{Files: optFiles.Unwrap(), ledger: l, provider: prov.Identity()}}, nil
}

// hold records an uploaded file of the owner.
func (l *Ledger) hold(prov provider.ID, id file.ID, owner file.Owner) error {
	now := l.now()
	entry := Entry{ID: entryID(prov, id), Provider: prov, File: id, Owner: owner, CreatedAt: now}
	if owner == file.TransientOwner {
		entry.DueAt = now.Add(TransientLifetime)
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.repo.Save(entry)
}

// Release makes all files of the owner due for deletion and wakes the worker. An error tells that the release has
// not been recorded, e.g. the deletion of the session should be aborted.
func (l *Ledger) Release(owner file.Owner) error {
	if owner == "" {
		return nil
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	now := l.now()
	released := 0
	for entry, err := range l.repo.All() {
		if err != nil {
			return fmt.Errorf("cannot read provider files: %w", err)
		}

		if entry.Owner != owner || !entry.DueAt.IsZero() {
			continue
		}

		entry.DueAt = now
		if err := l.repo.Save(entry); err != nil {
			return fmt.Errorf("cannot release provider file: %w", err)
		}

		released++
	}

	if released > 0 {
		l.notify()
	}

	return nil
}

func (l *Ledger) notify() {
	select {
	case l.wake <- struct{}{}:
	default:
	}
}

// Run deletes due files until the context is done.
func (l *Ledger) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		l.process()

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-l.wake:
		}
	}
}

// process deletes all due files once. A failure is retried with an exponential backoff, a file which the
// provider does not know anymore counts as deleted.
func (l *Ledger) process() {
	now := l.now()
	var due []Entry
	for entry, err := range l.repo.All() {
		if err != nil {
			slog.Error("cannot read provider files to delete", "err", err)
			return
		}

		if !entry.DueAt.IsZero() && !entry.DueAt.After(now) {
			due = append(due, entry)
		}
	}

	for _, entry := range due {
		err := l.delete(entry)
		if err == nil {
			if err := l.forget(entry); err != nil {
				slog.Error("cannot forget deleted provider file", "file", entry.ID, "err", err)
			}
			continue
		}

		l.retryLater(entry, err)
	}
}

func (l *Ledger) delete(entry Entry) error {
	optProv, err := l.findProvider(entry.Provider)
	if err != nil {
		return fmt.Errorf("cannot find provider: %w", err)
	}

	if optProv.IsNone() {
		return fmt.Errorf("provider %s is not configured", entry.Provider)
	}

	optFiles := optProv.Unwrap().Files()
	if optFiles.IsNone() {
		return fmt.Errorf("provider %s cannot delete files", entry.Provider)
	}

	err = optFiles.Unwrap().Delete(l.subject(), entry.File)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

// forget removes the entry of a deleted file.
func (l *Ledger) forget(entry Entry) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.repo.DeleteByID(entry.ID)
}

func (l *Ledger) retryLater(entry Entry, cause error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	entry.Attempts++
	entry.LastError = cause.Error()
	entry.DueAt = l.now().Add(backoff(entry.Attempts))
	if err := l.repo.Save(entry); err != nil {
		slog.Error("cannot schedule the retry of a provider file deletion", "file", entry.ID, "err", err)
	}

	if entry.Attempts >= complainAfter {
		slog.Error("cannot delete provider file, will retry", "file", entry.ID, "attempts", entry.Attempts, "next", entry.DueAt, "err", cause)
		return
	}

	slog.Warn("cannot delete provider file, will retry", "file", entry.ID, "attempts", entry.Attempts, "next", entry.DueAt, "err", cause)
}

// backoff returns the time until the next attempt after the given number of failed attempts.
func backoff(attempts int) time.Duration {
	d := time.Minute
	for range attempts - 1 {
		d *= 2
		if d >= maxBackoff {
			return maxBackoff
		}
	}

	return d
}

type ownedProvider struct {
	provider.Provider
	files ownedFiles
}

func (p ownedProvider) Files() std.Option[provider.Files] {
	return std.Some[provider.Files](p.files)
}

type ownedFiles struct {
	provider.Files
	ledger   *Ledger
	provider provider.ID
}

// Put records the upload of an owner. If that fails, the file is deleted again, because nobody would delete it
// otherwise.
func (f ownedFiles) Put(subject auth.Subject, opts file.CreateOptions) (file.File, error) {
	created, err := f.Files.Put(subject, opts)
	if err != nil || opts.Owner == "" {
		return created, err
	}

	if err := f.ledger.hold(f.provider, created.ID, opts.Owner); err != nil {
		if derr := f.Files.Delete(subject, created.ID); derr != nil {
			slog.Error("cannot delete unrecorded provider file", "file", created.ID, "err", derr)
		}

		return file.File{}, fmt.Errorf("cannot record provider file: %w", err)
	}

	return created, nil
}
