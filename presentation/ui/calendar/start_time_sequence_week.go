// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"maps"
	"slices"
	"strconv"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/localization/rstring"
)

type startTimeClusterWeek struct {
	week   isoCalWeek
	events []Event
	first  bool
	last   bool
}

func (s startTimeClusterWeek) String(bnd i18n.Bundler) string {
	return rstring.LabelWeek.Get(bnd) + " " + strconv.Itoa(s.week.week)
}

func (s startTimeClusterWeek) Events() []Event {
	return s.events
}

func (s startTimeClusterWeek) TimeStyle() seqPillTimeHint {
	return pillTimeYYYYMMDD
}

type isoCalWeek struct {
	year int
	week int
}

func (i isoCalWeek) String() string {
	return strconv.Itoa(i.week)
}

func mapStartTimeClusterCalWeeks(vp ViewPort, events []Event) []startTimeClusterWeek {
	tmp := map[isoCalWeek]startTimeClusterWeek{}
	for _, event := range events {
		// ignore, if out of range in the future
		if event.From.At.After(vp.To) {
			continue
		}

		// ignore, if out of range in the past
		if event.To.At.Before(vp.From) {
			continue
		}

		y, w := event.From.At.ISOWeek()
		iso := isoCalWeek{y, w}
		l := tmp[iso]
		l.week = iso
		l.events = append(l.events, event)
		slices.SortFunc(l.events, func(a, b Event) int {
			return a.From.At.Compare(b.From.At)
		})
		tmp[iso] = l
	}

	res := make([]startTimeClusterWeek, 0, len(tmp))
	for _, k := range slices.SortedFunc(maps.Keys(tmp), func(week isoCalWeek, week2 isoCalWeek) int {
		if week.year != week2.year {
			return week.year - week2.year
		}

		return week.week - week2.week
	}) {
		res = append(res, tmp[k])
	}

	if len(res) > 0 {
		res[0].first = true
		res[len(res)-1].last = true
	}

	return res
}
