// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

func DialogCreate[T any](wnd core.Window, name string, presented *core.State[bool], onCreate func(subject auth.Subject, value T) error) core.View {
	if !presented.Get() {
		return nil
	}

	modelState := core.StateOf[T](wnd, "create-"+name)

	return alert.Dialog(
		name,
		Auto(AutoOptions{}, modelState),
		presented,
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := onCreate(wnd.Subject(), modelState.Get()); err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}

			return true
		}),
	)
}

func DialogEdit[T any](wnd core.Window, name string, presented *core.State[bool], modelState *core.State[T], onSave func(subject auth.Subject) error) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		name,
		Auto(AutoOptions{}, modelState),
		presented,
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := onSave(wnd.Subject()); err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}

			return true
		}),
	)
}
