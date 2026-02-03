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

var _ ViewRenderer = (*TextFieldRenderer)(nil)

type TextFieldRenderer struct {
}

func (r *TextFieldRenderer) Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply) {
	f, ok := ctx.Workspace().Forms.ByView(parent)
	if !ok {
		return alert.BannerError(fmt.Errorf("parent has no form: %s", parent)), nil
	}

	state := core.StateOf[flow.AddFormTextFieldCmd](ctx.Window(), fmt.Sprintf("%s-%s", parent, after)).Init(func() flow.AddFormTextFieldCmd {
		return flow.AddFormTextFieldCmd{
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

	myCtx := core.WithContext(ctx.Context, core.ContextValue("nago.flow.fields", form.AnyUseCaseListReadOnly(func(subject auth.Subject) iter.Seq2[flow.Field, error] {
		return func(yield func(flow.Field, error) bool) {

			for f := range structType.Fields.All() {
				if f, ok := f.(*flow.StringField); ok {
					if !yield(f, nil) {
						return
					}
				}

				if f, ok := f.(*flow.TypeField); ok && f.StringType(ctx.Workspace()) {
					if !yield(f, nil) {
						return
					}
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
			return ctx.Handle(ctx.Window().Subject(), state.Get())
		}
}

func (r *TextFieldRenderer) Update(ctx RContext, view flow.FormView) core.View {
	return ui.Text("Edit Text field")
}

func (r *TextFieldRenderer) TeaserPreview(ctx RContext) core.View {
	return ui.VStack(
		ui.Text("Text field").Font(ui.BodySmall),
		ui.TextField("Label", "").Disabled(true).SupportingText("Supporting Text"),
	)
}

func (r *TextFieldRenderer) Preview(ctx RContext, view flow.FormView) core.View {
	text := view.(*flow.FormTextField)
	return ui.VStack(
		ui.TextField(text.Label(), "").Disabled(true).SupportingText(text.SupportingText()).FullWidth(),
	).Action(ctx.EditorAction(view)).Frame(text.Frame())
}

func (r *TextFieldRenderer) Bind(ctx ViewerRenderContext, view flow.FormView, state *core.State[*jsonptr.Obj]) core.View {
	tf := view.(*flow.FormTextField)
	structType := ctx.StructType()
	field, ok := structType.Fields.ByID(tf.Field())
	if !ok {
		return alert.BannerError(fmt.Errorf("field not found: %s", tf.Field()))
	}

	jsonName := field.JSONName()
	myState := core.DerivedState[string](state, string(tf.Field())).Init(func() string {
		val, ok := state.Get().Get(jsonName)
		if !ok {
			return ""
		}

		return val.String()
	})

	myState.Observe(func(newValue string) {
		obj := state.Get()
		obj.Put(jsonName, jsonptr.String(newValue))
		state.Set(obj)
		state.Notify()
	})

	return ui.TextField(tf.Label(), myState.Get()).SupportingText(tf.SupportingText()).InputValue(myState).Frame(tf.Frame())
}
