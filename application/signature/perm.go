// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import "go.wdy.de/nago/application/permission"

var (
	PermFindSignaturesByUser     = permission.Declare[FindSignaturesByUser]("nago.signature.find_by_user", "Elektronische Unterschriften eines Nutzers finden", "Träger dieser Berechtigungen können alle von einem Nutzer unterschriebenen Element-Referenzen einsehen.")
	PermFindSignaturesByResource = permission.Declare[FindSignaturesByResource]("nago.signature.find_by_resource", "Elektronische Unterschriften einer Resource finden", "Träger dieser Berechtigungen können alle Signaturen einer Resource einsehen.")
	PermFindByID                 = permission.Declare[FindByID]("nago.signature.find_by_id", "Elektronische Unterschriften per Id finden", "Träger dieser Berechtigungen können Signaturen per ID einsehen.")
)
