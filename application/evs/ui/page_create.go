// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uievs

import (
	"fmt"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/form"
)

type PageCreateOptions[Evt any] struct {
	Perms      evs.Permissions
	EntityName string
	Pages      Pages
	Prefix     permission.ID
	Create     func(wnd core.Window, uc evs.UseCases[Evt]) core.View
}

func PageCreate[Evt any](wnd core.Window, uc evs.UseCases[Evt], opts PageCreateOptions[Evt]) core.View {
	if opts.Create == nil {
		opts.Create = newDefaultCreate(wnd, opts)
	}

	return opts.Create(wnd, uc)
}

func newDefaultCreate[Evt any](wnd core.Window, opts PageCreateOptions[Evt]) func(wnd core.Window, uc evs.UseCases[Evt]) core.View {
	return func(wnd core.Window, uc evs.UseCases[Evt]) core.View {
		discriminator := evs.Discriminator(wnd.Path().Base())

		state := core.StateOf[Evt](wnd, string(discriminator)+"-create-form").Init(func() Evt {
			v, err := uc.MakeType(discriminator)
			if err != nil {
				var zero Evt
				alert.ShowBannerError(wnd, fmt.Errorf("programming error: %w", err))
				return zero
			}

			return v
		})

		errState := core.StateOf[error](wnd, string(discriminator)+"-create-form-err")
		view := form.Auto[Evt](form.AutoOptions{
			Window: wnd,
			Errors: errState.Get(),
		}, state).FullWidth()

		typeName := wnd.Path().Base()

		return ui.VStack(
			ui.Space(ui.L16),
			breadcrumb.Breadcrumbs(
				ui.TertiaryButton(func() {
					wnd.Navigation().BackwardTo("admin", wnd.Values().Put("#", string(opts.Prefix)))
				}).Title(StrDataManagement.Get(wnd)),
				ui.TertiaryButton(func() {
					wnd.Navigation().BackwardTo(opts.Pages.Audit, wnd.Values())
				}).Title(opts.EntityName+" "+StrElements.Get(wnd)),
			).ClampLeading(),

			ui.H1(StrCreateEvtX.Get(wnd, i18n.String("name", typeName))),
			ui.Text(StrCreateDisclaimer.Get(wnd)),
			ui.Space(ui.L16),
			
			view,
			ui.HLine(),
			ui.HStack(
				ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo(opts.Pages.Audit, wnd.Values())
				}).Title(rstring.ActionCancel.Get(wnd)),
				ui.PrimaryButton(func() {
					_, err := uc.Store(wnd.Subject(), state.Get(), evs.StoreOptions{})
					if err != nil {
						errState.Set(err)
						return
					}

					wnd.Navigation().ForwardTo(opts.Pages.Audit, wnd.Values())
				}).Title(rstring.ActionCreate.Get(wnd)),
			).Gap(ui.L8).FullWidth().Alignment(ui.Trailing),
		).Alignment(ui.Leading).Frame(ui.Frame{}.Larger())
	}
}
