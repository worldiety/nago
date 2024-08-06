package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"time"
)

type TPasswordField struct {
	label           string
	value           string
	inputValue      *core.State[string]
	supportingText  string
	errorText       string
	disabled        bool
	style           ora.TextFieldStyle
	disableDebounce bool
	debounceTime    time.Duration
	invisible       bool
	frame           ora.Frame
	lines           int
}

func PasswordField(label string) TPasswordField {
	c := TPasswordField{
		label: label,
	}

	return c
}

func (c TPasswordField) SupportingText(text string) TPasswordField {
	c.supportingText = text
	return c
}

func (c TPasswordField) ErrorText(text string) TPasswordField {
	c.errorText = text
	return c
}

// Style sets the wanted style. If empty, [ora.TextFieldOutlined] is applied.
func (c TPasswordField) Style(s TextFieldStyle) TPasswordField {
	c.style = s.ora()
	return c
}

// DebounceTime sets a custom debouncing time when entering text. By default, this is 500ms and always applied.
// You can disable debouncing, but be very careful with that, as it may break your server, the client or network.
func (c TPasswordField) DebounceTime(d time.Duration) TPasswordField {
	c.debounceTime = d
	return c
}

// Debounce is enabled by default. See also DebounceTime.
func (c TPasswordField) Debounce(enabled bool) TPasswordField {
	c.disableDebounce = !enabled
	return c
}

func (c TPasswordField) Label(label string) {
	c.label = label
}

func (c TPasswordField) InputValue(input *core.State[string]) TPasswordField {
	c.inputValue = input
	return c
}

func (c TPasswordField) Disabled(disabled bool) TPasswordField {
	c.disabled = disabled
	return c
}

func (c TPasswordField) Frame(frame ora.Frame) TPasswordField {
	c.frame = frame
	return c
}

// Lines are by default at 0 and enforces a single line text field. Otherwise, a text area is created.
func (c TPasswordField) Lines(lines int) TPasswordField {
	c.lines = lines
	return c
}

func (c TPasswordField) Visible(v bool) TPasswordField {
	c.invisible = !v
	return c
}

func (c TPasswordField) Render(ctx core.RenderContext) ora.Component {

	return ora.PasswordField{
		Type:            ora.PasswordFieldT,
		Label:           c.label,
		SupportingText:  c.supportingText,
		ErrorText:       c.errorText,
		Value:           c.value,
		InputValue:      c.inputValue.Ptr(),
		Disabled:        c.disabled,
		Style:           c.style,
		DebounceTime:    c.debounceTime,
		DisableDebounce: c.disableDebounce,
		Invisible:       c.invisible,
		Frame:           c.frame,
		Lines:           c.lines,
	}
}
