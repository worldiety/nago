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
	"log/slog"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type ViewerRenderContext struct {
	ctx              context.Context
	wnd              core.Window
	workspace        *flow.Workspace
	form             *flow.Form
	structType       *flow.StructType
	handle           flow.HandleCommand
	readOnly         bool
	renderersById    map[reflect.Type]ViewRenderer
	state            *core.State[*jsonptr.Obj]
	compileExprCache map[flow.Expression]*vm.Program
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
	return ViewerRenderContext{
		ctx:              ctx,
		wnd:              wnd,
		workspace:        ws,
		form:             form,
		readOnly:         readOnly,
		renderersById:    renderersById,
		state:            state,
		structType:       structType,
		compileExprCache: make(map[flow.Expression]*vm.Program),
	}
}

func (c ViewerRenderContext) ReadOnly() bool {
	return c.readOnly
}

func (c ViewerRenderContext) EvaluateVisibility(view flow.FormView) bool {
	estr := view.VisibleExpr()
	if estr == "" {
		return true
	}

	env := map[string]any{
		"self": c.state.Get(),
		"NULL": jsonptr.Null{},
		"field": func(name string) jsonptr.Value {
			return c.state.Get().GetOr(name, jsonptr.Null{})
		},
	}

	if _, ok := c.compileExprCache[estr]; !ok {
		p, err := expr.Compile(string(estr), expr.Env(env))
		if err != nil {
			slog.Error("cannot compile expression", "view", view.Identity(), "err", err)
			return false
		}

		c.compileExprCache[estr] = p
	}

	prg := c.compileExprCache[estr]

	output, err := expr.Run(prg, env)
	if err != nil {
		slog.Error("cannot evaluate expression", "view", view.Identity(), "err", err)
		return false
	}

	if b, ok := output.(bool); ok {
		return b
	}

	return true
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

	if !c.EvaluateVisibility(view) {
		return nil
	}

	return r.Bind(c, view, c.state)
}
