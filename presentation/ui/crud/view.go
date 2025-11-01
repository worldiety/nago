// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"slices"

	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
)

// TView is a crud component(CRUD View).
type TView[Entity data.Aggregate[ID], ID data.IDType] struct {
	opts TOptions[Entity, ID]

	padding            ui.Padding
	frame              ui.Frame
	border             ui.Border
	accessibilityLabel string
	invisible          bool
}

// deprecated: use [entities.NewUseCases]
func View[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) TView[Entity, ID] {
	return TView[Entity, ID]{
		opts:  opts,
		frame: ui.Frame{MinWidth: ui.L400},
	}
}

func (t TView[Entity, ID]) Render(ctx core.RenderContext) core.RenderNode {
	var dataView core.View

	if t.opts.viewMode == ViewStyleDefault {
		dataView = ui.ViewThatMatches(t.opts.wnd,
			ui.SizeClass(core.SizeClassSmall, func() core.View { return Cards[Entity, ID](t.opts).Frame(ui.Frame{MaxWidth: ui.L480}.FullWidth()) }),
			ui.SizeClass(core.SizeClassMedium, func() core.View { return Table[Entity, ID](t.opts).Frame(ui.Frame{}.FullWidth()) }),
		)
	} else {
		dataView = List(t.opts)
	}

	searchbarAndActions := slices.Collect[core.View](func(yield func(core.View) bool) {
		yield(ui.ImageIcon(heroSolid.MagnifyingGlass))
		yield(ui.TextField("", t.opts.queryState.String()).InputValue(t.opts.queryState).Style(ui.TextFieldReduced))
		if len(t.opts.actions) > 0 {
			yield(ui.Space(ui.L16))
		}

		for _, action := range t.opts.actions {
			yield(action)
		}
	})

	isSmall := t.opts.wnd.Info().SizeClass <= core.SizeClassSmall

	return ui.VStack(
		ui.IfFunc(isSmall, func() core.View {
			return ui.VStack(
				ui.HStack(ui.If(t.opts.title != "", ui.H1(t.opts.title))).FullWidth().Alignment(ui.Leading),
				ui.HStack(searchbarAndActions...).Padding(ui.Padding{Bottom: ui.L16}),
			).FullWidth().Alignment(ui.Trailing)
		}),
		ui.IfFunc(!isSmall, func() core.View {
			return ui.HStack(
				ui.If(t.opts.title != "", ui.H1(t.opts.title)),
				ui.Spacer(),
				ui.HStack(searchbarAndActions...).Padding(ui.Padding{Bottom: ui.L16}),
			).FullWidth()
		}),

		dataView,
	).Visible(!t.invisible).
		Frame(t.frame).
		Border(t.border).
		Padding(t.padding).
		AccessibilityLabel(t.accessibilityLabel).
		Render(ctx)
}

func (t TView[Entity, ID]) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TView[Entity, ID]) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

func (t TView[Entity, ID]) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TView[Entity, ID]) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

func (t TView[Entity, ID]) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

func (t TView[Entity, ID]) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}
