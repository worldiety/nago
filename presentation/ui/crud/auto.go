// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
)

type AutoOptions struct {
}

type AutoRootViewOptions struct {
	Title          string
	CreateDisabled bool
}

// TAutoRootView is a crud component(CRUD Auto Root View).
type TAutoRootView[E form.Aggregate[E, ID], ID ~string] struct {
	opts     AutoRootViewOptions
	useCases rcrud.UseCases[E, ID]

	frame ui.Frame
}

func AutoRootView[E form.Aggregate[E, ID], ID ~string](opts AutoRootViewOptions, useCases rcrud.UseCases[E, ID]) func(wnd core.Window) core.View {
	return func(wnd core.Window) core.View {
		return TAutoRootView[E, ID]{
			opts:     opts,
			useCases: useCases,
			frame:    ui.Frame{}.FullWidth(),
		}
	}
}

func (t TAutoRootView[E, ID]) Render(ctx core.RenderContext) core.RenderNode {
	bnd := AutoBinding[E](AutoBindingOptions{}, ctx.Window(), t.useCases)
	return ui.VStack(
		AutoView(AutoViewOptions{Title: t.opts.Title, CreateDisabled: t.opts.CreateDisabled}, bnd, t.useCases),
	).Frame(t.frame).
		Render(ctx)
}

type AutoViewOptions struct {
	Title                 string
	CreateButtonForwardTo core.NavigationPath
	CreateDisabled        bool
}

func canFindAll[E form.Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases rcrud.UseCases[E, ID]) bool {
	return bnd.wnd.Subject().HasPermission(useCases.PermFindAll())
}

func canCreate[E form.Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases rcrud.UseCases[E, ID]) bool {
	return bnd.wnd.Subject().HasPermission(useCases.PermCreate())
}

func canDelete[E form.Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases rcrud.UseCases[E, ID]) bool {
	return bnd.wnd.Subject().HasPermission(useCases.PermDeleteByID())
}

func canUpdate[E form.Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases rcrud.UseCases[E, ID]) bool {
	return bnd.wnd.Subject().HasPermission(useCases.PermUpdate())
}

// TAutoView is a crud component(CRUD Auto View).
// AutoView creates a real simple default CRUD view for rapid prototyping.
type TAutoView[E form.Aggregate[E, ID], ID ~string] struct {
	opts     AutoViewOptions
	bnd      *Binding[E]
	useCases rcrud.UseCases[E, ID]

	padding   ui.Padding
	alignment ui.Alignment
	frame     ui.Frame
}

func AutoView[E form.Aggregate[E, ID], ID ~string](opts AutoViewOptions, bnd *Binding[E], usecases rcrud.UseCases[E, ID]) core.View {
	return TAutoView[E, ID]{
		opts:      opts,
		bnd:       bnd,
		useCases:  usecases,
		padding:   ui.Padding{Top: ui.L40},
		alignment: ui.Leading,
		frame:     ui.Frame{}.FullWidth(),
	}
}

func (t TAutoView[E, ID]) Render(ctx core.RenderContext) core.RenderNode {
	if !canFindAll(t.bnd, t.useCases) {
		return ui.VStack(

			ui.H1(t.opts.Title),

			alert.IfPermissionDenied(t.bnd.wnd, "."),
		).Alignment(t.alignment).
			Padding(t.padding).
			Frame(t.frame).
			Render(ctx)
	}

	var zeroE E
	var actions []core.View
	if !t.opts.CreateDisabled && canCreate(t.bnd, t.useCases) {
		if t.opts.CreateButtonForwardTo == "" {
			actions = append(actions, ButtonCreate(t.bnd, zeroE, func(d E) (errorText string, infrastructureError error) {
				if !t.bnd.Validates(d) {
					return "Bitte prüfen Sie Ihre Eingaben", nil
				}

				_, err := t.useCases.Create(t.bnd.wnd.Subject(), d)
				return "", err
			}))
		} else {
			actions = append(actions, ButtonCreateForwardTo(t.bnd, t.opts.CreateButtonForwardTo, nil))
		}
	}

	return ui.VStack(

		ui.H1(t.opts.Title),

		View[E, ID](
			Options[E](t.bnd).
				FindAll(t.useCases.FindAll(t.bnd.wnd.Subject())).
				Actions(
					actions...,
				),
		).Frame(ui.Frame{}.FullWidth()),
	).Alignment(t.alignment).
		Padding(t.padding).
		Frame(t.frame).
		Render(ctx)
}

func NewUseCases[E form.Aggregate[E, ID], ID ~string](permissionPrefix permission.ID, repo data.Repository[E, ID]) rcrud.UseCases[E, ID] {
	return rcrud.UseCasesFrom(rcrud.DecorateRepository(rcrud.DecoratorOptions{PermissionPrefix: permissionPrefix}, repo))
}
