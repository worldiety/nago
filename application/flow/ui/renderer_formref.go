// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"
	"iter"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
)

var _ ViewRenderer = (*FormRefRenderer)(nil)

type FormRefRenderer struct {
}

func (r *FormRefRenderer) Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply) {
	f, ok := ctx.Workspace().Forms.ByView(parent)
	if !ok {
		return alert.BannerError(fmt.Errorf("parent has no form: %s", parent)), nil
	}

	state := core.StateOf[flow.AddFormRefCmd](ctx.Window(), fmt.Sprintf("%s-%s", parent, after)).Init(func() flow.AddFormRefCmd {
		return flow.AddFormRefCmd{
			Workspace: ctx.Workspace().Identity(),
			Form:      f.ID,
			Parent:    parent,
			After:     after,
		}
	})

	errState := core.DerivedState[error](state, "err")

	structType, ok := ctx.Workspace().Packages.StructTypeByID(f.Type())
	if !ok {
		return alert.BannerError(fmt.Errorf("form referenced a struct type over repository which cannot be resolved: %s.%s", f.Type(), f.ID)), nil
	}

	myCtx := core.WithContext(ctx.Context, core.ContextValue("nago.flow.forms", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[*flow.Form, error] {
		return func(yield func(*flow.Form, error) bool) {

			for f := range ctx.Workspace().Forms.All() {
				if f.ID == ctx.Form().ID {
					continue
				}

				fStruct, ok := ctx.Workspace().Packages.StructTypeByID(f.Type())
				if !ok {
					continue
				}

				if fStruct.ID != structType.ID {
					continue
				}

				if !yield(f, nil) {
					return
				}
			}
		}
	})))

	return ui.VStack(
			form.Auto(form.AutoOptions{
				Errors:  errState.Get(),
				Context: myCtx,
			}, state),
		).FullWidth(), func() error {
			err := ctx.Handle(ctx.Window().Subject(), state.Get())
			errState.Set(err)
			return err
		}
}

func (r *FormRefRenderer) Update(ctx RContext, view flow.FormView) core.View {
	return ui.Text("Edit Form Ref")
}

func (r *FormRefRenderer) TeaserPreview(ctx RContext) core.View {
	return ui.VStack(
		ui.Text("Embedded Form").Font(ui.BodySmall),
		ui.VStack(ui.Text("Other Form")).Padding(ui.Padding{}.All(ui.L8)).Border(ui.Border{BorderStyle: ui.BorderStyleDashed}.Width(ui.L1).Color(ui.ColorIconsMuted)),
	)
}

func (r *FormRefRenderer) Preview(ctx RContext, view flow.FormView) core.View {
	ctx = ctx.InheritRef()
	ref := view.(*flow.FormRef)
	refForm, ok := ctx.Workspace().Forms.ByID(ref.Ref())
	if !ok {
		return ui.VStack(
			ui.Text(fmt.Sprintf("referenced form not found: %s", ref.Ref())),
		).Action(ctx.EditorAction(view))
	}

	return ui.VStack(
		ctx.RenderPreview(refForm.Root),
		ui.VStack(
			ui.Text(refForm.String()).Color(ui.ColorWhite),
		).BackgroundColor("#00000055").
			Position(ui.Position{Type: ui.PositionAbsolute, Bottom: "0rem", Left: "0rem", Right: "0rem", Top: "0rem"}).
			Border(ui.Border{}.Radius(ui.L8)),
	).Action(ctx.EditorAction(view)).
		Position(ui.Position{Type: ui.PositionRelative}).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L8))
}

func (r *FormRefRenderer) Bind(ctx ViewerRenderContext, view flow.FormView, state *core.State[*jsonptr.Obj]) core.View {
	ref := view.(*flow.FormRef)
	refForm, ok := ctx.Workspace().Forms.ByID(ref.Ref())
	if !ok {
		return ui.VStack(
			ui.Text(fmt.Sprintf("referenced form not found: %s", ref.Ref())),
		)
	}

	return ctx.Render(refForm.Root)
}
