package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"runtime/debug"
)

type TButton struct {
	title              string
	accessibilityLabel string
	preIcon            ora.SVG
	postIcon           ora.SVG
	frame              Frame
	preset             StylePreset
	action             func()
	trace              string
}

// PrimaryButton uses an internal preset to represent a primary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func PrimaryButton(action func()) TButton {
	return initButton(action, StyleButtonPrimary)
}

// SecondaryButton uses an internal preset to represent a secondary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func SecondaryButton(action func()) TButton {
	return initButton(action, StyleButtonSecondary)
}

// TertiaryButton uses an internal preset to represent a tertiary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func TertiaryButton(action func()) TButton {
	return initButton(action, StyleButtonTertiary)
}

func initButton(action func(), preset StylePreset) TButton {
	btn := TButton{action: action, preset: preset}
	if core.Debug {
		btn.trace = string(debug.Stack())
	}
	return btn
}

func (c TButton) Title(text string) TButton {
	c.title = text
	return c
}

func (c TButton) AccessibilityLabel(label string) TButton {
	c.accessibilityLabel = label
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
	alabel := c.title
	if alabel == "" {
		alabel = c.accessibilityLabel
	}

	if core.Debug {
		if alabel == "" {
			panic(fmt.Errorf("the ora guidelines forbid buttons without accessibility label, allocated here: %s", c.trace))
		}
	}

	return HStack(
		If(len(c.preIcon) != 0, Image().Embed(c.preIcon).Frame(Frame{}.Size(L16, L16))),
		If(c.title != "", Text(c.title)),
		If(len(c.postIcon) != 0, Image().Embed(c.postIcon).Frame(Frame{}.Size(L16, L16))),
	).Gap(L4).
		Action(c.action).
		StylePreset(c.preset).
		Frame(c.frame).
		AccessibilityLabel(alabel). // this is redundant and requires the text twice, however we are "just" an hstack
		Render(context)
}
