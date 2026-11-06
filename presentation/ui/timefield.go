// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"
	"strings"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

// StrTimeFieldInvalid is the error text of a [TimeField], whose state holds no time of day.
var StrTimeFieldInvalid = i18n.MustString(
	"nago.timefield.invalid",
	i18n.Values{
		language.German:  "Keine gültige Uhrzeit, z. B. 8:30 oder 0830.",
		language.English: "Not a valid time, e.g. 8:30 or 0830.",
	},
	i18n.LocalizationHint("This text is shown below a field, into which a user types a time of day, if the typed text is no time."),
)

// TTimeField is a basic component (Time Field).
// It is a text field for a time of day, which the user types, e.g. 8:30, 8.30, 0830 or 8. See [TimeField].
type TTimeField struct {
	label          string
	state          *core.State[string]
	supportingText string
	errorText      string
	placeholder    string
	disabled       bool
	frame          Frame
}

// TimeField returns a text field for a time of day, which the user types, bound to the state.
//
// The state holds the time as HH:MM from 00:00 to 23:59, "" for an empty field or, if the text is no time, the text as
// typed; the field then shows an error text, so a use case must check the state before it saves it. The field reads
// (spaces at the edges do not count): with ":", "." or "," as hour and minute (8:30, 8.30, "8:" is 08:00); without a
// separator one or two digits as the hour, if it is at most 23, else as hour and minute (73 is 07:03); three digits as
// H:MM, if MM is at most 59, else as HH:M (830 is 08:30, 083 is 08:03); four digits as HH:MM. So each beginning of a
// valid time is a time, and typing 0830 shows no error on the way.
//
// The field shows what has been typed and rewrites it only at the complete forms H:MM, HH:MM and four digits, e.g.
// 8:30 and 0830 to 08:30, but not 8 or 830, which are still being typed. The observers of the state are called, when
// its value changes. A state, which has been set from outside, is shown as it is. The field does not debounce, it
// shows the keyboard for phone numbers, which keeps the text as typed, and it sets no state, while it renders.
func TimeField(label string, state *core.State[string]) TTimeField {
	return TTimeField{label: label, state: state}
}

// SupportingText sets the helper text below the field.
func (c TTimeField) SupportingText(text string) TTimeField {
	c.supportingText = text
	return c
}

// ErrorText sets the error message below the field. A text, which is not empty, replaces the error text of the
// field, an empty text does not.
func (c TTimeField) ErrorText(text string) TTimeField {
	c.errorText = text
	return c
}

// Placeholder sets the text, which an empty field shows.
func (c TTimeField) Placeholder(placeholder string) TTimeField {
	c.placeholder = placeholder
	return c
}

// Disabled enables or disables the field.
func (c TTimeField) Disabled(disabled bool) TTimeField {
	c.disabled = disabled
	return c
}

// Frame sets the layout frame of the field.
func (c TTimeField) Frame(frame Frame) TTimeField {
	c.frame = frame
	return c
}

// Render builds the text field.
func (c TTimeField) Render(ctx core.RenderContext) core.RenderNode {
	var input *core.State[string]
	var value, text string
	if c.state != nil {
		state := c.state
		// the text as typed, which the field sends; only its observer writes the bound state
		input = core.DerivedState[string](state, ".timefield")
		input.Observe(func(typed string) {
			if v := timeFieldValue(typed); v != state.Get() {
				state.Set(v)
				state.Notify() // delegate the observable event, because it was caused by the ui
			}
		})

		value = state.Get()
		text = value
		if typed := input.Get(); timeFieldValue(typed) == value {
			// the state is still what has been typed, otherwise it has been set from outside and is shown as it is
			text = timeFieldText(typed)
		}
	}

	errorText := c.errorText
	if errorText == "" {
		if _, _, ok := parseTimeOfDay(value); !ok && strings.TrimSpace(value) != "" {
			errorText = StrTimeFieldInvalid.Get(ctx.Window())
		}
	}

	return TextField(c.label, text).
		InputValue(input).
		KeyboardType(KeyboardPhone).
		Debounce(false).
		SupportingText(c.supportingText).
		ErrorText(errorText).
		Placeholder(c.placeholder).
		Disabled(c.disabled).
		Frame(c.frame).
		Render(ctx)
}

// timeFieldValue returns the value of the state of a [TimeField] for the typed text: HH:MM for a time, "" for an
// empty text or one of spaces, and otherwise the text as typed.
func timeFieldValue(typed string) string {
	if strings.TrimSpace(typed) == "" {
		return ""
	}

	if h, m, ok := parseTimeOfDay(typed); ok {
		return fmt.Sprintf("%02d:%02d", h, m)
	}

	return typed
}

// timeFieldText returns the text, which a [TimeField] shows for the typed text: HH:MM for the complete forms H:MM,
// HH:MM (with ":", "." or ",") and four digits, otherwise the text as typed.
func timeFieldText(typed string) string {
	h, m, ok := parseTimeOfDay(typed)
	if !ok {
		return typed
	}

	s := strings.TrimSpace(typed)
	complete := len(s) == 4 && isDigits(s)
	if i := strings.IndexAny(s, ":.,"); i >= 0 {
		complete = len(s)-i-1 == 2
	}

	if !complete {
		return typed
	}

	return fmt.Sprintf("%02d:%02d", h, m)
}

// parseTimeOfDay reads a typed time of day after the rules of [TimeField]. An empty text is no time.
func parseTimeOfDay(typed string) (hour, minute int, ok bool) {
	s := strings.TrimSpace(typed)
	if i := strings.IndexAny(s, ":.,"); i >= 0 {
		h, m := s[:i], s[i+1:]
		if len(h) < 1 || len(h) > 2 || len(m) > 2 || !isDigits(h) || !isDigits(m) {
			return 0, 0, false
		}

		hour, minute = digitsValue(h), digitsValue(m)
	} else {
		if !isDigits(s) {
			return 0, 0, false
		}

		switch len(s) {
		case 1, 2:
			if n := digitsValue(s); n <= 23 {
				hour = n
			} else {
				hour, minute = digitsValue(s[:1]), digitsValue(s[1:])
			}
		case 3:
			if mm := digitsValue(s[1:]); mm <= 59 {
				hour, minute = digitsValue(s[:1]), mm
			} else {
				hour, minute = digitsValue(s[:2]), digitsValue(s[2:])
			}
		case 4:
			hour, minute = digitsValue(s[:2]), digitsValue(s[2:])
		default:
			return 0, 0, false
		}
	}

	if hour > 23 || minute > 59 {
		return 0, 0, false
	}

	return hour, minute, true
}

// isDigits reports whether s consists of the ASCII digits 0 to 9 only; an empty s does.
func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// digitsValue returns the value of at most a few ASCII digits, see isDigits; an empty s is 0.
func digitsValue(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n = n*10 + int(s[i]-'0')
	}

	return n
}
