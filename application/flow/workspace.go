// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

type CreateWorkspaceCmd struct {
	Name        string `label:"nago.common.label.name"`
	Description string `label:"nago.common.label.description" lines:"3"`
}

type WorkspaceCreated struct {
	Workspace   WorkspaceID `json:"workspace,omitempty"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
}

func (e WorkspaceCreated) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e WorkspaceCreated) event() {}
