package uient

import (
	"reflect"

	"go.wdy.de/nago/application/ent"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/form"
)

type PageCreateOptions[T ent.Aggregate[T, ID], ID ~string] struct {
	Perms      ent.Permissions
	EntityName string
	Pages      Pages
	Prefix     permission.ID
	Create     func(wnd core.Window, uc ent.UseCases[T, ID]) core.View
}

func PageCreate[T ent.Aggregate[T, ID], ID ~string](wnd core.Window, uc ent.UseCases[T, ID], opts PageCreateOptions[T, ID]) core.View {
	if opts.Create == nil {
		opts.Create = newDefaultCreate(wnd, opts)
	}

	return opts.Create(wnd, uc)
}

func newDefaultCreate[T ent.Aggregate[T, ID], ID ~string](wnd core.Window, opts PageCreateOptions[T, ID]) func(wnd core.Window, uc ent.UseCases[T, ID]) core.View {
	return func(wnd core.Window, uc ent.UseCases[T, ID]) core.View {
		state := core.StateOf[T](wnd, reflect.TypeFor[T]().Name()+"-create-form")
		errState := core.StateOf[error](wnd, reflect.TypeFor[T]().Name()+"-create-form-err")
		view := form.Auto[T](form.AutoOptions{
			Window: wnd,
			Errors: errState.Get(),
		}, state).FullWidth()

		return ui.VStack(
			ui.Space(ui.L16),
			breadcrumb.Breadcrumbs(
				ui.TertiaryButton(func() {
					wnd.Navigation().BackwardTo("admin", wnd.Values().Put("#", string(opts.Prefix)))
				}).Title(StrDataManagement.Get(wnd)),
				ui.TertiaryButton(func() {
					wnd.Navigation().BackwardTo(opts.Pages.List, wnd.Values())
				}).Title(opts.EntityName+" "+StrElements.Get(wnd)),
			).ClampLeading(),

			ui.H1(opts.EntityName),
			
			view,
			ui.HLine(),
			ui.HStack(
				ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo(opts.Pages.List, wnd.Values())
				}).Title(rstring.ActionCancel.Get(wnd)),
				ui.PrimaryButton(func() {
					id, err := uc.Create(wnd.Subject(), state.Get())
					if err != nil {
						errState.Set(err)
						return
					}

					wnd.Navigation().ForwardTo(opts.Pages.List, wnd.Values().Put(string(opts.Prefix), string(id)))
				}).Title(rstring.ActionCreate.Get(wnd)),
			).Gap(ui.L8).FullWidth().Alignment(ui.Trailing),
		).Alignment(ui.Leading).Frame(ui.Frame{}.Larger())
	}
}
