// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"iter"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

// fakeCompletions is a deterministic, in-memory [completion.Completions] used for tests. It records the last
// request it received (so we can assert the full history is resent) and echoes a canned assistant answer.
type fakeCompletions struct {
	lastOptions completion.Options
	reply       string
}

func (f *fakeCompletions) Models(subject auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) {
		yield(model.Model{ID: "fake-model"}, nil)
	}
}

func (f *fakeCompletions) Complete(subject auth.Subject, opts completion.Options) (completion.Result, error) {
	f.lastOptions = opts
	reply := f.reply
	if reply == "" {
		reply = "ok"
	}
	return completion.Result{
		Message: completion.Message{
			Role:    completion.Assistant,
			Content: []completion.Content{completion.Text{Text: reply}},
		},
		StopReason: completion.StopEndTurn,
		Usage:      completion.Usage{InputTokens: 10, OutputTokens: 5},
		Model:      opts.Model,
	}, nil
}

func (f *fakeCompletions) Stream(subject auth.Subject, opts completion.Options) iter.Seq2[completion.Delta, error] {
	return func(yield func(completion.Delta, error) bool) {}
}

// newTestRDB creates an in-memory ReBAC database with the session static rules registered, mirroring the
// wiring in application/ai/cfg so grantOwner may write its triples.
func newTestRDB(t *testing.T) *rebac.DB {
	t.Helper()
	rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatalf("cannot create rebac db: %v", err)
	}

	rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Owner, Target: Namespace})
	for _, pid := range InstancePermissions {
		rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Relation(pid), Target: Namespace})
	}
	return rdb
}

func newTestUseCases(t *testing.T) (UseCases, Repository, *rebac.DB) {
	repo := Repository(json.NewSloppyJSONRepository[Session, ID](mem.NewBlobStore(string(Namespace))))
	rdb := newTestRDB(t)
	return NewUseCases(repo, rdb), repo, rdb
}

func firstText(t *testing.T, msg completion.Message) string {
	t.Helper()
	for _, c := range msg.Content {
		if tx, ok := c.(completion.Text); ok {
			return tx.Text
		}
	}
	t.Fatalf("message %+v has no text content", msg)
	return ""
}

// TestCreateAppendPersists exercises the core flow: create a session, append two turns and assert the history
// is persisted, grows correctly and is fully resent to the provider on the second turn.
func TestCreateAppendPersists(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	subject := user.SU()

	fake := &fakeCompletions{reply: "first-answer"}

	session, err := uc.Create(subject, CreateOptions{
		Title:  "my chat",
		Model:  model.ID("fake-model"),
		System: "be nice",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if session.ID == "" {
		t.Fatal("expected a generated session id")
	}

	// First turn.
	updated, err := uc.Append(subject, session.ID, AppendOptions{
		Completions: fake,
		Input:       []completion.Content{completion.Text{Text: "hello"}},
	})
	if err != nil {
		t.Fatalf("append 1: %v", err)
	}

	// user + assistant
	if len(updated.Messages) != 2 {
		t.Fatalf("expected 2 messages after first turn, got %d", len(updated.Messages))
	}
	if got := firstText(t, updated.Messages[1]); got != "first-answer" {
		t.Fatalf("unexpected assistant answer %q", got)
	}
	// System prompt must be forwarded.
	if fake.lastOptions.System != "be nice" {
		t.Fatalf("system prompt not forwarded, got %q", fake.lastOptions.System)
	}
	// Usage accumulated.
	if updated.Usage.InputTokens != 10 || updated.Usage.OutputTokens != 5 {
		t.Fatalf("unexpected usage after first turn: %+v", updated.Usage)
	}

	// Second turn.
	fake.reply = "second-answer"
	updated, err = uc.Append(subject, session.ID, AppendOptions{
		Completions: fake,
		Input:       []completion.Content{completion.Text{Text: "how are you?"}},
	})
	if err != nil {
		t.Fatalf("append 2: %v", err)
	}

	// user, assistant, user, assistant
	if len(updated.Messages) != 4 {
		t.Fatalf("expected 4 messages after second turn, got %d", len(updated.Messages))
	}

	// The provider must have received the FULL history (3 messages: hello, first-answer, how are you?).
	if len(fake.lastOptions.Messages) != 3 {
		t.Fatalf("expected full history of 3 messages resent, got %d", len(fake.lastOptions.Messages))
	}
	if got := firstText(t, fake.lastOptions.Messages[0]); got != "hello" {
		t.Fatalf("history[0] mismatch: %q", got)
	}
	if got := firstText(t, fake.lastOptions.Messages[1]); got != "first-answer" {
		t.Fatalf("history[1] mismatch: %q", got)
	}

	// Usage must be cumulative across both turns.
	if updated.Usage.InputTokens != 20 || updated.Usage.OutputTokens != 10 {
		t.Fatalf("unexpected cumulative usage: %+v", updated.Usage)
	}

	// And it must actually be persisted: reload straight from the repository.
	reloaded, err := repo.FindByID(session.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.IsNone() {
		t.Fatal("session vanished from repository")
	}
	if len(reloaded.Unwrap().Messages) != 4 {
		t.Fatalf("persisted history has %d messages, want 4", len(reloaded.Unwrap().Messages))
	}
	if firstText(t, reloaded.Unwrap().Messages[3]) != "second-answer" {
		t.Fatal("persisted last message mismatch")
	}
}

// TestCreateWithInitialInput verifies that an initial input is stored but not yet completed.
func TestCreateWithInitialInput(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	session, err := uc.Create(subject, CreateOptions{
		Model: model.ID("fake-model"),
		Input: []completion.Content{completion.Text{Text: "seed"}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if len(session.Messages) != 1 {
		t.Fatalf("expected 1 seeded message, got %d", len(session.Messages))
	}
	if session.Messages[0].Role != completion.User {
		t.Fatal("seeded message should be a user turn")
	}
}

// TestAppendRequiresModel ensures a completion cannot run without a model.
func TestAppendRequiresModel(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	session, err := uc.Create(subject, CreateOptions{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = uc.Append(subject, session.ID, AppendOptions{
		Completions: &fakeCompletions{},
		Input:       []completion.Content{completion.Text{Text: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error when no model is set")
	}
}

// TestRenameAndDelete covers the remaining lifecycle use cases.
func TestRenameAndDelete(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	subject := user.SU()

	session, err := uc.Create(subject, CreateOptions{Model: model.ID("fake-model")})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := uc.Rename(subject, session.ID, "renamed"); err != nil {
		t.Fatalf("rename: %v", err)
	}

	optSession, err := uc.FindByID(subject, session.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if optSession.IsNone() || optSession.Unwrap().Title != "renamed" {
		t.Fatal("rename did not persist")
	}

	if err := uc.Delete(subject, session.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	reloaded, err := repo.FindByID(session.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.IsSome() {
		t.Fatal("session was not deleted")
	}
}

// TestCreateGrantsOwnerAndDeleteRevokes verifies the ReBAC side effects: on create the creator is granted the
// owner relation plus every per-instance permission, and on delete those grants are removed again.
func TestCreateGrantsOwnerAndDeleteRevokes(t *testing.T) {
	uc, _, rdb := newTestUseCases(t)
	subject := user.SU() // sysUser.ID() == "" - we assert the triples for exactly that source instance

	session, err := uc.Create(subject, CreateOptions{Model: model.ID("fake-model")})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	src := rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(subject.ID())}
	target := rebac.Entity{Namespace: Namespace, Instance: rebac.Instance(session.ID)}

	// Owner relation present.
	if ok, err := rdb.Contains(rebac.Triple{Source: src, Relation: rebac.Owner, Target: target}); err != nil || !ok {
		t.Fatalf("expected owner grant, ok=%v err=%v", ok, err)
	}

	// Every instance permission present.
	for _, pid := range InstancePermissions {
		ok, err := rdb.Contains(rebac.Triple{Source: src, Relation: rebac.Relation(pid), Target: target})
		if err != nil || !ok {
			t.Fatalf("expected instance grant %q, ok=%v err=%v", pid, ok, err)
		}
	}

	// Delete revokes all grants targeting the instance.
	if err := uc.Delete(subject, session.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if ok, err := rdb.Contains(rebac.Triple{Source: src, Relation: rebac.Owner, Target: target}); err != nil || ok {
		t.Fatalf("owner grant should be revoked after delete, ok=%v err=%v", ok, err)
	}
	for _, pid := range InstancePermissions {
		ok, err := rdb.Contains(rebac.Triple{Source: src, Relation: rebac.Relation(pid), Target: target})
		if err != nil || ok {
			t.Fatalf("instance grant %q should be revoked after delete, ok=%v err=%v", pid, ok, err)
		}
	}

	// Assert the session is really gone from the read side too.
	if reloaded, err := uc.FindByID(subject, session.ID); err != nil {
		t.Fatalf("find after delete: %v", err)
	} else if reloaded.IsSome() {
		t.Fatal("session should be gone after delete")
	}
}
