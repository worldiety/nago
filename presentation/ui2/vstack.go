package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TVStack struct {
	children               []core.View
	alignment              ora.Alignment
	backgroundColor        ora.Color
	hoveredBackgroundColor ora.Color
	pressedBackgroundColor ora.Color
	focusedBackgroundColor ora.Color
	frame                  ora.Frame
	gap                    ora.Length
	padding                ora.Padding
	border                 ora.Border
	hoveredBorder          ora.Border
	focusedBorder          ora.Border
	pressedBorder          ora.Border

	invisible bool
	font      ora.Font
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	accessibilityLabel string
	action             func()
}

func VStack(children ...core.View) TVStack {
	c := TVStack{
		children: children,
	}
	return c
}

func (c TVStack) Gap(gap ora.Length) TVStack {
	c.gap = gap
	return c
}

func (c TVStack) BackgroundColor(backgroundColor ora.Color) core.DecoredView {
	c.backgroundColor = backgroundColor
	return c
}

func (c TVStack) HoveredBackgroundColor(backgroundColor ora.Color) TVStack {
	c.hoveredBackgroundColor = backgroundColor
	return c
}

func (c TVStack) PressedBackgroundColor(backgroundColor ora.Color) TVStack {
	c.pressedBackgroundColor = backgroundColor
	return c
}

func (c TVStack) FocusedBackgroundColor(backgroundColor ora.Color) TVStack {
	c.focusedBackgroundColor = backgroundColor
	return c
}

func (c TVStack) Action(f func()) TVStack {
	c.action = f
	return c
}

func (c TVStack) Alignment(alignment ora.Alignment) TVStack {
	c.alignment = alignment
	return c
}

func (c TVStack) Font(font ora.Font) TVStack {
	c.font = font
	return c
}

func (c TVStack) Frame(f ora.Frame) core.DecoredView {
	c.frame = f
	return c
}

func (c TVStack) Padding(padding ora.Padding) core.DecoredView {
	c.padding = padding
	return c
}

func (c TVStack) Border(border ora.Border) core.DecoredView {
	c.border = border
	return c
}

func (c TVStack) HoveredBorder(border ora.Border) TVStack {
	c.hoveredBorder = border
	return c
}

func (c TVStack) PressedBorder(border ora.Border) TVStack {
	c.pressedBorder = border
	return c
}

func (c TVStack) FocusedBorder(border ora.Border) TVStack {
	c.focusedBorder = border
	return c
}

func (c TVStack) Visible(visible bool) core.DecoredView {
	c.invisible = !visible
	return c
}

func (c TVStack) AccessibilityLabel(label string) core.DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TVStack) Render(ctx core.RenderContext) ora.Component {

	return ora.VStack{
		Type:               ora.VStackT,
		Children:           renderComponents(ctx, c.children),
		Frame:              c.frame,
		Alignment:          c.alignment,
		BackgroundColor:    c.backgroundColor,
		Gap:                c.gap,
		Padding:            c.padding,
		Border:             c.border,
		AccessibilityLabel: c.accessibilityLabel,
		Invisible:          c.invisible,
		Font:               c.font,

		HoveredBackgroundColor: c.hoveredBackgroundColor,
		PressedBackgroundColor: c.pressedBackgroundColor,
		FocusedBackgroundColor: c.focusedBackgroundColor,
		HoveredBorder:          c.hoveredBorder,
		FocusedBorder:          c.focusedBorder,
		PressedBorder:          c.pressedBorder,
		Action:                 ctx.MountCallback(c.action),
	}
}
