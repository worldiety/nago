// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"
	"os"
	"reflect"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

// TFormViewer is a component. It renders a flow form for editing or viewing and tracks its state using the given
// state.
type TFormViewer struct {
	loader        flow.LoadWorkspace
	wsID          flow.WorkspaceID
	formID        flow.FormID
	renderersById map[reflect.Type]ViewRenderer
	state         *core.State[*jsonptr.Obj]
	readOnly      bool
}

func FormViewer(loader flow.LoadWorkspace, workspace flow.WorkspaceID, form flow.FormID, state *core.State[*jsonptr.Obj]) TFormViewer {
	return TFormViewer{
		loader:        loader,
		wsID:          workspace,
		formID:        form,
		renderersById: DefaultRenderers,
		state:         state,
	}
}

// Renderers sets custom renderers for the form viewer instead of using the [DefaultRenderers].
func (c TFormViewer) Renderers(renderers map[reflect.Type]ViewRenderer) TFormViewer {
	c.renderersById = renderers
	return c
}

// Disabled makes the form viewer read-only.
func (c TFormViewer) Disabled(b bool) TFormViewer {
	c.readOnly = b
	return c
}

func (c TFormViewer) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	optWs, err := c.loader(wnd.Subject(), c.wsID)
	if err != nil {
		return alert.BannerError(err).Render(ctx)
	}

	if optWs.IsNone() {
		return alert.BannerError(fmt.Errorf("workspace not found: %s: %w", c.wsID, os.ErrNotExist)).Render(ctx)
	}

	ws := optWs.Unwrap()

	form, ok := ws.Forms.ByID(c.formID)
	if !ok {
		return alert.BannerError(fmt.Errorf("form not found: %s: %w", c.formID, os.ErrNotExist)).Render(ctx)
	}

	structType, ok := ws.Packages.StructTypeByID(form.RepositoryType())
	if !ok {
		return alert.BannerError(fmt.Errorf("form referenced a struct type over repository which cannot be resolved: %s.%s", form.Repository(), form.ID)).Render(ctx)
	}

	rctx := NewViewerRenderContext(
		wnd.Context(),
		wnd,
		ws,
		form,
		structType,
		c.renderersById,
		c.readOnly,
		c.state,
	)

	return rctx.Render(form.Root).Render(ctx)
}
