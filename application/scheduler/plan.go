// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// LastDayOfMonth selects the last day of each month in [Settings.CronMonthDays].
const LastDayOfMonth = -1

// maxPlanSearchDays bounds the search for the next or previous day of a cron plan. Every valid plan has a day
// within about two months, so that a search which reaches the bound is a programming error and fails loudly.
const maxPlanSearchDays = 400

// cronPlan is the wall clock time of a [Cron] scheduler on the selected days in its location.
type cronPlan struct {
	hour      int
	minute    int
	weekdays  []time.Weekday
	monthDays []int
	loc       *time.Location
}

func newCronPlan(settings Settings, loc *time.Location) cronPlan {
	if loc == nil {
		loc = time.Local
	}

	return cronPlan{
		hour:      settings.CronHour,
		minute:    settings.CronMinute,
		weekdays:  settings.CronWeekdays,
		monthDays: settings.CronMonthDays,
		loc:       loc,
	}
}

// matches reports, if the plan is due on the given calendar day.
func (p cronPlan) matches(y int, m time.Month, d int) bool {
	if len(p.weekdays) > 0 {
		return slices.Contains(p.weekdays, time.Date(y, m, d, 12, 0, 0, 0, time.UTC).Weekday())
	}

	if len(p.monthDays) > 0 {
		last := time.Date(y, m+1, 0, 12, 0, 0, 0, time.UTC).Day()
		for _, md := range p.monthDays {
			if md == d || (md == LastDayOfMonth && d == last) {
				return true
			}
		}

		return false
	}

	return true
}

// at returns the instant of the wall clock time of the plan on the given day. A time which does not exist on that
// day is normalized like by [time.Date], e.g. 02:30 to 03:30, when the daylight saving time begins. A time which
// exists twice, when the daylight saving time ends, is the first one, i.e. still in daylight saving time.
func (p cronPlan) at(y int, m time.Month, d int) time.Time {
	t := time.Date(y, m, d, p.hour, p.minute, 0, 0, p.loc)
	return firstInstant(t)
}

// firstInstant returns the earlier instant, if the wall clock time of t exists twice in its location, and t
// otherwise. [time.Date] does not guarantee, which of both it returns.
func firstInstant(t time.Time) time.Time {
	_, off := t.Zone()
	_, offBefore := t.Add(-3 * time.Hour).Zone()
	if offBefore <= off {
		return t
	}

	earlier := t.Add(-time.Duration(offBefore-off) * time.Second)
	if sameWallClock(earlier, t) {
		return earlier
	}

	return t
}

func sameWallClock(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd && a.Hour() == b.Hour() && a.Minute() == b.Minute() && a.Second() == b.Second()
}

// next returns the first due which is not before from.
func (p cronPlan) next(from time.Time) (time.Time, error) {
	y, m, d := from.In(p.loc).Date()
	for i := 0; i <= maxPlanSearchDays; i++ {
		dy, dm, dd := time.Date(y, m, d+i, 12, 0, 0, 0, time.UTC).Date()
		if !p.matches(dy, dm, dd) {
			continue
		}

		if t := p.at(dy, dm, dd); !t.Before(from) {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("no due within %d days after %v", maxPlanSearchDays, from)
}

// prev returns the latest due which is before the given time.
func (p cronPlan) prev(before time.Time) (time.Time, error) {
	y, m, d := before.In(p.loc).Date()
	for i := 0; i <= maxPlanSearchDays; i++ {
		dy, dm, dd := time.Date(y, m, d-i, 12, 0, 0, 0, time.UTC).Date()
		if !p.matches(dy, dm, dd) {
			continue
		}

		if t := p.at(dy, dm, dd); t.Before(before) {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("no due within %d days before %v", maxPlanSearchDays, before)
}

// usesPlan returns true, if the options use a cron plan instead of [nextCronAt], which is kept unchanged for the
// schedulers without the new options.
func usesPlan(opts Options, settings Settings) bool {
	return opts.Durable || opts.Location != nil || len(settings.CronWeekdays) > 0 || len(settings.CronMonthDays) > 0
}

// nextCronDue returns the first due of a cron scheduler which is not before from.
func nextCronDue(from time.Time, opts Options, settings Settings) (time.Time, error) {
	if !usesPlan(opts, settings) {
		return nextCronAt(from, settings), nil
	}

	if settings.PauseTime > 0 {
		// the repetition from the cron time, as before, but in the location of the scheduler
		return nextCronAt(from.In(locationOf(opts)), settings), nil
	}

	return newCronPlan(settings, opts.Location).next(from)
}

func locationOf(opts Options) *time.Location {
	if opts.Location == nil {
		return time.Local
	}

	return opts.Location
}

var errInvalidPlan = errors.New("invalid plan")

// validatePlan checks the new options and settings. The fields which already existed before are not checked, so
// that existing schedulers keep working as before.
func validatePlan(opts Options, settings Settings) error {
	if len(settings.CronWeekdays) > 0 && len(settings.CronMonthDays) > 0 {
		return fmt.Errorf("%w: weekdays and days of the month must not be combined", errInvalidPlan)
	}

	for _, wd := range settings.CronWeekdays {
		if wd < time.Sunday || wd > time.Saturday {
			return fmt.Errorf("%w: invalid weekday %d", errInvalidPlan, wd)
		}
	}

	for _, md := range settings.CronMonthDays {
		if md != LastDayOfMonth && (md < 1 || md > 31) {
			return fmt.Errorf("%w: invalid day of the month %d", errInvalidPlan, md)
		}
	}

	if (len(settings.CronWeekdays) > 0 || len(settings.CronMonthDays) > 0) && settings.PauseTime > 0 {
		return fmt.Errorf("%w: a pause time cannot be combined with weekdays or days of the month", errInvalidPlan)
	}

	if settings.KeepRunLogs < 0 {
		return fmt.Errorf("%w: negative amount of kept run logs", errInvalidPlan)
	}

	if settings.CatchUp && !opts.Durable {
		return fmt.Errorf("%w: catching up requires a durable scheduler", errInvalidPlan)
	}

	if opts.Durable {
		switch opts.Kind {
		case Schedule, Manual:
		case Cron:
			if settings.PauseTime > 0 {
				return fmt.Errorf("%w: a durable cron scheduler has no pause time", errInvalidPlan)
			}
		default:
			return fmt.Errorf("%w: a durable scheduler must be of kind Schedule, Cron or Manual", errInvalidPlan)
		}
	}

	return nil
}

// keepRunLogs returns the effective amount of kept run logs.
func keepRunLogs(settings Settings) int {
	if settings.KeepRunLogs <= 0 {
		return KeepRunLogs
	}

	return settings.KeepRunLogs
}

// planOf returns everything which determines the dues, to detect a changed plan.
func planOf(opts Options, settings Settings) string {
	return fmt.Sprintf("%d|%v|%d|%d|%d|%v|%v|%v|%v", opts.Kind, locationOf(opts), settings.PauseTime, settings.CronHour, settings.CronMinute, settings.CronWeekdays, settings.CronMonthDays, settings.Disabled, settings.StartDelay)
}
