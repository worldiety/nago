package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

// Str represents a constant string without any events or styling. Just text for performance sake.
// It has around 1/10 of the overhead of the Text component.
type Str struct {
	id         ora.Ptr
	value      String
	properties []core.Property
}

func (c *Str) ID() ora.Ptr {
	return c.id
}

func NewStr(v string) *Str {
	c := &Str{
		id: nextPtr(),
	}

	c.value = NewShared[string]("value")

	c.properties = []core.Property{c.value}
	c.value.Set(v)

	return c
}

func (c *Str) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Str) Render() ora.Component {
	return ora.Str{
		Type:  ora.StrT,
		Value: c.value.v,
	}
}
