// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"iter"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

type FindWorkspaces func(subject auth.Subject) iter.Seq2[WorkspaceID, error]
type LoadWorkspace func(subject auth.Subject, id WorkspaceID) (option.Opt[*Workspace], error)

type CreateWorkspace func(subject auth.Subject, cmd CreateWorkspaceCmd) (WorkspaceCreated, error)

type CreatePackage func(subject auth.Subject, cmd CreatePackageCmd) (PackageCreated, error)

type CreateStringType func(subject auth.Subject, cmd CreateStringTypeCmd) (StringTypeCreated, error)

type CreateStructType func(subject auth.Subject, cmd CreateStructTypeCmd) (StructTypeCreated, error)

type AppendStringField func(subject auth.Subject, cmd AppendStringFieldCmd) (StringFieldAppended, error)

type UseCases struct {
	FindWorkspaces    FindWorkspaces
	LoadWorkspace     LoadWorkspace
	CreateWorkspace   CreateWorkspace
	CreatePackage     CreatePackage
	CreateStringType  CreateStringType
	CreateStructType  CreateStructType
	AppendStringField AppendStringField
}

func NewUseCases(repoName string, storeEvent evs.Store[WorkspaceEvent], replayWorkspace evs.ReplayWithIndex[WorkspaceID, WorkspaceEvent], wsIndex *evs.StoreIndex[WorkspaceID, WorkspaceEvent]) UseCases {
	cache := map[WorkspaceID]*Workspace{}
	var mutex sync.Mutex

	loadFn := NewLoadWorkspace(&mutex, repoName, replayWorkspace, cache)
	return UseCases{
		FindWorkspaces:    NewFindWorkspace(repoName, wsIndex),
		LoadWorkspace:     loadFn,
		CreateWorkspace:   NewCreateWorkspace(storeEvent),
		CreatePackage:     NewCreatePackage(newHandleCmd[PackageCreated](repoName, loadFn, storeEvent)),
		CreateStringType:  NewCreateStringType(newHandleCmd[StringTypeCreated](repoName, loadFn, storeEvent)),
		CreateStructType:  NewCreateStructType(newHandleCmd[StructTypeCreated](repoName, loadFn, storeEvent)),
		AppendStringField: NewAppendStringField(newHandleCmd[StringFieldAppended](repoName, loadFn, storeEvent)),
	}
}
