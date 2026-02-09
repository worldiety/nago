// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rstring

import (
	"math"
	"time"

	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	strRelativeDurationSecondsPast = i18n.MustQuantityString(
		"nago.common.duration.past.seconds",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "{x} second ago",
				Other: "{x} seconds ago",
			},
			language.German: i18n.Quantities{
				One:   "Vor einer Sekunde",
				Other: "Vor {x} Sekunden",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is less than one minute."),
	)
	strRelativeDurationSecondsFuture = i18n.MustQuantityString(
		"nago.common.duration.future.seconds",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "In {x} second",
				Other: "In {x} seconds",
			},
			language.German: i18n.Quantities{
				One:   "In einer Sekunde",
				Other: "In {x} Sekunden",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is less than one minute."),
	)
	strRelativeDurationMinutesPast = i18n.MustQuantityString(
		"nago.common.duration.past.minutes",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "{x} minute ago",
				Other: "{x} minutes ago",
			},
			language.German: i18n.Quantities{
				One:   "Vor einer Minute",
				Other: "Vor {x} Minuten",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between one and 59 minutes."),
	)
	strRelativeDurationMinutesFuture = i18n.MustQuantityString(
		"nago.common.duration.future.minutes",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "In {x} minute",
				Other: "In {x} minutes",
			},
			language.German: i18n.Quantities{
				One:   "In einer Minute",
				Other: "In {x} Minuten",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between one and 59 minutes."),
	)
	relativeDurationHoursPast = i18n.MustQuantityString(
		"nago.common.duration.past.hours",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "{x} hour ago",
				Other: "{x} hours ago",
			},
			language.German: i18n.Quantities{
				One:   "Vor einer Stunde",
				Other: "Vor {x} Stunden",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between one and 23 hours."),
	)
	relativeDurationHoursFuture = i18n.MustQuantityString(
		"nago.common.duration.future.hours",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "In {x} hour",
				Other: "In {x} hours",
			},
			language.German: i18n.Quantities{
				One:   "In einer Stunde",
				Other: "In {x} Stunden",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between one and 23 hours."),
	)
	strRelativeDurationDaysPast = i18n.MustQuantityString(
		"nago.common.duration.past.days",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "{x} day ago",
				Other: "{x} days ago",
			},
			language.German: i18n.Quantities{
				One:   "Vor einem Tag",
				Other: "Vor {x} Tagen",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between one and 7 days."),
	)
	strRelativeDurationDaysFuture = i18n.MustQuantityString(
		"nago.common.duration.future.days",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "In {x} day",
				Other: "In {x} days",
			},
			language.German: i18n.Quantities{
				One:   "In einem Tag",
				Other: "In {x} Tagen",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between one and 7 days."),
	)
	strRelativeDurationWeeksPast = i18n.MustQuantityString(
		"nago.common.duration.past.weeks",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "{x} week ago",
				Other: "{x} weeks ago",
			},
			language.German: i18n.Quantities{
				One:   "Vor einer Woche",
				Other: "Vor {x} Wochen",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between 7 and 28 days."),
	)
	strRelativeDurationWeeksFuture = i18n.MustQuantityString(
		"nago.common.duration.future.weeks",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "In {x} week",
				Other: "In {x} weeks",
			},
			language.German: i18n.Quantities{
				One:   "In einer Woche",
				Other: "In {x} Wochen",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between 7 and 28 days."),
	)
	strRelativeDurationMonthsPast = i18n.MustQuantityString(
		"nago.common.duration.past.months",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "{x} month ago",
				Other: "{x} months ago",
			},
			language.German: i18n.Quantities{
				One:   "Vor einem Monat",
				Other: "Vor {x} Monaten",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between one and 11 months."),
	)
	strRelativeDurationMonthsFuture = i18n.MustQuantityString(
		"nago.common.duration.future.months",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "In {x} month",
				Other: "In {x} months",
			},
			language.German: i18n.Quantities{
				One:   "In einem Monat",
				Other: "In {x} Monaten",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is between one and 11 months."),
	)
	strRelativeDurationYearsPast = i18n.MustQuantityString(
		"nago.common.duration.past.years",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "{x} year ago",
				Other: "{x} years ago",
			},
			language.German: i18n.Quantities{
				One:   "Vor einem Jahr",
				Other: "Vor {x} Jahren",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is one or more years."),
	)
	strRelativeDurationYearsFuture = i18n.MustQuantityString(
		"nago.common.duration.future.years",
		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "In {x} year",
				Other: "In {x} years",
			},
			language.German: i18n.Quantities{
				One:   "In einem Jahr",
				Other: "In {x} Jahren",
			},
		},
		i18n.LocalizationHint("Displayed when the duration is one or more years."),
	)
	strRelativeDurationAgesPast = i18n.MustString(
		"nago.common.duration.past.ages",
		i18n.Values{
			language.English: "Ages ago",
			language.German:  "Vor langer Zeit",
		},
		i18n.LocalizationHint("Displayed when the duration is more than 50 years."),
	)
	strRelativeDurationAgesFuture = i18n.MustString(
		"nago.common.duration.future.ages",
		i18n.Values{
			language.English: "In the distant future",
			language.German:  "In weiter Zukunft",
		},
		i18n.LocalizationHint("Displayed when the duration is more than 50 years."),
	)
	strRelativeDurationNow = i18n.MustString(
		"nago.common.duration.now",
		i18n.Values{
			language.English: "Now",
			language.German:  "Jetzt",
		},
		i18n.LocalizationHint("Displayed when the duration is exactly zero."),
	)
)

// RelativeTimeFromNow formats the duration between now and the given time in the locale of the user
// using relative wording with decreasing accuracy similar to the 'RelativeTimeFormat' found in JavaScript.
// The function considers a negative duration to be in the past and a positive duration to be in the future
// and returns the appropriate localized string accordingly. Passing a duration of 0 will return "now".
//
// For example: Passing a duration of 00:00:13 will print "13 seconds ago" and passing 13:32:27 will print "13 hours ago".
func RelativeTimeFromNow(bundler i18n.Bundler, t time.Time) string {
	return RelativeTimeOfDuration(bundler, t.Sub(time.Now()))
}

// RelativeTimeOfDuration formats a given duration in the locale of the user using relative wording
// with decreasing accuracy similar to the 'RelativeTimeFormat' found in JavaScript.
// The function considers a negative duration to be in the past and a positive duration to be in the future
// and returns the appropriate localized string accordingly. Passing a duration of 0 will return "now".
//
// For example: Passing a duration of 00:00:13 will print "13 seconds ago" and passing 13:32:27 will print "13 hours ago".
func RelativeTimeOfDuration(bundler i18n.Bundler, d time.Duration) string {
	if d < 0 {
		return relativeTimeOfDurationInPast(bundler, d.Abs())
	} else if d > 0 {
		return relativeTimeOfDurationInFuture(bundler, d)
	}

	return strRelativeDurationNow.Get(bundler)
}

func relativeTimeOfDurationInFuture(bundler i18n.Bundler, d time.Duration) string {
	if d < time.Minute {
		value := math.Floor(d.Seconds())
		return strRelativeDurationSecondsFuture.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour {
		value := math.Floor(d.Minutes())
		return strRelativeDurationMinutesFuture.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24 {
		value := math.Floor(d.Hours())
		return relativeDurationHoursFuture.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24*7 {
		value := math.Floor(d.Hours() / 24)
		return strRelativeDurationDaysFuture.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24*28 {
		value := math.Floor(d.Hours() / 24 / 7)
		return strRelativeDurationWeeksFuture.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24*365 {
		value := math.Floor(d.Hours() / 24 / 28)
		return strRelativeDurationMonthsFuture.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24*365*50 {
		value := math.Floor(d.Hours() / 24 / 365)
		return strRelativeDurationYearsFuture.Get(bundler, value, i18n.Int("x", int(value)))
	}

	return strRelativeDurationAgesFuture.Get(bundler)
}

func relativeTimeOfDurationInPast(bundler i18n.Bundler, d time.Duration) string {
	if d < time.Minute {
		value := math.Floor(d.Seconds())
		return strRelativeDurationSecondsPast.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour {
		value := math.Floor(d.Minutes())
		return strRelativeDurationMinutesPast.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24 {
		value := math.Floor(d.Hours())
		return relativeDurationHoursPast.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24*7 {
		value := math.Floor(d.Hours() / 24)
		return strRelativeDurationDaysPast.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24*28 {
		value := math.Floor(d.Hours() / 24 / 7)
		return strRelativeDurationWeeksPast.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24*365 {
		value := math.Floor(d.Hours() / 24 / 28)
		return strRelativeDurationMonthsPast.Get(bundler, value, i18n.Int("x", int(value)))
	} else if d < time.Hour*24*365*50 {
		value := math.Floor(d.Hours() / 24 / 365)
		return strRelativeDurationYearsPast.Get(bundler, value, i18n.Int("x", int(value)))
	}

	return strRelativeDurationAgesPast.Get(bundler)
}
