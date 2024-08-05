package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TButton struct {
	title    string
	preIcon  ora.SVG
	postIcon ora.SVG
	frame    Frame
	preset   StylePreset
	action   func()
}

// PrimaryButton uses an internal preset to represent a primary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func PrimaryButton(action func()) TButton {
	return TButton{action: action, preset: StyleButtonPrimary}
}

// Secondary uses an internal preset to represent a secondary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func Secondary(action func()) TButton {
	return TButton{action: action, preset: StyleButtonSecondary}
}

// Tertiary uses an internal preset to represent a tertiary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func Tertiary(action func()) TButton {
	return TButton{action: action, preset: StyleButtonTertiary}
}

func (c TButton) Title(text string) TButton {
	c.title = text
	return c
}

func (c TButton) PreIcon(svg core.SVG) TButton {
	c.preIcon = ora.SVG(svg)
	return c
}

func (c TButton) PostIcon(svg core.SVG) TButton {
	c.postIcon = ora.SVG(svg)
	return c
}

func (c TButton) Frame(frame Frame) TButton {
	c.frame = frame
	return c
}

func (c TButton) Render(context core.RenderContext) ora.Component {
	return HStack(
		If(len(c.preIcon) != 0, Image().Embed(c.preIcon).Frame(Frame{}.Size(L16, L16))),
		If(c.title != "", Text(c.title)),
		If(len(c.postIcon) != 0, Image().Embed(c.postIcon).Frame(Frame{}.Size(L16, L16))),
	).Gap(L4).
		Action(c.action).
		StylePreset(c.preset).
		Frame(c.frame).
		Render(context)
}
