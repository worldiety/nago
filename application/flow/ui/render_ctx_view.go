// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"context"
	"fmt"
	"reflect"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type ViewerRenderContext struct {
	ctx           context.Context
	wnd           core.Window
	workspace     *flow.Workspace
	form          *flow.Form
	structType    *flow.StructType
	handle        flow.HandleCommand
	readOnly      bool
	renderersById map[reflect.Type]ViewRenderer
	state         *core.State[*jsonptr.Obj]
}

func NewViewerRenderContext(
	ctx context.Context,
	wnd core.Window,
	ws *flow.Workspace,
	form *flow.Form,
	structType *flow.StructType,
	renderersById map[reflect.Type]ViewRenderer,
	readOnly bool,
	state *core.State[*jsonptr.Obj],
) ViewerRenderContext {
	return ViewerRenderContext{ctx: ctx, wnd: wnd, workspace: ws, form: form, readOnly: readOnly, renderersById: renderersById, state: state, structType: structType}
}

func (c ViewerRenderContext) ReadOnly() bool {
	return c.readOnly
}

func (c ViewerRenderContext) Context() context.Context {
	return c.ctx
}

func (c ViewerRenderContext) Window() core.Window {
	return c.wnd
}

func (c ViewerRenderContext) Workspace() *flow.Workspace {
	return c.workspace
}

func (c ViewerRenderContext) Form() *flow.Form {
	return c.form
}

func (c ViewerRenderContext) StructType() *flow.StructType {
	return c.structType
}

func (c ViewerRenderContext) Render(view flow.FormView) core.View {
	r, ok := c.renderersById[reflect.TypeOf(view)]
	if !ok {
		return ui.Text(fmt.Sprintf("%T has no renderer", view))
	}

	return r.Bind(c, view, c.state)
}
