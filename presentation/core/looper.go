// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"go.wdy.de/nago/pkg/std/concurrent"
)

// An EventLoop handles messages respective executes functions.
// Everything which looks like a concurrent or parallel situation within the backend-frontend relation
// must be single-threaded through the EventLoop to avoid race conditions. Especially these are
//   - sending or receiving events to the frontend
//   - async domain events
//   - upload
//   - download
type EventLoop struct {
	// the queue is unbounded, because a Post must never block. A bounded channel would block any poster, as soon
	// as a single function hangs, and it would deadlock the loop itself when it posts into its own full queue.
	mutex     sync.Mutex
	queue     []func()
	destroyed bool

	// wake signals the loop that the queue may contain new functions or that the loop has been destroyed.
	wake chan struct{}
	// done is closed, when the loop has exited.
	done    chan struct{}
	onPanic concurrent.Value[func(p any)]
	// pending counts posted but not yet completed functions.
	pending atomic.Int64
}

func NewEventLoop() *EventLoop {
	l := &EventLoop{
		wake: make(chan struct{}, 1),
		done: make(chan struct{}),
	}

	go l.loop()

	return l
}

func (l *EventLoop) loop() {
	defer close(l.done)

	var batch []func()
	for {
		l.mutex.Lock()
		batch, l.queue = l.queue, batch[:0]
		destroyed := l.destroyed
		l.mutex.Unlock()

		if len(batch) == 0 {
			if destroyed {
				return
			}

			<-l.wake
			continue
		}

		for i, fn := range batch {
			l.saveExec(fn)
			batch[i] = nil // release the closure for the GC, the slice is reused
			l.pending.Add(-1)
		}
	}
}

func (l *EventLoop) saveExec(f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
			slog.Error("recovered from panic in EventLoop", slog.String("func", fmt.Sprintf("%#p", f)))

			if panicHandler := l.onPanic.Value(); panicHandler != nil {
				panicHandler(r)
			}
		}
	}()

	f()
}

func (l *EventLoop) SetOnPanicHandler(f func(p any)) {
	l.onPanic.SetValue(f)
}

// Post appends f to the internal queue. The functions are executed in the order of their posting (FiFo).
// A Post never blocks and keeps allocating space for messages infinitely. It may be called from any goroutine,
// including the loop itself. It returns false, if the loop has been destroyed. Otherwise, f is guaranteed to be
// executed.
func (l *EventLoop) Post(f func()) bool {
	l.mutex.Lock()
	if l.destroyed {
		l.mutex.Unlock()
		return false
	}

	l.queue = append(l.queue, f)
	l.pending.Add(1)
	l.mutex.Unlock()

	l.notify()
	return true
}

func (l *EventLoop) notify() {
	select {
	case l.wake <- struct{}{}:
	default:
		// the loop has already been notified and has not yet taken the queue
	}
}

// Pending returns the amount of posted functions which have not been completed yet. A function which is
// currently executed is included.
func (l *EventLoop) Pending() int64 {
	return l.pending.Load()
}

// Destroy stops accepting functions, thus future Post calls are ignored. The already posted functions and then
// the given final functions are executed in order, afterward the loop exits. Destroy never blocks and may also be
// called from the loop itself. It returns false, if the loop has already been destroyed.
func (l *EventLoop) Destroy(final ...func()) bool {
	l.mutex.Lock()
	if l.destroyed {
		l.mutex.Unlock()
		return false
	}

	l.destroyed = true
	l.queue = append(l.queue, final...)
	l.pending.Add(int64(len(final)))
	l.mutex.Unlock()

	l.notify()
	return true
}

// Done returns a channel which is closed, after the loop has been destroyed and has executed its last function.
func (l *EventLoop) Done() <-chan struct{} {
	return l.done
}
