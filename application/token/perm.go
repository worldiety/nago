// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import "go.wdy.de/nago/application/permission"

var (
	PermRotate             = permission.Declare[Rotate]("nago.token.rotate", "Access Token rotieren", "Träger dieser Berechtigung können vorhandene Tokens rotieren.")
	PermCreate             = permission.Declare[Create]("nago.token.create", "Access Token anlegen", "Träger dieser Berechtigung können neue Tokens anlegen.")
	PermDelete             = permission.Declare[Delete]("nago.token.delete", "Access Token entfernen", "Träger dieser Berechtigung können Tokens entfernen.")
	PermFindAll            = permission.Declare[FindAll]("nago.token.find_all", "Access Token finden", "Träger dieser Berechtigung können die Metadaten vorhandener Tokens sehen.")
	PermResolveTokenRights = permission.Declare[ResolvedTokenRights]("nago.token.resolve_token_rights", "Access Token Rechte einsehen", "Träger dieser Berechtigung können rekursiv die Metadaten vorhandener Tokens abrufen.")
)
