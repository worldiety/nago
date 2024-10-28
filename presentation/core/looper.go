package core

import (
	"fmt"
	"go.wdy.de/nago/pkg/std/concurrent"
	"log/slog"
	"runtime/debug"
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
	// Also, we want to control when and if at all we want to execute in batches or pause execution.
	queue concurrent.Slice[msg]

	batchChan chan []msg
	done      chan bool
	destroyed concurrent.Value[bool]
	onPanic   concurrent.Value[func(p any)]
	reloop    concurrent.Value[bool]
}

func NewEventLoop() *EventLoop {
	l := &EventLoop{
		batchChan: make(chan []msg, 1024),
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

			//if concurrent.CompareAndSwap(&l.reloop, true, false) {
			l.tickWithoutLock()
			//}

		}
	}()

	return l
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

// Post appends f to the internal queue. It will be executed in the next tick cycle.
// This is a LiFo, so the oldest message is evaluated first.
// A Post never blocks and keeps allocating space for messages infinitely.
func (l *EventLoop) Post(f func()) {
	// this is inprecise, because we may end up with neither this message nor an ever executed func
	if l.destroyed.Value() {
		debug.PrintStack()
		slog.Error("someone posted to a destroyed EventLoop")
		return
	}

	/*
		l.queue.Append(msg{
			typ: msgFunc,
			fn:  f,
		})*/
	l.batchChan <- []msg{msg{msgFunc, f}}

}

// pull allocates a copy of the queue and returns it. The original queue is cleared.
func (l *EventLoop) pull() []msg {
	if l.destroyed.Value() {
		return nil
	}

	tmp := l.queue.PopAll()
	return tmp
}

// Tick triggers the next batch of events to be processed. Without Tick, no messages will ever be processed.
// The current implementation blocks, until the message batch has been received from the looper.
// Thus, ticks may start accumulating in a blocking way.
// This is intentional, because it throttles e.g. a sender in a natural way (to many ticks and messages).
// A Tick can never block a Post or vice versa.
func (l *EventLoop) Tick() {
	l.tickWithoutLock()
}

func (l *EventLoop) tickWithoutLock() {
	/*
		messages := l.pull()
		if len(messages) == 0 {
			return
		}

		select {
		case l.batchChan <- messages:
			// as normal
		default:
			// the looper channel is busy and cannot accept.
			// this happens if the looper triggers itself a Tick
			l.queue.Append(messages...)
			l.reloop.SetValue(true)
		}
	*/
}

// Destroy stops the internal looper thread and releases all resources.
// Unprocessed messages are discarded.
// Future Post calls are ignored.
func (l *EventLoop) Destroy() {
	if l.destroyed.Value() {
		return
	}
	/*
		l.tickWithoutLock() // tick is posting messages into batchChan, but it is not clear if this has a defined timing regarding the eolDone channel

		l.eolDone <- true
		close(l.eolDone)
		close(l.batchChan)
		l.queue.Clear() //this is imprecise, but we want to reduce locks and therefore potential deadlocks
	*/
	l.Post(func() {
		if !concurrent.CompareAndSwap(&l.destroyed, false, true) {
			return
		}

		select {
		case l.done <- true:
		default:
			panic("eolDone cannot accept destruction twice")
		}

		close(l.done)
		close(l.batchChan)
	})
}

/*
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
*/

// Executor returns an isolated adapter, so that no one can apply other interface assertions to screw up the looper.
func (l *EventLoop) Executor() Executor {
	return execAdapter{looper: l}
}

type execAdapter struct {
	looper *EventLoop
}

func (e execAdapter) Execute(task func()) {
	e.looper.Post(task)
	e.looper.Tick() // ensure to trigger processing
}
