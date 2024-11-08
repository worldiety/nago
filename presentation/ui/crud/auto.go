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
	"iter"
	"log/slog"
	"reflect"
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
		)
	}
}

type AutoBindingOptions struct {
}

// AutoBinding takes the crud use cases and creates a naive binding for it.
// You can additionally tweak the binding, using the following field tags:
//   - label for an alternative name
//   - hidden to omit it completely
//
// To automatically also create a CRUD component e.g. for an entire page, see also [AutoView].
func AutoBinding[E Aggregate[E, ID], ID ~string](opts AutoBindingOptions, wnd core.Window, useCases UseCases[E, ID]) *Binding[E] {
	var zero E
	bnd := NewBinding[E](wnd)
	for _, field := range reflect.VisibleFields(reflect.TypeOf(zero)) {
		if flag, ok := field.Tag.Lookup("visible"); ok && flag == "false" {
			continue
		}

		label := field.Name
		if name, ok := field.Tag.Lookup("label"); ok {
			label = name
		}

		switch field.Type.Kind() {
		case reflect.String:
			switch field.Type {
			case reflect.TypeFor[ui.Color]():
				bnd.Add(PickOneColor(PickOneColorOptions{Label: label}, PropertyFuncs(
					func(e *E) std.Option[ui.Color] {
						value := reflect.ValueOf(e).Elem().FieldByName(field.Name).String()
						if value == "" {
							return std.None[ui.Color]()
						}

						return std.Some[ui.Color](ui.Color(value))
					},
					func(dst *E, v std.Option[ui.Color]) {
						reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(string(v.UnwrapOr("")))
					},
				)))

			default:
				bnd.Add(Text[E, string](TextOptions{Label: label}, PropertyFuncs(
					func(obj *E) string {
						return reflect.ValueOf(obj).Elem().FieldByName(field.Name).String()
					}, func(dst *E, v string) {
						reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(v)
					})))
			}

		case reflect.Int:
			fallthrough
		case reflect.Int64:
			bnd.Add(Int[E, int64](IntOptions{Label: label}, PropertyFuncs(
				func(obj *E) int64 {
					return reflect.ValueOf(obj).Elem().FieldByName(field.Name).Int()
				}, func(dst *E, v int64) {
					reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetInt(v)
				})))
		case reflect.Float64:
			bnd.Add(Float[E, float64](FloatOptions{Label: label}, PropertyFuncs(
				func(obj *E) float64 {
					return reflect.ValueOf(obj).Elem().FieldByName(field.Name).Float()
				}, func(dst *E, v float64) {
					reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetFloat(v)
				})))
		case reflect.Bool:
			bnd.Add(Bool[E, bool](BoolOptions{Label: label}, PropertyFuncs(
				func(obj *E) bool {
					return reflect.ValueOf(obj).Elem().FieldByName(field.Name).Bool()
				}, func(dst *E, v bool) {
					reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetBool(v)
				})))
		default:
			slog.Info("unsupported auto binding field type", "type", reflect.TypeOf(zero), "field", field.Name, "type", field.Type)
		}
	}

	var aggregateOpts []ElementViewFactory[E]
	if canSave(bnd, useCases) {
		aggregateOpts = append(aggregateOpts, ButtonEdit[E, ID](bnd, func(model E) (errorText string, infrastructureError error) {

			_, err := useCases.Save(wnd.Subject(), model)
			if err != nil {
				return "", err // The UI will hide the error
				// from the user and will show a general tracking.SupportRequestDialog
			}

			return "", nil
		}))
	}

	if canDelete(bnd, useCases) {
		aggregateOpts = append(aggregateOpts, ButtonDelete[E, ID](wnd, func(model E) error {

			err := useCases.DeleteByID(wnd.Subject(), model.Identity())
			if err != nil {
				return err // The UI will hide the error
				// from the user and will show a general tracking.SupportRequestDialog
			}

			return nil
		}))
	}

	if len(aggregateOpts) > 0 {
		bnd.Add(
			AggregateActions[E]("Optionen", aggregateOpts...),
		)
	}

	return bnd
}

type AutoViewOptions struct {
	Title string
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

			ui.Text("Leider haben Sie nicht die nötigen Rechte, um auf diese Daten zuzugreifen."),
		).Alignment(ui.Leading).Padding(ui.Padding{Top: ui.L40}).Frame(ui.Frame{}.FullWidth())
	}

	var zeroE E
	var actions []core.View
	if canCreate(bnd, usecases) {
		actions = append(actions, ButtonCreate(bnd, zeroE, func(d E) (errorText string, infrastructureError error) {
			if !bnd.Validates(d) {
				return "Bitte prüfen Sie Ihre Eingaben", nil
			}

			_, err := usecases.Save(bnd.wnd.Subject(), d)
			return "", err
		}))
	}

	return ui.VStack(

		ui.VStack(
			ui.Text(opts.Title).TextAlignment(ui.TextAlignCenter).Font(ui.Font{
				Name:   "",
				Size:   "2rem",
				Style:  "",
				Weight: ui.BoldFontWeight,
			}),

			ui.HLine(),
		).Alignment(ui.Leading).Padding(ui.Padding{Bottom: ui.Length("2rem").Negate()}),

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
