package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"time"
)

type ViewTextField struct {
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
	with                   func(field *ViewTextField)
}

func TextField(text *core.State[string], with func(tField *ViewTextField)) *ViewTextField {
	c := &ViewTextField{
		text: text,
	}

	c.debounceTime = time.Millisecond * 500
	c.with = with

	return c
}

func (c *ViewTextField) OnTextChanged(f func()) {
	c.onTextChanged = f
}

func (c *ViewTextField) OnDebouncedTextChanged(f func()) {
	c.onDebouncedTextChanged = f
}

func (c *ViewTextField) DebounceTime(d time.Duration) {
	c.debounceTime = d
}

func (c *ViewTextField) Placeholder(p string) {
	c.placeholder = p
}

func (c *ViewTextField) Label(label string) {
	c.label = label
}

func (c *ViewTextField) Hint(hint string) {
	c.hint = hint
}

func (c *ViewTextField) Help(help string) {
	c.help = help
}

func (c *ViewTextField) Error(error string) {
	c.error = error
}

func (c *ViewTextField) Disabled(disabled bool) {
	c.disabled = disabled
}

func (c *ViewTextField) Frame(frame ora.Frame) {
	c.frame = frame
}

func (c *ViewTextField) Visible(v bool) {
	c.invisible = !v
}

func (c *ViewTextField) Render(ctx core.RenderContext) ora.Component {
	if c.with != nil {
		c.with(c)
	}

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
