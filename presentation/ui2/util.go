package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

func renderComponents(ctx core.RenderContext, c []core.View) []ora.Component {
	res := make([]ora.Component, 0, len(c))
	for _, component := range c {
		res = append(res, component.Render(ctx))
	}

	return res
}

func propertyOf[T any](ctx core.RenderContext, s *core.State[T]) ora.Property[T] {
	return ora.Property[T]{
		Ptr:   s.Ptr(),
		Value: s.Get(),
	}
}

func reset(c *[]core.View) {
	if c == nil {
		return
	}

	tmp := *c
	for i := range tmp {
		tmp[i] = nil // let the gc free slice elements
	}

	tmp = tmp[:0]
	*c = tmp
}
