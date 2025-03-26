// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package group

import "go.wdy.de/nago/application/permission"

// System group denotes an internal group for system usages. Usually, a User should never belong to this group.
// The System group should be always valid and available after launch.
const System = "nago.group.system"

var (
	PermFindByID = permission.Declare[FindByID]("nago.group.find_by_id", "Eine Gruppe anzeigen", "Träger dieser Berechtigung können eine Gruppe über ihre EID anzeigen.")
	PermFindAll  = permission.Declare[FindAll]("nago.group.find_all", "Alle Gruppen anzeigen", "Träger dieser Berechtigung können alle vorhandenen Gruppen anzeigen.")
	PermCreate   = permission.Declare[Create]("nago.group.create", "Gruppen erstellen", "Träger dieser Berechtigung können neue Gruppen anlegen.")
	PermUpdate   = permission.Declare[Update]("nago.group.update", "Gruppen aktualisieren", "Träger dieser Berechtigung können vorhandene Gruppen aktualisieren.")
	PermDelete   = permission.Declare[Delete]("nago.group.delete", "Gruppen löschen", "Träger dieser Berechtigung können vorhandene Gruppen löschen.")
)
