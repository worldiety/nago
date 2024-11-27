package crud

import (
	"fmt"
	"go.wdy.de/nago/annotation"
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

type AutoRootViewOptions struct {
	Title string
}

func AutoRootView[E Aggregate[E, ID], ID ~string](opts AutoRootViewOptions, useCases UseCases[E, ID]) func(wnd core.Window) core.View {
	return func(wnd core.Window) core.View {
		bnd := AutoBinding[E](AutoBindingOptions{}, wnd, useCases)
		return ui.VStack(
			ui.WindowTitle(opts.Title),
			AutoView(AutoViewOptions{Title: opts.Title}, bnd, useCases),
		).FullWidth()
	}
}

type AutoViewOptions struct {
	Title                 string
	CreateButtonForwardTo core.NavigationPath
}

func canFindAll[E Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases UseCases[E, ID]) bool {
	if repoUC, ok := useCases.(repositoryUsecases[E, ID]); ok {
		return bnd.wnd.Subject().HasPermission(repoUC.permFindAll.Identity())
	}

	return false
}

func canCreate[E Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases UseCases[E, ID]) bool {
	if repoUC, ok := useCases.(repositoryUsecases[E, ID]); ok {
		return bnd.wnd.Subject().HasPermission(repoUC.permCreate.Identity())
	}

	return false
}

func canDelete[E Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases UseCases[E, ID]) bool {
	if repoUC, ok := useCases.(repositoryUsecases[E, ID]); ok {
		return bnd.wnd.Subject().HasPermission(repoUC.permDelete.Identity())
	}

	return false
}

func canSave[E Aggregate[E, ID], ID ~string](bnd *Binding[E], useCases UseCases[E, ID]) bool {
	if repoUC, ok := useCases.(repositoryUsecases[E, ID]); ok {
		return bnd.wnd.Subject().HasPermission(repoUC.permSave.Identity())
	}

	return false
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
	if canCreate(bnd, usecases) {
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

func NewUseCases[E Aggregate[E, ID], ID ~string](permissionPrefix annotation.PermissionID, repo data.Repository[E, ID]) UseCases[E, ID] {
	if !permissionPrefix.Valid() {
		panic(fmt.Sprintf("invalid permission prefix: %v", permissionPrefix))
	}

	if !strings.HasSuffix(string(permissionPrefix), ".") {
		permissionPrefix += "."
	}

	return repositoryUsecases[E, ID]{
		repo:        repo,
		permFindAll: annotation.Permission[findAll[E]](string(permissionPrefix) + "find"),
		permSave:    annotation.Permission[saveOne[E]](string(permissionPrefix) + "save"),
		permCreate:  annotation.Permission[createOne[E]](string(permissionPrefix) + "create"),
		permDelete:  annotation.Permission[deleteOne[E]](string(permissionPrefix) + "delete"),
	}
}

type findAll[E any] func()
type deleteOne[E any] func()
type saveOne[E any] func()
type createOne[E any] func()

type repositoryUsecases[E Aggregate[E, ID], ID ~string] struct {
	permFindAll annotation.SubjectPermission
	permSave    annotation.SubjectPermission
	permCreate  annotation.SubjectPermission
	permDelete  annotation.SubjectPermission
	repo        data.Repository[E, ID]
}

func (r repositoryUsecases[E, ID]) FindByID(subject auth.Subject, id ID) (std.Option[E], error) {
	if err := subject.Audit(r.permFindAll.Identity()); err != nil {
		return std.None[E](), err
	}

	return r.repo.FindByID(id)
}

func (r repositoryUsecases[E, ID]) All(subject auth.Subject) iter.Seq2[E, error] {
	if err := subject.Audit(r.permFindAll.Identity()); err != nil {
		return xslices.ValuesWithError([]E(nil), err)
	}

	return r.repo.All()
}

func (r repositoryUsecases[E, ID]) DeleteByID(subject auth.Subject, id ID) error {
	if err := subject.Audit(r.permDelete.Identity()); err != nil {
		return err
	}

	return r.repo.DeleteByID(id)
}

func (r repositoryUsecases[E, ID]) Save(subject auth.Subject, entity E) (ID, error) {
	var zeroID ID
	if err := subject.Audit(r.permSave.Identity()); err != nil {
		return zeroID, err
	}

	if entity.Identity() == zeroID {
		entity = entity.WithIdentity(data.RandIdent[ID]())
	}

	return entity.Identity(), r.repo.Save(entity)
}
