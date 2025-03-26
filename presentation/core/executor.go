// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

// An Executor takes a task and executes it eventually.
type Executor interface {
	Execute(task func())
}

// TODO not sure if we ever need that again, because we have OnAppear and OnDisappear which is has a smaller and more dynamic lifecycle.
//// Post executes the given task from within the event loop in the next cycle.
//// It is not executed, if the Window is destroyed before. It invalidates the view root automatically.
//func Post(wnd Window, task func()) {
//	if wnd != nil && task != nil {
//		wnd.Execute(func() {
//			if task != nil {
//				task()
//			}
//
//			wnd.Invalidate()
//		})
//
//	}
//}
//
//// PostDelayed executes the given task from within the event loop after the given duration.
//// It is not executed, if the Window is destroyed before. It invalidates the view root automatically.
//// The lifetime is scoped to the window.
//func PostDelayed(wnd Window, after time.Duration, task func()) {
//	if wnd != nil {
//		var timer *time.Timer
//		timer = time.AfterFunc(after, func() {
//			wnd.Execute(func() {
//				if task != nil {
//					task()
//				}
//				wnd.Invalidate()
//			})
//		})
//
//		wnd.AddDestroyObserver(func() {
//			timer.Stop()
//		})
//	}
//}

// Schedule repeats invocations of the given task within the event looper, until the Window is destroyed or
// the schedule is cancelled explicitly. It invalidates the view root automatically.
//func Schedule(wnd Window, d time.Duration, task func()) (cancel func()) {
//	if wnd != nil {
//		eolTicker := time.NewTicker(d)
//		eolDone := make(chan bool)
//		closed := false
//		wnd.AddDestroyObserver(func() {
//			eolTicker.Stop()
//			if closed {
//				return
//			}
//			eolDone <- true
//		})
//
//		go func() {
//			defer func() {
//				close(eolDone)
//				closed = true
//			}()
//			for {
//				select {
//				case <-eolDone:
//					return
//				case <-eolTicker.C:
//					wnd.Execute(func() {
//						if task != nil {
//							task()
//						}
//						wnd.Invalidate()
//					})
//				}
//			}
//		}()
//
//		return func() {
//			if closed {
//				return
//			}
//
//			eolDone <- true
//		}
//	}
//
//	return func() {
//		// nop
//	}
//}
