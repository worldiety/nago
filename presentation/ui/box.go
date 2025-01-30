package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type alignedComponent struct {
	Component core.View
	Alignment proto.Alignment
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
	backgroundColor    proto.Color
	frame              proto.Frame
	padding            proto.Padding
	font               proto.Font
	border             proto.Border
	accessibilityLabel string
	invisible          bool
}

func BoxAlign(alignment Alignment, child core.View) TBox {
	var bl BoxLayout
	switch alignment {
	case Top:
		bl.Top = child
	case Center:
		bl.Center = child
	case Bottom:
		bl.Bottom = child
	case Leading:
		bl.Leading = child
	case Trailing:
		bl.Trailing = child
	case TopLeading:
		bl.TopLeading = child
	case TopTrailing:
		bl.TopTrailing = child
	case BottomLeading:
		bl.BottomLeading = child
	case BottomTrailing:
		bl.BottomTrailing = child

	default:
		bl.Center = child
	}

	return Box(bl)
}

// Box is a container, in which the given children will be layout to the according BoxLayout
// rules. Note, that per definition the container clips its children. Thus, if working with shadows,
// you need to apply additional padding. Important: this container requires usually absolute height and width
// attributes and cannot work properly using wrap content semantics, because it intentionally allows overlapping.
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

func (c TBox) FullWidth() TBox {
	c.frame.Width = "100%"
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

func (c TBox) Render(ctx core.RenderContext) core.RenderNode {
	var tmp []proto.AlignedComponent
	for _, child := range c.children {
		tmp = append(tmp, proto.AlignedComponent{
			Component: child.Component.Render(ctx),
			Alignment: child.Alignment,
		})
	}

	return &proto.Box{
		Children:        tmp,
		Frame:           c.frame,
		BackgroundColor: c.backgroundColor,
		Padding:         c.padding,
		Border:          c.border,
	}
}
