// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"log/slog"
	"strconv"
	"time"
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

type TTextField struct {
	label           string
	value           string
	inputValue      *core.State[string]
	supportingText  string
	errorText       string
	disabled        bool
	leading         core.View
	trailing        core.View
	style           proto.TextFieldStyle
	disableDebounce bool
	debounceTime    time.Duration
	invisible       bool
	frame           Frame
	lines           int
	keyboardOptions TKeyboardOptions
	id              string
	keydownEnter    func()
}

func (c TTextField) Padding(padding Padding) DecoredView {
	// TODO implement me or reduce interface
	return c
}

func (c TTextField) Border(border Border) DecoredView {
	// TODO implement me or reduce interface
	return c
}

func (c TTextField) AccessibilityLabel(label string) DecoredView {
	// TODO implement me or reduce interface
	return c
}

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

func (c TTextField) SupportingText(text string) TTextField {
	c.supportingText = text
	return c
}

func (c TTextField) ErrorText(text string) TTextField {
	c.errorText = text
	return c
}

func (c TTextField) Leading(v core.View) TTextField {
	c.leading = v
	return c
}

func (c TTextField) Trailing(v core.View) TTextField {
	c.trailing = v
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

func (c TTextField) Label(label string) {
	c.label = label
}

func (c TTextField) InputValue(input *core.State[string]) TTextField {
	c.inputValue = input
	return c
}

func (c TTextField) Disabled(disabled bool) TTextField {
	c.disabled = disabled
	return c
}

func (c TTextField) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

func (c TTextField) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

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

func (c TTextField) Visible(v bool) DecoredView {
	c.invisible = !v
	return c
}

func (c TTextField) KeyboardOptions(options TKeyboardOptions) TTextField {
	c.keyboardOptions = options
	return c
}

func (c TTextField) KeyboardType(keyboardType KeyboardType) TTextField {
	c.keyboardOptions.keyboardType = keyboardType
	return c
}

func (c TTextField) ID(id string) TTextField {
	c.id = id
	return c
}

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
	}
}
