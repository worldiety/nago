// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
)

// fakeTB records a failure instead of failing the test.
type fakeTB struct {
	mutex    sync.Mutex
	failure  string
	cleanups []func()
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Errorf(format string, args ...any) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.failure = fmt.Sprintf(format, args...)
}

func (f *fakeTB) Fatalf(format string, args ...any) {
	f.Errorf(format, args...)
	runtime.Goexit()
}

func (f *fakeTB) Logf(format string, args ...any) {}

func (f *fakeTB) Cleanup(fn func()) {
	f.cleanups = append(f.cleanups, fn)
}

// run executes fn like a test and returns its failure, if any.
func (f *fakeTB) run(fn func()) string {
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	<-done

	for _, fn := range f.cleanups {
		fn()
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.failure
}

func counterApp(t *testing.T) *nagotest.App {
	return nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("counter", func(wnd core.Window) core.View {
			count := core.AutoState[int](wnd)
			return ui.VStack(
				ui.Text(fmt.Sprintf("count: %d", count.Get())),
				ui.PrimaryButton(func() { count.Set(count.Get() + 1) }).Title("increment"),
			)
		})
	})
}

// rerender renders the window again, e.g. like a frame ticker, which invalidates all callbacks of its tree.
func rerender(t *testing.T, w *nagotest.Window) {
	t.Helper()
	if err := w.Scope().Dispatch(&proto.RootViewRenderingRequested{RID: 9000}); err != nil {
		t.Fatal(err)
	}
	w.Settle()
}

func TestStaleActionIsRepeated(t *testing.T) {
	w := counterApp(t).Open(t, nil, "counter")

	sel := w.Find(nagotest.Text("increment"))
	rerender(t, w)

	before := w.Scope().StaleCalls()
	w.Click(sel)

	if n := w.Scope().StaleCalls() - before; n != 1 {
		t.Fatalf("expected a single stale click, got %d", n)
	}

	// the repeated click has been executed exactly once
	w.Find(nagotest.Text("count: 1"))
}

func TestStaleActionFailsWithoutRetries(t *testing.T) {
	app := counterApp(t)

	var tb fakeTB
	failure := tb.run(func() {
		w := app.Open(&tb, nil, "counter", nagotest.StaleRetries(0))
		sel := w.Find(nagotest.Text("increment"))
		rerender(t, w)
		w.Click(sel)
	})

	if !strings.Contains(failure, "stale") {
		t.Fatalf("expected a stale failure, got %q", failure)
	}
}

func TestRenderLoopDoesNotSettle(t *testing.T) {
	old := nagotest.SettleTimeout
	nagotest.SettleTimeout = time.Second
	defer func() { nagotest.SettleTimeout = old }()

	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("loop", func(wnd core.Window) core.View {
			renders := core.AutoState[int](wnd)
			renders.Set(renders.Get() + 1) // a changing value in each render
			return ui.Text(fmt.Sprintf("renders: %d", renders.Get()))
		})
	})

	var tb fakeTB
	failure := tb.run(func() {
		app.Open(&tb, nil, "loop")
	})

	if !strings.Contains(failure, "did not settle") {
		t.Fatalf("expected a settle failure, got %q", failure)
	}
}

// TestTransientStateAfterNavigation ensures that the transient states of a scope, like the banner messages,
// are not considered dirty after a navigation, just because the new window starts with a lower generation.
func TestTransientStateAfterNavigation(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("a", func(wnd core.Window) core.View {
			count := core.AutoState[int](wnd)
			// the banners allocate the transient state, whose generation follows the renders of this page
			return ui.VStack(
				ui.PrimaryButton(func() { count.Set(count.Get() + 1) }).Title("increment"),
				ui.PrimaryButton(func() {
					alert.ShowBannerMessage(wnd, alert.Message{Title: "saved", Message: "saved"})
					wnd.Navigation().ForwardTo("b", nil)
				}).Title("go"),
				alert.BannerMessages(wnd),
			)
		})
		cfg.RootView("b", func(wnd core.Window) core.View {
			return ui.VStack(ui.Text("page b"), alert.BannerMessages(wnd))
		})
	})

	w := app.Open(t, nil, "a")
	for range 40 {
		w.Click(w.Find(nagotest.Text("increment")))
	}

	renders := func() int {
		n := 0
		for _, evt := range w.Events() {
			if _, ok := evt.(*proto.RootViewInvalidated); ok {
				n++
			}
		}
		return n
	}

	before := renders()
	w.Click(w.Find(nagotest.Text("go")))
	w.Find(nagotest.Text("page b"))

	// render a, allocate b and a few more, but not one per render of the former window
	if n := renders() - before; n > 5 {
		t.Fatalf("expected a few renders, got %d", n)
	}

	settled := renders()
	time.Sleep(300 * time.Millisecond)
	if n := renders() - settled; n != 0 {
		t.Fatalf("expected no renders while idle, got %d", n)
	}
}

func TestMultiStepsWithoutSteps(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("wizard", func(wnd core.Window) core.View {
			return ui.VStack(ui.Text("empty wizard"), form.MultiSteps())
		})
	})

	w := app.Open(t, nil, "wizard")
	w.Find(nagotest.Text("empty wizard"))
}

func TestCloseRemovesScopeAndResolvesScreenshot(t *testing.T) {
	var wnd core.Window
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("shot", func(w core.Window) core.View {
			wnd = w
			return ui.Text("shot")
		})
	})

	w := app.Open(t, nil, "shot")
	id := w.Scope().ID()

	// the window records the capture request but never answers it
	fut := wnd.Screenshot(core.ScreenshotOptions{})
	w.Settle()
	if fut.Done() {
		t.Fatal("screenshot must not be done yet")
	}

	w.Close()

	if _, ok := app.Core().Scope(id); ok {
		t.Fatal("closed scope is still registered")
	}

	if !fut.Done() || !errors.Is(fut.Err(), core.ErrScreenshotUnavailable) {
		t.Fatalf("expected an unavailable screenshot, got %v", fut.Err())
	}
}

// shortSettle lowers the settle timeout for the test.
func shortSettle(t *testing.T, d time.Duration) {
	old := nagotest.SettleTimeout
	nagotest.SettleTimeout = d
	t.Cleanup(func() { nagotest.SettleTimeout = old })
}

// A transient state set during a render, even with a changing value like a banner, must not cause a render loop.
func TestTransientStateSetInRender(t *testing.T) {
	shortSettle(t, 2*time.Second)

	var renders atomic.Int32
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("same", func(wnd core.Window) core.View {
			core.TransientStateOf[string](wnd, "last-page").Set("same")
			return ui.Text("same")
		})
		cfg.RootView("banner", func(wnd core.Window) core.View {
			renders.Add(1)
			alert.ShowBannerError(wnd, fmt.Errorf("query failed at %s", time.Now().Format(time.RFC3339Nano)))
			return ui.VStack(ui.Text("banner"), alert.BannerMessages(wnd))
		})
	})

	app.Open(t, nil, "same").Find(nagotest.Text("same"))
	app.Open(t, nil, "banner").Find(nagotest.Text("banner"))

	if n := renders.Load(); n > 3 {
		t.Fatalf("expected a few renders, got %d", n)
	}
}

// Settle waits for the background work of nago, like [core.OnAppear] and [core.State.AsyncInit].
func TestSettleWaitsForBackground(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("bg", func(wnd core.Window) core.View {
			a := core.AutoState[string](wnd)
			b := core.AutoState[string](wnd).AsyncInit(func() string {
				time.Sleep(time.Duration(rand.IntN(3000)) * time.Microsecond)
				return "async-done"
			})
			core.OnAppear(wnd, "", func(ctx context.Context) {
				time.Sleep(time.Duration(rand.IntN(3000)) * time.Microsecond)
				a.Set("appear-done")
			})
			return ui.VStack(ui.Text("a="+a.Get()), ui.Text("b="+b.Get()))
		})
	})

	for range 50 {
		w := app.Open(t, nil, "bg")
		w.Find(nagotest.Text("a=appear-done"))
		w.Find(nagotest.Text("b=async-done"))
		w.Close()
	}
}

// An initialization which finished while a render is queued must not run again.
func TestAsyncInitRunsOnce(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	returned := make(chan struct{})
	var wnd core.Window
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("init", func(w core.Window) core.View {
			wnd = w
			show := core.StateOf[bool](w, "show")
			value := "hidden"
			if show.Get() {
				value = core.AutoState[string](w).AsyncInit(func() string {
					calls.Add(1)
					<-release
					defer close(returned)
					return "loaded"
				}).Get()
			}
			return ui.VStack(ui.Toggle(show.Get()).InputChecked(show).ID("show"), ui.Text("value="+value))
		})
	})

	w := app.Open(t, nil, "init")
	toggle := w.Find(nagotest.ID("show")).Node().Component.(*proto.Toggle).InputValue
	if err := w.Scope().Dispatch(&proto.UpdateStateValueRequested{StatePointer: toggle, Value: "true", RID: 100}); err != nil {
		t.Fatal(err)
	}

	for calls.Load() == 0 {
		time.Sleep(time.Millisecond)
	}

	// block the loop and queue a render before the initialized value
	loopRelease := make(chan struct{})
	wnd.Post(func() { <-loopRelease })
	if err := w.Scope().Dispatch(&proto.RootViewRenderingRequested{RID: 101}); err != nil {
		t.Fatal(err)
	}

	close(release)
	<-returned
	time.Sleep(50 * time.Millisecond) // the goroutine of the initialization ends
	close(loopRelease)

	w.Settle()
	w.Find(nagotest.Text("value=loaded"))
	if n := calls.Load(); n != 1 {
		t.Fatalf("initialization ran %d times", n)
	}
}

func TestScreenshotOfNavigatedWindow(t *testing.T) {
	var wndA core.Window
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("a", func(wnd core.Window) core.View {
			wndA = wnd
			return ui.PrimaryButton(func() {
				wnd.Navigation().ForwardTo("b", nil)
			}).Title("go")
		})
		cfg.RootView("b", func(wnd core.Window) core.View {
			return ui.Text("page b")
		})
	})

	// requested before the navigation
	w := app.Open(t, nil, "a")
	pending := wndA.Screenshot(core.ScreenshotOptions{})
	w.Settle()
	if pending.Done() {
		t.Fatal("must not be done yet")
	}

	w.Click(w.Find(nagotest.Text("go")))
	w.Find(nagotest.Text("page b"))
	if !pending.Done() || !errors.Is(pending.Err(), core.ErrScreenshotUnavailable) {
		t.Fatalf("expected unavailable after navigation, done=%v err=%v", pending.Done(), pending.Err())
	}

	// requested after the navigation
	calls := len(w.AsyncCalls())
	late := wndA.Screenshot(core.ScreenshotOptions{})
	w.Settle()
	if !late.Done() || !errors.Is(late.Err(), core.ErrScreenshotUnavailable) {
		t.Fatalf("expected unavailable for a closed window, done=%v err=%v", late.Done(), late.Err())
	}

	if n := len(w.AsyncCalls()) - calls; n != 0 {
		t.Fatalf("the closed window asked the frontend of the new page %d times", n)
	}
}

// A state update which the frontend sends for a former page must not hit a state of the current one.
func TestStateUpdateOfFormerPageIsStale(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("a", func(wnd core.Window) core.View {
			name := core.AutoState[string](wnd)
			return ui.VStack(
				ui.TextField("Name", name.Get()).InputValue(name),
				ui.PrimaryButton(func() { wnd.Navigation().ForwardTo("b", nil) }).Title("go"),
			)
		})
		cfg.RootView("b", func(wnd core.Window) core.View {
			secret := core.AutoState[string](wnd).Init(func() string { return "unchanged" })
			return ui.VStack(ui.Text("secret: "+secret.Get()), ui.TextField("Secret", secret.Get()).InputValue(secret))
		})
	})

	w := app.Open(t, nil, "a")
	ptr := w.Find(nagotest.Label("Name")).Node().Component.(*proto.TextField).InputValue
	w.Click(w.Find(nagotest.Text("go")))
	w.Find(nagotest.Text("secret: unchanged"))

	before := w.Scope().StaleCalls()
	if err := w.Scope().Dispatch(&proto.UpdateStateValueRequested{StatePointer: ptr, Value: "hacked", RID: 100}); err != nil {
		t.Fatal(err)
	}
	w.Settle()

	w.Find(nagotest.Text("secret: unchanged"))
	if n := w.Scope().StaleCalls() - before; n != 1 {
		t.Fatalf("expected a stale update, got %d", n)
	}
}

// Settle fails within its timeout, even if a function blocks the event loop.
func TestSettleTimesOutOnBlockedLoop(t *testing.T) {
	shortSettle(t, 500*time.Millisecond)

	release := make(chan struct{})
	defer close(release)
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("block", func(wnd core.Window) core.View {
			return ui.PrimaryButton(func() { <-release }).Title("block")
		})
	})

	var tb fakeTB
	start := time.Now()
	failure := tb.run(func() {
		w := app.Open(&tb, nil, "block")
		w.Click(w.Find(nagotest.Text("block")))
	})

	if !strings.Contains(failure, "did not settle") {
		t.Fatalf("expected a settle failure, got %q", failure)
	}

	if d := time.Since(start); d > 3*time.Second {
		t.Fatalf("settle took %v", d)
	}
}

// A browser keeps its scope id across a reconnect, thus it may end up with a new scope of the same id while
// it still shows the tree of the former one, e.g. after the former scope has been reaped. Nothing of that tree
// may hit the new scope.
func TestFormerScopeOfSameIDIsStale(t *testing.T) {
	var hits atomic.Int32
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("page", func(wnd core.Window) core.View {
			name := core.AutoState[string](wnd).Init(func() string { return "unchanged" })
			return ui.VStack(
				ui.PrimaryButton(func() { hits.Add(1) }).Title("hit"),
				ui.Text("name: "+name.Get()),
				ui.TextField("Name", name.Get()).InputValue(name),
			)
		})
	})

	w := app.Open(t, nil, "page")
	action := w.Find(nagotest.Text("hit")).Node()
	callback := action.Parents[len(action.Parents)-1].(*proto.Stack).Action
	state := w.Find(nagotest.Label("Name")).Node().Component.(*proto.TextField).InputValue

	// the scope is reaped, and the reconnect creates a new one of the same id
	former := w.Scope()
	if !app.Core().DestroyScope(former.ID()) {
		t.Fatal("cannot destroy the scope")
	}
	w.Reconnect()
	if w.Scope() == former || w.Scope().ID() != former.ID() {
		t.Fatal("expected a new scope of the same id")
	}

	before := w.Scope().StaleCalls()
	if err := w.Scope().Dispatch(&proto.FunctionCallRequested{Ptr: callback, RID: 100}); err != nil {
		t.Fatal(err)
	}
	if err := w.Scope().Dispatch(&proto.UpdateStateValueRequested{StatePointer: state, Value: "hacked", RID: 101}); err != nil {
		t.Fatal(err)
	}
	w.Settle()

	if hits.Load() != 0 {
		t.Fatal("a callback of the former scope has been called")
	}
	w.Find(nagotest.Text("name: unchanged"))
	if n := w.Scope().StaleCalls() - before; n != 2 {
		t.Fatalf("expected two stale requests, got %d", n)
	}
}

// A late answer of the frontend for a call of a former page must not hit a call of the current one.
func TestAsyncAnswerOfFormerPageIsIgnored(t *testing.T) {
	var answered atomic.Int32
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("a", func(wnd core.Window) core.View {
			return ui.PrimaryButton(func() {
				core.AsyncCall(wnd, &proto.CallRequestFocus{ID: "x"}, func(ret proto.CallRet) {})
				wnd.Navigation().ForwardTo("b", nil)
			}).Title("go")
		})
		cfg.RootView("b", func(wnd core.Window) core.View {
			core.OnAppear(wnd, "", func(ctx context.Context) {
				wnd.Post(func() {
					core.AsyncCall(wnd, &proto.CallRequestFocus{ID: "y"}, func(ret proto.CallRet) { answered.Add(1) })
				})
			})
			return ui.Text("page b")
		})
	})

	w := app.Open(t, nil, "a")
	w.Click(w.Find(nagotest.Text("go")))
	w.Find(nagotest.Text("page b"))
	w.Settle()

	calls := w.AsyncCalls()
	if len(calls) != 2 {
		t.Fatalf("expected the calls of both pages, got %d", len(calls))
	}

	// the answer for page a arrives late
	if err := w.Scope().Dispatch(&proto.CallResolved{CallPtr: calls[0].CallPtr, Ret: &proto.RetError{}}); err != nil {
		t.Fatal(err)
	}
	w.Settle()

	if answered.Load() != 0 {
		t.Fatal("the answer for page a has been delivered to page b")
	}
}

// The callback pointers of consecutive trees form disjoint, ascending intervals.
func TestCallbackPointersAreIntervals(t *testing.T) {
	w := counterApp(t).Open(t, nil, "counter")

	pointers := func() []proto.Ptr {
		var res []proto.Ptr
		for _, n := range w.FindAll(nagotest.Type[*proto.Stack]()).Nodes() {
			if ptr := n.Component.(*proto.Stack).Action; ptr != 0 {
				res = append(res, ptr)
			}
		}
		return res
	}

	first := pointers()
	w.Click(w.Find(nagotest.Text("increment")))
	second := pointers()

	if len(first) == 0 || len(first) != len(second) {
		t.Fatalf("unexpected pointers %v %v", first, second)
	}

	if first[0] < 1<<32 || second[0] <= first[len(first)-1] {
		t.Fatalf("expected ascending callback pointers above 2^32, got %v %v", first, second)
	}
}

// slowStep is how long the next button of the keyed app takes.
var slowStep time.Duration

// rowRemover changes the rows of the keyed app from outside, see keyedApp.
type rowRemover struct {
	wnd  core.Window
	rows *core.State[[]string]
}

func (r *rowRemover) remove(id string) {
	r.wnd.Post(func() {
		r.rows.Set(slices.DeleteFunc(slices.Clone(r.rows.Get()), func(s string) bool { return s == id }))
	})
}

// keyedApp has a text field which is saved by a keyed button, a keyed wizard and a keyed list.
func keyedApp(t *testing.T, remove *rowRemover) *nagotest.App {
	return nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("form", func(wnd core.Window) core.View {
			name := core.AutoState[string](wnd)
			saved := core.AutoState[string](wnd)
			step := core.AutoState[int](wnd)
			return ui.VStack(
				ui.TextField("Name", name.Get()).InputValue(name),
				ui.PrimaryButton(func() { saved.Set(name.Get()) }).Title("save").Key("save", "").Enabled(name.Get() != "invalid"),
				ui.PrimaryButton(func() { saved.Set("id:" + name.Get()) }).Title("save by id").ID("save-by-id"),
				ui.Text("saved: "+saved.Get()),
				ui.PrimaryButton(func() {
					time.Sleep(slowStep)
					step.Set(step.Get() + 1)
				}).Title("next").Key("next", ""),
				ui.Text(fmt.Sprintf("step: %d", step.Get())),
				ui.Text("item: "+wnd.Values()["item"]),
				ui.PrimaryButton(func() { saved.Set("deleted " + wnd.Values()["item"]) }).Title("delete").Key("delete", ""),
			)
		})
		cfg.RootView("rows", func(wnd core.Window) core.View {
			rows := core.AutoState[[]string](wnd).Init(func() []string { return []string{"a", "b", "c"} })
			// lets the test change the rows from outside, like a domain event would
			remove.wnd, remove.rows = wnd, rows

			var list []core.View
			for _, row := range rows.Get() {
				list = append(list, ui.HStack(
					ui.Text("row "+row),
					ui.SecondaryButton(func() {
						rows.Set(slices.DeleteFunc(slices.Clone(rows.Get()), func(s string) bool { return s == row }))
					}).Title("remove "+row).Key("remove", row),
				))
			}
			return ui.VStack(list...)
		})
	})
}

func actionOf(t *testing.T, w *nagotest.Window, text string) proto.Ptr {
	t.Helper()
	n := w.Find(nagotest.Text(text)).Node()
	return n.Parents[len(n.Parents)-1].(*proto.Stack).Action
}

func call(t *testing.T, w *nagotest.Window, ptr proto.Ptr) {
	t.Helper()
	if err := w.Scope().Dispatch(&proto.FunctionCallRequested{Ptr: ptr, RID: 500}); err != nil {
		t.Fatal(err)
	}
	w.Settle()
}

// The value of a text field is sent when it loses its focus, right before the click on save: the click refers
// to the tree before the value, but its key redirects it to the current save, which reads the new value.
func TestKeyedClickAfterInputIsNotLost(t *testing.T) {
	w := keyedApp(t, nil).Open(t, nil, "form")

	save := actionOf(t, w, "save")
	var saveByID proto.Ptr
	before := w.Scope().StaleCalls()

	w.Type(w.Find(nagotest.Label("Name")), "Alice")
	call(t, w, save)
	w.Find(nagotest.Text("saved: Alice"))

	// the same with a key derived from the id of the button
	saveByID = actionOf(t, w, "save by id")
	w.Type(w.Find(nagotest.Label("Name")), "Bob")
	call(t, w, saveByID)
	w.Find(nagotest.Text("saved: id:Bob"))

	if n := w.Scope().StaleCalls() - before; n != 0 {
		t.Fatalf("expected no stale calls, got %d", n)
	}
}

func TestKeyedDoubleClickAdvancesOnce(t *testing.T) {
	for _, slow := range []time.Duration{0, 700 * time.Millisecond} {
		slowStep = slow
		w := keyedApp(t, nil).Open(t, nil, "form")

		next := actionOf(t, w, "next")
		before := w.Scope().StaleCalls()
		call(t, w, next)
		call(t, w, next)

		w.Find(nagotest.Text("step: 1"))
		if n := w.Scope().StaleCalls() - before; n != 1 {
			t.Fatalf("expected the second click to be stale with a %v callback, got %d", slow, n)
		}
		w.Close()
	}
	slowStep = 0
}

// A call is not redirected to an action which the current tree disables or hides.
func TestKeyedClickIsNotRedirectedToDisabledButton(t *testing.T) {
	w := keyedApp(t, nil).Open(t, nil, "form")

	save := actionOf(t, w, "save")
	before := w.Scope().StaleCalls()
	w.Type(w.Find(nagotest.Label("Name")), "invalid")
	call(t, w, save)

	w.FindAll(nagotest.Text("saved: invalid")).None()
	if n := w.Scope().StaleCalls() - before; n != 1 {
		t.Fatalf("expected a stale click, got %d", n)
	}
}

// The same route may show another item after a navigation, so a key like ("delete", "") means something else.
func TestKeyedClickIsNotRedirectedAcrossValues(t *testing.T) {
	w := keyedApp(t, nil).Open(t, nil, "form", nagotest.Values(core.Values{"item": "1"}))
	w.Find(nagotest.Text("item: 1"))

	del := actionOf(t, w, "delete")
	before := w.Scope().StaleCalls()

	// a navigation to item 2 within the same route reuses the window
	if err := w.Scope().Dispatch(&proto.RootViewAllocationRequested{Factory: "form", Values: proto.RootViewParameters{"item": "2"}, RID: 600}); err != nil {
		t.Fatal(err)
	}
	w.Settle()
	w.Find(nagotest.Text("item: 2"))

	call(t, w, del)
	w.FindAll(nagotest.TextContains("deleted")).None()
	if n := w.Scope().StaleCalls() - before; n != 1 {
		t.Fatalf("expected a stale click, got %d", n)
	}
}

// A keyed click on a row hits that row even after the rows have been rendered again, and never another one.
func TestKeyedClickHitsTheRow(t *testing.T) {
	var remove rowRemover
	w := keyedApp(t, &remove).Open(t, nil, "rows")

	removeA := actionOf(t, w, "remove a")
	removeB := actionOf(t, w, "remove b")
	removeC := actionOf(t, w, "remove c")

	// a background change removes b and renders again
	remove.remove("b")
	w.Settle()
	w.FindAll(nagotest.Text("row b")).None()

	before := w.Scope().StaleCalls()
	call(t, w, removeB) // gone
	w.Find(nagotest.Text("row a"))
	w.Find(nagotest.Text("row c"))
	if n := w.Scope().StaleCalls() - before; n != 1 {
		t.Fatalf("expected a stale click on the removed row, got %d", n)
	}

	call(t, w, removeC) // moved up, still hit
	w.Find(nagotest.Text("row a"))
	w.FindAll(nagotest.Text("row c")).None()

	// the former tree has been used, so its other clicks are stale
	call(t, w, removeA)
	w.Find(nagotest.Text("row a"))
	if n := w.Scope().StaleCalls() - before; n != 2 {
		t.Fatalf("expected two stale clicks, got %d", n)
	}
}

// A click on the save button of a dialog right after typing refers to the tree before the typed value. The
// predefined dialog buttons are keyed, so the click is not lost.
func TestDialogSaveAfterInputIsNotLost(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("dialog", func(wnd core.Window) core.View {
			open := core.AutoState[bool](wnd).Init(func() bool { return true })
			name := core.AutoState[string](wnd)
			saved := core.AutoState[string](wnd)
			return ui.VStack(
				alert.Dialog("Rename", ui.TextField("Name", name.Get()).InputValue(name), open, alert.Save(func() bool {
					saved.Set(name.Get())
					return true
				}), alert.Cancel(nil)),
				ui.Text("saved: "+saved.Get()),
			)
		})
	})

	w := app.Open(t, nil, "dialog")
	save := actionOf(t, w, "Speichern")
	before := w.Scope().StaleCalls()

	w.Type(w.Find(nagotest.Label("Name")), "Alice")
	call(t, w, save)

	w.Find(nagotest.Text("saved: Alice"))
	if n := w.Scope().StaleCalls() - before; n != 0 {
		t.Fatalf("expected no stale calls, got %d", n)
	}
}
