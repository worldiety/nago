// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import "time"

// Clock provides the time and the timers of a [Manager] and its schedulers. A nil Clock means the system clock.
// Tests insert their own clock to control the time instead of waiting for it.
type Clock interface {
	Now() time.Time
	NewTimer(d time.Duration) Timer
}

// Timer is a single shot timer of a [Clock], like [time.Timer].
type Timer interface {
	C() <-chan time.Time
	Stop() bool
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

func (systemClock) NewTimer(d time.Duration) Timer {
	return systemTimer{t: time.NewTimer(d)}
}

type systemTimer struct {
	t *time.Timer
}

func (t systemTimer) C() <-chan time.Time {
	return t.t.C
}

func (t systemTimer) Stop() bool {
	return t.t.Stop()
}

func clockOrSystem(c Clock) Clock {
	if c == nil {
		return systemClock{}
	}

	return c
}
