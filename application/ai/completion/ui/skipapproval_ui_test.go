// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	stdjson "encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/proto"
)

// buttonText matches a button by its German or English caption.
func buttonText(de, en string) nagotest.Matcher {
	return nagotest.Where(de, func(n nagotest.Node) bool {
		tv, ok := n.Component.(*proto.TextView)
		return ok && (string(tv.Value) == de || string(tv.Value) == en)
	})
}

// calls returns a model turn with several tool calls at once.
func calls(cs ...completion.ToolCall) completion.Result {
	content := make([]completion.Content, 0, len(cs))
	for _, c := range cs {
		content = append(content, c)
	}
	return completion.Result{
		Message:    completion.Message{Role: completion.Assistant, Content: content},
		StopReason: completion.StopToolUse,
	}
}

// Several changes of one turn can be approved at once.
func TestChat_ApproveAll(t *testing.T) {
	var writes atomic.Int32
	write := completion.NewTool("write", "writes", func(struct{}) (struct{}, error) {
		writes.Add(1)
		return struct{}{}, nil
	}).AsMutating("writes")

	fake := &scripted{
		parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			if hasToolResult(opts) {
				return text("Alles erledigt."), nil
			}
			return calls(
				completion.ToolCall{ID: "w1", Name: "write", Arguments: stdjson.RawMessage(`{}`)},
				completion.ToolCall{ID: "w2", Name: "write", Arguments: stdjson.RawMessage(`{}`)},
				completion.ToolCall{ID: "w3", Name: "write", Arguments: stdjson.RawMessage(`{}`)},
			), nil
		},
	}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		ConfirmMutations:   true,
		Agents:             []uicompletion.Agent{{Tools: []completion.Tool{write}}},
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte schreiben")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(buttonText("Alle ausführen (3)", "Execute all (3)"), 10*time.Second)
	if writes.Load() != 0 {
		t.Fatalf("the changes must wait, got %d", writes.Load())
	}

	w.Click(w.Find(buttonText("Alle ausführen (3)", "Execute all (3)")))
	w.WaitFor(richText("Alles erledigt."), 10*time.Second)
	if writes.Load() != 3 {
		t.Fatalf("all approved changes must run, got %d", writes.Load())
	}
}

// A user who lets the conversation run without asking approves no further ordinary change, but still the marked
// ones.
func TestChat_SkipApproval(t *testing.T) {
	var writes, drops atomic.Int32
	write := completion.NewTool("write", "writes", func(struct{}) (struct{}, error) {
		writes.Add(1)
		return struct{}{}, nil
	}).AsMutating("writes")
	drop := completion.NewTool("drop", "drops", func(struct{}) (struct{}, error) {
		drops.Add(1)
		return struct{}{}, nil
	}).AsMutating("Alles löschen").RequireApproval()

	var mutex sync.Mutex
	step := 0
	fake := &scripted{
		parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			mutex.Lock()
			defer mutex.Unlock()
			step++
			switch step {
			case 1:
				return call("w1", "write", `{}`), nil
			case 2:
				return call("w2", "write", `{}`), nil
			case 3:
				return call("x1", "drop", `{}`), nil
			default:
				return text("Alles erledigt."), nil
			}
		},
	}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		ConfirmMutations:   true,
		Agents:             []uicompletion.Agent{{Tools: []completion.Tool{write, drop}}},
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte aufräumen")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(nagotest.ID("uicompletion-skip-approval"), 10*time.Second)

	w.Type(w.Find(nagotest.ID("uicompletion-skip-approval")), "true")
	w.Click(w.Find(buttonText("Ausführen", "Execute")))

	// the second write runs without asking, the marked drop waits
	w.WaitFor(nagotest.Text("Alles löschen"), 10*time.Second)
	if writes.Load() != 2 || drops.Load() != 0 {
		t.Fatalf("further writes must run and the marked change wait, writes=%d drops=%d", writes.Load(), drops.Load())
	}

	w.Click(w.Find(buttonText("Ausführen", "Execute")))
	w.WaitFor(richText("Alles erledigt."), 10*time.Second)
	if drops.Load() != 1 {
		t.Fatalf("the approved marked change must run once, got %d", drops.Load())
	}
}

// The switch below the input lets a new conversation run its changes without asking from the first message on.
// The conversation keeps the setting once it is created, and a new chat asks again.
func TestChat_SkipApprovalBeforeFirstMessage(t *testing.T) {
	var writes atomic.Int32
	write := completion.NewTool("write", "writes", func(struct{}) (struct{}, error) {
		writes.Add(1)
		return struct{}{}, nil
	}).AsMutating("writes")

	fake := &scripted{
		parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			if hasToolResult(opts) {
				return text("Alles erledigt."), nil
			}
			return call("w1", "write", `{}`), nil
		},
	}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Sessions:           testSessions(t),
		History:            true,
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		ConfirmMutations:   true,
		Agents:             []uicompletion.Agent{{Tools: []completion.Tool{write}}},
	})

	toggle := nagotest.ID("uicompletion-skip-approval-input")
	w.Type(w.Find(toggle), "true")
	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte schreiben")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(richText("Alles erledigt."), 10*time.Second)
	if writes.Load() != 1 {
		t.Fatalf("the change must run without asking, got %d", writes.Load())
	}

	// the created conversation took the setting over
	if !w.Find(toggle).Node().Component.(*proto.Toggle).Value {
		t.Fatal("the switch must stay on in the created conversation")
	}

	w.Click(w.Find(nagotest.Text("Neuer Chat")))
	if w.Find(toggle).Node().Component.(*proto.Toggle).Value {
		t.Fatal("a new chat must ask again")
	}
}
