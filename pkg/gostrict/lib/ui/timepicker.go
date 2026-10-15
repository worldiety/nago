// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import (
	"time"

	"go.wdy.de/nago/presentation/ui/timepicker"
)

// TimepickerPickerFormat is the display format of a duration, like timepicker.PickerFormat of nago. The picker itself
// is not mirrored, because it binds a state of a time.Duration.
type TimepickerPickerFormat int32

const (
	// TimepickerClockFormat shows a duration as a clock, e.g. 18:30, like timepicker.ClockFormat of nago.
	TimepickerClockFormat TimepickerPickerFormat = 0
	// TimepickerDecomposedFormat shows a duration as days, hours, minutes and seconds, like timepicker.DecomposedFormat
	// of nago.
	TimepickerDecomposedFormat TimepickerPickerFormat = 1
)

// TimepickerFromDuration returns the days, hours, minutes and seconds of a duration in milliseconds, like
// timepicker.FromDuration of nago.
//
//gostrict:host
func TimepickerFromDuration(millis int32) (days int32, hours int32, minutes int32, seconds int32) {
	d, h, m, s := timepicker.FromDuration(time.Duration(millis) * time.Millisecond)
	return int32(d), int32(h), int32(m), int32(s)
}

// TimepickerDuration returns a duration in milliseconds from the days, hours, minutes and seconds, like
// timepicker.Duration of nago. A duration of more than about 24 days does not fit into the milliseconds.
//
//gostrict:host
func TimepickerDuration(days int32, hours int32, minutes int32, seconds int32) int32 {
	return int32(timepicker.Duration(int(days), int(hours), int(minutes), int(seconds)) / time.Millisecond)
}

// TimepickerFormat returns the text of a duration in milliseconds, like timepicker.Format of nago.
//
//gostrict:host
func TimepickerFormat(showDays bool, showHours bool, showMinutes bool, showSeconds bool, format TimepickerPickerFormat, millis int32) string {
	return timepicker.Format(showDays, showHours, showMinutes, showSeconds, timepicker.PickerFormat(format), time.Duration(millis)*time.Millisecond)
}

// TimepickerMinutes returns the truncated minutes of a duration in milliseconds, like timepicker.Minutes of nago.
//
//gostrict:host
func TimepickerMinutes(millis int32) int32 {
	return int32(timepicker.Minutes(time.Duration(millis) * time.Millisecond))
}

// TimepickerHours returns the truncated hours of a duration in milliseconds, like timepicker.Hours of nago.
//
//gostrict:host
func TimepickerHours(millis int32) int32 {
	return int32(timepicker.Hours(time.Duration(millis) * time.Millisecond))
}
