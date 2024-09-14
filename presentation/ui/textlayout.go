package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TTextLayout struct {
	children        []core.View
	alignment       ora.TextAlignment
	backgroundColor ora.Color
	frame           ora.Frame
	gap             ora.Length
	padding         ora.Padding
	border          ora.Border
	invisible       bool
	font            ora.Font
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	accessibilityLabel string
	action             func()
}

func (c TTextLayout) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

func (c TTextLayout) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TTextLayout) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

// TextLayout performs an inline layouting of multiple text elements. The alignment properties of each
// Text are ignored. Any implementation must support an arbitrary amount of text elements with different
// font settings. However, implementations are also open to support images and any other views, as long
// as they can be rendered inline.
func TextLayout(views ...core.View) TTextLayout {
	return TTextLayout{
		children: views,
	}
}

func (c TTextLayout) Action(f func()) TTextLayout {
	c.action = f
	return c
}

func (c TTextLayout) Alignment(alignment TextAlignment) TTextLayout {
	c.alignment = ora.TextAlignment(alignment)
	return c
}

func (c TTextLayout) Font(font Font) TTextLayout {
	c.font = font.ora()
	return c
}

func (c TTextLayout) Frame(f Frame) DecoredView {
	c.frame = f.ora()
	return c
}

func (c TTextLayout) FullWidth() TTextLayout {
	c.frame.Width = "100%"
	return c
}

func (c TTextLayout) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

func (c TTextLayout) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TTextLayout) Render(ctx core.RenderContext) ora.Component {

	return ora.TextLayout{
		Type:               ora.TextLayoutT,
		Children:           renderComponents(ctx, c.children),
		Frame:              c.frame,
		TextAlignment:      c.alignment,
		BackgroundColor:    c.backgroundColor,
		Padding:            c.padding,
		AccessibilityLabel: c.accessibilityLabel,
		Invisible:          c.invisible,
		Font:               c.font,
		Border:             c.border,

		Action: ctx.MountCallback(c.action),
	}
}
