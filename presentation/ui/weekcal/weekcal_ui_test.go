// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package weekcal_test

// Acceptance tests of the week calendar on the server (wokoda N7, R-VIEWS-WEEKCAL, L-N7-2): every action of the
// client is checked against the calendar of the last render of its window, before a callback runs; whatever does not
// pass is discarded without a callback, the sequence number is acknowledged in the next tree also after a rejection,
// and an error of OnSelect or OnMove stays as the message until the next action. Written against the public API
// (weekcal, proto.WeekCalendar, nagotest.Window.WeekCalendarSend), not against the implementation.
//
// The calendar of most tests has two days, the range 360 to 1320 in a raster of 15 minutes and the events A day 0
// 08:00 to 10:30, B day 1 09:00 to 09:15 (fixed), C 08:07 to 08:52 and D 05:00 to 07:00 (beyond the range). Each
// rejection names a row, which passes and differs only in what the rule checks.

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/weekcal"
)

// interval is the base of an action.
type interval struct {
	Day  int `json:"day"`
	From int `json:"from"`
	To   int `json:"to"`
}

// action is an action of the client, raw replaces its json.
type action struct {
	Seq  int64     `json:"seq"`
	Kind string    `json:"kind"`
	ID   string    `json:"id"`
	Day  int       `json:"day"`
	From int       `json:"from"`
	To   int       `json:"to"`
	Base *interval `json:"base,omitempty"`
	raw  string
}

var seq atomic.Int64

func (a action) text() string {
	if a.raw != "" {
		return a.raw
	}
	buf, _ := json.Marshal(a)
	return string(buf)
}

func sel(d, f, t int) action { return action{Kind: "select", Day: d, From: f, To: t} }
func click(id string, b interval) action {
	return action{Kind: "click", ID: id, Day: b.Day, From: b.From, To: b.To, Base: &b}
}
func move(id string, d, f, t int, b interval) action {
	return action{Kind: "move", ID: id, Day: d, From: f, To: t, Base: &b}
}
func resize(id string, d, f, t int, b interval) action {
	return action{Kind: "resize", ID: id, Day: d, From: f, To: t, Base: &b}
}

var (
	ivA = interval{0, 480, 630}
	ivB = interval{1, 540, 555}
	ivC = interval{0, 487, 532}
	ivD = interval{0, 300, 420}
)

// model is the application: it records each callback and may return an error.
type model struct {
	mutex sync.Mutex
	calls []string
	fail  bool
}

func (m *model) record(s string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls = append(m.calls, s)
}

func (m *model) recorded() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]string(nil), m.calls...)
}

func (m *model) failing() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.fail
}

func (m *model) setFail(b bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.fail = b
}

// openCalendar opens the window "cal" with the calendar of the table and buttons, which toggle what the next render
// shows.
func openCalendar(t *testing.T, m *model) *nagotest.Window {
	t.Helper()
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.weekcaltest")
		cfg.RootView("cal", func(wnd core.Window) core.View {
			disabled := core.StateOf[bool](wnd, "disabled")
			withoutA := core.StateOf[bool](wnd, "withoutA")
			hidden := core.StateOf[bool](wnd, "hidden")
			noSelect := core.StateOf[bool](wnd, "noSelect")
			noMove := core.StateOf[bool](wnd, "noMove")
			noClick := core.StateOf[bool](wnd, "noClick")
			renders := core.StateOf[int](wnd, "renders")

			events := []weekcal.Event{
				{ID: "A", Day: 0, From: 480, To: 630, Title: "A"},
				{ID: "B", Day: 1, From: 540, To: 555, Title: "B", Fixed: true},
				{ID: "C", Day: 0, From: 487, To: 532, Title: "C"},
				{ID: "D", Day: 0, From: 300, To: 420, Title: "D"},
			}
			if withoutA.Get() {
				events = events[1:]
			}

			cal := weekcal.WeekCalendar("cal", weekcal.Day{Title: "Montag"}, weekcal.Day{Title: "Dienstag"}).
				Range(360, 1320).Step(15).Append(events...).Disabled(disabled.Get())
			if !noClick.Get() {
				cal = cal.OnClick(func(id string) { m.record("CLICK " + id) })
			}
			if !noSelect.Get() {
				cal = cal.OnSelect(func(s weekcal.Span) error {
					m.record(fmt.Sprintf("SELECT %d %d %d", s.Day, s.From, s.To))
					if m.failing() {
						return errors.New("belegt")
					}
					return nil
				})
			}
			if !noMove.Get() {
				cal = cal.OnMove(func(mv weekcal.Move) error {
					m.record(fmt.Sprintf("MOVE %s %d %d %d %d %d %d", mv.ID, mv.Day, mv.From, mv.To, mv.OldDay, mv.OldFrom, mv.OldTo))
					if m.failing() {
						return errors.New("belegt")
					}
					return nil
				})
			}

			toggle := func(s *core.State[bool], title string) core.View {
				return ui.PrimaryButton(func() { s.Set(!s.Get()) }).Title(title)
			}
			views := []core.View{
				ui.PrimaryButton(func() { renders.Set(renders.Get() + 1) }).Title("RENDER"),
				toggle(disabled, "DISABLED"), toggle(withoutA, "WITHOUT A"), toggle(hidden, "HIDDEN"),
				toggle(noSelect, "NO SELECT"), toggle(noMove, "NO MOVE"), toggle(noClick, "NO CLICK"),
			}
			if !hidden.Get() {
				views = append(views, cal)
			}
			return ui.VStack(views...)
		})
	})

	return app.Open(t, nil, "cal", nagotest.StaleRetries(0))
}

func calendars(w *nagotest.Window) []*proto.WeekCalendar {
	var res []*proto.WeekCalendar
	for _, n := range w.FindAll(nagotest.Type[*proto.WeekCalendar]()).Nodes() {
		res = append(res, n.Component.(*proto.WeekCalendar))
	}
	return res
}

func calendarOf(t *testing.T, w *nagotest.Window) *proto.WeekCalendar {
	t.Helper()
	cs := calendars(w)
	if len(cs) != 1 {
		t.Fatalf("premise: %d calendars instead of one", len(cs))
	}
	return cs[0]
}

// send sends a with a new sequence number to cal and returns the number.
func send(w *nagotest.Window, cal *proto.WeekCalendar, a action) int64 {
	a.Seq = seq.Add(1)
	w.WeekCalendarSend(cal, a.text())
	return a.Seq
}

func press(w *nagotest.Window, title string) { w.Click(w.Find(nagotest.Text(title))) }

// callbackOf is the action of the node or of its nearest ancestor, like the frontend sends a click.
func callbackOf(n nagotest.Node) proto.Ptr {
	chain := []proto.Component{n.Component}
	for i := len(n.Parents) - 1; i >= 0; i-- {
		chain = append(chain, n.Parents[i])
	}
	for _, c := range chain {
		v := reflect.ValueOf(c)
		for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
			if v.IsNil() {
				break
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			continue
		}
		for _, name := range []string{"Action", "OnClick"} {
			if f := v.FieldByName(name); f.IsValid() && f.Type() == reflect.TypeFor[proto.Ptr]() && f.Uint() != 0 {
				return proto.Ptr(f.Uint())
			}
		}
	}
	return 0
}

// row is a row of the table: calls is the recorded callback, "" means none, and then passes names the row, which
// passes without the rule.
type row struct {
	name   string
	a      action
	calls  string
	passes string
	noAck  bool
}

var rows = []row{
	{"select day 0 09:00-10:00", sel(0, 540, 600), "SELECT 0 540 600", "", false},
	{"select 06:00-07:00 (from at the start of the range)", sel(0, 360, 420), "SELECT 0 360 420", "", false},
	{"select 21:00-22:00 (to at the end of the range)", sel(0, 1260, 1320), "SELECT 0 1260 1320", "", false},
	{"select day 1 09:00-10:00", sel(1, 540, 600), "SELECT 1 540 600", "", false},
	{"select 09:05-10:00", sel(0, 545, 600), "", "select day 0 09:00-10:00", false},
	{"select 09:00-10:05", sel(0, 540, 605), "", "select day 0 09:00-10:00", false},
	{"select 10:00-09:00", sel(0, 600, 540), "", "select day 0 09:00-10:00", false},
	{"select day 2", sel(2, 540, 600), "", "select day 1 09:00-10:00", false},
	{"select day -1", sel(-1, 540, 600), "", "select day 0 09:00-10:00", false},
	{"select 05:00-06:30", sel(0, 300, 390), "", "select 06:00-07:00 (from at the start of the range)", false},
	{"select 21:00-22:15", sel(0, 1260, 1335), "", "select 21:00-22:00 (to at the end of the range)", false},
	{"select 09:00-09:00", sel(0, 540, 540), "", "select day 0 09:00-10:00", false},

	{"click A", click("A", ivA), "CLICK A", "", false},
	{"click B (fixed)", click("B", ivB), "CLICK B", "", false},
	{"click X (unknown)", click("X", ivA), "", "click A", false},

	{"move A to day 1 09:15-11:45", move("A", 1, 555, 705, ivA), "MOVE A 1 555 705 0 480 630", "", false},
	{"move A to day 0 09:15-11:45", move("A", 0, 555, 705, ivA), "MOVE A 0 555 705 0 480 630", "", false},
	{"move A to 19:30-22:00 (end of the range)", move("A", 0, 1170, 1320, ivA), "MOVE A 0 1170 1320 0 480 630", "", false},
	{"move A by 15 minutes", move("A", 0, 495, 645, ivA), "MOVE A 0 495 645 0 480 630", "", false},
	{"move A with another duration", move("A", 1, 555, 720, ivA), "", "move A to day 1 09:15-11:45", false},
	{"move A to 09:20", move("A", 0, 560, 710, ivA), "", "move A to day 0 09:15-11:45", false},
	{"move A to 21:00-23:30 (outside)", move("A", 0, 1260, 1410, ivA), "", "move A to 19:30-22:00 (end of the range)", false},
	{"move A to day 2", move("A", 2, 480, 630, ivA), "", "move A to day 1 09:15-11:45", false},
	{"move B (fixed)", move("B", 1, 555, 570, ivB), "", "move A by 15 minutes", false},
	{"move A with base 09:00-11:30", move("A", 0, 555, 705, interval{0, 540, 690}), "", "move A to day 0 09:15-11:45", false},

	{"resize A to 11:00", resize("A", 0, 480, 660, ivA), "MOVE A 0 480 660 0 480 630", "", false},
	{"resize A to 08:15", resize("A", 0, 480, 495, ivA), "MOVE A 0 480 495 0 480 630", "", false},
	{"resize A to 22:00", resize("A", 0, 480, 1320, ivA), "MOVE A 0 480 1320 0 480 630", "", false},
	{"resize A to 08:00", resize("A", 0, 480, 480, ivA), "", "resize A to 08:15", false},
	{"resize B (fixed)", resize("B", 1, 540, 570, ivB), "", "resize A to 11:00", false},
	{"resize A with another day", resize("A", 1, 480, 660, ivA), "", "resize A to 11:00", false},
	{"resize A with another from", resize("A", 0, 495, 660, ivA), "", "resize A to 11:00", false},
	{"resize A to 11:05", resize("A", 0, 480, 665, ivA), "", "resize A to 11:00", false},
	{"resize A to 22:15", resize("A", 0, 480, 1335, ivA), "", "resize A to 22:00", false},
	{"resize A with base to 11:30", resize("A", 0, 480, 660, interval{0, 480, 690}), "", "resize A to 11:00", false},

	{"move C 08:07-08:52 by 15 minutes", move("C", 0, 502, 547, ivC), "MOVE C 0 502 547 0 487 532", "", false},
	{"move C by 10 minutes", move("C", 0, 497, 542, ivC), "", "move C 08:07-08:52 by 15 minutes", false},
	{"move A by 60 minutes", move("A", 0, 540, 690, ivA), "MOVE A 0 540 690 0 480 630", "", false},
	{"move D beyond the range", move("D", 0, 360, 480, ivD), "", "move A by 60 minutes", false},
	{"resize D to 07:30", resize("D", 0, 300, 450, ivD), "", "resize A to 11:00", false},
	{"click D", click("D", ivD), "CLICK D", "", false},

	{"broken json", action{raw: `{"seq":1,"kind":"click","id":"A"`}, "", "click A", true},
	{"unknown kind", action{Kind: "delete", ID: "A", Day: 0, From: 480, To: 630, Base: &interval{0, 480, 630}}, "", "click A", false},
}

// check sends a to the calendar of the window and checks the callback and the acknowledgement.
func check(t *testing.T, w *nagotest.Window, m *model, cal *proto.WeekCalendar, r row) bool {
	t.Helper()
	before := len(m.recorded())
	n := send(w, cal, r.a)
	calls := m.recorded()[before:]
	ok := true
	if r.calls != "" {
		if len(calls) != 1 || calls[0] != r.calls {
			t.Errorf("%s (%s): callbacks %q, expected %q", r.name, r.a.text(), calls, r.calls)
			ok = false
		}
	} else if len(calls) != 0 {
		t.Errorf("%s (%s): callbacks %q, expected none", r.name, r.a.text(), calls)
		ok = false
	}
	if cs := calendars(w); len(cs) == 1 && !r.noAck && int64(cs[0].Ack) != n {
		t.Errorf("%s: ack %d in the next tree, expected %d", r.name, cs[0].Ack, n)
	}
	return ok
}

// Each row of the table invokes its callback or none; each rejection is acknowledged and has a row, which passes.
func TestWeekCalendarChecksEachAction(t *testing.T) {
	byName := map[string]row{}
	for _, r := range rows {
		byName[r.name] = r
	}
	for _, r := range rows {
		if p, ok := byName[r.passes]; r.calls == "" && (!ok || p.calls == "") {
			t.Errorf("instrument: the rejection %q names no passing row (%q)", r.name, r.passes)
		}
	}

	m := &model{}
	w := openCalendar(t, m)
	passed := map[string]bool{}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) { passed[r.name] = check(t, w, m, calendarOf(t, w), r) })
	}
	for _, r := range rows {
		if r.calls == "" && !passed[r.passes] {
			t.Errorf("premise: %s: the row %q did not pass, so the rejection proves nothing", r.name, r.passes)
		}
	}
}

// Without a callback, its kind invokes nothing, also from the tree before; with it, the same action passes.
func TestWeekCalendarWithoutCallback(t *testing.T) {
	m := &model{}
	w := openCalendar(t, m)
	for _, c := range []struct {
		toggle string
		r      row
	}{
		{"NO SELECT", row{"select", sel(0, 540, 600), "SELECT 0 540 600", "", false}},
		{"NO MOVE", row{"move", move("A", 1, 555, 705, ivA), "MOVE A 1 555 705 0 480 630", "", false}},
		{"NO CLICK", row{"click", click("A", ivA), "CLICK A", "", false}},
	} {
		t.Run(c.toggle, func(t *testing.T) {
			old := calendarOf(t, w)
			press(w, c.toggle)
			without := c.r
			without.calls = ""
			check(t, w, m, calendarOf(t, w), without)
			check(t, w, m, old, without)
			press(w, c.toggle)
			if !check(t, w, m, calendarOf(t, w), c.r) {
				t.Errorf("counter row: the same action with the callback does not pass")
			}
		})
	}
}

// A disabled calendar invokes nothing, also for an action from the tree before it was disabled; enabled again, the
// same actions pass.
func TestWeekCalendarDisabled(t *testing.T) {
	m := &model{}
	w := openCalendar(t, m)
	all := []row{
		{"select", sel(0, 540, 600), "SELECT 0 540 600", "", false},
		{"click", click("A", ivA), "CLICK A", "", false},
		{"move", move("A", 1, 555, 705, ivA), "MOVE A 1 555 705 0 480 630", "", false},
		{"resize", resize("A", 0, 480, 660, ivA), "MOVE A 0 480 660 0 480 630", "", false},
	}
	old := calendarOf(t, w)
	press(w, "DISABLED")
	for _, r := range all {
		r.calls = ""
		check(t, w, m, calendarOf(t, w), r)
		check(t, w, m, old, r)
	}
	press(w, "DISABLED")
	for _, r := range all {
		if !check(t, w, m, calendarOf(t, w), r) {
			t.Errorf("counter row: %s without disabled does not pass", r.name)
		}
	}
}

// An action from the tree before a render, which keeps A, passes; a button of the same tree is stale (counter row).
func TestWeekCalendarActionOfFormerTree(t *testing.T) {
	m := &model{}
	w := openCalendar(t, m)
	old := calendarOf(t, w)
	button := callbackOf(w.Find(nagotest.Text("RENDER")).Node())
	if button == 0 {
		t.Fatal("premise: RENDER without a callback")
	}
	renders := func() int {
		n := 0
		for _, e := range w.Events() {
			if _, ok := e.(*proto.RootViewInvalidated); ok {
				n++
			}
		}
		return n
	}
	before := renders()
	press(w, "RENDER")
	if renders() == before {
		t.Fatalf("premise: RENDER did not render again")
	}
	for _, r := range []row{
		{"select from the former tree", sel(0, 540, 600), "SELECT 0 540 600", "", false},
		{"move from the former tree", move("A", 1, 555, 705, ivA), "MOVE A 1 555 705 0 480 630", "", false},
	} {
		check(t, w, m, old, r)
	}
	stale := w.Scope().StaleCalls()
	if err := w.Scope().Dispatch(&proto.FunctionCallRequested{Ptr: button, RID: 990001}); err != nil {
		t.Fatal(err)
	}
	w.Settle()
	if w.Scope().StaleCalls() != stale+1 {
		t.Errorf("counter row: the button of the former tree was not stale (%d stale calls)", w.Scope().StaleCalls()-stale)
	}
}

// After a render without A, an action on A invokes nothing, also from the former tree; after a render without the
// calendar, no action invokes anything.
func TestWeekCalendarAfterRemoval(t *testing.T) {
	m := &model{}
	w := openCalendar(t, m)
	old := calendarOf(t, w)
	if !check(t, w, m, old, row{"counter row: click A while shown", click("A", ivA), "CLICK A", "", false}) {
		t.Fatal("premise: click A does not pass")
	}
	press(w, "WITHOUT A")
	for _, a := range []action{click("A", ivA), move("A", 1, 555, 705, ivA), resize("A", 0, 480, 660, ivA)} {
		check(t, w, m, calendarOf(t, w), row{a.Kind + " A after removal", a, "", "", false})
		check(t, w, m, old, row{a.Kind + " A after removal, former tree", a, "", "", false})
	}
	if !check(t, w, m, calendarOf(t, w), row{"counter row: click C", click("C", ivC), "CLICK C", "", false}) {
		t.Errorf("counter row: the calendar without A invokes nothing at all")
	}
	press(w, "HIDDEN")
	if len(calendars(w)) != 0 {
		t.Fatalf("premise: the calendar is still shown")
	}
	for _, a := range []action{sel(0, 540, 600), click("C", ivC), move("C", 0, 502, 547, ivC)} {
		check(t, w, m, old, row{a.Kind + " to a calendar, which is not shown", a, "", "", true})
	}
}

// An error of OnMove or OnSelect is the message of the next tree, also after a render without an action, until the
// next action.
func TestWeekCalendarMessage(t *testing.T) {
	m := &model{}
	w := openCalendar(t, m)
	if !check(t, w, m, calendarOf(t, w), row{"counter row: move without error", move("A", 1, 555, 705, ivA), "MOVE A 1 555 705 0 480 630", "", false}) {
		t.Fatal("premise: move does not pass")
	}
	if msg := string(calendarOf(t, w).Message); msg != "" {
		t.Errorf("counter row: message %q without an error", msg)
	}
	for _, r := range []row{
		{"move", move("A", 1, 555, 705, ivA), "MOVE A 1 555 705 0 480 630", "", false},
		{"select", sel(0, 540, 600), "SELECT 0 540 600", "", false},
	} {
		t.Run(r.name, func(t *testing.T) {
			m.setFail(true)
			check(t, w, m, calendarOf(t, w), r)
			if msg := strings.TrimSpace(string(calendarOf(t, w).Message)); msg != "belegt" {
				t.Errorf("message %q, expected belegt", msg)
			}
			press(w, "RENDER")
			if msg := strings.TrimSpace(string(calendarOf(t, w).Message)); msg != "belegt" {
				t.Errorf("after a render without an action: message %q, expected belegt until the next action", msg)
			}
			m.setFail(false)
			check(t, w, m, calendarOf(t, w), row{"next action", click("A", ivA), "CLICK A", "", false})
			if msg := string(calendarOf(t, w).Message); msg != "" {
				t.Errorf("after the next action: message %q, expected none", msg)
			}
		})
	}
}

// Validate names each violation of the rules and accepts the row next to it.
func TestWeekCalendarValidate(t *testing.T) {
	day := weekcal.Day{Title: "Montag"}
	days := func(n int) []weekcal.Day {
		var r []weekcal.Day
		for range n {
			r = append(r, day)
		}
		return r
	}
	ab := []weekcal.Event{{ID: "A", Day: 0, From: 480, To: 630}, {ID: "B", Day: 1, From: 540, To: 555, Fixed: true}}
	with := func(e ...weekcal.Event) []weekcal.Event { return append(append([]weekcal.Event(nil), ab...), e...) }
	many := func(n int) []weekcal.Event {
		var r []weekcal.Event
		for i := range n {
			from := (i * 13) % 1380
			r = append(r, weekcal.Event{ID: fmt.Sprintf("e%d", i), Day: i % 2, From: from, To: from + 60})
		}
		return r
	}
	cal := func(n int, ev []weekcal.Event) weekcal.TWeekCalendar {
		return weekcal.WeekCalendar("cal", days(n)...).Append(ev...)
	}
	two := func() weekcal.TWeekCalendar { return cal(2, ab) }
	for _, c := range []struct {
		name  string
		cal   weekcal.TWeekCalendar
		valid bool
	}{
		{"0 days", cal(0, nil), false}, {"1 day", cal(1, nil), true},
		{"8 days", cal(8, ab), false}, {"7 days", cal(7, ab), true},
		{"day 2 of two", cal(2, with(weekcal.Event{ID: "E", Day: 2, From: 600, To: 660})), false},
		{"day -1", cal(2, with(weekcal.Event{ID: "E", Day: -1, From: 600, To: 660})), false},
		{"day 1 of two", cal(2, with(weekcal.Event{ID: "E", Day: 1, From: 600, To: 660})), true},
		{"from = to", cal(2, with(weekcal.Event{ID: "E", From: 600, To: 600})), false},
		{"from < to", cal(2, with(weekcal.Event{ID: "E", From: 600, To: 615})), true},
		{"from -1", cal(2, with(weekcal.Event{ID: "E", From: -1, To: 60})), false},
		{"from 0", cal(2, with(weekcal.Event{ID: "E", From: 0, To: 60})), true},
		{"to 1441", cal(2, with(weekcal.Event{ID: "E", From: 1380, To: 1441})), false},
		{"to 1440", cal(2, with(weekcal.Event{ID: "E", From: 1380, To: 1440})), true},
		{"empty id", cal(2, with(weekcal.Event{ID: "", From: 600, To: 660})), false},
		{"id X", cal(2, with(weekcal.Event{ID: "X", From: 600, To: 660})), true},
		{"duplicate id", cal(2, with(weekcal.Event{ID: "A", Day: 1, From: 600, To: 660})), false},
		{"id A2", cal(2, with(weekcal.Event{ID: "A2", Day: 1, From: 600, To: 660})), true},
		{"range 1320-360", two().Range(1320, 360), false},
		{"range -15-1320", two().Range(-15, 1320), false},
		{"range 365-1320", two().Range(365, 1320), false},
		{"range 0-1440", two().Range(0, 1440), true},
		// the result depends only on the calendar, not on the order of the modifiers
		{"range 365-1320, then step 5", two().Range(365, 1320).Step(5), true},
		{"step 5, then range 365-1320", two().Step(5).Range(365, 1320), true},
		{"step 5, range 365-1320, then step 15", two().Step(5).Range(365, 1320).Step(15), false},
		{"click span 50, then step 10", two().ClickSpan(50).Step(10), true},
		{"step 10, then click span 50", two().Step(10).ClickSpan(50), true},
		{"click span 50 at step 15", two().ClickSpan(50), false},
		{"step 10, click span 50, then step 15", two().Step(10).ClickSpan(50).Step(15), false},
		{"step 7", two().Step(7), false}, {"step 0", two().Step(0), false}, {"step 5", two().Step(5), true},
		{"click span 45 at step 30", two().Step(30).ClickSpan(45), false},
		{"click span 0", two().Step(30).ClickSpan(0), false},
		{"click span 30 at step 30", two().Step(30).ClickSpan(30), true},
		{"hour height 23", two().HourHeight(23), false}, {"hour height 24", two().HourHeight(24), true},
		{"hour height 241", two().HourHeight(241), false}, {"hour height 240", two().HourHeight(240), true},
		{"hour height NaN", two().HourHeight(math.NaN()), false},
		{"hour height +Inf", two().HourHeight(math.Inf(1)), false},
		{"hour height -Inf", two().HourHeight(math.Inf(-1)), false},
		{"now day 2", two().Now(2, 600), false}, {"now day -1", two().Now(-1, 600), false},
		{"now minute -1", two().Now(0, -1), false}, {"now minute 1441", two().Now(0, 1441), false},
		{"now 1, 1440", two().Now(1, 1440), true}, {"now 0, 0", two().Now(0, 0), true},
		{"1000 events", cal(2, many(1000)), true},
	} {
		err := c.cal.Validate()
		if (err == nil) != c.valid {
			t.Errorf("%s: Validate() = %v, expected valid %v", c.name, err, c.valid)
		}
	}
}

// openView opens a window with the view of render.
func openView(t *testing.T, render func(wnd core.Window) core.View) *nagotest.Window {
	t.Helper()
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.weekcaltest")
		cfg.RootView("view", render)
	})
	return app.Open(t, nil, "view", nagotest.StaleRetries(0))
}

// eventIDs are the ids of the events of a calendar in the tree.
func eventIDs(cal *proto.WeekCalendar) []string {
	var ids []string
	for _, e := range cal.Events {
		ids = append(ids, string(e.Id))
	}
	return ids
}

// Append appends, also twice and in a loop, and returns a new calendar: the receiver and another branch stay unchanged.
func TestWeekCalendarAppend(t *testing.T) {
	ev := func(id string, day, from int) weekcal.Event {
		return weekcal.Event{ID: id, Day: day, From: from, To: from + 60}
	}
	w := openView(t, func(wnd core.Window) core.View {
		looped := weekcal.WeekCalendar("looped", weekcal.Day{Title: "Montag"})
		for _, e := range []weekcal.Event{ev("E1", 0, 480), ev("E2", 0, 600), ev("E3", 0, 720)} {
			looped = looped.Append(e)
		}
		twice := weekcal.WeekCalendar("twice", weekcal.Day{Title: "Montag"}).Append(ev("E1", 0, 480)).Append(ev("E2", 0, 600))
		base := weekcal.WeekCalendar("branch", weekcal.Day{Title: "Montag"}, weekcal.Day{Title: "Dienstag"}).
			Append(ev("C1", 0, 480)).Append(ev("C2", 0, 600)).Append(ev("C3", 1, 480))
		x := base.Append(ev("D", 1, 600))
		y := base.Append(ev("E", 1, 720))
		return ui.VStack(looped, twice, base, x, y)
	})
	cs := calendars(w)
	if len(cs) != 5 {
		t.Fatalf("premise: %d calendars instead of five", len(cs))
	}
	for i, want := range []string{"E1,E2,E3", "E1,E2", "C1,C2,C3", "C1,C2,C3,D", "C1,C2,C3,E"} {
		if got := strings.Join(eventIDs(cs[i]), ","); got != want {
			t.Errorf("calendar %d (%s): events %s, expected %s", i, cs[i].Id, got, want)
		}
	}
}

// Per state, nago discards a number, which is not greater than the acknowledged one, without a callback: number 5
// twice and then 3 call once, and the same calendar twice in the tree calls once for one action.
func TestWeekCalendarSequenceNumbers(t *testing.T) {
	m := &model{}
	w := openCalendar(t, m)
	sendAs := func(cal *proto.WeekCalendar, a action, n int64) {
		a.Seq = n
		w.WeekCalendarSend(cal, a.text())
	}
	calls := func() int { return len(m.recorded()) }
	sendAs(calendarOf(t, w), sel(0, 540, 600), 5)
	if c, ack := calls(), int64(calendarOf(t, w).Ack); c != 1 || ack != 5 {
		t.Errorf("number 5: %d callbacks, ack %d; expected 1 and 5", c, ack)
	}
	sendAs(calendarOf(t, w), sel(1, 540, 600), 5)
	if c, ack := calls(), int64(calendarOf(t, w).Ack); c != 1 || ack != 5 {
		t.Errorf("number 5 again: %d callbacks, ack %d; expected 1 and 5", c, ack)
	}
	sendAs(calendarOf(t, w), sel(0, 600, 660), 3)
	if c, ack := calls(), int64(calendarOf(t, w).Ack); c != 1 || ack != 5 {
		t.Errorf("number 3 afterward: %d callbacks, ack %d; expected 1 and 5", c, ack)
	}
	sendAs(calendarOf(t, w), sel(0, 600, 660), 6)
	if c, ack := calls(), int64(calendarOf(t, w).Ack); c != 2 || ack != 6 {
		t.Errorf("counter row number 6: %d callbacks, ack %d; expected 2 and 6", c, ack)
	}
}

func TestWeekCalendarTwiceInTree(t *testing.T) {
	m := &model{}
	w := openView(t, func(wnd core.Window) core.View {
		cal := weekcal.WeekCalendar("cal", weekcal.Day{Title: "Montag"}).OnSelect(func(s weekcal.Span) error {
			m.record(fmt.Sprintf("SELECT %d %d %d", s.Day, s.From, s.To))
			return nil
		})
		return ui.VStack(cal, cal)
	})
	cs := calendars(w)
	if len(cs) != 2 {
		t.Fatalf("premise: %d calendars instead of two", len(cs))
	}
	a := sel(0, 540, 600)
	a.Seq = seq.Add(1)
	w.WeekCalendarSend(cs[0], a.text())
	w.WeekCalendarSend(calendars(w)[1], a.text())
	if got := m.recorded(); len(got) != 1 {
		t.Errorf("one action from both occurrences: callbacks %q, expected exactly one", got)
	}
	send(w, calendars(w)[1], sel(0, 600, 660))
	if got := m.recorded(); len(got) != 2 {
		t.Errorf("counter row: a new number from the second occurrence: callbacks %q, expected two", got)
	}
}
