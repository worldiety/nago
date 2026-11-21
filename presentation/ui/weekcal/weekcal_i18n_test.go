// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package weekcal_test

// The reasons of an invalid week calendar for the user (wokoda N28 part 2, L-N28b-10): each violation of Validate
// reaches OnInvalid as a std.LocalizedError with a sentence in the language of the window, German and English, with
// the numbers, but without the id of the calendar or an event, and with the reason in English as its cause. A valid
// calendar beside each group renders without OnInvalid.

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/weekcal"
)

func TestWeekCalendarInvalidLocalized(t *testing.T) {
	day := weekcal.Day{Title: "Montag"}
	days := func(n int) []weekcal.Day {
		var r []weekcal.Day
		for range n {
			r = append(r, day)
		}
		return r
	}
	ab := []weekcal.Event{{ID: "evA", Day: 0, From: 480, To: 630}, {ID: "evB", Day: 1, From: 540, To: 555, Fixed: true}}
	with := func(e ...weekcal.Event) []weekcal.Event { return append(append([]weekcal.Event(nil), ab...), e...) }
	cal := func(n int, ev []weekcal.Event) weekcal.TWeekCalendar {
		return weekcal.WeekCalendar("calX", days(n)...).Append(ev...)
	}
	two := func() weekcal.TWeekCalendar { return cal(2, ab) }

	rows := []struct {
		name   string
		cal    weekcal.TWeekCalendar
		de, en string // "" for a valid calendar
		reason string
	}{
		{"8 days", cal(8, ab), "Ein Wochenkalender hat 1 bis 7 Tage, nicht 8.", "A week calendar has 1 to 7 days, not 8.", "a week calendar has 1 to 7 days, not 8"},
		{"0 days", cal(0, nil), "Ein Wochenkalender hat 1 bis 7 Tage, nicht 0.", "A week calendar has 1 to 7 days, not 0.", "a week calendar has 1 to 7 days, not 0"},
		{"7 days", cal(7, ab), "", "", ""},
		{"step 7", two().Step(7), "Das Raster eines Wochenkalenders beträgt 5, 10, 15, 20, 30 oder 60 Minuten, nicht 7.", "The step of a week calendar is 5, 10, 15, 20, 30 or 60 minutes, not 7.", "the step is one of 5, 10, 15, 20, 30 or 60 minutes, not 7"},
		{"step 5", two().Step(5), "", "", ""},
		{"range 1320-360", two().Range(1320, 360), "Die sichtbare Zeit eines Wochenkalenders liegt zwischen Minute 0 und 1440 und beginnt vor ihrem Ende, nicht von 1320 bis 360.", "The visible time of a week calendar lies between minute 0 and 1440 and starts before its end, not from 1320 to 360.", "the range is 0 <= from < to <= 1440, not 1320 to 360"},
		{"range 0-1440", two().Range(0, 1440), "", "", ""},
		{"range 365-1320", two().Range(365, 1320), "Die sichtbare Zeit von Minute 365 bis 1320 liegt nicht im Raster von 15 Minuten.", "The visible time from minute 365 to 1320 is not in the step of 15 minutes.", "the range 365 to 1320 is not in the raster of 15 minutes"},
		{"range 365-1320, step 5", two().Range(365, 1320).Step(5), "", "", ""},
		{"click span 50", two().ClickSpan(50), "Die Dauer eines Klicks ist ein positives Vielfaches des Rasters von 15 Minuten, nicht 50.", "The duration of a click is a positive multiple of the step of 15 minutes, not 50.", "the click span is a positive multiple of the step of 15 minutes, not 50"},
		{"click span 45", two().ClickSpan(45), "", "", ""},
		{"hour height 23.5", two().HourHeight(23.5), "Eine Stunde ist 24 bis 240 dp hoch, nicht 23,5.", "An hour is 24 to 240 dp high, not 23.5.", "the hour height is 24 to 240 dp, not 23.5"},
		{"hour height NaN", two().HourHeight(math.NaN()), "Eine Stunde ist 24 bis 240 dp hoch, nicht NaN.", "An hour is 24 to 240 dp high, not NaN.", "the hour height is 24 to 240 dp, not NaN"},
		{"hour height 24", two().HourHeight(24), "", "", ""},
		{"now day 2", two().Now(2, 600), "Die aktuelle Zeit liegt an keinem der 2 Tage des Wochenkalenders.", "The current time lies on none of the 2 days of the week calendar.", "the current time is at a day from 0 to 1 and a minute from 0 to 1440, not at day 2 and minute 600"},
		{"now minute 1441", two().Now(0, 1441), "Die aktuelle Zeit liegt bei Minute 1441, nicht zwischen 0 und 1440.", "The current time lies at minute 1441, not between 0 and 1440.", "the current time is at a day from 0 to 1 and a minute from 0 to 1440, not at day 0 and minute 1441"},
		{"now 1, 1440", two().Now(1, 1440), "", "", ""},
		{"3 events, at most 2", two().Append(weekcal.Event{ID: "evC", From: 600, To: 660}).MaxEvents(2), "Ein Wochenkalender hat höchstens 2 Termine, nicht 3.", "A week calendar has at most 2 events, not 3.", "a week calendar has at most 2 events, not 3"},
		{"2 events, at most 2", two().MaxEvents(2), "", "", ""},
		{"empty id", cal(2, with(weekcal.Event{ID: "", From: 600, To: 660})), "Ein Termin des Wochenkalenders hat keine Kennung.", "An event of the week calendar has no id.", "the event at index 2 has an empty id"},
		{"duplicate id", cal(2, with(weekcal.Event{ID: "evA", Day: 1, From: 600, To: 660})), "Zwei Termine des Wochenkalenders haben dieselbe Kennung.", "Two events of the week calendar have the same id.", `the id "evA" of an event is not unique`},
		{"id evC", cal(2, with(weekcal.Event{ID: "evC", Day: 1, From: 600, To: 660})), "", "", ""},
		{"day 2 of two", cal(2, with(weekcal.Event{ID: "evE", Day: 2, From: 600, To: 660})), "Ein Termin liegt an keinem der 2 Tage des Wochenkalenders.", "An event lies on none of the 2 days of the week calendar.", `the event "evE" has the day 2, but the calendar has the days 0 to 1`},
		{"day 1 of two", cal(2, with(weekcal.Event{ID: "evE", Day: 1, From: 600, To: 660})), "", "", ""},
		{"from = to", cal(2, with(weekcal.Event{ID: "evE", From: 600, To: 600})), "Ein Termin von Minute 600 bis 600 liegt nicht zwischen 0 und 1440 oder beginnt nicht vor seinem Ende.", "An event from minute 600 to 600 does not lie between 0 and 1440 or does not start before its end.", `the event "evE" is not 0 <= from < to <= 1440, but 600 to 600`},
		{"to 1441", cal(2, with(weekcal.Event{ID: "evE", From: 1380, To: 1441})), "Ein Termin von Minute 1380 bis 1441 liegt nicht zwischen 0 und 1440 oder beginnt nicht vor seinem Ende.", "An event from minute 1380 to 1441 does not lie between 0 and 1440 or does not start before its end.", `the event "evE" is not 0 <= from < to <= 1440, but 1380 to 1441`},
		{"to 1440", cal(2, with(weekcal.Event{ID: "evE", From: 1380, To: 1440})), "", "", ""},
	}

	// one application with a root view per row, each with OnInvalid, which records the error of its last render
	var mutex sync.Mutex
	got := map[string]error{}
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.weekcaltest")
		for i, r := range rows {
			path := core.NavigationPath(fmt.Sprintf("row%d", i))
			cfg.RootView(path, func(wnd core.Window) core.View {
				return r.cal.OnInvalid(func(err error) core.View {
					mutex.Lock()
					defer mutex.Unlock()
					got[string(path)+"/"+wnd.Bundle().Tag().String()] = err
					return ui.Text("INVALID")
				})
			})
		}
	})

	for i, r := range rows {
		path := fmt.Sprintf("row%d", i)
		for _, lang := range []struct{ locale, want string }{{"de", r.de}, {"en", r.en}} {
			w := app.Open(t, nil, core.NavigationPath(path), nagotest.Locale(lang.locale), nagotest.StaleRetries(0))
			mutex.Lock()
			err, called := got[path+"/"+lang.locale]
			mutex.Unlock()

			if lang.want == "" {
				if called {
					t.Errorf("%s (%s): OnInvalid(%v), expected a valid calendar", r.name, lang.locale, err)
				}
				if n := w.FindAll(nagotest.Type[*proto.WeekCalendar]()).Len(); n != 1 {
					t.Errorf("%s (%s): %d calendars, expected 1", r.name, lang.locale, n)
				}
				w.Close()
				continue
			}

			if !called {
				t.Errorf("%s (%s): OnInvalid not called", r.name, lang.locale)
				w.Close()
				continue
			}

			var localized std.LocalizedError
			if !errors.As(err, &localized) {
				t.Errorf("%s (%s): %T is no std.LocalizedError", r.name, lang.locale, err)
				w.Close()
				continue
			}

			if d := localized.Description(); d != lang.want {
				t.Errorf("%s (%s): description %q, expected %q", r.name, lang.locale, d, lang.want)
			}

			title := map[string]string{"de": "Der Wochenkalender ist ungültig.", "en": "The week calendar is invalid."}[lang.locale]
			if localized.Title() != title {
				t.Errorf("%s (%s): title %q, expected %q", r.name, lang.locale, localized.Title(), title)
			}

			for _, id := range []string{"calX", "evA", "evB", "evC", "evE"} {
				if strings.Contains(localized.Title()+localized.Description(), id) {
					t.Errorf("%s (%s): the sentence names the id %q", r.name, lang.locale, id)
				}
			}

			if cause := errors.Unwrap(localized); cause == nil || cause.Error() != r.reason {
				t.Errorf("%s (%s): cause %v, expected %q", r.name, lang.locale, cause, r.reason)
			}

			if w.FindAll(nagotest.Type[*proto.WeekCalendar]()).Len() != 0 || w.FindAll(nagotest.Text("INVALID")).Len() != 1 {
				t.Errorf("%s (%s): the view of OnInvalid is not shown instead of the calendar", r.name, lang.locale)
			}
			w.Close()
		}

		// Validate without a window: the sentence in English, the same cause
		err := r.cal.Validate()
		if (err == nil) != (r.en == "") {
			t.Errorf("%s: Validate() = %v, expected valid %v", r.name, err, r.en == "")
			continue
		}
		if err != nil {
			var localized std.LocalizedError
			if !errors.As(err, &localized) || localized.Description() != r.en || errors.Unwrap(localized).Error() != r.reason {
				t.Errorf("%s: Validate() = %#v, expected %q with the cause %q", r.name, err, r.en, r.reason)
			}
		}
	}
}
