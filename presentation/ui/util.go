// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"iter"
	"slices"
)

func render(ctx core.RenderContext, c core.View) proto.Component {
	if c == nil {
		return nil
	}

	return c.Render(ctx)
}

func renderComponents(ctx core.RenderContext, c []core.View) []proto.Component {
	res := make([]proto.Component, 0, len(c))
	for _, component := range c {
		if component == nil {
			continue
		}
		v := component.Render(ctx)
		if v != nil {
			res = append(res, v)
		}

	}

	return res
}

// If conditionally returns the view or nil. This can be used as a kind of inline ternary operator
func If(b bool, t core.View) core.View {
	return IfElse(b, t, nil)
}

// IfFunc conditionally returns the view or nil. This can be used as a kind of inline ternary operator
func IfFunc(b bool, fn func() core.View) core.View {
	return IfElse(b, Lazy(fn), nil)
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

/*// Yield is a shortcut of iter.Collect(iter.Seq(func(core.View))
func Yield[T any](seq iter.Seq[T]) []T {
	var res []T
	seq(func(view T) bool {
		res = append(res, view)
		return true
	})

	return res
}*/

func ForEach[T any, V any](seq []T, m func(T) V, more ...V) []V {
	return Each(slices.Values(seq), m, more...)
}

func Each[T any, V any](seq iter.Seq[T], m func(T) V, more ...V) []V {
	var res []V
	seq(func(t T) bool {
		res = append(res, m(t))
		return true
	})

	res = append(res, more...)

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
