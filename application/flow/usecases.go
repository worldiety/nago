// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"iter"
	"os"
	"slices"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

type FindWorkspaces func(subject auth.Subject) iter.Seq2[WorkspaceID, error]
type LoadWorkspace func(subject auth.Subject, id WorkspaceID) (option.Opt[*Workspace], error)

type DeleteWorkspace func(subject auth.Subject, id WorkspaceID) error
type HandleCommand func(subject auth.Subject, cmd WorkspaceCommand) error

// ExportedWorkspace represents a json decoded Export. See also [ExportWorkspace] and [ImportWorkspace].
type ExportedWorkspace struct {
	Events []evs.JsonEnvelope `json:"events"`
}
type ExportWorkspace func(subject auth.Subject, id WorkspaceID) ([]byte, error)
type ImportWorkspace func(subject auth.Subject, data []byte) error

type ShareFormOptions struct {
	AllowUnauthenticated bool
	AllowedUsers         []user.ID `source:"nago.user"`
}

type FormShareID string
type FormShare struct {
	ID                   FormShareID `json:"id" visible:"false"`
	Workspace            WorkspaceID `json:"workspace,omitempty" visible:"false"`
	Form                 FormID      `json:"form,omitempty" visible:"false"`
	AllowUnauthenticated bool        `json:"allowUnauthenticated,omitempty"`
	AllowedUsers         []user.ID   `json:"allowedUsers,omitempty" source:"nago.users"`
}

func (s FormShare) clone() FormShare {
	return FormShare{
		ID:                   s.ID,
		Workspace:            s.Workspace,
		Form:                 s.Form,
		AllowUnauthenticated: s.AllowUnauthenticated,
		AllowedUsers:         slices.Clone(s.AllowedUsers),
	}
}

func (s FormShare) Identity() FormShareID {
	return s.ID
}

func (s FormShare) WithIdentity(id FormShareID) FormShare {
	s.ID = id
	return s
}

type UpdateFormShare func(subject auth.Subject, workspace WorkspaceID, form FormID, opts ShareFormOptions) (FormShareID, error)

type DeleteFormShare func(subject auth.Subject, id FormShareID) error

type FindFormShare func(subject auth.Subject, workspace WorkspaceID, form FormID) (option.Opt[FormShare], error)
type FindFormShareByID func(subject auth.Subject, shareID FormShareID) (option.Opt[FormShare], error)

type FormShareRepository data.Repository[FormShare, FormShareID]

type UseCases struct {
	FindWorkspaces    FindWorkspaces
	LoadWorkspace     LoadWorkspace
	HandleCommand     HandleCommand
	DeleteWorkspace   DeleteWorkspace
	ExportWorkspace   ExportWorkspace
	ImportWorkspace   ImportWorkspace
	UpdateFormShare   UpdateFormShare
	FindFormShare     FindFormShare
	FindFormShareByID FindFormShareByID
}

func NewUseCases(handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID], formShareRepo FormShareRepository) UseCases {
	loadFn := NewLoadWorkspace(handler)
	idx := &shareIndex{repo: formShareRepo, load: loadFn}

	return UseCases{
		FindWorkspaces:    NewFindWorkspace(handler),
		LoadWorkspace:     loadFn,
		HandleCommand:     NewHandleCommand(handler),
		DeleteWorkspace:   NewDeleteWorkspace(handler),
		ExportWorkspace:   NewExportWorkspace(handler),
		ImportWorkspace:   NewImportWorkspace(handler),
		UpdateFormShare:   NewUpdateFormShare(idx, loadFn),
		FindFormShare:     NewFindFormShare(idx),
		FindFormShareByID: NewFindFormShareByID(idx),
	}
}

type shareIndex struct {
	mutex         sync.Mutex
	reverseLookup map[FormID]FormShare
	lookup        map[FormShareID]FormShare
	repo          FormShareRepository
	load          LoadWorkspace
	loaded        bool
}

func (idx *shareIndex) byForm(workspace WorkspaceID, form FormID, fn func(ws *Workspace, form *Form) error) error {
	return idx.mut(func() error {
		optWs, err := idx.load(user.SU(), workspace)
		if err != nil {
			return err
		}

		if optWs.IsNone() {
			return fmt.Errorf("workspace %s does not exist: %w", workspace, os.ErrNotExist)
		}

		ws := optWs.Unwrap()

		fr, ok := ws.Forms.ByID(form)
		if !ok {
			return fmt.Errorf("form %s does not exist: %w", form, os.ErrNotExist)
		}

		return fn(ws, fr)
	})
}

func (idx *shareIndex) mut(fn func() error) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()
	if !idx.loaded {
		idx.reverseLookup = map[FormID]FormShare{}
		idx.lookup = map[FormShareID]FormShare{}
		for share, err := range idx.repo.All() {
			if err != nil {
				return err
			}

			idx.reverseLookup[share.Form] = share
			idx.lookup[share.Identity()] = share
		}

		idx.loaded = true
	}

	return fn()
}
