// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgflow

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/evs"
	cfgevs "go.wdy.de/nago/application/evs/cfg"
	"go.wdy.de/nago/application/flow"
	uiflow "go.wdy.de/nago/application/flow/ui"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

type Module struct {
	Mutex *sync.Mutex // Mutex used by the UseCases to protect critical write sections
	//UseCases    evs.UseCases
	Pages       uiflow.Pages
	Permissions evs.Permissions
	UseCases    flow.UseCases
}

type Options struct {
	Renderers map[reflect.Type]uiflow.ViewRenderer
	Events    []flow.WorkspaceEvent
}

type Opt func(*Options)

func (o Options) WithOptions(opts ...Opt) Options {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

func Enable(cfg *application.Configurator, opts Options) (Module, error) {
	mod, ok := core.FromContext[Module](cfg.Context(), "")
	if ok {
		return mod, nil
	}

	prefix := permission.ID("nago.flow")

	if !prefix.Valid() {
		return Module{}, fmt.Errorf("prefix is not valid")
	}

	stores, err := cfg.Stores()
	if err != nil {
		return Module{}, err
	}

	mod = Module{Mutex: &sync.Mutex{}}
	mod.Pages = uiflow.Pages{
		Workspaces:       "admin/" + core.NavigationPath(makeFactoryID(prefix)) + "/workspaces",
		Editor:           "admin/" + core.NavigationPath(makeFactoryID(prefix)) + "/workspace/edit",
		FormViewerCreate: "admin/" + core.NavigationPath(makeFactoryID(prefix)) + "/workspace/form/create",
	}

	if len(opts.Renderers) == 0 {
		opts.Renderers = maps.Clone(uiflow.DefaultRenderers)
	}

	renderers := map[reflect.Type]uiflow.ViewRenderer{}
	for k, r := range opts.Renderers {

		renderers[k] = r
	}

	if len(opts.Events) == 0 {
		opts.Events = slices.Collect(flow.DefaultEvents)
	}

	evsSchemas := map[evs.Discriminator]reflect.Type{}
	for _, ev := range opts.Events {
		if other, ok := evsSchemas[ev.Discriminator()]; ok {
			return Module{}, fmt.Errorf("duplicate event discriminator: %s defined by both %v and %T", ev.Discriminator(), other, ev)
		}

		evsSchemas[ev.Discriminator()] = reflect.TypeOf(ev)
	}

	var replayByWorkspace evs.ReplayWithIndex[flow.WorkspaceID, flow.WorkspaceEvent]
	var idxByWorkspace *evs.StoreIndex[flow.WorkspaceID, flow.WorkspaceEvent]
	modEvs, err := cfgevs.Enable[flow.WorkspaceEvent](cfg, "nago.flow.workspace", "Flow Workspace", cfgevs.Options[flow.WorkspaceEvent]{Schema: evsSchemas}.WithOptions(
		cfgevs.Index[flow.WorkspaceID, flow.WorkspaceEvent](func(e evs.Envelope[flow.WorkspaceEvent]) (flow.WorkspaceID, error) {
			return e.Data.WorkspaceID(), nil
		}, func(ctx cfgevs.IndexContext[flow.WorkspaceID, flow.WorkspaceEvent]) error {
			replayByWorkspace = ctx.Replay
			idxByWorkspace = ctx.Index
			return nil
		}),
	))

	if err != nil {
		return mod, err
	}

	workspaceHandler := evs.NewHandler[*flow.Workspace](
		modEvs.UseCases,
		replayByWorkspace,
		idxByWorkspace,
	)

	for _, event := range opts.Events {
		workspaceHandler.RegisterEvents(event)
	}

	uc := flow.NewUseCases(string(prefix), workspaceHandler, idxByWorkspace)
	mod.UseCases = uc

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		var grp admin.Group
		if !subject.HasPermission(flow.PermFindWorkspaces) {
			return grp
		}

		grp.Title = uiflow.StrGroupTitle.Get(subject)
		grp.Entries = append(grp.Entries, admin.Card{
			Title:  uiflow.StrWorkspaces.Get(subject),
			Text:   uiflow.StrGroupDescription.Get(subject),
			Target: mod.Pages.Workspaces,
		})

		return grp
	})

	cfg.RootViewWithDecoration(mod.Pages.Workspaces, func(wnd core.Window) core.View {
		return uiflow.PageWorkspaces(wnd, mod.Pages, mod.UseCases)
	})

	cfg.RootViewWithDecoration(mod.Pages.Editor, func(wnd core.Window) core.View {
		return uiflow.PageEditor(wnd, uiflow.PageEditorOptions{
			UseCases:  mod.UseCases,
			Renderers: renderers,
		})
	})

	cfg.RootViewWithDecoration(mod.Pages.FormViewerCreate, func(wnd core.Window) core.View {
		return uiflow.PageFormViewerCreate(wnd, uc.LoadWorkspace, stores)
	})

	cfg.AddContextValue(core.ContextValue(string("module-"+prefix), mod))

	return mod, nil
}

func makeFactoryID(prefix permission.ID) core.NavigationPath {
	return core.NavigationPath(strings.ReplaceAll(string(prefix), ".", "-"))
}
