// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"iter"
	"reflect"
	
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
)

func dialogCmd[T flow.WorkspaceCommand, Evt flow.WorkspaceEvent](wnd core.Window, ws *flow.Workspace, title string, presented *core.State[bool], handler func(subject auth.Subject, cmd T) (Evt, error), init func() T) core.View {
	if !presented.Get() {
		return nil
	}

	cmdState := core.DerivedState[T](presented, reflect.TypeFor[T]().String()).Init(init)
	errState := core.DerivedState[error](cmdState, "err")

	ctx := core.WithContext(wnd.Context(), core.ContextValue("nago.flow.packages", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[*flow.Package, error] {
		return func(yield func(*flow.Package, error) bool) {
			for pkg := range ws.Packages() {
				if !yield(pkg, nil) {
					return
				}
			}
		}
	})))

	ctx = core.WithContext(ctx, core.ContextValue("nago.flow.structs", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[*flow.StructType, error] {
		return func(yield func(*flow.StructType, error) bool) {
			for pkg := range ws.StructTypes() {
				if !yield(pkg, nil) {
					return
				}
			}
		}
	})))

	ctx = core.WithContext(ctx, core.ContextValue("nago.flow.pkstructs", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[*flow.StructType, error] {
		return func(yield func(*flow.StructType, error) bool) {
			for t := range ws.StructTypes() {
				if !t.DocumentStoreReady() {
					continue
				}

				if !yield(t, nil) {
					return
				}
			}
		}
	})))

	return alert.Dialog(
		title,
		form.Auto[T](form.AutoOptions{Context: ctx, Errors: errState.Get()}, cmdState),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			errState.Set(nil)
			if _, err := handler(wnd.Subject(), cmdState.Get()); err != nil {
				errState.Set(err)
				return false
			}

			return true
		}),
	)
}
