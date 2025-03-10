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

func AutoRootView[E form.Aggregate[E, ID], ID ~string](opts AutoRootViewOptions, useCases rcrud.UseCases[E, ID]) func(wnd core.Window) core.View {
	return func(wnd core.Window) core.View {
		bnd := AutoBinding[E](AutoBindingOptions{}, wnd, useCases)
		return ui.VStack(
			AutoView(AutoViewOptions{Title: opts.Title, CreateDisabled: opts.CreateDisabled}, bnd, useCases),
		).FullWidth()
	}
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

// AutoView creates a real simple default CRUD view for rapid prototyping.
func AutoView[E form.Aggregate[E, ID], ID ~string](opts AutoViewOptions, bnd *Binding[E], usecases rcrud.UseCases[E, ID]) core.View {
	if !canFindAll(bnd, usecases) {
		return ui.VStack(

			ui.H1(opts.Title),

			alert.IfPermissionDenied(bnd.wnd, "."),
		).Alignment(ui.Leading).Padding(ui.Padding{Top: ui.L40}).Frame(ui.Frame{}.FullWidth())
	}

	var zeroE E
	var actions []core.View
	if !opts.CreateDisabled && canCreate(bnd, usecases) {
		if opts.CreateButtonForwardTo == "" {
			actions = append(actions, ButtonCreate(bnd, zeroE, func(d E) (errorText string, infrastructureError error) {
				if !bnd.Validates(d) {
					return "Bitte prüfen Sie Ihre Eingaben", nil
				}

				_, err := usecases.Create(bnd.wnd.Subject(), d)
				return "", err
			}))
		} else {
			actions = append(actions, ButtonCreateForwardTo(bnd, opts.CreateButtonForwardTo, nil))
		}
	}

	return ui.VStack(

		ui.H1(opts.Title),

		View[E, ID](
			Options[E](bnd).
				FindAll(usecases.FindAll(bnd.wnd.Subject())).
				Actions(
					actions...,
				),
		).Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).Padding(ui.Padding{Top: ui.L40}).Frame(ui.Frame{}.FullWidth())
}

func NewUseCases[E form.Aggregate[E, ID], ID ~string](permissionPrefix permission.ID, repo data.Repository[E, ID]) rcrud.UseCases[E, ID] {
	return rcrud.UseCasesFrom(rcrud.DecorateRepository(rcrud.DecoratorOptions{PermissionPrefix: permissionPrefix}, repo))
}
