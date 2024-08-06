package uilegacy

import (
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type VStack struct {
	Bla             iter.Seq[core.View]
	Children        []core.View //iter.Seq[core.View]
	alignment       ora.Alignment
	backgroundColor ora.Color
	frame           ora.Frame
	gap             ora.Length
	padding         ora.Padding
	with            func(stack *VStack)

	blub *SharedList[core.View]
}

func NewVStackF(with func(hstack *VStack)) func() core.View {
	return func() core.View {
		return NewVStack(with)
	}
}

func NewVStack(with func(hstack *VStack)) *VStack {
	c := &VStack{
		blub: NewSharedList[core.View]("asd"),
	}

	c.alignment = "" // if nothing is defined, ora.Center must be applied by renderer
	c.with = with

	return c
}

func (c *VStack) Gap() ora.Length {
	return c.gap
}

func (c *VStack) SetGap(gap ora.Length) {
	c.gap = gap
}

func (c *VStack) BackgroundColor() ora.Color {
	return c.backgroundColor
}

func (c *VStack) SetBackgroundColor(backgroundColor ora.Color) {
	c.backgroundColor = backgroundColor
}

func (c *VStack) Alignment() ora.Alignment {
	return c.alignment
}

func (c *VStack) SetAlignment(alignment ora.Alignment) {
	c.alignment = alignment
}

func (c *VStack) ID() ora.Ptr {
	return 0
}

func (c *VStack) Properties(yield func(core.Property) bool) {
	if !c.blub.frozen {
		c.blub.values = c.Children
	}
	yield(c.blub)
}

func (c *VStack) Frame() *ora.Frame {
	return &c.frame
}

func (c *VStack) Append(...core.View) {

}

func (c *VStack) Render() ora.Component {
	if c.with != nil {
		c.with(c)
	}

	return ora.VStack{
		Type:            ora.VStackT,
		Children:        renderComponents(slices.Values(c.Children)),
		Frame:           c.frame,
		Alignment:       c.alignment,
		BackgroundColor: c.backgroundColor,
		Gap:             c.gap,
		Padding:         c.padding,
	}
}
