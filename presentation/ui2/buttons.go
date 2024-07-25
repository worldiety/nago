package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TButton struct {
	title    string
	preIcon  ora.SVG
	postIcon ora.SVG
	frame    ora.Frame
	preset   ora.StylePreset
	action   func()
}

// PrimaryButton uses an internal preset to represent a primary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func PrimaryButton(action func()) TButton {
	return TButton{action: action, preset: ora.StyleButtonPrimary}
}

// Secondary uses an internal preset to represent a secondary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func Secondary(action func()) TButton {
	return TButton{action: action, preset: ora.StyleButtonSecondary}
}

// Tertiary uses an internal preset to represent a tertiary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func Tertiary(action func()) TButton {
	return TButton{action: action, preset: ora.StyleButtonTertiary}
}

func (c TButton) Title(text string) TButton {
	c.title = text
	return c
}

func (c TButton) PreIcon(svg ora.SVG) TButton {
	c.preIcon = svg
	return c
}

func (c TButton) PostIcon(svg ora.SVG) TButton {
	c.postIcon = svg
	return c
}

func (c TButton) Frame(frame ora.Frame) TButton {
	c.frame = frame
	return c
}

func (c TButton) Render(context core.RenderContext) ora.Component {
	return HStack(
		If(len(c.preIcon) != 0, Image().Embed(c.preIcon).Frame(ora.Frame{}.Size(ora.L16, ora.L16))),
		If(c.title != "", Text(c.title)),
		If(len(c.postIcon) != 0, Image().Embed(c.postIcon).Frame(ora.Frame{}.Size(ora.L16, ora.L16))),
	).Gap(ora.L4).
		Action(c.action).
		StylePreset(c.preset).
		Frame(c.frame).
		Render(context)
}
