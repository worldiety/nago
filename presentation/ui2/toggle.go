package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TToggle struct {
	value      bool
	inputValue *core.State[bool]
	disabled   bool
	invisible  bool
}

func Toggle(checked bool) *TToggle {
	c := &TToggle{
		value: checked,
	}

	return c
}

func (c TToggle) InputChecked(input *core.State[bool]) TToggle {
	c.inputValue = input
	return c
}

func (c TToggle) Disabled(disabled bool) TToggle {
	c.disabled = disabled
	return c
}

func (c TToggle) Visible(v bool) TToggle {
	c.invisible = !v
	return c
}

func (c TToggle) Render(ctx core.RenderContext) ora.Component {

	return ora.Toggle{
		Type:       ora.ToggleT,
		Value:      c.value,
		InputValue: c.inputValue.Ptr(),
		Disabled:   c.disabled,
		Invisible:  c.invisible,
	}
}
