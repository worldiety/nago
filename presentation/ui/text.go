package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TextAlignment string

const (
	TextAlignStart   TextAlignment = "s"
	TextAlignEnd     TextAlignment = "e"
	TextAlignCenter  TextAlignment = "c"
	TextAlignJustify TextAlignment = "j"
)

type TText struct {
	content                string
	color                  ora.Color
	backgroundColor        ora.Color
	hoveredBackgroundColor ora.Color
	pressedBackgroundColor ora.Color
	focusedBackgroundColor ora.Color
	font                   ora.Font
	invisible              bool
	onClick                func()
	onHoverStart           func()
	onHoverEnd             func()
	padding                ora.Padding
	frame                  ora.Frame
	border                 ora.Border
	hoveredBorder          ora.Border
	focusedBorder          ora.Border
	pressedBorder          ora.Border
	accessibilityLabel     string
	textAlignment          TextAlignment
	action                 func()
}

func Text(content string) TText {
	return TText{content: content}
}

func (c TText) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

func (c TText) Frame(frame Frame) DecoredView {
	c.frame = frame.ora()
	return c
}

func (c TText) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

func (c TText) HoveredBorder(border Border) TText {
	c.hoveredBorder = border.ora()
	return c
}

func (c TText) PressedBorder(border Border) TText {
	c.pressedBorder = border.ora()
	return c
}

func (c TText) FocusedBorder(border Border) TText {
	c.focusedBorder = border.ora()
	return c
}

func (c TText) TextAlignment(align TextAlignment) TText {
	c.textAlignment = align
	return c
}

func (c TText) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TText) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TText) Font(font Font) TText {
	c.font = font.ora()
	return c
}

func (c TText) Color(color Color) TText {
	c.color = color.ora()
	return c
}

func (c TText) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TText) Action(f func()) TText {
	c.action = f
	return c
}

func (c TText) Render(ctx core.RenderContext) ora.Component {

	return ora.Text{
		Type:               ora.TextT,
		Value:              c.content,
		Color:              c.color,
		BackgroundColor:    c.backgroundColor,
		Font:               c.font,
		OnClick:            ctx.MountCallback(c.onClick),
		OnHoverStart:       ctx.MountCallback(c.onHoverStart),
		OnHoverEnd:         ctx.MountCallback(c.onHoverEnd),
		Invisible:          c.invisible,
		Border:             c.border,
		Padding:            c.padding,
		Frame:              c.frame,
		AccessibilityLabel: c.accessibilityLabel,

		HoveredBackgroundColor: c.hoveredBackgroundColor,
		PressedBackgroundColor: c.pressedBackgroundColor,
		FocusedBackgroundColor: c.focusedBackgroundColor,
		HoveredBorder:          c.hoveredBorder,
		FocusedBorder:          c.focusedBorder,
		PressedBorder:          c.pressedBorder,
		TextAlignment:          ora.TextAlignment(c.textAlignment),
		Action:                 ctx.MountCallback(c.action),
	}
}
