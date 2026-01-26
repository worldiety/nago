// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/markdown"
)

var _ ViewRenderer = (*TextRenderer)(nil)

type TextRenderer struct {
}

func (r *TextRenderer) Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply) {
	f, ok := ctx.Workspace.Forms.ByView(parent)
	if !ok {
		return alert.BannerError(fmt.Errorf("parent has no form: %s", parent)), nil
	}

	state := core.StateOf[flow.AddFormTextCmd](ctx.Window, fmt.Sprintf("%s-%s", parent, after)).Init(func() flow.AddFormTextCmd {
		return flow.AddFormTextCmd{
			Workspace: ctx.Workspace.Identity(),
			Form:      f.ID,
			Parent:    parent,
			After:     after,
		}
	})

	errState := core.DerivedState[error](state, "err")

	return ui.VStack(
			form.Auto(form.AutoOptions{
				Errors:  errState.Get(),
				Context: ctx.Context,
			}, state),
		).FullWidth(), func() error {
			return ctx.Handle(ctx.Window.Subject(), state.Get())
		}
}

func (r *TextRenderer) Update(ctx RContext, view flow.ViewID) (core.View, Apply) {
	//TODO implement me
	panic("implement me")
}

func (r *TextRenderer) Bind(ctx RContext, view flow.ViewID, state *core.State[*jsonptr.Obj]) core.View {
	//TODO implement me
	panic("implement me")
}

func (r *TextRenderer) TeaserPreview(ctx RContext) core.View {
	return ui.VStack(
		ui.Text("Const Text").Font(ui.BodySmall),
		ui.Text("H1, H2, H6, Text, Markdown"),
	)
}

func (r *TextRenderer) Preview(ctx RContext, view flow.FormView) core.View {
	text := view.(*flow.FormText)
	switch text.Style() {
	case flow.FormTextStyleDefault:
		return ui.Text(text.Value())
	case flow.FormTextStyleH1:
		return ui.H1(text.Value())
	case flow.FormTextStyleH2:
		return ui.H2(text.Value())
	case flow.FormTextStyleH3:
		return ui.Heading(3, text.Value())
	case flow.FormTextStyleMarkdown:
		return markdown.RichText(text.Value())
	default:
		panic("unknown text style")
	}
}
