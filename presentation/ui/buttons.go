package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"runtime/debug"
)

type TButton struct {
	title              string
	accessibilityLabel string
	preIcon            proto.SVG
	postIcon           proto.SVG
	frame              Frame
	preset             StylePreset
	font               Font
	action             func()
	trace              string
	invisible          bool
	disabled           bool
	id                 string
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

// Enabled has only an effect for StylePreset otherwise it is ignored.
func (c TButton) Enabled(b bool) TButton {
	c.disabled = !b
	return c
}

func (c TButton) Title(text string) TButton {
	c.title = text
	return c
}

func (c TButton) Visible(b bool) TButton {
	c.invisible = !b
	return c
}

func (c TButton) Font(font Font) TButton {
	c.font = font
	return c
}

func (c TButton) AccessibilityLabel(label string) TButton {
	c.accessibilityLabel = label
	return c
}

func (c TButton) PreIcon(svg core.SVG) TButton {
	c.preIcon = proto.SVG(svg)
	return c
}

func (c TButton) PostIcon(svg core.SVG) TButton {
	c.postIcon = proto.SVG(svg)
	return c
}

func (c TButton) Frame(frame Frame) TButton {
	c.frame = frame
	return c
}

func (c TButton) ID(id string) TButton {
	c.id = id
	return c
}

func (c TButton) Render(context core.RenderContext) proto.Component {
	alabel := c.accessibilityLabel
	if alabel == "" {
		alabel = c.title
	}

	if core.Debug {
		if alabel == "" {
			panic(fmt.Errorf("the ora guidelines forbid buttons without accessibility label, allocated here: %s", c.trace))
		}
	}

	return HStack(
		If(len(c.preIcon) != 0, Image().Embed(c.preIcon).Frame(Frame{}.Size(L16, L16))),
		If(c.title != "", Text(c.title).Font(c.font)),
		If(len(c.postIcon) != 0, Image().Embed(c.postIcon).Frame(Frame{}.Size(L16, L16))),
	).Gap(L4).
		ID(c.id).
		Enabled(!c.disabled).
		Action(c.action).
		StylePreset(c.preset).
		Frame(c.frame).
		Visible(!c.invisible).
		AccessibilityLabel(alabel). // this is redundant and requires the text twice, however we are "just" an hstack
		Render(context)
}
