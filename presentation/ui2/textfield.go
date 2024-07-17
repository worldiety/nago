package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"time"
)

type TTextField struct {
	label                  string
	text                   *core.State[string]
	placeholder            string
	hint                   string
	help                   string
	error                  string
	disabled               bool
	simple                 bool
	invisible              bool
	onTextChanged          func()
	onDebouncedTextChanged func()
	debounceTime           time.Duration
	frame                  ora.Frame
}

func TextField(label string, text *core.State[string]) *TTextField {
	c := &TTextField{
		text:  text,
		label: label,
	}

	c.debounceTime = time.Millisecond * 500

	return c
}

func (c *TTextField) OnTextChanged(f func()) {
	c.onTextChanged = f
}

func (c *TTextField) OnDebouncedTextChanged(f func()) {
	c.onDebouncedTextChanged = f
}

func (c *TTextField) DebounceTime(d time.Duration) {
	c.debounceTime = d
}

func (c *TTextField) Placeholder(p string) {
	c.placeholder = p
}

func (c *TTextField) Label(label string) {
	c.label = label
}

func (c *TTextField) Hint(hint string) {
	c.hint = hint
}

func (c *TTextField) Help(help string) {
	c.help = help
}

func (c *TTextField) Error(error string) {
	c.error = error
}

func (c *TTextField) Disabled(disabled bool) {
	c.disabled = disabled
}

func (c *TTextField) Frame(frame ora.Frame) {
	c.frame = frame
}

func (c *TTextField) Visible(v bool) {
	c.invisible = !v
}

func (c *TTextField) Render(ctx core.RenderContext) ora.Component {

	return ora.TextField{
		Type:                   ora.TextFieldT,
		Label:                  c.label,
		Hint:                   c.hint,
		Help:                   c.help,
		Error:                  c.error,
		Text:                   propertyOf(ctx, c.text),
		Placeholder:            c.placeholder,
		Disabled:               c.disabled,
		Simple:                 c.simple,
		Invisible:              c.invisible,
		DebounceTime:           c.debounceTime,
		OnDebouncedTextChanged: ctx.MountCallback(c.onDebouncedTextChanged),
		OnTextChanged:          ctx.MountCallback(c.onTextChanged),
		Frame:                  c.frame,
	}
}
