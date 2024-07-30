package ui

import (
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

func render(ctx core.RenderContext, c core.View) ora.Component {
	if c == nil {
		return nil
	}

	return c.Render(ctx)
}

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

// If conditionally returns the view or nil. This can be used as a kind of inline ternary operator
func If(b bool, t core.View) core.View {
	return IfElse(b, t, nil)
}

// IfElse conditionally returns one or the other view. This can be used as a kind of inline ternary operator.
// This is intentionally not generic, because the zero value of our value view types are not nil and therefore
// we cannot distinguish between an absent or zero value view (e.g. an empty text)
func IfElse(b bool, ifTrue, ifFalse core.View) core.View {
	if b {
		return ifTrue
	}

	return ifFalse
}

// With can be used to simply intercept a builder chain without resorting to local variables.
func With[T any](t T, with func(T) T) T {
	if with == nil {
		return t
	}

	return with(t)
}

func Each[T any](seq iter.Seq[T], m func(T) core.View) []core.View {
	var res []core.View
	seq(func(t T) bool {
		res = append(res, m(t))
		return true
	})

	return res
}

func Each2[K, V any](seq iter.Seq2[K, V], m func(K, V) core.View) []core.View {
	var res []core.View
	seq(func(k K, v V) bool {
		res = append(res, m(k, v))
		return true
	})

	return res
}
