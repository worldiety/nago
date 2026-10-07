// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package filecleanup

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	datamem "go.wdy.de/nago/pkg/data/mem"
	"go.wdy.de/nago/pkg/std"
)

// fakeFiles keeps the uploaded files like a provider and fails deletions on demand.
type fakeFiles struct {
	mutex    sync.Mutex
	files    map[file.ID]bool
	next     int
	failWith error
}

func (f *fakeFiles) All(auth.Subject) iter.Seq2[file.File, error] { return nil }

func (f *fakeFiles) FindByID(auth.Subject, file.ID) (std.Option[file.File], error) {
	return std.None[file.File](), nil
}

func (f *fakeFiles) Get(auth.Subject, file.ID) (std.Option[io.ReadCloser], error) {
	return std.None[io.ReadCloser](), nil
}

func (f *fakeFiles) Put(_ auth.Subject, opts file.CreateOptions) (file.File, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.next++
	id := file.ID(fmt.Sprintf("file-%d", f.next))
	f.files[id] = true
	return file.File{ID: id, Name: opts.Name}, nil
}

func (f *fakeFiles) Delete(_ auth.Subject, id file.ID) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.failWith != nil {
		return f.failWith
	}

	if !f.files[id] {
		return fmt.Errorf("file %s: %w", id, os.ErrNotExist)
	}

	delete(f.files, id)
	return nil
}

func (f *fakeFiles) exists(id file.ID) bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.files[id]
}

type fakeProvider struct{ files *fakeFiles }

func (p fakeProvider) Identity() provider.ID { return "anthropic-1" }
func (p fakeProvider) Name() string          { return "fake" }
func (p fakeProvider) Description() string   { return "" }
func (p fakeProvider) Files() std.Option[provider.Files] {
	return std.Some[provider.Files](p.files)
}
func (p fakeProvider) Completions() std.Option[completion.Completions] {
	return std.None[completion.Completions]()
}

type fixture struct {
	ledger *Ledger
	repo   *datamem.Repository[Entry, string]
	files  *fakeFiles
	prov   provider.Provider // decorated
	clock  *time.Time
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	files := &fakeFiles{files: map[file.ID]bool{}}
	repo := &datamem.Repository[Entry, string]{}
	var known provider.Provider = fakeProvider{files: files}
	ledger := New(repo, func(id provider.ID) (std.Option[provider.Provider], error) {
		if id != known.Identity() {
			return std.None[provider.Provider](), nil
		}
		return std.Some(known), nil
	}, func() auth.Subject { return user.SU() })

	clock := time.Now()
	ledger.now = func() time.Time { return clock }

	prov, err := ledger.Decorate(known)
	if err != nil {
		t.Fatal(err)
	}

	return fixture{ledger: ledger, repo: repo, files: files, prov: prov, clock: &clock}
}

func (f fixture) upload(t *testing.T, owner file.Owner) file.ID {
	t.Helper()
	created, err := f.prov.Files().Unwrap().Put(user.SU(), file.CreateOptions{
		Name:  "a.pdf",
		Owner: owner,
		Open:  func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("pdf")), nil },
	})
	if err != nil {
		t.Fatal(err)
	}

	return created.ID
}

func (f fixture) advance(d time.Duration) {
	*f.clock = f.clock.Add(d)
}

func (f fixture) entries(t *testing.T) int {
	t.Helper()
	n, err := f.repo.Count()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// The files of a session stay at the provider while the session exists and are deleted after its release, but no
// file of another session and no file without an owner.
func TestReleaseDeletesTheFilesOfTheOwner(t *testing.T) {
	f := newFixture(t)
	mine := f.upload(t, "session-1")
	other := f.upload(t, "session-2")
	unowned := f.upload(t, "")

	if f.entries(t) != 2 {
		t.Fatalf("expected two recorded files, got %d", f.entries(t))
	}

	f.ledger.process()
	if !f.files.exists(mine) {
		t.Fatal("a file must stay while its owner holds it")
	}

	if err := f.ledger.Release("session-1"); err != nil {
		t.Fatal(err)
	}

	f.ledger.process()
	if f.files.exists(mine) || !f.files.exists(other) || !f.files.exists(unowned) {
		t.Fatal("expected only the file of the released session to be deleted")
	}

	if f.entries(t) != 1 {
		t.Fatalf("expected the deleted file to be forgotten, got %d entries", f.entries(t))
	}
}

// A provider which cannot delete right now is asked again later, with growing pauses, until it succeeds.
func TestFailedDeletionIsRetried(t *testing.T) {
	f := newFixture(t)
	id := f.upload(t, "session-1")
	f.files.failWith = errors.New("provider unavailable")
	if err := f.ledger.Release("session-1"); err != nil {
		t.Fatal(err)
	}

	f.ledger.process()
	f.ledger.process() // not due again yet
	entry := std.Must(f.repo.FindByID(entryID("anthropic-1", id))).Unwrap()
	if entry.Attempts != 1 || entry.LastError == "" || !f.files.exists(id) {
		t.Fatalf("expected one failed attempt, got %+v", entry)
	}

	f.advance(backoff(1))
	f.ledger.process()
	entry = std.Must(f.repo.FindByID(entryID("anthropic-1", id))).Unwrap()
	if entry.Attempts != 2 || !entry.DueAt.Equal(f.clock.Add(backoff(2))) {
		t.Fatalf("expected a second attempt with a longer pause, got %+v", entry)
	}

	f.files.failWith = nil
	f.advance(backoff(2))
	f.ledger.process()
	if f.files.exists(id) || f.entries(t) != 0 {
		t.Fatal("expected the file to be deleted once the provider recovered")
	}
}

// A file the provider does not know anymore, e.g. deleted by hand, counts as deleted.
func TestUnknownFileCountsAsDeleted(t *testing.T) {
	f := newFixture(t)
	id := f.upload(t, "session-1")
	if err := f.files.Delete(user.SU(), id); err != nil {
		t.Fatal(err)
	}

	if err := f.ledger.Release("session-1"); err != nil {
		t.Fatal(err)
	}

	f.ledger.process()
	if f.entries(t) != 0 {
		t.Fatal("expected an unknown file to be forgotten")
	}
}

// Files of a conversation without a session are deleted a day after their upload, because nobody releases them.
func TestTransientFilesExpire(t *testing.T) {
	f := newFixture(t)
	id := f.upload(t, file.TransientOwner)

	f.advance(TransientLifetime - time.Minute)
	f.ledger.process()
	if !f.files.exists(id) {
		t.Fatal("a transient file must stay for its lifetime")
	}

	f.advance(time.Minute)
	f.ledger.process()
	if f.files.exists(id) {
		t.Fatal("a transient file must be deleted after its lifetime")
	}
}

// A provider which has been removed from the configuration keeps its files in the ledger, so that they are
// deleted once it is back.
func TestUnknownProviderIsRetried(t *testing.T) {
	f := newFixture(t)
	if err := f.repo.Save(Entry{ID: "gone/file-1", Provider: "gone", File: "file-1", Owner: "session-1", DueAt: f.clock.Add(-time.Second)}); err != nil {
		t.Fatal(err)
	}

	f.ledger.process()
	entry := std.Must(f.repo.FindByID("gone/file-1")).Unwrap()
	if entry.Attempts != 1 || !strings.Contains(entry.LastError, "not configured") {
		t.Fatalf("expected a retry for an unknown provider, got %+v", entry)
	}
}

func TestBackoffIsBounded(t *testing.T) {
	if backoff(1) != time.Minute || backoff(2) != 2*time.Minute || backoff(100) != maxBackoff {
		t.Fatalf("unexpected backoff %v %v %v", backoff(1), backoff(2), backoff(100))
	}
}

// If an upload cannot be recorded, nobody would ever delete it, so it is deleted at once and the upload fails.
func TestUnrecordedUploadIsRolledBack(t *testing.T) {
	f := newFixture(t)
	f.ledger.repo = failingRepo{Repository: f.repo}

	_, err := f.prov.Files().Unwrap().Put(user.SU(), file.CreateOptions{
		Name:  "a.pdf",
		Owner: "session-1",
		Open:  func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("pdf")), nil },
	})
	if err == nil {
		t.Fatal("expected the upload to fail")
	}

	if f.files.exists("file-1") {
		t.Fatal("expected the unrecorded file to be deleted")
	}
}

type failingRepo struct {
	Repository
}

func (failingRepo) Save(Entry) error {
	return errors.New("disk full")
}
