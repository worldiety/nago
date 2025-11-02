package uient

import (
	"fmt"
	"reflect"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ent"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
)

type PageListOptions[T ent.Aggregate[T, ID], ID ~string] struct {
	Perms      ent.Permissions
	EntityName string
	List       func(wnd core.Window, uc ent.UseCases[T, ID]) core.View
	Pages      Pages
	Prefix     permission.ID
}

func PageList[T ent.Aggregate[T, ID], ID ~string](wnd core.Window, uc ent.UseCases[T, ID], opts PageListOptions[T, ID]) core.View {
	if opts.List == nil {
		opts.List = newDefaultList[T, ID](opts)
	}

	return opts.List(wnd, uc)
}

func newDefaultList[T ent.Aggregate[T, ID], ID ~string](opts PageListOptions[T, ID]) func(wnd core.Window, uc ent.UseCases[T, ID]) core.View {
	return func(wnd core.Window, uc ent.UseCases[T, ID]) core.View {
		dv := dataview.FromData(wnd, dataview.Data[T, ID]{
			FindAll: uc.FindAllIdentifiers(wnd.Subject()),
			FindByID: func(id ID) (option.Opt[T], error) {
				return uc.FindByID(wnd.Subject(), id)
			},
			Fields: autoFields[T](wnd),
		}).Action(func(e T) {
			wnd.Navigation().ForwardTo(opts.Pages.Update, wnd.Values().Put(string(opts.Prefix), string(e.Identity())))
		}).NextActionIndicator(true)

		if wnd.Subject().HasPermission(opts.Perms.Create) {
			dv = dv.NewAction(func() {
				wnd.Navigation().ForwardTo(opts.Pages.Create, nil)
			})
		}

		if wnd.Subject().HasPermission(opts.Perms.DeleteByID) {
			dv = dv.SelectOptions(dataview.NewSelectOptionDelete(wnd, func(selected []ID) error {
				for _, id := range selected {
					if err := uc.DeleteByID(wnd.Subject(), id); err != nil {
						return err
					}
				}

				return nil
			}))
		}

		return ui.VStack(
			ui.Space(ui.L16),
			breadcrumb.Breadcrumbs(
				ui.TertiaryButton(func() {
					wnd.Navigation().BackwardTo("admin", wnd.Values().Put("#", string(opts.Prefix)))
				}).Title(StrDataManagement.Get(wnd)),
			).ClampLeading(),
			ui.H1(opts.EntityName),
			dv,
		).FullWidth().Alignment(ui.Leading)
	}
}

func autoFields[T any](wnd core.Window) []dataview.Field[T] {
	var res []dataview.Field[T]
	rtype := reflect.TypeFor[T]()
	for _, field := range reflect.VisibleFields(rtype) {
		if v, ok := field.Tag.Lookup("visible"); ok && v == "false" {
			continue
		}

		if _, ok := field.Tag.Lookup("source"); ok {
			// currently, we can only look up the entire data set and not individual items which
			// becomes most expensive in our display loop here. Thus discard these entirely.
			continue
		}

		switch field.Type.Kind() {
		case reflect.Slice:
			// is is usually garbage to render long lists
			continue
		default:
			// just render
		}

		name := field.Name
		if v, ok := field.Tag.Lookup("label"); ok {
			name = v
		}

		name = wnd.Bundle().Resolve(name)

		res = append(res, dataview.Field[T]{
			Name: name,
			Map: func(obj T) core.View {
				iface := reflect.ValueOf(obj).FieldByName(field.Name).Interface()
				return ui.Text(fmt.Sprintf("%v", iface))
			},
		})
	}

	return res
}
