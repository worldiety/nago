// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgflow

import (
	"fmt"
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
	underlyingTypes map[TID]UnderlyingType[any]
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

	mod = Module{Mutex: &sync.Mutex{}}
	mod.Pages = uiflow.Pages{
		Workspaces: "admin/" + core.NavigationPath(makeFactoryID(prefix)) + "/workspaces",
		Editor:     "admin/" + core.NavigationPath(makeFactoryID(prefix)) + "/workspace/edit",
	}

	var replayByWorkspace evs.ReplayWithIndex[flow.WorkspaceID, flow.WorkspaceEvent]
	var idxByWorkspace *evs.StoreIndex[flow.WorkspaceID, flow.WorkspaceEvent]
	modEvs, err := cfgevs.Enable[flow.WorkspaceEvent](cfg, "nago.flow.workspace", "Flow Workspace", cfgevs.Options[flow.WorkspaceEvent]{}.WithOptions(
		cfgevs.Schema[flow.WorkspaceCreated, flow.WorkspaceEvent]("WorkspaceCreated"),
		cfgevs.Schema[flow.PackageCreated, flow.WorkspaceEvent]("PackageCreated"),
		cfgevs.Schema[flow.StringTypeCreated, flow.WorkspaceEvent]("StringTypeCreated"),
		cfgevs.Schema[flow.StructTypeCreated, flow.WorkspaceEvent]("StructTypeCreated"),
		cfgevs.Schema[flow.StringFieldAppended, flow.WorkspaceEvent]("StringFieldAppended"),
		cfgevs.Schema[flow.BoolFieldAppended, flow.WorkspaceEvent]("BoolFieldAppended"),
		cfgevs.Schema[flow.RepositoryAssigned, flow.WorkspaceEvent]("RepositoryAssigned"),
		cfgevs.Schema[flow.PrimaryKeySelected, flow.WorkspaceEvent]("PrimaryKeySelected"),
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

	uc := flow.NewUseCases(string(prefix), modEvs.UseCases.Store, replayByWorkspace, idxByWorkspace)
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
		return uiflow.PageEditor(wnd, mod.UseCases)
	})

	cfg.AddContextValue(core.ContextValue(string("module-"+prefix), mod))

	return mod, nil
}

// A Workspace shares multiple forms and fields.
type Workspace struct {
	Fields []FieldID
}

type FieldID string

// A Field is like a Textinput (1 line or multiple), bool value, multiple choice (single or multi select and the same model but different design like checkboxes or picker),
// foreign key (multi or single), embedded struct (multi or single). Also multiple choice often requires adding additional freetext
type Field struct {
	// Name must be unique and a valid Go identifier for optional code generation.
	Name string `json:"name"`
	Type TID    `json:"type"`
	// Label is only used when rendering the field.
	Label string `json:"label,omitempty"`
	// SupportingText is only used when rendering the field.
	SupportingText string `json:"supportingText,omitempty"`

	// JsonName must be unique within the scope. If empty, Name is used. Used as the json-Tag in code generation
	JsonName string `json:"jsonName,omitempty"`

	// Required could be part of the validation chain, however, it cannot be evaluated
	// programmatically, and we need this information to decide quickly to display an indicator.
	Required bool `json:"required,omitempty"`

	// References to validation functions.
	Validation []VID `json:"validation,omitempty"`
}

type Struct struct {
	Fields []Field `json:"fields,omitempty"`
}

type Style int

const (
	StyleRow Style = iota + 1
	StyleOptional
	StyleCard
)

type Form struct {
	Groups []Group // TODO what about dialogs
}

// TODO what about repositories?
// TODO some may be shared after insertion (login is enough, e.g. a wiki)?
// TODO some may be shared after another user has proofed it (contract? docusign?)
// TODO some may be always local (questionaire about used frameworks etc.)
// TODO what about stepper/pages?
// TODO what about readonly? roles? states? (e.g. invoices or anträge after einreichung)
// TODO there are optional groups where the flag itself is relevant (e.g. catering and catering comment)
type Group struct {
	Style          Style
	Label          string  // Depending on the Style, this label may get rendered differently
	SupportingText string  // Depending on the Style, this text may get rendered differently
	Fields         []Field `json:"fields,omitempty"`
	//Groups  []Group // ??? TODO do we need nesting?
	Visible func(any) bool // TODO we must reference the visible predicate and what is the parameter?
	// TODO how to support self-referencing cardinality?
	// TODO are there fields based on a group? what about recursion of groups?
}

// / SCRATCH
type MyForm struct {
	ErbringtDerKundeDienstleistungen bool
	VerkauftDerKundeProdukte         bool
	ErfolgtEinVertragsabschluss      bool
	//...
}

func (f MyForm) Validate() error {
	return fmt.Errorf("if all false, not answered")
	//return fmt.Errorf("if A && B || C || D, answerer D")
}

func makeFactoryID(prefix permission.ID) core.NavigationPath {
	return core.NavigationPath(strings.ReplaceAll(string(prefix), ".", "-"))
}
