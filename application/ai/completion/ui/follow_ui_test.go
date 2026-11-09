// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// followWorld is an application whose windows continue one kept conversation, as the floating assistant does from
// page to page.
type followWorld struct {
	app      *nagotest.App
	sessions session.UseCases

	mu   sync.Mutex
	kept session.ID
}

func newFollowWorld(t *testing.T, comps completion.Completions, sessions session.UseCases, maxTurns int) *followWorld {
	t.Helper()
	f := &followWorld{sessions: sessions}
	f.app = nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uicompletion.followtest")
		cfg.RootView("chat", func(wnd core.Window) core.View {
			state := core.AutoState[session.ID](wnd).Init(func() session.ID {
				f.mu.Lock()
				defer f.mu.Unlock()
				return f.kept
			})
			return uicompletion.Chat(wnd, uicompletion.ChatOptions{
				Sessions: sessions, History: true, Completions: comps, Provider: echo.New("p", "p"),
				DisableCurrentTime: true, MaxTurns: maxTurns, Agents: noopAgent(), AskUser: true, Session: state,
				Delegation: &uicompletion.DelegationOptions{},
			})
		})
	})
	return f
}

func (f *followWorld) open(t *testing.T) *nagotest.Window {
	t.Helper()
	return f.app.Open(t, user.SU(), "chat")
}

// keep makes every window opened or reloaded from now on continue the only conversation there is.
func (f *followWorld) keep(t *testing.T) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	for s, err := range f.sessions.FindAll(user.SU(), session.FindAllOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		f.kept = s.ID
	}
	if f.kept == "" {
		t.Fatal("there is no conversation")
	}
}

// inBackground runs an action of a window which only settles with the end of the run it follows, see [core.OnAppear].
func inBackground(action func()) <-chan struct{} {
	settled := make(chan struct{})
	go func() {
		defer close(settled)
		action()
	}()
	return settled
}

// shown waits for what the tree of a window shows, without settling, because a following window does not settle.
func shown(t *testing.T, w *nagotest.Window, what string, m nagotest.Matcher) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for w.FindAll(m).Len() == 0 {
		if time.Now().After(deadline) {
			t.Fatalf("%s is not shown", what)
		}
		time.Sleep(time.Millisecond)
	}
}

func lockedField() nagotest.Matcher {
	return nagotest.Where("locked message field", func(n nagotest.Node) bool {
		f, ok := n.Component.(*proto.TextField)
		return ok && f.Label == "Nachricht" && bool(f.Disabled)
	})
}

// hold blocks a request to the model until the test lets it go on, so that the test decides what a run has done
// when a window looks at it.
type hold struct {
	arrived chan struct{}
	release chan struct{}
}

func newHold() hold { return hold{arrived: make(chan struct{}), release: make(chan struct{})} }

func (h hold) wait(ctx context.Context) error {
	close(h.arrived)
	select {
	case <-h.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// heldThen answers a held request with what end returns.
func heldThen(h hold, end func() (completion.Result, error)) func(ctx context.Context) (completion.Result, error) {
	return func(ctx context.Context) (completion.Result, error) {
		if err := h.wait(ctx); err != nil {
			return completion.Result{}, err
		}
		return end()
	}
}

// script answers the requests in order: the first one with the given answer, the rest with the steps.
func script(steps ...func(ctx context.Context) (completion.Result, error)) *recording {
	return &recording{next: func(ctx context.Context, n int) (completion.Result, error) {
		return steps[n-1](ctx)
	}}
}

func answer(s string) func(context.Context) (completion.Result, error) {
	return func(context.Context) (completion.Result, error) { return text(s), nil }
}

func named(id, label string) func(context.Context) (completion.Result, error) {
	return func(context.Context) (completion.Result, error) { return stepWithText(id, label), nil }
}

func awaitHold(t *testing.T, h hold) {
	t.Helper()
	select {
	case <-h.arrived:
	case <-time.After(10 * time.Second):
		t.Fatal("the run did not reach the held request")
	}
}

// A window which shows a running conversation shows the message and the steps at once and follows every further step.
func TestChat_RunningRun_FollowsStepByStep(t *testing.T) {
	first, second := newHold(), newHold()
	fake := script(
		answer("Erste Antwort"),
		named("c1", "Schritt eins"),
		heldThen(first, func() (completion.Result, error) { return stepWithText("c2", "Schritt zwei"), nil }),
		heldThen(second, func() (completion.Result, error) { return text("Alles fertig."), nil }),
	)

	world := newFollowWorld(t, fake, testSessions(t), 8)
	starter := world.open(t)
	ask(starter, "Erste Frage")
	starter.WaitFor(richText("Erste Antwort"), 10*time.Second)
	world.keep(t)
	page := world.open(t)

	ask(starter, "Bitte arbeiten")
	awaitHold(t, first)

	settled := inBackground(page.Reload)
	shown(t, page, "the message", richText("Bitte arbeiten"))
	shown(t, page, "the step", richText("Schritt eins"))
	shown(t, page, "the stop button", nagotest.Text("Stopp"))
	shown(t, page, "the locked field", lockedField())

	close(first.release)
	awaitHold(t, second)
	shown(t, page, "the next step", richText("Schritt zwei"))

	close(second.release)
	<-settled
	page.WaitFor(richText("Alles fertig."), 10*time.Second)
	for _, want := range []string{"Erste Frage", "Bitte arbeiten", "Schritt eins", "Schritt zwei", "Alles fertig."} {
		page.Find(richText(want)).Exactly(1)
		starter.WaitFor(richText(want), 10*time.Second).Exactly(1)
	}
	page.FindAll(nagotest.Text("Stopp")).None()
}

// The message which started a run is part of what a window shows, although nothing has been persisted yet.
func TestChat_RunningRun_ShowsTheMessageBeforeTheFirstStep(t *testing.T) {
	h := newHold()
	fake := script(heldThen(h, func() (completion.Result, error) { return text("Alles fertig."), nil }))

	world := newFollowWorld(t, fake, testSessions(t), 8)
	starter := world.open(t)
	page := world.open(t)
	ask(starter, "Bitte arbeiten")
	awaitHold(t, h)
	world.keep(t)

	settled := inBackground(page.Reload)
	shown(t, page, "the message", richText("Bitte arbeiten"))
	shown(t, page, "the stop button", nagotest.Text("Stopp"))

	close(h.release)
	<-settled
	page.WaitFor(richText("Alles fertig."), 10*time.Second)
	page.Find(richText("Bitte arbeiten")).Exactly(1)
}

// However a run ends, a window which follows it shows what the run has persisted.
func TestChat_RunningRun_Ends(t *testing.T) {
	boom := errors.New("provider down")
	type row struct {
		name string
		// withStep lets the run have a step before it is held
		withStep bool
		end      func() (completion.Result, error)
		// stop ends the held run with the button of the following window instead
		stop     bool
		want     []string
		wantGone []string
		wantText []string
	}

	tests := []row{
		{
			name:     "the model answers",
			withStep: true,
			end:      func() (completion.Result, error) { return text("Alles fertig."), nil },
			want:     []string{"Erste Frage", "Bitte arbeiten", "Schritt eins", "Alles fertig."},
		},
		{
			name:     "the run ends at its turn limit",
			withStep: true,
			end:      func() (completion.Result, error) { return step("c2"), nil },
			want:     []string{"Bitte arbeiten", "Schritt eins"},
			wantText: []string{"Fortsetzen"},
		},
		{
			name:     "the run waits on a question",
			withStep: true,
			end: func() (completion.Result, error) {
				return call("q1", completion.AskUserToolName, `{"question":"Welche?"}`), nil
			},
			want:     []string{"Bitte arbeiten", "Schritt eins", "Welche?"},
			wantText: []string{"Antworten"},
		},
		{
			name:     "the run fails after a step",
			withStep: true,
			end:      func() (completion.Result, error) { return completion.Result{}, boom },
			want:     []string{"Erste Frage", "Bitte arbeiten", "Schritt eins"},
		},
		{
			name:     "the run fails without a step",
			end:      func() (completion.Result, error) { return completion.Result{}, boom },
			want:     []string{"Erste Frage", "Erste Antwort"},
			wantGone: []string{"Bitte arbeiten"},
		},
		{
			name:     "the run is stopped after a step",
			withStep: true,
			stop:     true,
			want:     []string{"Erste Frage", "Bitte arbeiten", "Schritt eins"},
		},
		{
			name:     "the run is stopped without a step",
			stop:     true,
			want:     []string{"Erste Frage", "Erste Antwort"},
			wantGone: []string{"Bitte arbeiten"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHold()
			end := tc.end
			if tc.stop {
				end = func() (completion.Result, error) { return completion.Result{}, context.Canceled }
			}
			steps := []func(context.Context) (completion.Result, error){answer("Erste Antwort")}
			if tc.withStep {
				steps = append(steps, named("c1", "Schritt eins"))
			}
			steps = append(steps, heldThen(h, end))
			fake := script(steps...)

			world := newFollowWorld(t, fake, testSessions(t), 2)
			starter := world.open(t)
			ask(starter, "Erste Frage")
			starter.WaitFor(richText("Erste Antwort"), 10*time.Second)
			world.keep(t)
			page := world.open(t)

			ask(starter, "Bitte arbeiten")
			awaitHold(t, h)

			settled := inBackground(page.Reload)
			shown(t, page, "the stop button", nagotest.Text("Stopp"))

			if tc.stop {
				page.Click(page.Find(nagotest.Text("Stopp")))
			} else {
				close(h.release)
			}
			<-settled

			page.WaitFor(richText(tc.want[0]), 10*time.Second)
			page.FindAll(nagotest.Text("Stopp")).None()
			for _, want := range tc.want {
				page.Find(richText(want)).Exactly(1)
			}
			for _, gone := range tc.wantGone {
				page.FindAll(richText(gone)).None()
			}
			for _, want := range tc.wantText {
				page.Find(nagotest.Text(want)).Exactly(1)
			}
		})
	}
}

// Every result counts as a step of the run, also of a call without a view: a sub-agent which is finished does not look
// as if it were working, in the window which started the run as in one which follows it.
func TestChat_RunningRun_FinishedSubAgent(t *testing.T) {
	h := newHold()
	fake := script(
		func(context.Context) (completion.Result, error) {
			return call("d1", completion.DelegateToolName, `{"tasks":[{"title":"Erste","task":"A"}]}`), nil
		},
		heldThen(h, func() (completion.Result, error) { return text("Alles fertig."), nil }),
	)

	world := newFollowWorld(t, fake, testSessions(t), 8)
	starter := world.open(t)
	page := world.open(t)
	ask(starter, "Bitte parallel arbeiten")
	awaitHold(t, h)
	world.keep(t)

	settled := inBackground(page.Reload)
	for _, w := range []*nagotest.Window{starter, page} {
		shown(t, w, "the finished sub-agent", nagotest.TextContains("• erledigt"))
		shown(t, w, "its answer", richText("Antwort der Teilaufgabe"))
		w.FindAll(nagotest.TextContains("• läuft")).None()
	}

	close(h.release)
	<-settled
	page.WaitFor(richText("Alles fertig."), 10*time.Second)
}

// Several windows show the same steps, none twice and none missing.
func TestChat_RunningRun_SeveralWindows(t *testing.T) {
	first, second := newHold(), newHold()
	fake := script(
		named("c1", "Schritt eins"),
		heldThen(first, func() (completion.Result, error) { return stepWithText("c2", "Schritt zwei"), nil }),
		heldThen(second, func() (completion.Result, error) { return text("Alles fertig."), nil }),
	)

	world := newFollowWorld(t, fake, testSessions(t), 8)
	starter := world.open(t)
	one := world.open(t)
	two := world.open(t)
	ask(starter, "Bitte arbeiten")
	awaitHold(t, first)
	world.keep(t)

	settledOne, settledTwo := inBackground(one.Reload), inBackground(two.Reload)
	windows := []*nagotest.Window{starter, one, two}
	for _, w := range windows {
		shown(t, w, "the step", richText("Schritt eins"))
		w.Find(richText("Bitte arbeiten")).Exactly(1)
	}

	close(first.release)
	awaitHold(t, second)
	for _, w := range windows {
		shown(t, w, "the next step", richText("Schritt zwei"))
		w.Find(richText("Schritt eins")).Exactly(1)
		w.Find(richText("Schritt zwei")).Exactly(1)
	}

	close(second.release)
	<-settledOne
	<-settledTwo
	for _, w := range windows {
		w.WaitFor(richText("Alles fertig."), 10*time.Second)
		w.Find(richText("Schritt eins")).Exactly(1)
		w.Find(richText("Schritt zwei")).Exactly(1)
	}
}

// A window shows the whole conversation of a running run, even after the run shortened it for the model.
func TestChat_RunningRun_ShowsTheWholeConversationAfterACompaction(t *testing.T) {
	question := "ORIGINAL " + strings.Repeat("Ein langer Auftrag. ", 30)
	h := newHold()
	fake := script(
		named("c1", "Schritt 1"), named("c2", "Schritt 2"), named("c3", "Schritt 3"), named("c4", "Schritt 4"),
		func(context.Context) (completion.Result, error) { return overflowed() },
		heldThen(h, func() (completion.Result, error) { return text("Alles fertig."), nil }),
	)

	world := newFollowWorld(t, fake, testSessions(t), 16)
	starter := world.open(t)
	page := world.open(t)
	ask(starter, question)
	awaitHold(t, h)
	world.keep(t)

	settled := inBackground(page.Reload)
	shown(t, page, "the question", richText("ORIGINAL"))
	shown(t, page, "the first step", richText("Schritt 1"))
	page.FindAll(richText("SUMMARY")).None()

	close(h.release)
	<-settled
	page.WaitFor(richText("Alles fertig."), 10*time.Second)
	page.FindAll(richText("SUMMARY")).None()
	page.Find(richText("ORIGINAL")).Exactly(1)
}

// readerSubject may read a conversation but not continue it.
type readerSubject struct{ auth.Subject }

func (s readerSubject) AuditResource(ns rebac.Namespace, id rebac.Instance, p permission.ID) error {
	if p == session.PermAppend {
		return errors.New("denied")
	}
	return s.Subject.AuditResource(ns, id, p)
}

// Who may only read a conversation sees what is persisted and cannot stop the run, who may continue it follows it.
func TestChat_RunningRun_OnlyWithTheRightToContinue(t *testing.T) {
	tests := []struct {
		name       string
		subject    auth.Subject
		wantFollow bool
	}{
		{name: "may continue the conversation", subject: user.SU(), wantFollow: true},
		{name: "may only read the conversation", subject: readerSubject{user.SU()}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHold()
			fake := script(
				answer("Erste Antwort"),
				named("c1", "Schritt eins"),
				heldThen(h, func() (completion.Result, error) { return text("Alles fertig."), nil }),
			)

			world := newFollowWorld(t, fake, testSessions(t), 8)
			starter := world.open(t)
			ask(starter, "Erste Frage")
			starter.WaitFor(richText("Erste Antwort"), 10*time.Second)
			world.keep(t)
			page := world.app.Open(t, tc.subject, "chat")
			ask(starter, "Bitte arbeiten")
			awaitHold(t, h)

			settled := inBackground(page.Reload)
			if !tc.wantFollow {
				<-settled
				page.Find(richText("Erste Frage")).Exactly(1)
				page.FindAll(richText("Bitte arbeiten")).None()
				page.FindAll(richText("Schritt eins")).None()
				page.FindAll(nagotest.Text("Stopp")).None()
				close(h.release)
				return
			}

			shown(t, page, "the message", richText("Bitte arbeiten"))
			shown(t, page, "the step", richText("Schritt eins"))
			shown(t, page, "the stop button", nagotest.Text("Stopp"))
			close(h.release)
			<-settled
			page.WaitFor(richText("Alles fertig."), 10*time.Second)
		})
	}
}

// A window which reads the conversation just before the run ends shows the result, not the state before. The run ends
// right after the window has read the conversation, the moment which would leave it on the older state, if it looked
// for the run afterwards.
func TestChat_RunningRun_WindowOpensWhileItEnds(t *testing.T) {
	h := newHold()
	fake := script(
		answer("Erste Antwort"),
		heldThen(h, func() (completion.Result, error) { return text("Alles fertig."), nil }),
	)

	sessions := testSessions(t)
	var afterRead atomic.Pointer[func()]
	read := sessions.FindByID
	sessions.FindByID = func(subject auth.Subject, id session.ID) (option.Opt[session.Session], error) {
		res, err := read(subject, id)
		if hook := afterRead.Swap(nil); hook != nil {
			(*hook)()
		}
		return res, err
	}

	world := newFollowWorld(t, fake, sessions, 8)
	starter := world.open(t)
	ask(starter, "Erste Frage")
	starter.WaitFor(richText("Erste Antwort"), 10*time.Second)
	world.keep(t)
	page := world.open(t)

	ask(starter, "Bitte arbeiten")
	awaitHold(t, h)

	endRun := func() {
		close(h.release)
		shown(t, starter, "the end of the run", richText("Alles fertig."))
	}
	afterRead.Store(&endRun)

	page.Reload()
	page.WaitFor(richText("Alles fertig."), 10*time.Second)
	page.Find(richText("Bitte arbeiten")).Exactly(1)
	page.FindAll(nagotest.Text("Stopp")).None()
}

// A conversation opened from the history shows its running run as well; a list which is older than the end of the run
// does not show an older state.
func TestChat_RunningRun_OpenedFromTheHistory(t *testing.T) {
	tests := []struct {
		name string
		// endBeforePick ends the run after the list has been read and before the conversation is picked
		endBeforePick bool
	}{
		{name: "while the run works"},
		{name: "the run ended after the list was read", endBeforePick: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHold()
			fake := script(
				answer("Erste Antwort"),
				named("c1", "Schritt eins"),
				heldThen(h, func() (completion.Result, error) { return text("Alles fertig."), nil }),
			)

			world := newFollowWorld(t, fake, testSessions(t), 8)
			starter := world.open(t)
			ask(starter, "Erste Frage")
			starter.WaitFor(richText("Erste Antwort"), 10*time.Second)
			page := world.open(t)

			ask(starter, "Bitte arbeiten")
			awaitHold(t, h)

			page.Click(page.Find(nagotest.Text("Verlauf")))
			page.WaitFor(nagotest.TextContains("Nachrichten"), 5*time.Second)
			card := page.Find(nagotest.Text("Erste Frage"))

			if tc.endBeforePick {
				close(h.release)
				shown(t, starter, "the end of the run", richText("Alles fertig."))
				page.Click(card)
				page.WaitFor(richText("Alles fertig."), 10*time.Second)
				page.Find(richText("Schritt eins")).Exactly(1)
				return
			}

			settled := inBackground(func() { page.Click(card) })
			shown(t, page, "the message", richText("Bitte arbeiten"))
			shown(t, page, "the step", richText("Schritt eins"))
			shown(t, page, "the stop button", nagotest.Text("Stopp"))
			close(h.release)
			<-settled
			page.WaitFor(richText("Alles fertig."), 10*time.Second)
			page.Find(richText("Bitte arbeiten")).Exactly(1)
		})
	}
}
