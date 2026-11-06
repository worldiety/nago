// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"time"

	"go.wdy.de/nago/presentation/proto"
)

// WeekCalendarSend writes the raw text into the input of the week calendar, like the frontend writes an action of
// the user, e.g. {"seq":1,"kind":"select","day":0,"from":540,"to":600}, see [proto.WeekCalendar.Input]. Unlike the
// other actions, it sends exactly once and to any calendar, also to a disabled one or to one of a former tree, so
// that a test can check, that the server rejects what a user cannot do, and does not repeat a stale action. The
// window settles afterward.
func (w *Window) WeekCalendarSend(cal *proto.WeekCalendar, raw string) {
	w.t.Helper()

	start := time.Now()
	w.dispatch(&proto.UpdateStateValueRequested{StatePointer: cal.Input, Value: proto.Str(raw), RID: w.nextRID()})
	w.Settle()
	w.observe("weekcal-send", string(cal.Id), start)
}
