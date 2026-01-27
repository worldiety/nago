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

var _ ViewRenderer = (*CheckboxRenderer)(nil)

type CheckboxRenderer struct {
}

func (r *CheckboxRenderer) Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply) {
	f, ok := ctx.Workspace().Forms.ByView(parent)
	if !ok {
		return alert.BannerError(fmt.Errorf("parent has no form: %s", parent)), nil
	}

	state := core.StateOf[flow.AddCheckboxCmd](ctx.Window(), fmt.Sprintf("%s-%s", parent, after)).Init(func() flow.AddCheckboxCmd {
		return flow.AddCheckboxCmd{
			Workspace: ctx.Workspace().Identity(),
			Form:      f.ID,
			Parent:    parent,
			After:     after,
		}
	})

	structType, ok := ctx.Workspace().Packages.StructTypeByID(f.RepositoryType())
	if !ok {
		return alert.BannerError(fmt.Errorf("form referenced a struct type over repository which cannot be resolved: %s.%s", f.Repository(), f.ID)), nil
	}

	myCtx := core.WithContext(ctx.Context, core.ContextValue("nago.flow.fields", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[flow.Field, error] {
		return func(yield func(flow.Field, error) bool) {

			for f := range structType.Fields.All() {
				if f, ok := f.(*flow.BoolField); ok {
					if !yield(f, nil) {
						return
					}
				}

			}
		}
	})))

	errState := core.DerivedState[error](state, "err")

	return ui.VStack(
			form.Auto(form.AutoOptions{
				Errors:  errState.Get(),
				Context: myCtx,
			}, state),
		).FullWidth(), func() error {
			return ctx.Handle(ctx.Window().Subject(), state.Get())
		}
}

func (r *CheckboxRenderer) Update(ctx RContext, view flow.FormView) core.View {
	return ui.Text("Edit Checkbox")
}

func (r *CheckboxRenderer) TeaserPreview(ctx RContext) core.View {
	return ui.VStack(
		ui.Text("Checkbox").Font(ui.BodySmall),
		ui.CheckboxField("Label", true).SupportingText("Supporting Text"),
	)
}

func (r *CheckboxRenderer) Preview(ctx RContext, view flow.FormView) core.View {
	box := view.(*flow.FormCheckbox)
	return ui.CheckboxField(box.Label(), false).SupportingText(box.SupportingText())
}

func (r *CheckboxRenderer) Bind(ctx ViewerRenderContext, view flow.FormView, state *core.State[*jsonptr.Obj]) core.View {
	box := view.(*flow.FormCheckbox)
	structType := ctx.StructType()
	field, ok := structType.Fields.ByID(box.Field())
	if !ok {
		return alert.BannerError(fmt.Errorf("field not found: %s", box.Field()))
	}

	jsonName := field.JSONName()
	myState := core.DerivedState[bool](state, string(box.Field())).Init(func() bool {
		val, ok := state.Get().Get(jsonName)
		if !ok {
			return false
		}

		return val.Bool()
	})

	myState.Observe(func(newValue bool) {
		obj := state.Get()
		obj.Put(jsonName, jsonptr.Bool(newValue))
		state.Set(obj)
		state.Notify()
	})

	return ui.CheckboxField(box.Label(), myState.Get()).SupportingText(box.SupportingText()).InputValue(myState)
}
