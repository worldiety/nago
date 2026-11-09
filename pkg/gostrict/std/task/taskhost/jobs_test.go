// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package taskhost_test

import (
	"sync/atomic"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/gostrict/std/task"
	"go.wdy.de/nago/pkg/gostrict/std/task/taskhost"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// queue is an event loop, which runs its functions only when a test tells it.
type queue struct {
	fns    []func()
	closed bool
}

func (q *queue) post(fn func()) bool {
	if q.closed {
		return false
	}
	q.fns = append(q.fns, fn)
	return true
}

func (q *queue) run() {
	fns := q.fns
	q.fns = nil
	for _, f := range fns {
		f()
	}
}

func TestEndBeforeWaitInTheSameTurn(t *testing.T) {
	q := &queue{}
	j := taskhost.New(q.post)
	id, _ := j.Start(0)

	// the work was faster than the rest of the call, which started it
	j.Finish(id, "boom")
	var got []string
	j.Wait(id, func(failure string) { got = append(got, failure) })
	if len(got) != 0 {
		t.Fatal("a function of a wait must not run during the wait")
	}

	q.run()
	if len(got) != 1 || got[0] != "boom" {
		t.Fatalf("unexpected ends: %v", got)
	}
	if j.Len() != 0 {
		t.Fatalf("the job must be gone: %d", j.Len())
	}
}

func TestWaitOfUnknownJob(t *testing.T) {
	q := &queue{}
	j := taskhost.New(q.post)
	var got []string
	j.Wait(task.ID("nope"), func(failure string) { got = append(got, failure) })
	if len(got) != 0 {
		t.Fatal("a function of a wait must not run during the wait")
	}
	q.run()
	if len(got) != 1 || got[0] != taskhost.UnknownFailure {
		t.Fatalf("unexpected ends: %v", got)
	}
}

func TestLastWaitWins(t *testing.T) {
	q := &queue{}
	j := taskhost.New(q.post)
	id, _ := j.Start(0)
	var got []string
	j.Wait(id, func(string) { got = append(got, "first") })
	j.Wait(id, func(string) { got = append(got, "second") })
	j.Finish(id, "")
	q.run()
	if len(got) != 1 || got[0] != "second" {
		t.Fatalf("unexpected ends: %v", got)
	}
}

func TestLimitAndClose(t *testing.T) {
	q := &queue{}
	j := taskhost.New(q.post)
	a, ok := j.Start(1)
	if !ok {
		t.Fatal("the first job must start")
	}
	if _, ok := j.Start(1); ok {
		t.Fatal("the second job exceeds the limit")
	}

	called := false
	j.Wait(a, func(string) { called = true })
	j.Close()
	j.Finish(a, "")
	q.run()
	if called || j.Len() != 0 {
		t.Fatal("a closed owner must drop its functions")
	}
	if _, ok := j.Start(0); ok {
		t.Fatal("a closed owner must not start jobs")
	}
	j.Wait(a, func(string) { called = true })
	q.run()
	if called {
		t.Fatal("a closed owner must not call a wait")
	}
}

func TestFailedPostForgetsTheJob(t *testing.T) {
	q := &queue{}
	j := taskhost.New(q.post)
	id, _ := j.Start(0)
	q.closed = true
	j.Finish(id, "")
	if j.Len() != 0 {
		t.Fatal("the job must be gone, if the event loop has ended")
	}
}

// TestJobsOfARealWindow checks the coupling with a window: a render keeps the wait, while leaving the window drops it,
// even though Post of the former window still succeeds.
func TestJobsOfARealWindow(t *testing.T) {
	var jobs atomic.Pointer[taskhost.Jobs]
	var id atomic.Pointer[task.ID]
	var ends atomic.Int32

	configure := func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.taskhost")
		cfg.RootView("start", func(wnd core.Window) core.View {
			counter := core.AutoState[int](wnd)
			return ui.VStack(
				ui.Text("count"),
				ui.PrimaryButton(func() {
					j := taskhost.Of(wnd)
					jobs.Store(j)
					v, _ := j.Start(0)
					id.Store(&v)
					j.Wait(v, func(string) { ends.Add(1) })
				}).Title("start"),
				ui.PrimaryButton(func() { counter.Set(counter.Get() + 1) }).Title("render"),
				ui.PrimaryButton(func() { wnd.Navigation().ForwardTo("other", nil) }).Title("leave"),
			)
		})
		cfg.RootView("other", func(wnd core.Window) core.View { return ui.Text("other") })
	}

	w := nagotest.New(t, configure).Open(t, nil, "start")

	w.Click(w.Find(nagotest.Text("start")))
	w.Click(w.Find(nagotest.Text("render")))
	jobs.Load().Finish(*id.Load(), "")
	w.Settle()
	if ends.Load() != 1 {
		t.Fatalf("a render must keep the wait, got %d ends", ends.Load())
	}

	w.Click(w.Find(nagotest.Text("start")))
	before := jobs.Load()
	open := *id.Load()
	w.Click(w.Find(nagotest.Text("leave")))
	w.Find(nagotest.Text("other"))
	if before.Len() != 0 {
		t.Fatalf("leaving the window must drop its jobs, got %d", before.Len())
	}
	before.Finish(open, "")
	w.Settle()
	if ends.Load() != 1 {
		t.Fatalf("a closed window must not get an end, got %d ends", ends.Load())
	}
}
