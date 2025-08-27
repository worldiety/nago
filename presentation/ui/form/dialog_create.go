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

// TDialogCreate is an feedback component (Dialog Create).
// This component presents a creation dialog bound to a visibility state,
// invoking a callback to create a new value of type T.
type TDialogCreate[T any] struct {
	wnd       core.Window                               // window context
	name      string                                    // dialog title or name
	presented *core.State[bool]                         // state controlling dialog visibility
	onCreate  func(subject auth.Subject, value T) error // creation callback
}

// DialogCreate creates a new TDialogCreate with the given window, name,
// visibility state, and creation callback.
func DialogCreate[T any](wnd core.Window, name string, presented *core.State[bool], onCreate func(subject auth.Subject, value T) error) TDialogCreate[T] {
	return TDialogCreate[T]{
		wnd:       wnd,
		name:      name,
		presented: presented,
		onCreate:  onCreate,
	}
}

// TDialogEdit is a overlay component (Dialog Edit).
// This component presents an editing dialog bound to a model state,
// invoking a callback to save changes.
type TDialogEdit[T any] struct {
	wnd        core.Window                      // window context
	name       string                           // dialog title or name
	presented  *core.State[bool]                // state controlling dialog visibility
	modelState *core.State[T]                   // state holding the model being edited
	onSave     func(subject auth.Subject) error // save callback
}

// DialogEdit creates a new TDialogEdit with the given window, name,
// visibility state, model state, and save callback.
func DialogEdit[T any](wnd core.Window, name string, presented *core.State[bool], modelState *core.State[T], onSave func(subject auth.Subject) error) TDialogEdit[T] {
	return TDialogEdit[T]{
		wnd:        wnd,
		name:       name,
		presented:  presented,
		modelState: modelState,
		onSave:     onSave,
	}
}

// Render builds and returns the RenderNode for the TDialogEdit.
// It displays the edit dialog if presented is true, binding the model state
// to an auto-generated form. On save, it calls the onSave callback and
// shows an error banner if saving fails.
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

// Render builds and returns the RenderNode for the TDialogCreate.
// It displays the create dialog if presented is true, binding a new
// model state to an auto-generated form. On save, it calls the onCreate
// callback with the new value and shows an error banner if creation fails.
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
