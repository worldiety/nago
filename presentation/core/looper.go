package core

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"slices"
	"sync"
)

type msgType int

const (
	msgFunc msgType = iota + 1
)

type msg struct {
	typ msgType
	fn  func()
}

// An EventLoop handles messages respective executes functions.
// Everything which looks like a concurrent or parallel situation within the backend-frontend relation
// must be single-threaded through the EventLoop to avoid race conditions. Especially these are
//   - sending or receiving events to the frontend
//   - async domain events
//   - upload
//   - download
type EventLoop struct {
	// using go channels seems to be a bad choice, because their capacity is allocated eagerly and
	// we may need to grow "unbound".
	// Also
	queue []msg

	mutex     sync.Mutex
	batchChan chan []msg
	done      chan bool
	destroyed bool
}

func NewEventLoop() *EventLoop {
	l := &EventLoop{
		batchChan: make(chan []msg),
		done:      make(chan bool),
	}

	go func() {
		for {
			select {
			case <-l.done:
				return
			case batch := <-l.batchChan:
				for _, m := range batch {
					l.saveExec(m.fn)
				}
			}
		}
	}()

	return l
}

func (l *EventLoop) saveExec(f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
			slog.Error("recovered from panic in EventLoop", slog.String("func", fmt.Sprintf("%#v", f)))
		}
	}()

	f()
}

// Post appends f to the internal queue. It will be executed in the next tick cycle.
// This is a LiFo, so the oldest message is evaluated first.
// A Post never blocks and keeps allocating space for messages infinitely.
func (l *EventLoop) Post(f func()) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.destroyed {
		debug.PrintStack()
		slog.Error("someone posted to a destroyed EventLoop")
		return
	}

	l.queue = append(l.queue, msg{
		typ: msgFunc,
		fn:  f,
	})
}

// pull allocates a copy of the queue and returns it. The original queue is cleared.
func (l *EventLoop) pull() []msg {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.destroyed {
		return nil
	}

	tmp := slices.Clone(l.queue)
	clear(l.queue)        // remove hidden leaks
	l.queue = l.queue[:0] // reset

	return tmp
}

// Tick triggers the next batch of events to be processed. Without Tick, no messages will ever be processed.
// The current implementation blocks, until the message batch has been received from the looper.
// Thus, ticks may start accumulating in a blocking way.
// This is intentional, because it throttles e.g. a sender in a natural way (to many ticks and messages).
// A Tick can never block a Post or vice versa.
func (l *EventLoop) Tick() {
	messages := l.pull()
	if len(messages) == 0 {
		return
	}

	l.batchChan <- messages
}

// Destroy stops the internal looper thread and releases all resources.
// Unprocessed messages are discarded.
// Future Post calls are ignored.
func (l *EventLoop) Destroy() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.destroyed {
		return
	}

	l.destroyed = true
	l.done <- true
	close(l.done)
	close(l.batchChan)
	clear(l.queue)
}

// Shutdown executes all added tasks until now and blocks until finished.
// This will deadlock if called from the EventLoop itself.
func (l *EventLoop) Shutdown() {
	l.mutex.Lock()
	if l.destroyed {
		// make Shutdown idempotent
		l.mutex.Unlock()
		return
	}
	l.mutex.Unlock()

	var wg sync.WaitGroup
	wg.Add(1)
	l.Post(func() {
		wg.Done()
	})

	wg.Wait()
	l.Destroy()
}

// Executor returns an isolated adapter, so that no one can apply other interface assertions to screw up the looper.
func (l *EventLoop) Executor() Executor {
	return execAdapter{looper: l}
}

type execAdapter struct {
	looper *EventLoop
}

func (e execAdapter) Execute(task func()) {
	e.looper.Post(task)
}
