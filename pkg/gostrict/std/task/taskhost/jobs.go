// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package taskhost keeps the open jobs of the windows of a host in Go, see package std/task.
package taskhost

import (
	"crypto/rand"
	"sync"

	"go.wdy.de/nago/pkg/gostrict/std/task"
	"go.wdy.de/nago/presentation/core"
)

// UnknownFailure is the failure of a wait for a job, which the window does not know, e.g. because it has ended
// before the wait, in another window or a former call. Start a job and wait for it in the same call.
const UnknownFailure = "the job is unknown or has already ended"

// Jobs are the open jobs of a window. A job lives from its start until its end was delivered on the event loop of the
// window or until the window was closed, so the jobs of a window are limited by the time their work may take.
//
// The end of a job always passes the event loop, so a wait in the same call as the start never misses it, even if the
// work was faster, and the function of a wait never runs on another goroutine than the window or during the call of
// the wait.
type Jobs struct {
	post func(func()) bool

	mu     sync.Mutex
	open   map[task.ID]func(failure string)
	closed bool
}

// New returns the jobs of an owner, whose event loop post schedules a function on or which returns false.
func New(post func(func()) bool) *Jobs {
	return &Jobs{post: post, open: map[task.ID]func(string){}}
}

var windows sync.Map // of core.Window to *Jobs

// Of returns the jobs of the window and creates them at the first use, which must happen on the event loop of the
// window, because its observers are not synchronized. The jobs end with the window, not with a render: closing the
// window drops all waits, while the work itself goes on.
func Of(wnd core.Window) *Jobs {
	if j, ok := windows.Load(wnd); ok {
		return j.(*Jobs)
	}
	j := New(wnd.Post)
	windows.Store(wnd, j)
	wnd.AddDestroyObserver(func() {
		windows.Delete(wnd)
		j.Close()
	}, core.DestroyOnClose)
	return j
}

// Start registers a new job, unless the owner is closed or already has limit open jobs, if limit is positive.
func (j *Jobs) Start(limit int) (task.ID, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed || (limit > 0 && len(j.open) >= limit) {
		return "", false
	}
	id := task.ID(rand.Text())
	j.open[id] = nil
	return id, true
}

// Len returns the number of the open jobs.
func (j *Jobs) Len() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return len(j.open)
}

// Finish ends the job with the failure, or "" if it succeeded, and calls the function of its wait on the event loop.
func (j *Jobs) Finish(id task.ID, failure string) {
	posted := j.post(func() {
		if fn := j.take(id); fn != nil {
			fn(failure)
		}
	})
	if !posted {
		j.take(id)
	}
}

func (j *Jobs) take(id task.ID) func(string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	fn := j.open[id]
	delete(j.open, id)
	return fn
}

// Wait calls fn on the event loop, when the job has ended, and replaces an earlier function of the job, so that a
// job holds at most one. A job, which is not open, calls fn with [UnknownFailure], so that the program does not wait
// forever, but never during this call.
func (j *Jobs) Wait(id task.ID, fn func(failure string)) {
	j.mu.Lock()
	_, ok := j.open[id]
	if ok {
		j.open[id] = fn
	}
	closed := j.closed
	j.mu.Unlock()
	if !ok && !closed {
		j.post(func() { fn(UnknownFailure) })
	}
}

// Close drops all jobs and their functions.
func (j *Jobs) Close() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.closed = true
	clear(j.open)
}
