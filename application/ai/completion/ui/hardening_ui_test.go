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
	"errors"
	"iter"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// scripted is a provider whose parent and sub-agent answers are scripted independently.
type scripted struct {
	parent func(ctx context.Context, opts completion.Options) (completion.Result, error)
	sub    func(ctx context.Context, opts completion.Options) (completion.Result, error)
}

func (f *scripted) Models(auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) { yield(model.Model{ID: "fake"}, nil) }
}

func (f *scripted) Complete(ctx context.Context, _ auth.Subject, opts completion.Options) (completion.Result, error) {
	if strings.HasPrefix(opts.System, "You are a sub-agent.") {
		return f.sub(ctx, opts)
	}
	return f.parent(ctx, opts)
}

func (f *scripted) Stream(context.Context, auth.Subject, completion.Options) iter.Seq2[completion.Delta, error] {
	return nil
}

func call(id, name, args string) completion.Result {
	return completion.Result{
		Message:    completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.ToolCall{ID: id, Name: name, Arguments: stdjson.RawMessage(args)}}},
		StopReason: completion.StopToolUse,
	}
}

func hasToolResult(opts completion.Options) bool {
	last := opts.Messages[len(opts.Messages)-1]
	return slices.ContainsFunc(last.Content, func(c completion.Content) bool {
		_, ok := c.(completion.ToolResult)
		return ok
	})
}

func openChatOptions(t *testing.T, opts uicompletion.ChatOptions) *nagotest.Window {
	t.Helper()
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uicompletion.test")
		cfg.RootView("chat", func(wnd core.Window) core.View {
			return uicompletion.Chat(wnd, opts)
		})
	})

	return app.Open(t, user.SU(), "chat")
}

// A custom delegation tool may let its sub-agents write. A read-only or confirming chat must keep them
// read-only nonetheless, because a sub-agent cannot ask anyone.
func TestChat_CustomDelegationRespectsReadOnlyAndConfirmation(t *testing.T) {
	cases := []struct {
		name       string
		readOnly   bool
		confirm    bool
		wantWrites int32
	}{
		{"writing", false, false, 1},
		{"read-only", true, false, 0},
		{"confirming", false, true, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var writes atomic.Int32
			write := completion.NewTool("write", "writes", func(struct{}) (struct{}, error) {
				writes.Add(1)
				return struct{}{}, nil
			}).AsMutating("writes")
			delegate := completion.NewDelegateTool(completion.DelegateConfig{Tools: []completion.Tool{write}, AllowMutating: true})

			fake := &scripted{
				parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
					if hasToolResult(opts) {
						return text("Alles erledigt."), nil
					}
					return call("d1", completion.DelegateToolName, `{"tasks":[{"title":"Schreiben","task":"write"}]}`), nil
				},
				sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
					offered := slices.ContainsFunc(opts.Tools, func(d completion.ToolDef) bool { return d.Name == "write" })
					if hasToolResult(opts) || !offered {
						return text("fertig"), nil
					}
					return call("w1", "write", `{}`), nil
				},
			}

			w := openChatOptions(t, uicompletion.ChatOptions{
				Completions:        fake,
				Provider:           echo.New("p", "p"),
				DisableCurrentTime: true,
				ReadOnly:           c.readOnly,
				ConfirmMutations:   c.confirm,
				Agents:             []uicompletion.Agent{{Tools: []completion.Tool{delegate}}},
			})

			w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte schreiben")
			w.Click(w.Find(nagotest.Text("Senden")))
			w.WaitFor(richText("Alles erledigt."), 10*time.Second)

			if got := writes.Load(); got != c.wantWrites {
				t.Fatalf("expected %d writes, got %d", c.wantWrites, got)
			}
		})
	}
}

// Background tasks of a failed run are cancelled: nobody would receive their results, and the chat would
// neither show them nor offer to stop them.
func TestChat_FailedRunCancelsBackgroundTasks(t *testing.T) {
	subStarted := make(chan struct{})
	subCancelled := make(chan struct{})
	fake := &scripted{
		parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			if hasToolResult(opts) {
				// the task has started, now the provider fails
				<-subStarted
				return completion.Result{}, errors.New("provider unavailable")
			}
			return call("s1", completion.StartTasksToolName, `{"tasks":[{"title":"Lange","task":"long"}]}`), nil
		},
		sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			close(subStarted)
			<-ctx.Done()
			close(subCancelled)
			return completion.Result{}, ctx.Err()
		},
	}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		Delegation:         &uicompletion.DelegationOptions{BackgroundTasks: true},
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte im Hintergrund")
	w.Click(w.Find(nagotest.Text("Senden")))

	select {
	case <-subCancelled:
	case <-time.After(10 * time.Second):
		t.Fatal("the background task of the failed run is still running")
	}

	w.WaitFor(nagotest.Text("Senden"), 5*time.Second)
}

// taskIDOf returns the id of the first background task found in the tool results of the messages.
func taskIDOf(msgs []completion.Message) string {
	for _, m := range msgs {
		for _, c := range m.Content {
			if r, ok := c.(completion.ToolResult); ok {
				if tasks, ok := completion.ParseTaskResults(r); ok && len(tasks) > 0 && tasks[0].ID != "" {
					return tasks[0].ID
				}
			}
		}
	}
	return ""
}

// A failed run keeps the results of finished background tasks, so a retry still receives them.
func TestChat_RetryReceivesResultsOfFailedRun(t *testing.T) {
	subDone := make(chan struct{})
	var calls atomic.Int32
	fake := &scripted{
		parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			switch calls.Add(1) {
			case 1:
				return call("s1", completion.StartTasksToolName, `{"tasks":[{"title":"Kurz","task":"short"}]}`), nil
			case 2:
				// the task has finished, now the provider fails
				<-subDone
				return completion.Result{}, errors.New("provider unavailable")
			case 3:
				return call("a1", completion.AwaitTasksToolName, `{"ids":["`+taskIDOf(opts.Messages)+`"],"timeout_seconds":5}`), nil
			default:
				last := opts.Messages[len(opts.Messages)-1]
				for _, c := range last.Content {
					if r, ok := c.(completion.ToolResult); ok {
						if tasks, ok := completion.ParseTaskResults(r); ok && len(tasks) == 1 && tasks[0].Status == completion.TaskCompleted {
							return text("Ergebnis: " + tasks[0].Answer), nil
						}
					}
				}
				return text("Ergebnis: verloren"), nil
			}
		},
		sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			defer close(subDone)
			return text("fertig"), nil
		},
	}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		Delegation:         &uicompletion.DelegationOptions{BackgroundTasks: true},
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte im Hintergrund")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(nagotest.Text("Senden"), 10*time.Second)

	w.Type(w.Find(nagotest.Label("Nachricht")), "Nochmal bitte")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(richText("Ergebnis: fertig"), 10*time.Second)
}

// A refused answer ends the run without handing over background tasks, so the running ones are cancelled.
func TestChat_RefusalCancelsBackgroundTasks(t *testing.T) {
	subStarted := make(chan struct{})
	subCancelled := make(chan struct{})
	fake := &scripted{
		parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			if hasToolResult(opts) {
				<-subStarted
				return completion.Result{
					Message:    completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "Nein."}}},
					StopReason: completion.StopRefusal,
				}, nil
			}
			return call("s1", completion.StartTasksToolName, `{"tasks":[{"title":"Lange","task":"long"}]}`), nil
		},
		sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			close(subStarted)
			<-ctx.Done()
			close(subCancelled)
			return completion.Result{}, ctx.Err()
		},
	}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		Delegation:         &uicompletion.DelegationOptions{BackgroundTasks: true},
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte im Hintergrund")
	w.Click(w.Find(nagotest.Text("Senden")))

	select {
	case <-subCancelled:
	case <-time.After(10 * time.Second):
		t.Fatal("the background task of the refused run is still running")
	}
}

// Without History, a run is cancelled together with the chat, because nobody receives its answer any more.
func TestChat_TransientRunIsCancelledWithTheChat(t *testing.T) {
	requested := make(chan struct{})
	cancelled := make(chan struct{})
	fake := &scripted{parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
		close(requested)
		<-ctx.Done()
		close(cancelled)
		return completion.Result{}, ctx.Err()
	}}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Das dauert")
	w.Click(w.Find(nagotest.Text("Senden")))
	select {
	case <-requested:
	case <-time.After(5 * time.Second):
		t.Fatal("the provider has not been requested")
	}

	// the user closes the tab
	w.Close()

	select {
	case <-cancelled:
	case <-time.After(5 * time.Second):
		t.Fatal("the run of the closed chat is still working")
	}
}

// Tasks the model never awaited show the result BeforeFinish handed over, including their transcript.
func TestChat_ForgottenTasksShowTheirResult(t *testing.T) {
	var calls atomic.Int32
	fake := &scripted{
		parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			switch calls.Add(1) {
			case 1:
				return call("s1", completion.StartTasksToolName, `{"tasks":[{"title":"Rechnen","task":"6*7"}]}`), nil
			case 2:
				return text("Ich antworte, ohne zu warten."), nil
			default:
				return text("Nachtrag: 42."), nil
			}
		},
		sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			return text("42"), nil
		},
	}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Sessions:           testSessions(t),
		History:            true,
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		Delegation:         &uicompletion.DelegationOptions{BackgroundTasks: true},
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Rechne im Hintergrund")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(richText("Nachtrag: 42."), 10*time.Second)

	// the start block shows the final state and offers the transcript
	w.Find(nagotest.TextContains("erledigt"))
	w.FindAll(nagotest.Label("Verlauf der Teilaufgabe anzeigen")).Exactly(1)
}

// Hiding a chat without History destroys its states, which cancels the run and must not block the window.
func TestChat_TransientRunIsCancelledWhenChatIsHidden(t *testing.T) {
	requested := make(chan struct{})
	cancelled := make(chan struct{})
	fake := &scripted{parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
		close(requested)
		<-ctx.Done()
		close(cancelled)
		return completion.Result{}, ctx.Err()
	}}

	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uicompletion.test")
		cfg.RootView("page", func(wnd core.Window) core.View {
			show := core.StateOf[bool](wnd, "show").Init(func() bool { return true })
			return ui.VStack(
				ui.PrimaryButton(func() { show.Set(false) }).Title("hide"),
				ui.Text("page"),
				ui.IfFunc(show.Get(), func() core.View {
					return uicompletion.Chat(wnd, uicompletion.ChatOptions{
						Completions:        fake,
						Provider:           echo.New("p", "p"),
						DisableCurrentTime: true,
					})
				}),
			)
		})
	})

	w := app.Open(t, user.SU(), "page")
	w.Type(w.Find(nagotest.Label("Nachricht")), "Das dauert")
	w.Click(w.Find(nagotest.Text("Senden")))
	select {
	case <-requested:
	case <-time.After(5 * time.Second):
		t.Fatal("the provider has not been requested")
	}

	// settles only if destroying the states does not block the window
	w.Click(w.Find(nagotest.Text("hide")))
	w.FindAll(nagotest.Text("Senden")).None()

	select {
	case <-cancelled:
	case <-time.After(5 * time.Second):
		t.Fatal("the run of the hidden chat is still working")
	}

	// the window still works
	w.Find(nagotest.Text("page"))
}

// A panic in a run of a persisted conversation is reported, and the chat is ready again.
func TestChat_PanickingRunIsReported(t *testing.T) {
	fake := &scripted{parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
		panic("boom")
	}}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Sessions:           testSessions(t),
		History:            true,
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte stürze ab")
	w.Click(w.Find(nagotest.Text("Senden")))

	// the process survives, the run is over and the chat is ready again
	w.WaitFor(nagotest.Text("Senden"), 10*time.Second)
	w.FindAll(nagotest.Text("Stopp")).None()
}
