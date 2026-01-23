// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"

	"go.wdy.de/nago/application/evs"
)

type StringEnumCaseAdded struct {
	Workspace   WorkspaceID `json:"workspace"`
	String      TypeID      `json:"string"`
	Name        Ident       `json:"name"`
	Value       string      `json:"value"`
	Description string      `json:"description"`
}

func (evt StringEnumCaseAdded) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt StringEnumCaseAdded) Discriminator() evs.Discriminator {
	return "StringEnumCaseAdded"
}

func (evt StringEnumCaseAdded) Evolve(ws *Workspace) error {
	st, ok := ws.Packages.StringTypeByID(evt.String)
	if !ok {
		return fmt.Errorf("string %s not found", evt.String)
	}

	st.Enumeration.Add(&Literal{
		Name:        evt.Name,
		Value:       evt.Value,
		Description: evt.Description,
	})

	return nil
}
