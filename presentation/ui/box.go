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
			Alignment: Center.ora(),
		})
	}

	if layout.Top != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Top,
			Alignment: Top.ora(),
		})
	}

	if layout.Bottom != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Bottom,
			Alignment: Bottom.ora(),
		})
	}

	if layout.Leading != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Leading,
			Alignment: Leading.ora(),
		})
	}

	if layout.Trailing != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Trailing,
			Alignment: Trailing.ora(),
		})
	}

	if layout.TopLeading != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.TopLeading,
			Alignment: TopLeading.ora(),
		})
	}

	if layout.BottomLeading != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.BottomLeading,
			Alignment: BottomLeading.ora(),
		})
	}

	if layout.TopTrailing != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.TopTrailing,
			Alignment: TopTrailing.ora(),
		})
	}

	if layout.BottomTrailing != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.BottomTrailing,
			Alignment: BottomTrailing.ora(),
		})
	}

	return c
}

func (c TBox) Padding(p Padding) DecoredView {
	c.padding = p.ora()
	return c
}

func (c TBox) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TBox) Frame(fr Frame) DecoredView {
	c.frame = fr.ora()
	return c
}

func (c TBox) Font(font Font) DecoredView {
	c.font = font.ora()
	return c
}

func (c TBox) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

func (c TBox) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TBox) AccessibilityLabel(label string) DecoredView {
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
