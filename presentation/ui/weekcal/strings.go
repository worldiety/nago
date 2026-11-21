// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package weekcal

import (
	"strconv"
	"strings"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

// The sentences of a violation of [TWeekCalendar.Validate] for the user of an application: each is a whole sentence
// with the numbers, but without the id of the calendar or of an event, because the user cannot change the calendar,
// only report it.
var (
	StrInvalid = core.DefaultStr("nago.weekcal.invalid", "The week calendar is invalid.", "Der Wochenkalender ist ungültig.")

	StrInvalidDays = i18n.MustVarString("nago.weekcal.invalid_days", i18n.Values{
		language.German:  "Ein Wochenkalender hat 1 bis {max} Tage, nicht {days}.",
		language.English: "A week calendar has 1 to {max} days, not {days}.",
	})

	StrInvalidStep = i18n.MustVarString("nago.weekcal.invalid_step", i18n.Values{
		language.German:  "Das Raster eines Wochenkalenders beträgt 5, 10, 15, 20, 30 oder 60 Minuten, nicht {step}.",
		language.English: "The step of a week calendar is 5, 10, 15, 20, 30 or 60 minutes, not {step}.",
	})

	StrInvalidRange = i18n.MustVarString("nago.weekcal.invalid_range", i18n.Values{
		language.German:  "Die sichtbare Zeit eines Wochenkalenders liegt zwischen Minute 0 und 1440 und beginnt vor ihrem Ende, nicht von {from} bis {to}.",
		language.English: "The visible time of a week calendar lies between minute 0 and 1440 and starts before its end, not from {from} to {to}.",
	})

	StrInvalidRaster = i18n.MustVarString("nago.weekcal.invalid_raster", i18n.Values{
		language.German:  "Die sichtbare Zeit von Minute {from} bis {to} liegt nicht im Raster von {step} Minuten.",
		language.English: "The visible time from minute {from} to {to} is not in the step of {step} minutes.",
	})

	StrInvalidClickSpan = i18n.MustVarString("nago.weekcal.invalid_click_span", i18n.Values{
		language.German:  "Die Dauer eines Klicks ist ein positives Vielfaches des Rasters von {step} Minuten, nicht {span}.",
		language.English: "The duration of a click is a positive multiple of the step of {step} minutes, not {span}.",
	})

	StrInvalidHourHeight = i18n.MustVarString("nago.weekcal.invalid_hour_height", i18n.Values{
		language.German:  "Eine Stunde ist 24 bis 240 dp hoch, nicht {height}.",
		language.English: "An hour is 24 to 240 dp high, not {height}.",
	})

	StrInvalidNowDay = i18n.MustVarString("nago.weekcal.invalid_now_day", i18n.Values{
		language.German:  "Die aktuelle Zeit liegt an keinem der {days} Tage des Wochenkalenders.",
		language.English: "The current time lies on none of the {days} days of the week calendar.",
	})

	StrInvalidNowMinute = i18n.MustVarString("nago.weekcal.invalid_now_minute", i18n.Values{
		language.German:  "Die aktuelle Zeit liegt bei Minute {minute}, nicht zwischen 0 und 1440.",
		language.English: "The current time lies at minute {minute}, not between 0 and 1440.",
	})

	StrInvalidEvents = i18n.MustVarString("nago.weekcal.invalid_events", i18n.Values{
		language.German:  "Ein Wochenkalender hat höchstens {max} Termine, nicht {events}.",
		language.English: "A week calendar has at most {max} events, not {events}.",
	})

	StrInvalidEventID = i18n.MustVarString("nago.weekcal.invalid_event_id", i18n.Values{
		language.German:  "Ein Termin des Wochenkalenders hat keine Kennung.",
		language.English: "An event of the week calendar has no id.",
	})

	StrInvalidEventIDTwice = i18n.MustVarString("nago.weekcal.invalid_event_id_twice", i18n.Values{
		language.German:  "Zwei Termine des Wochenkalenders haben dieselbe Kennung.",
		language.English: "Two events of the week calendar have the same id.",
	})

	StrInvalidEventDay = i18n.MustVarString("nago.weekcal.invalid_event_day", i18n.Values{
		language.German:  "Ein Termin liegt an keinem der {days} Tage des Wochenkalenders.",
		language.English: "An event lies on none of the {days} days of the week calendar.",
	})

	StrInvalidEventTime = i18n.MustVarString("nago.weekcal.invalid_event_time", i18n.Values{
		language.German:  "Ein Termin von Minute {from} bis {to} liegt nicht zwischen 0 und 1440 oder beginnt nicht vor seinem Ende.",
		language.English: "An event from minute {from} to {to} does not lie between 0 and 1440 or does not start before its end.",
	})
)

// violation is a rule, which a calendar breaks: the reason in English for the developer and the sentence for the user.
type violation struct {
	reason error
	text   i18n.VarStrHnd
	attrs  []i18n.Attr
	// height is the hour height of [StrInvalidHourHeight], written as a decimal number of the language
	height *float64
}

func violate(reason error, text i18n.VarStrHnd, attrs ...i18n.Attr) *violation {
	return &violation{reason: reason, text: text, attrs: attrs}
}

// localized returns the violation with the sentences of the bundle, English without one, and the reason as cause.
func (v *violation) localized(b *i18n.Bundle) std.LocalizedError {
	if b == nil {
		b, _ = i18n.Default.MatchBundle(language.English)
	}
	if b == nil {
		return std.NewLocalizedError(v.reason.Error(), v.reason.Error()).WithError(v.reason)
	}

	attrs := v.attrs
	if v.height != nil {
		attrs = append(attrs[:len(attrs):len(attrs)], i18n.String("height", decimal(b, *v.height)))
	}

	return std.NewLocalizedError(StrInvalid.Get(b), v.text.Get(b, attrs...)).WithError(v.reason)
}

// decimal writes f like 23.5, with a decimal comma in German.
func decimal(b *i18n.Bundle, f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if base, _ := b.Tag().Base(); base.String() == "de" {
		s = strings.Replace(s, ".", ",", 1)
	}

	return s
}
