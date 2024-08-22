package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"runtime/debug"
	"strings"
)

type THStack struct {
	children               []core.View
	alignment              ora.Alignment
	backgroundColor        ora.Color
	hoveredBackgroundColor ora.Color
	pressedBackgroundColor ora.Color
	focusedBackgroundColor ora.Color
	frame                  ora.Frame
	gap                    ora.Length
	padding                ora.Padding
	font                   ora.Font
	border                 ora.Border
	hoveredBorder          ora.Border
	focusedBorder          ora.Border
	pressedBorder          ora.Border
	accessibilityLabel     string
	invisible              bool
	action                 func()
	stylePreset            ora.StylePreset
	originTrace            string
	wrap                   bool
}

// HStack is a container, in which the given children will be layout in a row according to the applied
// alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
// you need to apply additional padding.
func HStack(children ...core.View) *THStack {
	c := &THStack{
		children: children,
	}

	if core.Debug {
		c.originTrace = strings.Split(string(debug.Stack()), "\n")[6]
	}

	return c
}

func (c THStack) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

func (c THStack) Gap(gap Length) THStack {
	c.gap = gap.ora()
	return c
}

// Wrap tries to reproduce the flex-box wrap behavior. This means, that if the HStack has a limited width,
// it must create multiple rows to place its children. Note, that the text layout behavior is unspecified
// (it may layout without word-wrap or use some sensible defaults). Each row and each element may have its own
// custom size, so this must not use a grid-like layouting.
func (c THStack) Wrap(wrap bool) THStack {
	c.wrap = wrap
	return c
}

func (c THStack) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c THStack) HoveredBackgroundColor(backgroundColor Color) THStack {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

func (c THStack) PressedBackgroundColor(backgroundColor Color) THStack {
	c.pressedBackgroundColor = backgroundColor.ora()
	return c
}

func (c THStack) FocusedBackgroundColor(backgroundColor Color) THStack {
	c.focusedBackgroundColor = ora.Color(backgroundColor)
	return c
}

func (c THStack) Alignment(alignment Alignment) THStack {
	c.alignment = alignment.ora()
	return c
}

func (c THStack) Frame(fr Frame) DecoredView {
	c.frame = fr.ora()
	return c
}

func (c THStack) FullWidth() THStack {
	c.frame.Width = "100%"
	return c
}

func (c THStack) Font(font Font) DecoredView {
	c.font = font.ora()
	return c
}

func (c THStack) StylePreset(preset StylePreset) THStack {
	c.stylePreset = preset.ora()
	return c
}

func (c THStack) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

func (c THStack) HoveredBorder(border Border) THStack {
	c.hoveredBorder = border.ora()
	return c
}

func (c THStack) PressedBorder(border Border) THStack {
	c.pressedBorder = border.ora()
	return c
}

func (c THStack) FocusedBorder(border Border) THStack {
	c.focusedBorder = border.ora()
	return c
}

func (c THStack) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c THStack) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c THStack) Action(f func()) THStack {
	c.action = f
	return c
}

func (c THStack) Render(ctx core.RenderContext) ora.Component {
	ptr := ctx.MountCallback(c.action)
	if core.Debug {
		fmt.Printf("hstack got %d @%s\n", ptr, c.originTrace)
	}
	return ora.HStack{
		Type:               ora.HStackT,
		Children:           renderComponents(ctx, c.children),
		Gap:                c.gap,
		Frame:              c.frame,
		Alignment:          c.alignment,
		BackgroundColor:    c.backgroundColor,
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
		Action:                 ptr,
		Wrap:                   c.wrap,

		StylePreset: c.stylePreset,
	}
}
