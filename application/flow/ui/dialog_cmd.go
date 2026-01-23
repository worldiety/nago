// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"iter"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
)

func dialogCmd(wnd core.Window, ws *flow.Workspace, title string, presented *core.State[bool], handler flow.HandleCommand, init func() flow.WorkspaceCommand) core.View {
	if !presented.Get() {
		return nil
	}

	cmdState := core.DerivedState[flow.WorkspaceCommand](presented, title).Init(init)
	errState := core.DerivedState[error](cmdState, "err")

	ctx := core.WithContext(wnd.Context(), core.ContextValue("nago.flow.packages", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[*flow.Package, error] {
		return func(yield func(*flow.Package, error) bool) {
			for pkg := range ws.Packages.All() {
				if !yield(pkg, nil) {
					return
				}
			}
		}
	})))

	ctx = core.WithContext(ctx, core.ContextValue("nago.flow.structs", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[*flow.StructType, error] {
		return func(yield func(*flow.StructType, error) bool) {
			for pkg := range ws.Packages.StructTypes() {
				if !yield(pkg, nil) {
					return
				}
			}
		}
	})))

	ctx = core.WithContext(ctx, core.ContextValue("nago.flow.types", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[flow.Type, error] {
		return func(yield func(flow.Type, error) bool) {
			for pkg := range ws.Packages.Types() {
				if !yield(pkg, nil) {
					return
				}
			}
		}
	})))

	ctx = core.WithContext(ctx, core.ContextValue("nago.flow.pkstructs", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[*flow.StructType, error] {
		return func(yield func(*flow.StructType, error) bool) {
			for t := range ws.Packages.StructTypes() {
				if !t.DocumentStoreReady() {
					continue
				}

				if !yield(t, nil) {
					return
				}
			}
		}
	})))

	ctx = core.WithContext(ctx, core.ContextValue("nago.flow.repositories", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[*flow.Repository, error] {
		return func(yield func(*flow.Repository, error) bool) {
			for t := range ws.Repositories.All() {

				if !yield(t, nil) {
					return
				}
			}
		}
	})))

	return alert.Dialog(
		title,
		form.Auto[flow.WorkspaceCommand](form.AutoOptions{Context: ctx, Errors: errState.Get()}, cmdState),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			errState.Set(nil)
			if err := handler(wnd.Subject(), cmdState.Get()); err != nil {
				errState.Set(err)
				return false
			}

			return true
		}),
	)
}
