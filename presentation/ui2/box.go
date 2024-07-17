package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type alignedComponent struct {
	Component core.View
	Alignment ora.Alignment
}

type BoxLayout struct {
	Top            core.View
	Center         core.View
	Bottom         core.View
	Leading        core.View
	Trailing       core.View
	TopLeading     core.View
	TopTrailing    core.View
	BottomLeading  core.View
	BottomTrailing core.View
}

type TBox struct {
	children           []alignedComponent
	backgroundColor    ora.Color
	frame              ora.Frame
	padding            ora.Padding
	font               ora.Font
	border             ora.Border
	accessibilityLabel string
	invisible          bool
}

func Box(layout BoxLayout) TBox {
	c := TBox{
		children: make([]alignedComponent, 0, 9),
	}

	if layout.Center != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Center,
			Alignment: ora.Center,
		})
	}

	if layout.Top != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Top,
			Alignment: ora.Top,
		})
	}

	if layout.Bottom != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Bottom,
			Alignment: ora.Bottom,
		})
	}

	if layout.Leading != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Leading,
			Alignment: ora.Leading,
		})
	}

	if layout.Trailing != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Trailing,
			Alignment: ora.Trailing,
		})
	}

	if layout.TopLeading != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.TopLeading,
			Alignment: ora.TopLeading,
		})
	}

	if layout.BottomLeading != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.BottomLeading,
			Alignment: ora.BottomLeading,
		})
	}

	if layout.TopTrailing != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.TopTrailing,
			Alignment: ora.TopTrailing,
		})
	}

	if layout.BottomTrailing != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.BottomTrailing,
			Alignment: ora.BottomTrailing,
		})
	}

	return c
}

func (c TBox) Padding(p ora.Padding) core.DecoredView {
	c.padding = p
	return c
}

func (c TBox) BackgroundColor(backgroundColor ora.Color) core.DecoredView {
	c.backgroundColor = backgroundColor
	return c
}

func (c TBox) Frame(fr ora.Frame) core.DecoredView {
	c.frame = fr
	return c
}

func (c TBox) Font(font ora.Font) core.DecoredView {
	c.font = font
	return c
}

func (c TBox) Border(border ora.Border) core.DecoredView {
	c.border = border
	return c
}

func (c TBox) Visible(visible bool) core.DecoredView {
	c.invisible = !visible
	return c
}

func (c TBox) AccessibilityLabel(label string) core.DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TBox) Render(ctx core.RenderContext) ora.Component {
	var tmp []ora.AlignedComponent
	for _, child := range c.children {
		tmp = append(tmp, ora.AlignedComponent{
			Component: child.Component.Render(ctx),
			Alignment: child.Alignment,
		})
	}

	return ora.Box{
		Type:            ora.BoxT,
		Children:        tmp,
		Frame:           c.frame,
		BackgroundColor: c.backgroundColor,
		Padding:         c.padding,
		Border:          c.border,
	}
}
