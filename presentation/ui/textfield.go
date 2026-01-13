// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type TextFieldStyle uint

func (t TextFieldStyle) ora() proto.TextFieldStyle {
	return proto.TextFieldStyle(t)
}

const (
	// TextFieldReduced has no outlines and thus less visual disruption in larger forms.
	TextFieldReduced TextFieldStyle = TextFieldStyle(proto.TextFieldReduced)

	// TextFieldOutlined is fine for smaller forms and helps to identify where to put text in the form.
	TextFieldOutlined TextFieldStyle = TextFieldStyle(proto.TextFieldOutlined)

	// TextFieldBasic removes as much as decorations as possible. There may be limitations based on the platform.
	// Note, that an implementation is allowed to ignore leading, trailing, supporting and errorText for this mode.
	// It may serve as a building-block for custom fields.
	TextFieldBasic TextFieldStyle = TextFieldStyle(proto.TextFieldBasic)
)

// TTextField is a basic component (Text Field).
// This component provides a text input field with optional supporting and error text,
// leading/trailing views (e.g., icons), debounce settings, styling, and keyboard options.
// It supports both controlled (via State) and uncontrolled (via value) modes.
//
// It is typically used in forms, search bars, and other user input scenarios.
type TTextField struct {
	label           string               // label displayed above or inside the field
	value           string               // static text value (used if InputValue is not set)
	inputValue      *core.State[string]  // bound state for controlled input
	supportingText  string               // helper text shown below the field
	errorText       string               // error message shown below the field
	disabled        bool                 // disables user interaction
	leading         core.View            // optional leading element (e.g., icon)
	trailing        core.View            // optional trailing element (e.g., button, icon)
	style           proto.TextFieldStyle // visual style of the field (outlined, filled, etc.)
	disableDebounce bool                 // disables debounce if true
	debounceTime    time.Duration        // custom debounce duration for input
	invisible       bool                 // hides the field when true
	frame           Frame                // layout constraints
	lines           int                  // number of lines (0 = single line, >0 = text area)
	keyboardOptions TKeyboardOptions     // platform-specific keyboard options
	id              string               // unique identifier for the field
	keydownEnter    func()               // callback for Enter key press
	textAlignment   TextAlignment
	showZero        bool    // show '0' character for empty/zero values in number fields
	step            int     // step size to increase/decrease number values
	maxValue        float64 // max value for number fields
	minValue        float64 // min value for number fields
}

// Padding is a placeholder implementation.
func (c TTextField) Padding(padding Padding) DecoredView {
	// TODO implement me or reduce interface
	return c
}

// Border is a placeholder implementation.
func (c TTextField) Border(border Border) DecoredView {
	// TODO implement me or reduce interface
	return c
}

// AccessibilityLabel is a placeholder implementation.
func (c TTextField) AccessibilityLabel(label string) DecoredView {
	// TODO implement me or reduce interface
	return c
}

// TextField creates a new text field with the given label and initial value.
// By default, it is single-line and uncontrolled until InputValue is set.
func TextField(label string, value string) TTextField {
	c := TTextField{
		label: label,
		value: value,
	}

	return c
}

// IntField is just a TextField using the according keyboard hints. Remember, that these IME hints are no guarantees
// and a user may enter non-integer stuff anyway. However, any
// incompatible inputs are ignored and the given int-state is just a kind of view on top of the string state.
func IntField(label string, value int64, state *core.State[int64]) TTextField {
	strState := core.StateOf[string](state.Window(), state.ID()+".int64").Init(func() string {
		return strconv.FormatInt(value, 10)
	})

	strState.Observe(func(newValue string) {
		v, err := strconv.ParseInt(newValue, 10, 64)
		if err != nil {
			slog.Error("cannot parse IntField value from TextField state", "strState", strState.ID(), "err", err)
		}

		if v != state.Get() {
			state.Set(v)
			state.Notify() // delegate the observable event, because it was caused by the ui
		}
	})

	state.Observe(func(newValue int64) {
		i := strconv.FormatInt(newValue, 10)
		if strState.Get() != i {
			strState.Set(i)
		}
	})

	return TextField(label, strState.Get()).InputValue(strState).KeyboardType(KeyboardInteger)
}

// FloatFieldValue just renders a non-stateful float value. See also [FloatField]. Due to the generic instantiation,
// one can influence the float rendering through the Stringer interface.
func FloatFieldValue[T ~float64 | ~float32](label string, value T) TTextField {
	return TextField(label, fmt.Sprintf("%v", value)).KeyboardType(KeyboardFloat)
}

// FloatField is just a TextField using the according keyboard hints. Remember, that these IME hints are no guarantees
// and a user may enter non-integer stuff anyway. However, any
// incompatible inputs are ignored and the given int-state is just a kind of view on top of the string state.
// See also [FloatFieldValue] if you just want to display a non-stateful float value.
func FloatField(label string, value float64, state *core.State[float64]) TTextField {
	var strState *core.State[string]
	var val string

	if state != nil {

		strState = core.StateOf[string](state.Window(), state.ID()+".float64").Init(func() string {
			return strconv.FormatFloat(value, 'f', -1, 64)
		})

		strState.Observe(func(newValue string) {
			v, err := strconv.ParseFloat(newValue, 64)
			if err != nil {
				slog.Error("cannot parse FloatField value from TextField state", "strState", strState.ID(), "err", err)
			}

			if v != state.Get() {
				state.Set(v)
				state.Notify() // delegate the observable event, because it was caused by the ui
			}
		})

		state.Observe(func(newValue float64) {
			i := strconv.FormatFloat(newValue, 'f', -1, 64)
			if strState.Get() != i {
				strState.Set(i)
			}
		})

		val = strState.Get()
	}

	return TextField(label, val).InputValue(strState).KeyboardType(KeyboardFloat)
}

// SupportingText sets helper text for the field.
// This text is displayed below the input and is typically used to provide hints or guidance.
func (c TTextField) SupportingText(text string) TTextField {
	c.supportingText = text
	return c
}

// ErrorText sets an error message for the field.
// When provided, this text is shown below the input in place of supporting text,
// usually styled to indicate an error state.
func (c TTextField) ErrorText(text string) TTextField {
	c.errorText = text
	return c
}

// Leading sets a leading view for the field.
// This view is displayed at the start of the input field, e.g., an icon.
func (c TTextField) Leading(v core.View) TTextField {
	c.leading = v
	return c
}

// Trailing sets a trailing view for the field.
// This view is displayed at the end of the input field, e.g., a clear button or icon.
func (c TTextField) Trailing(v core.View) TTextField {
	c.trailing = v
	return c
}

func (c TTextField) TextAlignment(v TextAlignment) TTextField {
	c.textAlignment = v
	return c
}

// Style sets the wanted style. If empty, [proto.TextFieldOutlined] is applied.
func (c TTextField) Style(s TextFieldStyle) TTextField {
	c.style = s.ora()
	return c
}

// DebounceTime sets a custom debouncing time when entering text. By default, this is 500ms and always applied.
// You can disable debouncing, but be very careful with that, as it may break your server, the client or network.
func (c TTextField) DebounceTime(d time.Duration) TTextField {
	c.debounceTime = d
	return c
}

// Debounce is enabled by default. See also DebounceTime.
func (c TTextField) Debounce(enabled bool) TTextField {
	c.disableDebounce = !enabled
	return c
}

// Label sets the label text of the field.
// Unlike other setters, this does not return a modified copy of TTextField.
func (c TTextField) Label(label string) {
	c.label = label
}

// InputValue binds the text field to a reactive state.
// This enables controlled input behavior where the state is updated as the user types.
func (c TTextField) InputValue(input *core.State[string]) TTextField {
	c.inputValue = input
	return c
}

// Disabled disables or enables the field.
// When disabled, the user cannot interact with the field.
func (c TTextField) Disabled(disabled bool) TTextField {
	c.disabled = disabled
	return c
}

// Frame sets the layout frame of the field (size, width, height, etc.).
func (c TTextField) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

// WithFrame updates the current frame of the field via a transformation function.
func (c TTextField) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// FullWidth expands the text field to take the full available width.
func (c TTextField) FullWidth() TTextField {
	c.frame = c.frame.FullWidth()
	return c
}

// KeydownEnter currently only works for one line text fields (lines=0) and not for text area.
// The enter key logic collides with the new line logic and it is currently not clear how this situation
// shall be handled:
//   - there must be a combined key gesture
//   - using shift and enter for new lines is surprising for any user
//   - using the inversion, which is shift and enter for submitting is also wrong, because that is already overloaded
//     in multiple ways (e.g. line break vs paragraph behavior or opening a new window in Chrome etc)
//   - same applies to Str + Enter which may also be overloaded, typically for a soft line break
func (c TTextField) KeydownEnter(fn func()) TTextField {
	c.keydownEnter = fn
	return c
}

// Lines are by default at 0 and enforces a single line text field. Otherwise, a text area is created.
// This is also true, if lines 1 to differentiate between subtile behavior of single line text fields and single
// line text areas, which may take even more lines, because e.g. a web browser allows to change that on demand.
func (c TTextField) Lines(lines int) TTextField {
	c.lines = lines
	return c
}

// Visible toggles the visibility of the text field.
// When set to false, the field is hidden from view but still part of the layout.
func (c TTextField) Visible(v bool) DecoredView {
	c.invisible = !v
	return c
}

// KeyboardOptions sets advanced keyboard behavior (type, capitalization, return key, etc.).
func (c TTextField) KeyboardOptions(options TKeyboardOptions) TTextField {
	c.keyboardOptions = options
	return c
}

// KeyboardType sets the type of keyboard to display (e.g., text, number, email).
func (c TTextField) KeyboardType(keyboardType KeyboardType) TTextField {
	c.keyboardOptions.keyboardType = keyboardType
	return c
}

// ID assigns a unique identifier to the text field.
// Useful for testing, accessibility, or programmatic interaction.
func (c TTextField) ID(id string) TTextField {
	c.id = id
	return c
}

// ShowZero defines wheter the '0' character should be displayed for empty/zero values in number fields.
func (c TTextField) ShowZero(showZero bool) TTextField {
	c.showZero = showZero
	return c
}

// Step defines the step size to increase/decrease number values stepwise
func (c TTextField) Step(step int) TTextField {
	c.step = step
	return c
}

// Max defines the max value of number fields
func (c TTextField) Max(max float64) TTextField {
	c.maxValue = max
	return c
}

// Min defines the min value of number fields
func (c TTextField) Min(min float64) TTextField {
	c.minValue = min
	return c
}

// Render builds and returns the protocol representation of the text field.
func (c TTextField) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.TextField{
		Label:           proto.Str(c.label),
		SupportingText:  proto.Str(c.supportingText),
		ErrorText:       proto.Str(c.errorText),
		Value:           proto.Str(c.value),
		InputValue:      c.inputValue.Ptr(),
		Disabled:        proto.Bool(c.disabled),
		Leading:         render(ctx, c.leading),
		Trailing:        render(ctx, c.trailing),
		Style:           c.style,
		DebounceTime:    proto.Duration(c.debounceTime),
		DisableDebounce: proto.Bool(c.disableDebounce),
		Invisible:       proto.Bool(c.invisible),
		Frame:           c.frame.ora(),
		Lines:           proto.Uint(c.lines),
		KeyboardOptions: c.keyboardOptions.ora(),
		Id:              proto.Str(c.id),
		KeydownEnter:    ctx.MountCallback(c.keydownEnter),
		TextAlignment:   proto.TextAlignment(c.textAlignment),
		ShowZero:        proto.Bool(c.showZero),
		Step:            proto.Uint(c.step),
		Max:             proto.Float(c.maxValue),
		Min:             proto.Float(c.minValue),
	}
}
