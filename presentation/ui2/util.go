package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

func renderComponents(ctx core.RenderContext, c []core.View) []ora.Component {
	res := make([]ora.Component, 0, len(c))
	for _, component := range c {
		if component == nil {
			continue
		}
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

func If(b bool, view core.View) core.View {
	if b {
		return view
	}

	// TODO implement non-nil view
	return nil
}
