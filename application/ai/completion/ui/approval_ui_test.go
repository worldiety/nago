// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/proto"
)

// A chat which confirms only marked tools runs ordinary changes without asking, asks before a marked tool and
// lets sub-agents use the unmarked mutating tools only. This holds for a transient chat and for one with a
// history, whose runs go through the session use cases.
func TestChat_ConfirmMarked(t *testing.T) {
	t.Run("transient", func(t *testing.T) { testConfirmMarked(t, false) })
	t.Run("history", func(t *testing.T) { testConfirmMarked(t, true) })
}

func testConfirmMarked(t *testing.T, history bool) {
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
	var subTools []string
	step := 0
	fake := &scripted{
		parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			mutex.Lock()
			defer mutex.Unlock()
			step++
			switch step {
			case 1:
				return call("d1", completion.DelegateToolName, `{"tasks":[{"title":"Schreiben","task":"write"}]}`), nil
			case 2:
				return call("w1", "write", `{}`), nil
			case 3:
				return call("x1", "drop", `{}`), nil
			default:
				return text("Alles erledigt."), nil
			}
		},
		sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			mutex.Lock()
			for _, d := range opts.Tools {
				subTools = append(subTools, d.Name)
			}
			mutex.Unlock()
			return text("fertig"), nil
		},
	}

	var sessions session.UseCases
	if history {
		sessions = testSessions(t)
	}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Sessions:           sessions,
		History:            history,
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		ConfirmMarked:      true,
		Delegation:         &uicompletion.DelegationOptions{AllowMutating: true},
		Agents:             []uicompletion.Agent{{Tools: []completion.Tool{write, drop}}},
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte aufräumen")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(nagotest.Text("Alles löschen"), 10*time.Second) // the effect of the marked tool in the dialog

	if writes.Load() != 1 || drops.Load() != 0 {
		t.Fatalf("the unmarked change must run and the marked one wait, writes=%d drops=%d", writes.Load(), drops.Load())
	}

	mutex.Lock()
	offered := slices.Clone(subTools)
	mutex.Unlock()
	if !slices.Contains(offered, "write") || slices.Contains(offered, "drop") {
		t.Fatalf("the sub-agent must get only the unmarked mutating tool, got %v", offered)
	}

	w.Click(w.Find(nagotest.Where("approve", func(n nagotest.Node) bool {
		tv, ok := n.Component.(*proto.TextView)
		return ok && (tv.Value == "Ausführen" || tv.Value == "Execute")
	})))
	w.WaitFor(richText("Alles erledigt."), 10*time.Second)
	if drops.Load() != 1 {
		t.Fatalf("the approved change must run once, got %d", drops.Load())
	}
}
