// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/weekcal"
	"go.wdy.de/nago/web/vuejs"
)

// bookings are kept by the application, the calendar only shows them and reports what the user does.
type bookings struct {
	mutex  sync.Mutex
	events []weekcal.Event
	next   int
}

func (b *bookings) list() []weekcal.Event {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return slices.Clone(b.events)
}

func (b *bookings) add(s weekcal.Span) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if b.overlapsFixed(s.Day, s.From, s.To, "") {
		return errors.New("Die Zeit ist belegt.")
	}
	b.next++
	b.events = append(b.events, weekcal.Event{ID: fmt.Sprintf("n%d", b.next), Day: s.Day, From: s.From, To: s.To, Title: "Neue Buchung", Color: "#2e7d32"})
	return nil
}

func (b *bookings) move(m weekcal.Move) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	i := slices.IndexFunc(b.events, func(e weekcal.Event) bool { return e.ID == m.ID })
	if i < 0 {
		return errors.New("Die Buchung gibt es nicht mehr.")
	}
	e := b.events[i]
	if e.Day != m.OldDay || e.From != m.OldFrom || e.To != m.OldTo {
		return errors.New("Die Buchung wurde inzwischen geändert.")
	}
	if b.overlapsFixed(m.Day, m.From, m.To, m.ID) {
		return errors.New("belegt")
	}
	e.Day, e.From, e.To = m.Day, m.From, m.To
	b.events[i] = e
	return nil
}

func (b *bookings) overlapsFixed(day, from, to int, except string) bool {
	for _, e := range b.events {
		if e.Fixed && e.ID != except && e.Day == day && e.From < to && from < e.To {
			return true
		}
	}
	return false
}

func main() {
	data := &bookings{events: []weekcal.Event{
		{ID: "a", Day: 0, From: 480, To: 630, Title: "Projekt A", Text: "Konzept"},
		{ID: "b", Day: 0, From: 540, To: 660, Title: "Projekt B", Color: "#c62828"},
		{ID: "c", Day: 0, From: 690, To: 750, Title: "Projekt C"},
		{ID: "d", Day: 1, From: 300, To: 420, Title: "Früh", Text: "ragt über den Beginn"},
		{ID: "e", Day: 1, From: 120, To: 180, Title: "Nacht"},
		{ID: "f", Day: 1, From: 60, To: 90, Title: "Nacht 2"},
		{ID: "g", Day: 1, From: 1350, To: 1400, Title: "Spät"},
		{ID: "h", Day: 2, From: 720, To: 780, Title: "Mittag", Fixed: true},
		{ID: "i", Day: 3, From: 487, To: 532, Title: "Außerhalb des Rasters"},
	}}

	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial.weekcal")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			clicked := core.AutoState[string](wnd)
			disabled := core.AutoState[bool](wnd)
			days := []weekcal.Day{
				{Title: "Montag", Subtitle: "3. Nov.", Today: true},
				{Title: "Dienstag", Subtitle: "4. Nov.\nUrlaub (halber Tag)"},
				{Title: "Mittwoch", Subtitle: "5. Nov."},
				{Title: "Donnerstag", Subtitle: "6. Nov."},
				{Title: "Freitag", Subtitle: "7. Nov."},
				{Title: "Samstag", Subtitle: "8. Nov.", Muted: true},
				{Title: "Sonntag", Subtitle: "9. Nov.", Muted: true},
			}

			// a tick renders the window again every second, like a refresh, without changing the tree before the
			// calendar: a drag goes on
			tick := core.StateOf[int](wnd, "tick")
			core.StateOf[bool](wnd, "ticker").Init(func() bool {
				go func() {
					for wnd.Context().Err() == nil {
						time.Sleep(time.Second)
						wnd.Post(func() {
							tick.Set(tick.Get() + 1)
							tick.Invalidate()
						})
					}
				}()
				return true
			})
			dialog := core.AutoState[bool](wnd)
			spanned := core.AutoState[string](wnd)
			small := func(id string) core.View {
				return weekcal.WeekCalendar(id, weekcal.Day{Title: "Heute"}).
					Range(480, 720).
					OnSelect(func(s weekcal.Span) error {
						spanned.Set(fmt.Sprintf("%s %d %d-%d", id, s.Day, s.From, s.To))
						return nil
					})
			}

			return VStack(
				alert.Dialog("Kalender im Dialog", small("dialog"), dialog, alert.Closeable()),
				Text("Wochenkalender").Font(Title),
				Text("Aufziehen in freier Zeit legt an, Ziehen eines Blocks verschiebt ihn, die Unterkante ändert die Dauer, ein Klick wählt aus. Mittwoch 12:00 ist fest; wer darauf schiebt, bekommt „belegt“."),
				HStack(
					Text("Zuletzt geklickt: "+clicked.Get()),
					Toggle(disabled.Get()).InputChecked(disabled),
					Text("deaktiviert"),
				).Gap(L8),
				weekcal.WeekCalendar("week", days...).
					Append(data.list()...).
					Now(0, 600).
					Disabled(disabled.Get()).
					OnClick(func(id string) { clicked.Set(id) }).
					OnSelect(data.add).
					OnMove(data.move),
				Text("Zuletzt aufgezogen: "+spanned.Get()),
				Text(fmt.Sprintf("Takt %d", tick.Get())),
				HStack(
					DnDArea(VStack(Text("Ziehbare Kachel"), small("dragged")).Frame(Frame{Width: L320})).ID("tile").CanDrag(true),
					SecondaryButton(func() { dialog.Set(true) }).Title("Dialog öffnen"),
				).Gap(L16).Alignment(Top),
			).Gap(L16).FullWidth().Padding(Padding{}.All(L16))
		})
	}).Run()
}
