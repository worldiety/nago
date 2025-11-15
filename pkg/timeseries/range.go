// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timeseries

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	dateTimeFormat    = "2006-01-02 15:04:05.000"
	dateTimeFormatSec = "2006-01-02 15:04:05"
)

// parserState represents a state in our finite state machine to parse the format as defined by Range.
type parserState int

const (
	psStart parserState = iota
	psMin
	psMax
	psTimeZone
)

// Range is a string representation of a range. ( or ] can be used to indicate inclusive and exclusive intervals.
// ( or ) means exclusive and [ or ] means inclusive.
// Format specification:
//
//	<[|(> <min>, <max> <]|)> @ <IANA time zone name>
//
// Examples:
//
//		[2038-01-19 03:14:07,2038-01-19 03:14:07]@Europe/Berlin
//		(2038-01-19 03:14:07,2038-01-19 03:14:07)@Europe/Berlin
//	 "" the empty string selects int64 min/max
type Range string

// NewRange creates a Range which includes the from and to values interpreted within the given timezone.
func NewRange(fromInc, toInc UnixMilli, tz *time.Location) Range {
	a := time.UnixMilli(int64(fromInc))
	b := time.UnixMilli(int64(toInc))
	return Range("[" + a.In(tz).Format(dateTimeFormat) + "," + b.In(tz).Format(dateTimeFormat) + "]@" + tz.String())
}

// Unbound returns true, if this Range is the empty string (zero value).
// Interval will return MinInt and MaxInt.
//
//goland:noinspection GoMixedReceiverTypes
func (r Range) Unbound() bool {
	return r == ""
}

// MarshalText implements encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (r Range) MarshalText() ([]byte, error) {
	if _, _, err := r.Interval(); err != nil {
		return nil, err
	}

	return []byte(r), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (r *Range) UnmarshalText(data []byte) error {
	if _, _, err := Range(data).Interval(); err != nil {
		return err
	}

	*r = Range(data)
	return nil
}

// Interval parses and returns the min and max unix timestamps, which have always 'inclusive' semantics.
// Min and max are represented as a unix timestamp in seconds.
//
//goland:noinspection GoMixedReceiverTypes
func (r Range) Interval() (min, max UnixMilli, err error) {
	if r == "" {
		return math.MinInt64, math.MaxInt64, nil
	}

	str := strings.TrimSpace(string(r))
	state := psStart
	minIsInclusive := false
	maxIsInclusive := false

	minString := ""
	maxString := ""
	tzString := ""

	offset := 0

parserLoop:
	for i, r := range str {
		switch r {
		case '[':
			if err := notInState(state, psStart, r, i); err != nil {
				return -1, -1, err
			}

			minIsInclusive = true
			state = psMin
			offset = i + 1
		case '(':
			if err := notInState(state, psStart, r, i); err != nil {
				return -1, -1, err
			}

			minIsInclusive = false
			state = psMin
			offset = i + 1
		case ',':
			if err := notInState(state, psMin, r, i); err != nil {
				return -1, -1, err
			}

			minString = strings.TrimSpace(str[offset:i])
			state = psMax
			offset = i + 1
		case ')':
			if err := notInState(state, psMax, r, i); err != nil {
				return -1, -1, err
			}

			maxString = strings.TrimSpace(str[offset:i])
			maxIsInclusive = false
			state = psTimeZone
		case ']':
			if err := notInState(state, psMax, r, i); err != nil {
				return -1, -1, err
			}

			maxString = strings.TrimSpace(str[offset:i])
			maxIsInclusive = true
			state = psTimeZone
		case '@':
			if err := notInState(state, psTimeZone, r, i); err != nil {
				return -1, -1, err
			}

			tzString = strings.TrimSpace(str[i+1:])
			break parserLoop
		}
	}

	if tzString == "" {
		return -1, -1, fmt.Errorf("time zone most be explicit")
	}

	loc, err := time.LoadLocation(tzString)
	if err != nil {
		return -1, -1, err
	}

	format := dateTimeFormat
	if !strings.Contains(minString, ".") {
		format = dateTimeFormatSec
	}

	minTime, err := time.ParseInLocation(format, minString, loc)
	if err != nil {
		return -1, -1, err
	}

	format = dateTimeFormat
	if !strings.Contains(maxString, ".") {
		format = dateTimeFormatSec
	}

	maxTime, err := time.ParseInLocation(format, maxString, loc)
	if err != nil {
		return -1, -1, err
	}

	minUnix := minTime.UnixMilli()
	maxUnix := maxTime.UnixMilli()

	if !minIsInclusive {
		minUnix++
	}

	if !maxIsInclusive {
		maxUnix--
	}

	return UnixMilli(minUnix), UnixMilli(maxUnix), nil
}

func notInState(state, expected parserState, r rune, pos int) error {
	if state != expected {
		return fmt.Errorf("1:%d: unexpected char '%s'", pos, string(r))
	}

	return nil
}
