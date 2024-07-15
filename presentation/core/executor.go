package core

import (
	"time"
)

// An Executor takes a task and executes it eventually.
type Executor interface {
	Execute(task func())
}

// Post executes the given task from within the event loop in the next cycle.
// It is not executed, if the Window is destroyed before. It invalidates the view root automatically.
func Post(wnd Window, task func()) {
	if wnd != nil && task != nil {
		wnd.Execute(func() {
			task()
			wnd.Invalidate()
		})

	}
}

// PostDelayed executes the given task from within the event loop after the given duration.
// It is not executed, if the Window is destroyed before. It invalidates the view root automatically.
func PostDelayed(wnd Window, after time.Duration, task func()) {
	if wnd != nil {
		var timer *time.Timer
		timer = time.AfterFunc(after, func() {
			wnd.Execute(func() {
				task()
				wnd.Invalidate()
			})
		})

		wnd.AddDestroyObserver(func() {
			timer.Stop()
		})
	}
}

// Schedule repeats invocations of the given task within the event looper, until the Window is destroyed or
// the schedule is cancelled explicitly. It invalidates the view root automatically.
func Schedule(wnd Window, d time.Duration, task func()) (cancel func()) {
	if wnd != nil {
		ticker := time.NewTicker(d)
		done := make(chan bool)
		closed := false
		wnd.AddDestroyObserver(func() {
			ticker.Stop()
			if closed {
				return
			}
			done <- true
		})

		go func() {
			defer func() {
				close(done)
				closed = true
			}()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					wnd.Execute(func() {
						task()
						wnd.Invalidate()
					})
				}
			}
		}()

		return func() {
			if closed {
				return
			}

			done <- true
		}
	}

	return func() {
		// nop
	}
}
