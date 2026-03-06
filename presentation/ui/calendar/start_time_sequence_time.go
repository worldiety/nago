// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/worldiety/i18n"
)

type startTimeClusterTime struct {
	time   isoCalTime
	events []Event
	first  bool
	last   bool
}

func (s startTimeClusterTime) String(bnd i18n.Bundler) string {
	return fmt.Sprintf("%02d:%02d", s.time.hour, s.time.minute)
}

func (s startTimeClusterTime) Events() []Event {
	return s.events
}

func (s startTimeClusterTime) TimeStyle() seqPillTimeHint {
	return pillTimeNone
}

type isoCalTime struct {
	year   int
	week   int
	month  time.Month
	day    int
	hour   int
	minute int
	start  time.Time
}

func (i isoCalTime) String() string {
	return fmt.Sprintf("%02d:%02d", i.hour, i.minute)
}

func mapStartTimeClusterCalTime(vp ViewPort, events []Event) []startTimeClusterTime {
	tmp := map[isoCalTime]startTimeClusterTime{}
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
		iso := isoCalTime{y, w, event.From.At.Month(), event.From.At.Day(), event.From.At.Hour(), event.From.At.Minute(), event.From.At}
		l := tmp[iso]
		l.time = iso
		l.events = append(l.events, event)
		slices.SortFunc(l.events, func(a, b Event) int {
			return a.From.At.Compare(b.From.At)
		})
		tmp[iso] = l
	}

	res := make([]startTimeClusterTime, 0, len(tmp))
	for _, k := range slices.SortedFunc(maps.Keys(tmp), func(a isoCalTime, b isoCalTime) int {
		return a.start.Compare(b.start)
	}) {
		res = append(res, tmp[k])
	}

	if len(res) > 0 {
		res[0].first = true
		res[len(res)-1].last = true
	}

	return res
}
