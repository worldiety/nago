package core

import (
	"time"
)

// An Executor takes a task and executes it eventually.
type Executor interface {
	Execute(task func())
}

// PostDelayed executes the given task from within the event loop after the given duration.
// It is not executed, if the Window is destroyed before. It invalidates the view root automatically.
func PostDelayed(wnd Window, after time.Duration, task func()) {
	if root := wnd.ViewRoot(); root != nil {
		var timer *time.Timer
		timer = time.AfterFunc(after, func() {
			wnd.Execute(func() {
				task()
				root.Invalidate()
			})
		})

		root.AddDestroyObserver(func() {
			timer.Stop()
		})
	}
}

// Schedule repeats invocations of the given task within the event looper, until the Window is destroyed or
// the schedule is cancelled explicitly. It invalidates the view root automatically.
func Schedule(wnd Window, d time.Duration, task func()) (cancel func()) {
	if root := wnd.ViewRoot(); root != nil {
		ticker := time.NewTicker(d)
		done := make(chan bool)
		root.AddDestroyObserver(func() {
			ticker.Stop()
			done <- true
		})

		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					wnd.Execute(func() {
						task()
						root.Invalidate()
					})
				}
			}
		}()

		return func() {
			done <- true
		}
	}

	return func() {
		// nop
	}
}
