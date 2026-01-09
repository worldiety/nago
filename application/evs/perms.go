// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"fmt"

	"go.wdy.de/nago/application/permission"
)

type Permissions struct {
	Store   permission.ID
	Load    permission.ID
	Replay  permission.ID
	ReadAll permission.ID
	Delete  permission.ID

	Prefix     permission.ID
	EntityName string
}

// DeclarePermissions is a factory to create a bunch of permissions. Use it at package level, so that permission
// identifiers are available already at package initialization time to avoid working with empty
// permission identifiers accidentally.
// The following identifier naming rules apply:
//   - [Store]: <prefix>.store
//   - [Load]: <prefix>.load
//   - [Replay]: <prefix>.replay
//   - [All]: <prefix>.all
func DeclarePermissions[Evt any](prefix permission.ID, eventSumTypeName string) Permissions {
	if !prefix.Valid() {
		panic(fmt.Errorf("invalid prefix: %s", prefix))
	}

	return Permissions{
		Store:      permission.DeclareCreate[Store[Evt]](prefix+".store", eventSumTypeName),
		Load:       permission.DeclareFindByID[Load[Evt]](prefix+".load", eventSumTypeName),
		Replay:     permission.DeclareReplay[Replay[Evt]](prefix+".replay", eventSumTypeName),
		ReadAll:    permission.DeclareReplay[ReadAll[Evt]](prefix+".readall", eventSumTypeName),
		Delete:     permission.DeclareDeleteByID[Delete[Evt]](prefix+".delete", eventSumTypeName),
		Prefix:     prefix,
		EntityName: eventSumTypeName,
	}
}
