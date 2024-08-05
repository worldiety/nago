package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
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
}

func TextField(label string, value string) TTextField {
	c := TTextField{
		label: label,
		value: value,
	}

	return c
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

func (c TTextField) Frame(frame ora.Frame) TTextField {
	c.frame = frame
	return c
}

// Lines are by default at 0 and enforces a single line text field. Otherwise, a text area is created.
func (c TTextField) Lines(lines int) TTextField {
	c.lines = lines
	return c
}

func (c TTextField) Visible(v bool) TTextField {
	c.invisible = !v
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
	}
}
