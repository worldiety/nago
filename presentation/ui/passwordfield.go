package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"time"
)

type TPasswordField struct {
	label               string
	value               string
	inputValue          *core.State[string]
	supportingText      string
	errorText           string
	disabled            bool
	style               ora.TextFieldStyle
	disableDebounce     bool
	disableAutocomplete bool
	debounceTime        time.Duration
	invisible           bool
	frame               ora.Frame
	lines               int
}

// PasswordField represents a secret entered by the user.
// It is very important for the security of your implementation, that you
// must not expose any secret through the value after the user has finished the secret step.
// So, consider the following situations:
//   - a user wants to change his password and fills out 3 password fields. Thus, you don't want, that the user requires
//     to enter his secret any time, e.g. if the new password does not match the guidelines
//   - by design, Nago requires to fill in the value whatever that was, so that it will not get lost between render
//     cycles.
//   - by security, you should never store any password, neither in plain text and also never encrypted and never using
//     just a simple cryptographic hash. Use a password derivation function instead, like argonid or others.
//   - if you need to store a secret in plain text, e.g. like an API token, you must not show that later again to
//     the user, after the insertion phase of you form flow is over.
func PasswordField(label string, value string) TPasswordField {
	c := TPasswordField{
		label: label,
		value: value,
	}

	return c
}

func (c TPasswordField) AutoComplete(autoComplete bool) TPasswordField {
	c.disableAutocomplete = !autoComplete
	return c
}

func (c TPasswordField) Padding(padding Padding) DecoredView {
	return c // TODO
}

func (c TPasswordField) Border(border Border) DecoredView {
	return c // TODO
}

func (c TPasswordField) AccessibilityLabel(label string) DecoredView {
	return c // TODO
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

func (c TPasswordField) Frame(frame Frame) DecoredView {
	c.frame = frame.ora()
	return c
}

// Lines are by default at 0 and enforces a single line text field. Otherwise, a text area is created.
func (c TPasswordField) Lines(lines int) TPasswordField {
	c.lines = lines
	return c
}

func (c TPasswordField) Visible(v bool) DecoredView {
	c.invisible = !v
	return c
}

func (c TPasswordField) Render(ctx core.RenderContext) ora.Component {

	return ora.PasswordField{
		Type:                ora.PasswordFieldT,
		Label:               c.label,
		SupportingText:      c.supportingText,
		ErrorText:           c.errorText,
		Value:               c.value,
		InputValue:          c.inputValue.Ptr(),
		Disabled:            c.disabled,
		Style:               c.style,
		DebounceTime:        c.debounceTime,
		DisableDebounce:     c.disableDebounce,
		Invisible:           c.invisible,
		Frame:               c.frame,
		Lines:               c.lines,
		DisableAutocomplete: c.disableAutocomplete,
	}
}
