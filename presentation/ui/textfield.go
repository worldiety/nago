package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
	"strconv"
	"time"
)

type TextFieldStyle string

func (t TextFieldStyle) ora() ora.TextFieldStyle {
	return ora.TextFieldStyle(t)
}

const (
	// TextFieldReduced has no outlines and thus less visual disruption in larger forms.
	TextFieldReduced TextFieldStyle = "r"

	// TextFieldOutlined is fine for smaller forms and helps to identify where to put text in the form.
	TextFieldOutlined TextFieldStyle = "o"

	// TextFieldBasic removes as much as decorations as possible. There may be limitations based on the platform.
	// Note, that an implementation is allowed to ignore leading, trailing, supporting and errorText for this mode.
	// It may serve as a building-block for custom fields.
	TextFieldBasic TextFieldStyle = "b"
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
	style           ora.TextFieldStyle
	disableDebounce bool
	debounceTime    time.Duration
	invisible       bool
	frame           ora.Frame
	lines           int
	keyboardOptions TKeyboardOptions
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
	strState := core.StateOf[string](state.Window(), state.ID()+".int64").From(func() string {
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

// FloatField is just a TextField using the according keyboard hints. Remember, that these IME hints are no guarantees
// and a user may enter non-integer stuff anyway. However, any
// incompatible inputs are ignored and the given int-state is just a kind of view on top of the string state.
func FloatField(label string, value float64, state *core.State[float64]) TTextField {
	strState := core.StateOf[string](state.Window(), state.ID()+".float64").From(func() string {
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

	return TextField(label, strState.Get()).InputValue(strState).KeyboardType(KeyboardFloat)
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

// Style sets the wanted style. If empty, [ora.TextFieldOutlined] is applied.
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
	c.frame = frame.ora()
	return c
}

// Lines are by default at 0 and enforces a single line text field. Otherwise, a text area is created.
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

func (c TTextField) Render(ctx core.RenderContext) ora.Component {

	return ora.TextField{
		Type:            ora.TextFieldT,
		Label:           c.label,
		SupportingText:  c.supportingText,
		ErrorText:       c.errorText,
		Value:           c.value,
		InputValue:      c.inputValue.Ptr(),
		Disabled:        c.disabled,
		Leading:         render(ctx, c.leading),
		Trailing:        render(ctx, c.trailing),
		Style:           c.style,
		DebounceTime:    c.debounceTime,
		DisableDebounce: c.disableDebounce,
		Invisible:       c.invisible,
		Frame:           c.frame,
		Lines:           c.lines,
		KeyboardOptions: c.keyboardOptions.ora(),
	}
}
