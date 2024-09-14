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
	stylePreset            ora.StylePreset

	invisible bool
	font      ora.Font
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	accessibilityLabel string
	action             func()
}

// VStack is a container, in which the given children will be layout in a column according to the applied
// alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
// you need to apply additional padding.
func VStack(children ...core.View) TVStack {
	c := TVStack{
		children: children,
	}
	return c
}

func (c TVStack) Gap(gap Length) TVStack {
	c.gap = gap.ora()
	return c
}

func (c TVStack) StylePreset(preset StylePreset) TVStack {
	c.stylePreset = preset.ora()
	return c
}

func (c TVStack) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TVStack) HoveredBackgroundColor(backgroundColor Color) TVStack {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

func (c TVStack) PressedBackgroundColor(backgroundColor Color) TVStack {
	c.pressedBackgroundColor = backgroundColor.ora()
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

func (c TVStack) Alignment(alignment Alignment) TVStack {
	c.alignment = alignment.ora()
	return c
}

func (c TVStack) Font(font Font) TVStack {
	c.font = font.ora()
	return c
}

func (c TVStack) Frame(f Frame) DecoredView {
	c.frame = f.ora()
	return c
}

func (c TVStack) FullWidth() TVStack {
	c.frame.Width = "100%"
	return c
}

func (c TVStack) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

func (c TVStack) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

func (c TVStack) HoveredBorder(border Border) TVStack {
	c.hoveredBorder = border.ora()
	return c
}

func (c TVStack) PressedBorder(border Border) TVStack {
	c.pressedBorder = border.ora()
	return c
}

func (c TVStack) FocusedBorder(border Border) TVStack {
	c.focusedBorder = border.ora()
	return c
}

func (c TVStack) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TVStack) AccessibilityLabel(label string) DecoredView {
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
		StylePreset:        c.stylePreset,

		HoveredBackgroundColor: c.hoveredBackgroundColor,
		PressedBackgroundColor: c.pressedBackgroundColor,
		FocusedBackgroundColor: c.focusedBackgroundColor,
		HoveredBorder:          c.hoveredBorder,
		FocusedBorder:          c.focusedBorder,
		PressedBorder:          c.pressedBorder,
		Action:                 ctx.MountCallback(c.action),
	}
}
