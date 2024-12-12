package crud

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"iter"
	"strings"
)

type AutoOptions struct {
}

type Aggregate[A any, ID comparable] interface {
	data.Aggregate[ID]
	WithIdentity(ID) A
}

// UseCases represent the most basic and simplest CRUD-based use cases. See also [NewUseCases] to automatically
// derive an instance from a [data.Repository]. This is only useful for rapid prototyping the most simple CRUD
// UIs. See also [AutoBinding].
type UseCases[E data.Aggregate[ID], ID ~string] interface {
	FindByID(subject auth.Subject, id ID) (std.Option[E], error)
	All(subject auth.Subject) iter.Seq2[E, error]
	DeleteByID(subject auth.Subject, id ID) error
	Save(subject auth.Subject, entity E) (ID, error)
}

func FuncsFromUseCases[E data.Aggregate[ID], ID ~string](uc UseCases[E, ID]) (
	findByID func(subject auth.Subject, id ID) (std.Option[E], error),
	all func(subject auth.Subject) iter.Seq2[E, error],
	deleteID func(subject auth.Subject, id ID) error,
	save func(subject auth.Subject, entity E) (ID, error),
) {
	return uc.FindByID, uc.All, uc.DeleteByID, uc.Save
}

func UseCasesFromFuncs[E data.Aggregate[ID], ID ~string](
	findByID func(subject auth.Subject, id ID) (std.Option[E], error),
	all func(subject auth.Subject) iter.Seq2[E, error],
	deleteID func(subject auth.Subject, id ID) error,
	save func(subject auth.Subject, entity E) (ID, error),
) UseCases[E, ID] {
	return useCasesFuncs[E, ID]{
		findByID: findByID,
		all:      all,
		deleteID: deleteID,
		save:     save,
	}
}

type AutoRootViewOptions struct {
	Title          string
	CreateDisabled bool
}

func AutoRootView[E Aggregate[E, ID], ID ~string](opts AutoRootViewOptions, useCases UseCases[E, ID]) func(wnd core.Window) core.View {
	return func(wnd core.Window) core.View {
		bnd := AutoBinding[E](AutoBindingOptions{}, wnd, useCases)
		return ui.VStack(
			ui.WindowTitle(opts.Title),
			AutoView(AutoViewOptions{Title: opts.Title, CreateDisabled: opts.CreateDisabled}, bnd, useCases),
		).FullWidth()
	}
}

type AutoViewOptions struct {
	Title                 string
	CreateButtonForwardTo core.NavigationPath
	CreateDisabled        bool
}

func canFindAll[E Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases UseCases[E, ID]) bool {
	if repoUC, ok := useCases.(repositoryUsecases[E, ID]); ok {
		return bnd.wnd.Subject().HasPermission(repoUC.permFindAll)
	}

	return true
}

func canCreate[E Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases UseCases[E, ID]) bool {
	if repoUC, ok := useCases.(repositoryUsecases[E, ID]); ok {
		return bnd.wnd.Subject().HasPermission(repoUC.permCreate)
	}

	return true
}

func canDelete[E Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases UseCases[E, ID]) bool {
	if repoUC, ok := useCases.(repositoryUsecases[E, ID]); ok {
		return bnd.wnd.Subject().HasPermission(repoUC.permDelete)
	}

	return true
}

func canSave[E Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases UseCases[E, ID]) bool {
	if repoUC, ok := useCases.(repositoryUsecases[E, ID]); ok {
		return bnd.wnd.Subject().HasPermission(repoUC.permSave)
	}

	return true
}

// AutoView creates a real simple default CRUD view for rapid prototyping.
func AutoView[E Aggregate[E, ID], ID ~string](opts AutoViewOptions, bnd *Binding[E], usecases UseCases[E, ID]) core.View {
	if !canFindAll(bnd, usecases) {
		return ui.VStack(

			ui.VStack(
				ui.Text(opts.Title).TextAlignment(ui.TextAlignCenter).Font(ui.Font{
					Name:   "",
					Size:   "2rem",
					Style:  "",
					Weight: ui.BoldFontWeight,
				}),

				ui.HLine(),
			).Alignment(ui.Leading).Padding(ui.Padding{Bottom: ui.Length("2rem")}),

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

				_, err := usecases.Save(bnd.wnd.Subject(), d)
				return "", err
			}))
		} else {
			actions = append(actions, ButtonCreateForwardTo(bnd, opts.CreateButtonForwardTo, nil))
		}
	}

	return ui.VStack(

		ui.VStack(
			ui.Text(opts.Title).TextAlignment(ui.TextAlignCenter).Font(ui.Font{
				Name:   "",
				Size:   "2rem",
				Style:  "",
				Weight: ui.BoldFontWeight,
			}),

			ui.HLineWithColor(ui.ColorAccent),
		).Alignment(ui.Leading),

		View[E, ID](
			Options[E](bnd).
				FindAll(usecases.All(bnd.wnd.Subject())).
				Actions(
					actions...,
				),
		).Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).Padding(ui.Padding{Top: ui.L40}).Frame(ui.Frame{}.FullWidth())
}

func NewUseCases[E Aggregate[E, ID], ID ~string](permissionPrefix permission.ID, repo data.Repository[E, ID]) UseCases[E, ID] {
	if !permissionPrefix.Valid() {
		panic(fmt.Sprintf("invalid permission prefix: %v", permissionPrefix))
	}

	if !strings.HasSuffix(string(permissionPrefix), ".") {
		permissionPrefix += "."
	}

	return repositoryUsecases[E, ID]{
		repo:        repo,
		permFindAll: permission.Make[findAll[E]](permissionPrefix + "find"),
		permSave:    permission.Make[saveOne[E]](permissionPrefix + "save"),
		permCreate:  permission.Make[createOne[E]](permissionPrefix + "create"),
		permDelete:  permission.Make[deleteOne[E]](permissionPrefix + "delete"),
	}
}

type findAll[E any] func()
type deleteOne[E any] func()
type saveOne[E any] func()
type createOne[E any] func()

type useCasesFuncs[E data.Aggregate[ID], ID ~string] struct {
	findByID func(subject auth.Subject, id ID) (std.Option[E], error)
	all      func(subject auth.Subject) iter.Seq2[E, error]
	deleteID func(subject auth.Subject, id ID) error
	save     func(subject auth.Subject, entity E) (ID, error)
}

func (u useCasesFuncs[E, ID]) FindByID(subject auth.Subject, id ID) (std.Option[E], error) {
	return u.findByID(subject, id)
}

func (u useCasesFuncs[E, ID]) All(subject auth.Subject) iter.Seq2[E, error] {
	return u.all(subject)
}

func (u useCasesFuncs[E, ID]) DeleteByID(subject auth.Subject, id ID) error {
	return u.deleteID(subject, id)
}

func (u useCasesFuncs[E, ID]) Save(subject auth.Subject, entity E) (ID, error) {
	return u.save(subject, entity)
}

type repositoryUsecases[E Aggregate[E, ID], ID ~string] struct {
	permFindAll permission.ID
	permSave    permission.ID
	permCreate  permission.ID
	permDelete  permission.ID
	repo        data.Repository[E, ID]
}

func (r repositoryUsecases[E, ID]) FindByID(subject auth.Subject, id ID) (std.Option[E], error) {
	if err := subject.Audit(r.permFindAll); err != nil {
		return std.None[E](), err
	}

	return r.repo.FindByID(id)
}

func (r repositoryUsecases[E, ID]) All(subject auth.Subject) iter.Seq2[E, error] {
	if err := subject.Audit(r.permFindAll); err != nil {
		return xslices.ValuesWithError([]E(nil), err)
	}

	return r.repo.All()
}

func (r repositoryUsecases[E, ID]) DeleteByID(subject auth.Subject, id ID) error {
	if err := subject.Audit(r.permDelete); err != nil {
		return err
	}

	return r.repo.DeleteByID(id)
}

func (r repositoryUsecases[E, ID]) Save(subject auth.Subject, entity E) (ID, error) {
	var zeroID ID
	if err := subject.Audit(r.permSave); err != nil {
		return zeroID, err
	}

	if entity.Identity() == zeroID {
		entity = entity.WithIdentity(data.RandIdent[ID]())
	}

	return entity.Identity(), r.repo.Save(entity)
}
