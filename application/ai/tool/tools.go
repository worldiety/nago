// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tool

import "go.wdy.de/nago/pkg/data"

type ID string

const (
	// DocumentLibrary is the standard name for the default rag library system.
	DocumentLibrary ID = "document_library"
)

type Tool struct {
	ID          ID     `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func (m Tool) Identity() ID {
	return m.ID
}

func (m Tool) WithIdentity(id ID) Tool {
	m.ID = id
	return m
}

func (m Tool) String() string {
	return m.Name + " (" + string(m.ID) + ")" + "\n" + m.Description
}

type Repository data.Repository[Tool, ID]
