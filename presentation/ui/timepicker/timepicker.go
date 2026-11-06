// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timepicker

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/accessibility"
	"go.wdy.de/nago/presentation/ui/alert"
)

type PickerFormat int

const (
	// ClockFormat interprets a duration as a clock e.g. 18:30. Note, that days and seconds
	// are also rendered the same way, so you may get also something like 01:18:30:42 to display.
	ClockFormat PickerFormat = iota

	// DecomposedFormat interprets a duration as distinct days, hours, minutes and seconds e.g.
	// 18 Std. 30 Min. At worst, you may get 1 T 18 Std. 30 Min. 42 Sek to display.
	DecomposedFormat
)

// TPicker is a composite component (Time Picker).
// It lets users choose a duration with optional granularity for days,
// hours, minutes, and seconds. The picker can be shown in a dialog,
// bind to external state, and format its value either as a clock or
// as decomposed units.
type TPicker struct {
	id                   string
	label                string                     // primary label shown with the control
	supportingText       string                     // helper or secondary text shown below the label
	errorText            string                     // validation or error message
	frame                ui.Frame                   // layout frame for size and positioning
	pickerPresented      *core.State[bool]          // controls whether the picker UI is currently shown
	targetSelectedState  *core.State[time.Duration] // external binding for the selected duration, which the field shows
	currentSelectedState *core.State[time.Duration] // working copy of the dialog, set to the bound value on each open
	title                string                     // title used when presenting the picker (e.g., in a dialog)
	showDays             bool                       // enables selection of days
	showHours            bool                       // enables selection of hours
	showMinutes          bool                       // enables selection of minutes
	showSeconds          bool                       // enables selection of seconds
	format               PickerFormat               // display format (clock or decomposed units)
	disabled             bool                       // when true, interaction is disabled
	// why is the scale-to-seconds removed? because it limits the picker to be in seconds, however we also need at least millisecond resolution in the near future.
}

// Picker renders a time.Duration either in clock time format or in decomposed format.
// Default is [ClockFormat]. By default, the Picker shows hours and minutes,
// but you can be specific by setting the according flags.
// Keep in mind, that the picker also clamps to the natural limits, e.g. you cannot set
// 25 hours, instead you must enable the day flag, so that the user can configure 1 day and 1 hour.
func Picker(label string, selectedState *core.State[time.Duration]) TPicker {
	p := TPicker{
		label:               label,
		format:              ClockFormat,
		targetSelectedState: selectedState,
	}

	if selectedState != nil {
		// keep the id stable across render cycles, otherwise labels lose their input reference
		p.id = core.DerivedState[string](selectedState, ".pck.id").Init(func() string {
			return fmt.Sprintf("timepicker-%d", rand.Intn(999999999))
		}).Get()

		p.pickerPresented = core.DerivedState[bool](selectedState, ".pck.pre")
		// The working copy is the value of the dialog: +/- and the fields change only the copy, which apply writes
		// into the bound state. Opening the dialog sets the copy to the bound value, so neither a cancel nor a set
		// of the bound state from outside leaves a former value in the dialog.
		p.currentSelectedState = core.DerivedState[time.Duration](selectedState, ".pck.tmp").Init(func() time.Duration {
			return selectedState.Get()
		})
	}
	return p
}

// Padding sets the inner spacing around the time picker content.
// (currently not implemented)
func (c TPicker) Padding(padding ui.Padding) ui.DecoredView {
	//TODO implement me
	return c
}

// Frame sets the layout frame of the time picker, including size and positioning.
func (c TPicker) Frame(frame ui.Frame) ui.DecoredView {
	c.frame = frame
	return c
}

// WithFrame applies a transformation function to the picker's frame
// and returns the updated component.
func (c TPicker) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	c.frame = fn(c.frame)
	return c
}

// Border sets the border style of the time picker.
// (currently not implemented)
func (c TPicker) Border(border ui.Border) ui.DecoredView {
	//TODO implement me
	return c
}

// Visible controls the visibility of the time picker; setting false hides it.
// (currently not implemented)
func (c TPicker) Visible(visible bool) ui.DecoredView {
	//TODO implement me
	return c
}

// AccessibilityLabel sets a label used by screen readers for accessibility.
// (currently not implemented)
func (c TPicker) AccessibilityLabel(label string) ui.DecoredView {
	//TODO implement me
	return c
}

// Disabled enables or disables user interaction with the time picker.
func (c TPicker) Disabled(disabled bool) TPicker {
	c.disabled = disabled
	return c
}

// Title sets the title of the picker, typically shown in dialogs.
func (c TPicker) Title(title string) TPicker {
	c.title = title
	return c
}

// Format sets the display format for the duration value (clock or decomposed).
func (c TPicker) Format(format PickerFormat) TPicker {
	c.format = format
	return c
}

// SupportingText sets helper or secondary text displayed below the picker label.
func (c TPicker) SupportingText(text string) TPicker {
	c.supportingText = text
	return c
}

// ErrorText sets the validation or error message displayed below the picker.
func (c TPicker) ErrorText(text string) TPicker {
	c.errorText = text
	return c
}

// Hours toggles whether the picker allows selecting hours.
func (c TPicker) Hours(showHours bool) TPicker {
	c.showHours = showHours
	return c
}

// Minutes toggles whether the picker allows selecting minutes.
func (c TPicker) Minutes(showMinutes bool) TPicker {
	c.showMinutes = showMinutes
	return c
}

// Days toggles whether the picker allows selecting days.
func (c TPicker) Days(showDays bool) TPicker {
	c.showDays = showDays
	return c
}

// Seconds toggles whether the picker allows selecting seconds.
func (c TPicker) Seconds(showSeconds bool) TPicker {
	c.showSeconds = showSeconds
	return c
}

// auto returns true if none of the time units are explicitly enabled,
// indicating that the picker should choose units automatically.
func auto(showDays, showHours, showMinutes, showSeconds bool) bool {
	return !showDays && !showHours && !showMinutes && !showSeconds
}

// fmtDurationTime formats a duration into a human-readable string,
// showing days, hours, minutes, and/or seconds depending on the flags.
// If no units are explicitly enabled, the function chooses appropriate
// ones based on the actual duration values.
func fmtDurationTime(showDays, showHours, showMinutes, showSeconds bool, d time.Duration) string {
	days, hours, minutes, seconds := FromDuration(d)

	if auto(showDays, showHours, showMinutes, showSeconds) {
		showHours = true
		showMinutes = true
	}

	var segments []string
	if showDays {
		segments = append(segments, fmt.Sprintf("%d T", days))
	}

	if showHours {
		segments = append(segments, fmt.Sprintf("%d Std.", hours))
	}

	if showMinutes {
		segments = append(segments, fmt.Sprintf("%d Min.", minutes))
	}

	if showSeconds {
		segments = append(segments, fmt.Sprintf("%d Sek.", seconds))
	}

	return strings.Join(segments, " ")
}

// fmtClockTime formats a duration in a clock-like style, using zero-padded
// values separated by thin spaces and colons. Units shown are determined
// by the flags or inferred from the duration.
func fmtClockTime(showDays, showHours, showMinutes, showSeconds bool, d time.Duration) string {
	days, hours, minutes, seconds := FromDuration(d)

	if auto(showDays, showHours, showMinutes, showSeconds) {
		showHours = true
		showMinutes = true
	}

	var segments []string
	if showDays {
		segments = append(segments, fmt.Sprintf("%02d", days))
	}

	if showHours {
		segments = append(segments, fmt.Sprintf("%02d", hours))
	}

	if showMinutes {
		segments = append(segments, fmt.Sprintf("%02d", minutes))
	}

	if showSeconds {
		segments = append(segments, fmt.Sprintf("%02d", seconds))
	}

	return strings.Join(segments, "\u202F:\u202F") // use thin space instead of space
}

// dayDown decreases the number of days in the current selection,
// wrapping around to 99 if it goes below 0.
func (c TPicker) dayDown() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	days--
	if days < 0 {
		days = 99
	}
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// dayUp increases the number of days in the current selection,
// wrapping back to 0 if it exceeds 99.
func (c TPicker) dayUp() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	days++
	if days > 99 {
		days = 0
	}
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// setDay sets the typed days, clamped to 0 to 99 like the buttons.
func (c TPicker) setDay(days int) {
	_, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	days = clamp(days, 99)
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// hourDown decreases the hours in the current selection,
// wrapping around to 23 if it goes below 0.
func (c TPicker) hourDown() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	hours--
	if hours < 0 {
		hours = 23
	}
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// hourUp increases the hours in the current selection,
// wrapping back to 0 if it reaches 24.
func (c TPicker) hourUp() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	hours++
	if hours >= 24 {
		hours = 0
	}
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// setHour sets the typed hours, clamped to 0 to 23.
func (c TPicker) setHour(hours int) {
	days, _, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	hours = clamp(hours, 23)
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// minDown decreases the minutes in the current selection,
// wrapping around to 59 if it goes below 0.
func (c TPicker) minDown() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	minutes--
	if minutes < 0 {
		minutes = 59
	}
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// minUp increases the minutes in the current selection,
// wrapping back to 0 if it exceeds 59.
func (c TPicker) minUp() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	minutes++
	if minutes > 59 {
		minutes = 0
	}
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// setMin sets the typed minutes, clamped to 0 to 59.
func (c TPicker) setMin(minutes int) {
	days, hours, _, seconds := FromDuration(c.currentSelectedState.Get())
	minutes = clamp(minutes, 59)
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// secDown decreases the seconds in the current selection,
// wrapping around to 59 if it goes below 0.
func (c TPicker) secDown() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	seconds--
	if seconds < 0 {
		seconds = 59
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

// secUp increases the seconds in the current selection,
// wrapping back to 0 if it exceeds 59.
func (c TPicker) secUp() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	seconds++
	if seconds > 59 {
		seconds = 0
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

// setSec sets the typed seconds, clamped to 0 to 59.
func (c TPicker) setSec(seconds int) {
	days, hours, minutes, _ := FromDuration(c.currentSelectedState.Get())
	seconds = clamp(seconds, 59)
	c.currentSelectedState.Set(c.round(Duration(days, hours, minutes, seconds)))
}

// clamp returns v within 0 and max: a typed value out of the bounds of a field is set to the nearest bound, instead
// of wrapping around like the buttons.
func clamp(v, max int) int {
	if v < 0 {
		return 0
	}

	if v > max {
		return max
	}

	return v
}

// round normalizes the given duration based on the picker's configuration.
// If seconds are not displayed, the duration is truncated to the nearest minute;
// otherwise, it is returned unchanged.
func (c TPicker) round(d time.Duration) time.Duration {
	if !c.showSeconds {
		return d.Truncate(time.Minute)
	}

	return d
}

// renderPicker builds the interactive picker view for adjusting the duration.
// It shows increment and decrement buttons with numeric labels for each
// enabled unit (days, hours, minutes, seconds). If no units are explicitly
// enabled, the picker automatically decides which units to display based on
// the current duration.
func (c TPicker) renderPicker() core.View {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())

	if auto(c.showDays, c.showHours, c.showMinutes, c.showSeconds) {
		c.showHours = true
		c.showMinutes = true
	}

	segments := make([]core.View, 0, 8)
	if c.showDays {
		inputId := fmt.Sprintf("%s-day", c.id)
		segments = append(segments,
			ui.VStack(
				ui.TertiaryButton(c.dayUp).PreIcon(heroSolid.ChevronUp).AccessibilityLabel(fmt.Sprintf("Tage erhöhen, aktuell %d", days)).Key(keyUp(unitDay), c.id),
				editInt(c.currentSelectedState, inputId, unitDay, days, func(v int) {
					c.setDay(v)
				}),
				ui.TertiaryButton(c.dayDown).PreIcon(heroSolid.ChevronDown).AccessibilityLabel(fmt.Sprintf("Tage verringern, aktuell %d", days)).Key(keyDown(unitDay), c.id),
			),
			ui.Text("Tage").LabelFor(inputId),
		)
	}

	if c.showHours {
		inputId := fmt.Sprintf("%s-hour", c.id)
		segments = append(segments,
			ui.VStack(
				ui.TertiaryButton(c.hourUp).PreIcon(heroSolid.ChevronUp).AccessibilityLabel(fmt.Sprintf("Stunden erhöhen, aktuell %d", hours)).Key(keyUp(unitHour), c.id),
				editInt(c.currentSelectedState, inputId, unitHour, hours, func(v int) {
					c.setHour(v)
				}),
				ui.TertiaryButton(c.hourDown).PreIcon(heroSolid.ChevronDown).AccessibilityLabel(fmt.Sprintf("Stunden verringern, aktuell %d", hours)).Key(keyDown(unitHour), c.id),
			),
			ui.Text("Std.").LabelFor(inputId),
		)
	}

	if c.showMinutes {
		inputId := fmt.Sprintf("%s-min", c.id)
		segments = append(segments,
			ui.VStack(
				ui.TertiaryButton(c.minUp).PreIcon(heroSolid.ChevronUp).AccessibilityLabel(fmt.Sprintf("Minuten erhöhen, aktuell %d", minutes)).Key(keyUp(unitMin), c.id),
				editInt(c.currentSelectedState, inputId, unitMin, minutes, func(v int) {
					c.setMin(v)
				}),
				ui.TertiaryButton(c.minDown).PreIcon(heroSolid.ChevronDown).AccessibilityLabel(fmt.Sprintf("Minuten verringern, aktuell %d", minutes)).Key(keyDown(unitMin), c.id),
			),
			ui.Text("Min.").LabelFor(inputId),
		)
	}

	if c.showSeconds {
		inputId := fmt.Sprintf("%s-sec", c.id)
		segments = append(segments,
			ui.VStack(
				ui.TertiaryButton(c.secUp).PreIcon(heroSolid.ChevronUp).AccessibilityLabel(fmt.Sprintf("Sekunden erhöhen, aktuell %d", seconds)).Key(keyUp(unitSec), c.id),
				editInt(c.currentSelectedState, inputId, unitSec, seconds, func(v int) {
					c.setSec(v)
				}),
				ui.TertiaryButton(c.secDown).PreIcon(heroSolid.ChevronDown).AccessibilityLabel(fmt.Sprintf("Sekunden verringern, aktuell %d", seconds)).Key(keyDown(unitSec), c.id),
			),
			ui.Text("Sek.").LabelFor(inputId),
		)
	}

	return ui.HStack(segments...).Frame(ui.Frame{}.FullWidth())
}

// The names of the units of the dialog, which name the fields and the keys of their buttons.
const (
	unitDay  = "day"
	unitHour = "hour"
	unitMin  = "min"
	unitSec  = "sec"
)

// keyApply, keyUp and keyDown name the buttons of the dialog, see [ui.TButton.Key], with the id of the picker as
// target, so that a click right after typing into a field of the dialog is not lost. They are not the keys of alert,
// so that the buttons of a picker in a dialog without a title do not collide with the buttons of that dialog. Two
// pickers of the same state share their id and thus their keys, which makes them ambiguous.
const keyApply = "nago.timepicker.apply"

func keyUp(unit string) string {
	return "nago.timepicker." + unit + ".up"
}

func keyDown(unit string) string {
	return "nago.timepicker." + unit + ".down"
}

// editInt is the field of a unit of the dialog. The field shows the value of the working copy; its state only
// carries the typed text to set, which parses it, so the render does not set it (which would render again).
// Text, which is no integer, leaves the value.
func editInt(state *core.State[time.Duration], inputId, name string, val int, set func(v int)) core.View {
	formattedVal := fmt.Sprintf("%02d", val)

	editVal := core.DerivedState[string](state, name+"-edit").Observe(func(newValue string) {
		if v, err := strconv.Atoi(newValue); err == nil {
			set(v)
		}
	})

	textField := ui.TextField("", formattedVal).
		ID(inputId).
		ShowZero(true).
		Min(0).
		Step(1).
		Style(ui.TextFieldBasic).
		InputValue(editVal).
		KeyboardOptions(ui.KeyboardOptions().KeyboardType(ui.KeyboardInteger)).
		TextAlignment(ui.TextAlignCenter)

	switch name {
	case unitDay:
		textField = textField.Max(99)
	case unitHour:
		textField = textField.Max(23)
	case unitMin, unitSec:
		textField = textField.Max(59)
	}

	return textField.Frame(ui.Frame{Width: ui.L32})
}

// Render builds and returns the UI representation of the time picker.
// It displays the current duration value, opens a dialog for selecting
// a new duration, and shows labels, supporting text, or error messages
// depending on its state. The picker can be disabled, in which case
// interaction is blocked.
func (c TPicker) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	// the field shows the bound state, not the working copy of the dialog
	durationText := Format(c.showDays, c.showHours, c.showMinutes, c.showSeconds, c.format, c.targetSelectedState.Get())

	return accessibility.Accessible(
		ui.VStack(
			ui.TextField(c.label, durationText).
				ReadOnly(true).
				SizeByContent(true).
				Trailing(ui.Image().Embed(heroSolid.ChevronRight).Frame(ui.Frame{}.Size(ui.L16, ui.L16))).
				Disabled(c.disabled).
				ErrorText(c.errorText).
				SupportingText(c.supportingText).
				Frame(ui.Frame{MinWidth: ui.L0}),

			// cancel, escape and a click beside the dialog only close it: the working copy is set again on the next
			// open, so they write nothing
			alert.Dialog(c.title, c.renderPicker(), c.pickerPresented, alert.Cancel(nil), alert.Custom(func(close func(closeDlg bool)) core.View {
				// positive case
				return ui.PrimaryButton(func() {
					c.targetSelectedState.Set(c.currentSelectedState.Get())
					c.targetSelectedState.Notify() // invoke observers
					close(true)
				}).Title(rstring.ActionApply.Get(wnd)).Key(keyApply, c.id)
			})),
		).
			Action(func() {
				if c.disabled {
					return
				}
				c.currentSelectedState.Set(c.targetSelectedState.Get())
				c.pickerPresented.Set(true)
			}).
			Alignment(ui.Leading).
			Frame(c.frame),
	).
		Role(accessibility.RoleButton).
		Disabled(c.disabled).
		Render(ctx)
}

// FromDuration returns the days, hours, minutes and seconds from the given duration.
func FromDuration(d time.Duration) (days, hours, minutes, seconds int) {
	days = int(d / (time.Hour * 24))
	d = d % (time.Hour * 24)

	hours = int(d / time.Hour)
	d = d % time.Hour

	minutes = int(d / time.Minute)
	d = d % time.Minute

	seconds = int(d / time.Second)

	return
}

// Duration creates time.Duration based on the decomposed individual durations.
func Duration(days, hours, minutes, seconds int) time.Duration {
	return time.Hour*24*time.Duration(days) + time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes) + time.Second*time.Duration(seconds)
}

func Format(showDays, showHours, showMinutes, showSeconds bool, format PickerFormat, duration time.Duration) string {
	dur := duration

	var durationText string
	switch format {
	case DecomposedFormat:
		durationText = fmtDurationTime(showDays, showHours, showMinutes, showSeconds, dur)
	default:
		durationText = fmtClockTime(showDays, showHours, showMinutes, showSeconds, dur)
	}

	return durationText
}

// Minutes returns only the exact truncated minutes fraction of the duration.
func Minutes(duration time.Duration) int {
	_, _, m, _ := FromDuration(duration)
	return m
}

// Hours returns only the exact truncated hours fraction of the duration.
func Hours(duration time.Duration) int {
	_, h, _, _ := FromDuration(duration)
	return h
}
