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

// TDialogCreate is an overlay component(Dialog Create).
type TDialogCreate[T any] struct {
	wnd       core.Window
	name      string
	presented *core.State[bool]
	onCreate  func(subject auth.Subject, value T) error
}

func DialogCreate[T any](wnd core.Window, name string, presented *core.State[bool], onCreate func(subject auth.Subject, value T) error) TDialogCreate[T] {
	return TDialogCreate[T]{
		wnd:       wnd,
		name:      name,
		presented: presented,
		onCreate:  onCreate,
	}
}

// TDialogEdit is an overlay component(Dialog Edit).
type TDialogEdit[T any] struct {
	wnd        core.Window
	name       string
	presented  *core.State[bool]
	modelState *core.State[T]
	onSave     func(subject auth.Subject) error
}

func DialogEdit[T any](wnd core.Window, name string, presented *core.State[bool], modelState *core.State[T], onSave func(subject auth.Subject) error) TDialogEdit[T] {
	return TDialogEdit[T]{
		wnd:        wnd,
		name:       name,
		presented:  presented,
		modelState: modelState,
		onSave:     onSave,
	}
}

func (t TDialogEdit[T]) Render(ctx core.RenderContext) core.RenderNode {
	if !t.presented.Get() {
		return nil
	}

	return alert.Dialog(
		t.name,
		Auto(AutoOptions{}, t.modelState),
		t.presented,
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := t.onSave(t.wnd.Subject()); err != nil {
				alert.ShowBannerError(t.wnd, err)
				return false
			}

			return true
		}),
	).Render(ctx)
}

func (t TDialogCreate[T]) Render(ctx core.RenderContext) core.RenderNode {
	if !t.presented.Get() {
		return nil
	}

	modelState := core.StateOf[T](t.wnd, "create-"+t.name)

	return alert.Dialog(
		t.name,
		Auto(AutoOptions{}, modelState),
		t.presented,
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := t.onCreate(t.wnd.Subject(), modelState.Get()); err != nil {
				alert.ShowBannerError(t.wnd, err)
				return false
			}

			return true
		}),
	).Render(ctx)
}
