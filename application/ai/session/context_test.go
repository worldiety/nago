// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"context"
	"encoding/json"
	"iter"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

// toolCallingCompletions calls the tool "probe" once and then answers.
type toolCallingCompletions struct{}

func (toolCallingCompletions) Models(auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) { yield(model.Model{ID: "fake-model"}, nil) }
}

func (toolCallingCompletions) Stream(context.Context, auth.Subject, completion.Options) iter.Seq2[completion.Delta, error] {
	return nil
}

func (toolCallingCompletions) Complete(_ context.Context, _ auth.Subject, opts completion.Options) (completion.Result, error) {
	last := opts.Messages[len(opts.Messages)-1]
	if _, ok := last.Content[0].(completion.ToolResult); ok {
		return completion.Result{Message: completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "done"}}}, StopReason: completion.StopEndTurn}, nil
	}

	return completion.Result{Message: completion.Message{Role: completion.Assistant, Content: []completion.Content{
		completion.ToolCall{ID: "p1", Name: "probe", Arguments: json.RawMessage(`{}`)},
	}}, StopReason: completion.StopToolUse}, nil
}

// A tool finds the conversation it runs in, also as a sub-agent in a child session, so that it can find what the
// application keeps per conversation, e.g. the attachments it took over.
func TestToolsKnowTheirSession(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	var root, running ID
	var known bool
	probe := completion.Tool{
		Def: completion.ToolDef{Name: "probe", Description: "probe"},
		Invoke: func(subject auth.Subject, _ json.RawMessage) (json.RawMessage, error) {
			root, known = IDOf(subject.Context())
			running, _ = RunningIDOf(subject.Context())
			return json.RawMessage(`"ok"`), nil
		},
	}

	parent, err := uc.Create(subject, CreateOptions{Model: model.ID("fake-model")})
	if err != nil {
		t.Fatal(err)
	}

	child, err := uc.Create(subject, CreateOptions{Model: model.ID("fake-model"), ParentID: parent.ID})
	if err != nil {
		t.Fatal(err)
	}

	run := func(id ID) {
		t.Helper()
		if _, err := uc.Append(subject, id, AppendOptions{
			Completions: toolCallingCompletions{},
			Input:       []completion.Content{completion.Text{Text: "go"}},
			Tools:       []completion.Tool{probe},
		}); err != nil {
			t.Fatal(err)
		}
	}

	run(parent.ID)
	if !known || root != parent.ID || running != parent.ID {
		t.Fatalf("expected the session in the tool, got %q %q %v", root, running, known)
	}

	run(child.ID)
	if root != parent.ID || running != child.ID {
		t.Fatalf("expected the root and the running child session, got %q %q", root, running)
	}

	if _, ok := IDOf(context.Background()); ok {
		t.Fatal("a context without session must report none")
	}
}
