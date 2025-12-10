package uient

import (
	"fmt"
	"os"
	"reflect"

	"go.wdy.de/nago/application/ent"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/form"
)

type PageUpdateOptions[T ent.Aggregate[T, ID], ID ~string] struct {
	Perms        ent.Permissions
	EntityName   string
	Pages        Pages
	Prefix       permission.ID
	Update       func(wnd core.Window, uc ent.UseCases[T, ID], id ID) core.View
	DecorateView func(wnd core.Window, state *core.State[T], view core.View) core.View
}

func PageUpdate[T ent.Aggregate[T, ID], ID ~string](wnd core.Window, uc ent.UseCases[T, ID], opts PageUpdateOptions[T, ID]) core.View {
	id := ID(wnd.Values()[string(opts.Prefix)])
	if opts.Update == nil {
		opts.Update = newDefaultUpdate(wnd, opts)
	}
	return opts.Update(wnd, uc, id)
}

func newDefaultUpdate[T ent.Aggregate[T, ID], ID ~string](wnd core.Window, opts PageUpdateOptions[T, ID]) func(wnd core.Window, uc ent.UseCases[T, ID], id ID) core.View {
	return func(wnd core.Window, uc ent.UseCases[T, ID], id ID) core.View {
		if err := wnd.Subject().AuditResource(string(opts.Prefix), string(id), opts.Perms.FindByID); err != nil {
			return alert.BannerError(err)
		}

		canReadOnly := !wnd.Subject().HasResourcePermission(string(opts.Prefix), string(id), opts.Perms.Update)

		state := core.StateOf[T](wnd, reflect.TypeFor[T]().Name()+"-update-form").Init(func() T {
			var zero T
			optT, err := uc.FindByID(wnd.Subject(), id)
			if err != nil {
				alert.ShowBannerError(wnd, err)
				canReadOnly = true
				return zero
			}

			if optT.IsNone() {
				alert.ShowBannerError(wnd, fmt.Errorf("entity not found: %s: %w", id, os.ErrNotExist))
				canReadOnly = true
				return zero
			}

			return optT.Unwrap()
		})

		errState := core.StateOf[error](wnd, reflect.TypeFor[T]().Name()+"-update-form-err")
		var view core.View
		view = form.Auto[T](form.AutoOptions{
			Window:   wnd,
			Errors:   errState.Get(),
			ViewOnly: canReadOnly,
		}, state).FullWidth()

		if opts.DecorateView != nil {
			view = opts.DecorateView(wnd, state, view)
		}

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
					if err := uc.Update(wnd.Subject(), state.Get()); err != nil {
						errState.Set(err)
						return
					}

					wnd.Navigation().ForwardTo(opts.Pages.List, wnd.Values().Put(string(opts.Prefix), string(id)))
				}).Title(rstring.ActionSave.Get(wnd)).Enabled(!canReadOnly),
			).Gap(ui.L8).FullWidth().Alignment(ui.Trailing),
		).Alignment(ui.Leading).Frame(ui.Frame{}.Larger())
	}
}
